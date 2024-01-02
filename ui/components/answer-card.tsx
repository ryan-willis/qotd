import { Badge, Box, Button, Card, Flex, Transition } from "@mantine/core";
import { useState } from "react";
import { Participant } from "../types";

export const AnswerCard: React.FC<{
  participant: Participant;
  answer?: string;
  showAnswer?: () => void;
}> = ({ participant, answer, showAnswer }) => {
  const [shown, setShown] = useState(false);

  return (
    <Transition
      mounted={showAnswer ? true : answer !== undefined}
      transition="slide-left"
      duration={300}
      timingFunction="ease"
    >
      {(styles) => (
        <Card style={{ ...styles, maxWidth: "40vw" }}>
          <Flex direction="column" gap="sm" justify="center" align="center">
            <Box
              style={{
                wordWrap: "break-word",
                textAlign: "center",
                maxWidth: "100%",
              }}
            >
              <strong>{participant.name}</strong>
            </Box>
            <Box
              style={{
                wordWrap: "break-word",
                maxWidth: "100%",
                whiteSpace: "pre-wrap",
              }}
            >
              {answer}
            </Box>
            {showAnswer && !shown ? (
              <Button
                color="pale-indigo"
                onClick={() => {
                  setShown(true);
                  showAnswer();
                }}
              >
                Show
              </Button>
            ) : showAnswer ? (
              <Badge color="lime">Shown</Badge>
            ) : null}
          </Flex>
        </Card>
      )}
    </Transition>
  );
};
