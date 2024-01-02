package api

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryan-willis/qotd/app/logger"
	"github.com/ryan-willis/qotd/app/room"
)

func (a *API) roomHandler(r *RoomRef) {
	for {
		if r.Dead {
			return
		}
		select {
		case msg := <-r.Channel:
			a.ctx.Log.Debug("broadcasting message", logger.Field("msg", string(msg)))
			if string(msg) == "answers" {
				allAnswers := map[string]string{}
				ownerAnswers := map[string]string{}
				ownerId := ""
				r.Room.ParticipantsMutext.RLock()
				for _, participant := range r.Room.Participants {
					ownerAnswers[participant.ID] = participant.Answer
					if participant.Owner {
						ownerId = participant.ID
					}
					if participant.AnswerShown {
						allAnswers[participant.ID] = participant.Answer
					}
				}
				r.Room.ParticipantsMutext.RUnlock()
				ownerJson, err := json.Marshal(map[string]interface{}{
					"answers": ownerAnswers,
				})
				if err != nil {
					a.ctx.Log.Error("error marshalling answers", logger.Field("err", err.Error()))
					continue
				}
				allJson, err := json.Marshal(map[string]interface{}{
					"answers": allAnswers,
				})
				if err != nil {
					a.ctx.Log.Error("error marshalling answers", logger.Field("err", err.Error()))
					continue
				}
				r.ClientsMutex.RLock()
				for _, client := range r.Clients {
					if client == r.Clients[ownerId] {
						*client <- ownerJson
					} else {
						*client <- allJson
					}
				}
				r.ClientsMutex.RUnlock()
				continue
			}
			if string(msg) == "refresh" {
				json, err := json.Marshal(map[string]interface{}{
					"room": r.Room,
				})
				if err != nil {
					a.ctx.Log.Error("error marshalling room", logger.Field("err", err.Error()))
					continue
				}
				r.ClientsMutex.RLock()
				for _, client := range r.Clients {
					*client <- json
				}
				r.ClientsMutex.RUnlock()
				continue
			}
			r.ClientsMutex.RLock()
			for _, client := range r.Clients {
				*client <- msg
			}
			r.ClientsMutex.RUnlock()
		case <-time.After(5 * time.Second):
			json, err := json.Marshal(map[string]interface{}{
				"room": r.Room,
			})
			if err != nil {
				a.ctx.Log.Error("error marshalling room", logger.Field("err", err.Error()))
				continue
			}
			r.ClientsMutex.RLock()
			for _, client := range r.Clients {
				*client <- json
			}
			r.ClientsMutex.RUnlock()
		}
	}
}

func (a *API) joinRoom(ws *websocket.Conn, bc *ByteChannel, participantId *string, dead bool, m map[string]interface{}) {
	if m["participantId"] == nil || m["participantToken"] == nil || m["name"] == nil || m["roomId"] == nil {
		a.CloseWithMessage(&dead, "missing required params", ws, websocket.ClosePolicyViolation, "MISSING_PARTICIPANT_TOKEN")
		a.CleanupCheck(bc)
		return
	}
	*participantId = m["participantId"].(string)
	participantToken := m["participantToken"].(string)
	name := m["name"].(string)
	roomId := m["roomId"].(string)
	// authorize the participantId with the participantToken
	tokenKey := "qotd:tokens:" + roomId + ":" + *participantId
	token, err := a.ctx.Storage.Retrieve(tokenKey)
	if err != nil {
		if err.Error() != "redis: nil" {
			a.CloseWithMessage(&dead, "error retrieving token", ws, websocket.CloseInternalServerErr, "INTERNAL_ERROR")
			a.CleanupCheck(bc)
			return
		} else {
			token = []byte(participantToken)
			if err := a.ctx.Storage.Store(tokenKey, token); err != nil {
				a.CloseWithMessage(&dead, "error writing token", ws, websocket.CloseInternalServerErr, "INTERNAL_ERROR")
				a.CleanupCheck(bc)
				return
			}
			a.ctx.Log.Debug("stored participant token", logger.Field("tokenKey", tokenKey))
		}
	} else {
		a.ctx.Log.Debug("checking participant token against stored token", logger.Field("tokenKey", tokenKey))
	}
	if string(token) != participantToken {
		a.CloseWithMessage(&dead, "invalid token", ws, websocket.ClosePolicyViolation, "INVALID_TOKEN")
		a.CleanupCheck(bc)
		return
	} else {
		a.ctx.Log.Debug("token is valid", logger.Field("tokenKey", tokenKey))
	}
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		a.ctx.Log.Debug("creating room")
		clients := make(map[string]*ByteChannel)
		clients[*participantId] = bc
		roomRef = &RoomRef{
			ClientsMutex: sync.RWMutex{},
			Channel:      make(ByteChannel, 1000),
			Room: &room.Room{
				ID:        roomId,
				State:     room.RoomStateWaiting,
				CreatedAt: time.Now().Format(time.RFC3339),
				Participants: []room.Participant{
					{
						ID:                *participantId,
						Name:              name,
						HasAnswered:       false,
						ActivelyConnected: true,
						Attentive:         true,
						Owner:             true,
						Answer:            "",
					},
				},
			},
			Clients: clients,
		}
		a.roomsMutex.Lock()
		a.rooms[roomId] = roomRef
		a.roomsMutex.Unlock()
		go a.roomHandler(roomRef)
	} else {
		participant := roomRef.Room.FindParticipant(*participantId)
		if participant == nil {
			a.ctx.Log.Debug("adding participant to room", logger.Field("participantId", *participantId))
			roomRef.Room.AddParticipant(&room.Participant{
				ID:                *participantId,
				Name:              name,
				HasAnswered:       false,
				ActivelyConnected: true,
				Attentive:         true,
				Owner:             len(roomRef.Room.Participants) == 0,
				Answer:            "",
			})
		} else {
			a.ctx.Log.Debug("reconnecting participant", logger.Field("participantId", *participantId))
			roomRef.Dead = false
			participant.ActivelyConnected = true
			participant.Attentive = true
		}
		a.ctx.Log.Debug("adding client to room", logger.Field("roomId", roomId), logger.Field("participantId", *participantId))
		roomRef.ClientsMutex.Lock()
		roomRef.Clients[*participantId] = bc
		roomRef.ClientsMutex.Unlock()
		a.ctx.Log.Debug("added client to room", logger.Field("roomId", roomId), logger.Field("participantId", *participantId))
	}
	roomRef.Channel <- []byte("refresh")
}
