package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ryan-willis/qotd/app/logger"
)

func (a *API) CleanupCheck(bc *ByteChannel) {
	a.ctx.Log.Debug("cleanup check")
	for roomId, room := range a.rooms {
		room.ClientsMutex.Lock()
		for clientId, client := range room.Clients {
			if client == bc {
				a.ctx.Log.Debug("disconnecting participant", logger.Field("participantId", clientId), logger.Field("roomId", roomId))
				delete(room.Clients, clientId)
				participant := room.Room.FindParticipant(clientId)
				if participant == nil {
					continue
				}
				participant.ActivelyConnected = false
				participant.Attentive = false
				go func(rId *string, pId *string) {
					time.Sleep(60 * time.Second)
					roomRef, exists := a.getRoomRef(*rId)
					if !exists {
						return
					}
					foundParticipant := roomRef.Room.FindParticipant(*pId)
					if foundParticipant == nil || foundParticipant.ActivelyConnected {
						return
					}
					roomRef.Room.RemoveParticipant(*pId)
					roomRef.Channel <- []byte("refresh")
				}(&roomId, &clientId)
				room.Channel <- []byte("refresh")
			}
		}
		if len(room.Clients) == 0 {
			room.Dead = true
		}
		room.ClientsMutex.Unlock()
		if room.Dead {
			a.ctx.Log.Debug("room is dead, cleaning up", logger.Field("roomId", room.Room.ID))
			delete(a.rooms, roomId)
		}
	}
}

func (a *API) CloseWithMessage(dead *bool, errMsg string, ws *websocket.Conn, closeCode int, msg string) {
	if errMsg != "" {
		a.ctx.Log.Error(errMsg)
	}
	*dead = true
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, msg))
	ws.Close()
}

func handleLatency(bc *ByteChannel, m map[string]interface{}) {
	*bc <- []byte(`{"latency":"` + m["stamp"].(string) + `"}`)
}

func (a *API) GetRoomWebsocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			a.ctx.Log.Error("error upgrading connection", logger.Field("err", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_code": "INTERNAL_ERROR",
			})
			return
		}
		a.ctx.Log.Debug("websocket connected")
		ws.WriteMessage(websocket.TextMessage, []byte("ok"))
		dead := false
		participantId := ""
		roomId := ""

		channel := make(ByteChannel)

		go func(c *ByteChannel, _ws *websocket.Conn, _dead *bool) {
			for {
				bytes := <-*c
				if *_dead {
					break
				}
				(*_ws).WriteMessage(websocket.TextMessage, bytes)
			}
		}(&channel, ws, &dead)

		for {
			if dead {
				a.ctx.Log.Debug("dead")
				break
			}
			_, payload, err := ws.ReadMessage()
			if err != nil {
				a.ctx.Log.Debug("error reading message", logger.Field("err", err.Error()))
				break
			}
			if len(payload) > 0 && payload[0] == '{' {
				var m map[string]interface{}
				if err := json.Unmarshal(payload, &m); err != nil {
					a.ctx.Log.Debug("error unmarshalling message", logger.Field("err", err.Error()))
					break
				}
				if m["do"] == nil {
					continue
				}
				if m["do"].(string) != "latency" {
					a.ctx.Log.Debug("ws message from client", logger.Field("payload", m))
				}
				switch m["do"] {
				case "latency":
					go handleLatency(&channel, m)
				case "join":
					if roomId != "" {
						go a.leaveRoom(""+roomId, &participantId)
					}
					roomId = m["roomId"].(string)
					go a.joinRoom(ws, &channel, &participantId, dead, m)
				case "leave":
					go a.leaveRoom(""+roomId, &participantId)
					roomId = ""
				case "participant_attentive":
					go a.participantAttentive(roomId, &participantId)
				case "participant_inattentive":
					go a.participantInattentive(roomId, &participantId)
				case "start":
					go a.startRound(roomId, &participantId)
				case "get_answers":
					go a.getAnswers(roomId, &participantId)
				case "close_answers":
					go a.closeAnswers(roomId, &participantId)
				case "end_room":
					go a.endRoom(roomId, &participantId)
				case "ask_question":
					go a.askQuestion(roomId, &participantId, m["question"].(string))
				case "answer":
					go a.answerQuestion(roomId, &participantId, m["answer"].(string))
				case "update_name":
					go a.updateName(roomId, &participantId, m["name"].(string))
				case "show_answer":
					go a.showAnswer(roomId, &participantId, m["participant_id"].(string))
				}
			} else {
				a.ctx.Log.Debug("invalid message", logger.Field("msg", string(payload)))
			}
		}
		a.ctx.Log.Debug("closing websocket")
		dead = true
		ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now())
		ws.Close()
		a.CleanupCheck(&channel)
	}
}
