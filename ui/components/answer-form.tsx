import { Box, Button, Flex, Textarea } from "@mantine/core";
import { useState } from "react";

export const AnswerForm: React.FC<{
  value: string;
  onChange: (answer: string) => void;
  onSubmit: () => void;
}> = ({ value, onChange, onSubmit }) => {
  const [submitted, setSubmitted] = useState(false);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        setSubmitted(true);
        onSubmit();
      }}
    >
      <Flex direction="column" gap="md" justify="center" align="center">
        {submitted ? (
          <>
            <strong>Answer submitted:</strong>
            <Box style={{ whiteSpace: "pre-wrap" }}>{value}</Box>
            <Button
              color="lime"
              onClick={(e) => {
                e.preventDefault();
                console.log("yo");
                setSubmitted(false);
              }}
              type="button"
            >
              Edit
            </Button>
          </>
        ) : (
          <>
            <Textarea
              value={value}
              autoFocus
              placeholder="Enter your answer"
              onChange={(e) => {
                onChange(e.currentTarget.value);
              }}
              autosize
              minRows={3}
              maxRows={5}
              w="30rem"
              maw="88vw"
            />
            <Button color="pale-indigo" type="submit">
              Answer
            </Button>
          </>
        )}
      </Flex>
    </form>
  );
};
