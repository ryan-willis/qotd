package api

import (
	"github.com/ryan-willis/qotd/app/logger"
	"github.com/ryan-willis/qotd/app/room"
)

func (a *API) participantAttentive(roomId string, participantId *string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}

	participant := roomRef.Room.FindParticipant(*participantId)
	if participant != nil {
		a.ctx.Log.Debug("participant attentive", logger.Field("participantId", *participantId))
		participant.ActivelyConnected = true
		participant.Attentive = true
		roomRef.Channel <- []byte("refresh")
	}
}

func (a *API) participantInattentive(roomId string, participantId *string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}

	participant := roomRef.Room.FindParticipant(*participantId)
	if participant != nil {
		a.ctx.Log.Debug("participant inattentive", logger.Field("participantId", *participantId))
		participant.Attentive = false
		roomRef.Channel <- []byte("refresh")
	}
}

func (a *API) askQuestion(roomId string, participantId *string, question string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	if roomRef.Room.State == room.RoomStateWaiting {
		return
	}
	if roomRef.Room.State == room.RoomStateAnswered {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	if !participant.Owner {
		return
	}
	a.ctx.Log.Debug("asking question", logger.Field("roomId", roomId), logger.Field("question", question))
	roomRef.Room.State = room.RoomStatePlaying
	roomRef.Room.Question = question
	roomRef.Channel <- []byte("refresh")
}

func (a *API) answerQuestion(roomId string, participantId *string, answer string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	if roomRef.Room.State == room.RoomStateWaiting {
		return
	}
	if roomRef.Room.State == room.RoomStateAnswered {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	a.ctx.Log.Debug("answering question", logger.Field("roomId", roomId), logger.Field("answer", answer))
	participant.HasAnswered = true
	participant.Answer = answer
	roomRef.Channel <- []byte("refresh")
}

func (a *API) startRound(roomId string, participantId *string) {
	a.ctx.Log.Debug("attempting to start round", logger.Field("roomId", roomId))
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		a.ctx.Log.Debug("room does not exist", logger.Field("roomId", roomId))
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		a.ctx.Log.Debug("cannot find participant", logger.Field("roomId", roomId))
		return
	}
	if !participant.Owner {
		a.ctx.Log.Debug("participant not owner", logger.Field("roomId", roomId))
		return
	}
	a.ctx.Log.Debug("starting round", logger.Field("roomId", roomId))
	roomRef.Room.State = room.RoomStatePlaying
	roomRef.Room.Question = ""
	roomRef.Room.ParticipantsMutext.Lock()
	for i := range roomRef.Room.Participants {
		roomRef.Room.Participants[i].HasAnswered = false
		roomRef.Room.Participants[i].AnswerShown = false
		roomRef.Room.Participants[i].Answer = ""
	}
	roomRef.Room.ParticipantsMutext.Unlock()
	roomRef.Channel <- []byte("refresh")
}

func (a *API) endRoom(roomId string, participantId *string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	if !participant.Owner {
		return
	}
	a.ctx.Log.Debug("ending room", logger.Field("roomId", roomId))
	roomRef.Room.State = room.RoomStateEnd
	roomRef.Channel <- []byte("refresh")
}

func (a *API) leaveRoom(roomId string, participantId *string) {
	a.ctx.Log.Debug("attempting to leave room", logger.Field("roomId", roomId), logger.Field("participantId", *participantId))
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	roomRef.Room.RemoveParticipant(*participantId)
	roomRef.ClientsMutex.RLock()
	client := roomRef.Clients[*participantId]
	roomRef.ClientsMutex.RUnlock()
	roomRef.Channel <- []byte("refresh")
	a.CleanupCheck(client)
}

func (a *API) closeAnswers(roomId string, participantId *string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	if !participant.Owner {
		return
	}
	a.ctx.Log.Debug("closing answers", logger.Field("roomId", roomId))
	roomRef.Room.State = room.RoomStateAnswered
	roomRef.Channel <- []byte("refresh")
	roomRef.Channel <- []byte("answers")
}

func (a *API) getAnswers(roomId string, participantId *string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	a.ctx.Log.Debug("getting answers", logger.Field("roomId", roomId))
	roomRef.Channel <- []byte("answers")
}

func (a *API) showAnswer(roomId string, participantId *string, answerParticipantId string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	requester := roomRef.Room.FindParticipant(*participantId)
	if requester == nil {
		return
	}
	if !requester.Owner {
		return
	}
	participant := roomRef.Room.FindParticipant(answerParticipantId)
	if participant == nil {
		return
	}
	a.ctx.Log.Debug("showing answer", logger.Field("roomId", roomId), logger.Field("participantId", answerParticipantId))
	participant.AnswerShown = true
	roomRef.Channel <- []byte("answers")
}

func (a *API) updateName(roomId string, participantId *string, name string) {
	roomRef, exists := a.getRoomRef(roomId)
	if !exists {
		return
	}
	participant := roomRef.Room.FindParticipant(*participantId)
	if participant == nil {
		return
	}
	a.ctx.Log.Debug("updating participant name", logger.Field("roomId", roomId), logger.Field("name", name), logger.Field("participantId", *participantId))
	participant.Name = name
	roomRef.Channel <- []byte("refresh")
}
