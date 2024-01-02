import { Button, Flex, Textarea } from "@mantine/core";

export const QuestionForm: React.FC<{
  value: string;
  onChange: (question: string) => void;
  onSubmit: () => void;
}> = ({ value, onChange, onSubmit }) => {
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <Flex direction="column" gap="md" justify="center" align="center">
        <Textarea
          value={value}
          autoFocus
          placeholder="Enter your question"
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
          Ask
        </Button>
      </Flex>
    </form>
  );
};
