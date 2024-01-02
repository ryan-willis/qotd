package room

import (
	"sync"
	"time"

	"github.com/ryan-willis/qotd/app/util"
)

type RoomState string

const (
	RoomStateWaiting  RoomState = "waiting"
	RoomStatePlaying  RoomState = "playing"
	RoomStateAnswered RoomState = "answered"
	RoomStateEnd      RoomState = "finished"
)

type Room struct {
	ID                 string        `json:"id"`
	State              RoomState     `json:"state"`
	Question           string        `json:"question"`
	CreatedAt          string        `json:"created_at"`
	ParticipantsMutext sync.RWMutex  `json:"-"`
	Participants       []Participant `json:"participants"`
}

func New() (*Room, error) {
	id := util.GenerateRoomCode()

	return &Room{
		ID:                 id,
		State:              RoomStateWaiting,
		CreatedAt:          time.Now().Format(time.RFC3339),
		ParticipantsMutext: sync.RWMutex{},
		Participants:       []Participant{},
	}, nil
}

func (r *Room) FindParticipant(id string) *Participant {
	r.ParticipantsMutext.RLock()
	for i, participant := range r.Participants {
		if participant.ID == id {
			r.ParticipantsMutext.RUnlock()
			return &r.Participants[i]
		}
	}
	r.ParticipantsMutext.RUnlock()
	return nil
}

func (r *Room) AddParticipant(p *Participant) {
	r.ParticipantsMutext.RLock()
	if len(r.Participants) == 0 {
		p.Owner = true
	}
	r.ParticipantsMutext.RUnlock()
	existingParticipant := r.FindParticipant(p.ID)
	if existingParticipant != nil {
		existingParticipant.Name = p.Name
		existingParticipant.Answer = p.Answer
		existingParticipant.HasAnswered = p.HasAnswered
		existingParticipant.AnswerShown = p.AnswerShown
		existingParticipant.Attentive = p.Attentive
		existingParticipant.ActivelyConnected = p.ActivelyConnected
		existingParticipant.Owner = p.Owner
	} else {
		r.ParticipantsMutext.Lock()
		r.Participants = append(r.Participants, *p)
		r.ParticipantsMutext.Unlock()
	}
}

func (r *Room) AskQuestion(question string) {
	r.Question = question
	r.State = RoomStatePlaying
	r.ParticipantsMutext.Lock()
	for i := range r.Participants {
		r.Participants[i].HasAnswered = false
	}
	r.ParticipantsMutext.Unlock()
}

func (r *Room) RemoveParticipant(id string) {
	r.ParticipantsMutext.Lock()
	if len(r.Participants) == 1 {
		r.Participants = []Participant{}
		r.State = RoomStateEnd
		r.ParticipantsMutext.Unlock()
		return
	}
	for i, participant := range r.Participants {
		if participant.ID == id {
			r.Participants = append(r.Participants[:i], r.Participants[i+1:]...)
			if participant.Owner {
				r.Participants[0].Owner = true
			}
			break
		}
	}
	r.ParticipantsMutext.Unlock()
}

func (r *Room) IsOwner(participantId string) bool {
	r.ParticipantsMutext.RLock()
	for _, participant := range r.Participants {
		if participant.Owner {
			r.ParticipantsMutext.RUnlock()
			return participant.ID == participantId
		}
		if participant.ID == participantId {
			r.ParticipantsMutext.RUnlock()
			return participant.Owner
		}
	}
	r.ParticipantsMutext.RUnlock()
	return false
}

func (r *Room) ChangeOwner(id string, to string) {
	var fromParticipant, toParticipant *Participant
	r.ParticipantsMutext.RLock()
	for i, participant := range r.Participants {
		if participant.ID == id {
			fromParticipant = &r.Participants[i]
		}
		if participant.ID == to {
			toParticipant = &r.Participants[i]
		}
	}
	r.ParticipantsMutext.RUnlock()
	if fromParticipant == nil || toParticipant == nil {
		return
	}
	fromParticipant.Owner = false
	toParticipant.Owner = true
}
