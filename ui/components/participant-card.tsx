import { Button, Card, Flex } from "@mantine/core";
import { Participant } from "../types";

export const ParticipantCard: React.FC<{
  name?: string;
  onNameChange?: () => void;
  participant?: Participant;
}> = ({ onNameChange, participant, name }) => {
  return (
    <Card
      style={{
        width: "7rem",
        opacity: onNameChange || participant?.attentive ? 1 : 0.5,
        // maxHeight: "9rem",
        textWrap: "wrap",
      }}
    >
      <Flex justify="center" align="center" direction="column" gap=".5rem">
        <strong>
          {participant && participant.owner
            ? "👑"
            : !participant
            ? ""
            : participant.actively_connected
            ? "👤"
            : "⛓️‍💥"}
        </strong>
        <strong
          style={{
            lineHeight: "1",
            textAlign: "center",
            maxWidth: "10rem",
          }}
        >
          {name || participant?.name}
        </strong>
        {participant ? (participant.has_answered ? "✅" : "⋯") : ""}
        {onNameChange && (
          <Button
            size="compact-xs"
            color="pale-indigo"
            onClick={() => onNameChange()}
          >
            Edit
          </Button>
        )}
      </Flex>
    </Card>
  );
};
