export type Participant = {
  id: string;
  name: string;
  has_answered: boolean;
  attentive: boolean;
  actively_connected: boolean;
  owner: boolean;
};

export type RoomState = {
  created_at: string;
  id: string;
  participants: Participant[];
  question: string;
  state: string;
};
