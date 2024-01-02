package room

type Participant struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Answer            string `json:"-"` // Answer is server-private
	HasAnswered       bool   `json:"has_answered"`
	AnswerShown       bool   `json:"-"` // AnswerShown is server-private
	Attentive         bool   `json:"attentive"`
	ActivelyConnected bool   `json:"actively_connected"`
	Owner             bool   `json:"owner"`
}

type ParticipantChannel struct {
	ID      string
	Channel chan string
}
