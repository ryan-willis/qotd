import { Button, Flex, TextInput } from "@mantine/core";
import { useEffect, useState } from "react";

type NameFormState = {
  playerName: string;
  editing: boolean;
};

export const NameForm: React.FC<{
  onSubmit: (playerName: string) => void;
  initialName: string;
  open?: boolean;
}> = ({ onSubmit, initialName, open = false }) => {
  const [formState, setFormState] = useState<NameFormState>({
    playerName: initialName,
    editing: open,
  });

  useEffect(() => {
    if (initialName) {
      setFormState({ playerName: initialName, editing: open ? true : false });
    }
  }, [initialName]);

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        if (formState.editing) {
          onSubmit(formState.playerName);
        }
      }}
    >
      {formState.editing ? (
        <Flex gap="xs">
          <TextInput
            autoFocus
            placeholder="Enter your name"
            value={formState.playerName}
            maxLength={20}
            onChange={(e) =>
              setFormState({ ...formState, playerName: e.currentTarget.value })
            }
          />
          <Button color="pale-indigo" type="submit">
            Save
          </Button>
          {!open && (
            <Button
              type="button"
              variant="outline"
              color="red"
              onClick={() => setFormState({ ...formState, editing: false })}
            >
              Cancel
            </Button>
          )}
        </Flex>
      ) : (
        <Flex align="center" gap="xs">
          Your name: <strong>{formState.playerName}</strong>
          <Button
            variant="outline"
            onClick={() => setFormState({ ...formState, editing: true })}
          >
            Edit
          </Button>
        </Flex>
      )}
    </form>
  );
};
