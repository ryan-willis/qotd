import { Button, TextInput, Text, Group } from "@mantine/core";
import { useState } from "react";
import { isValidRoomCode } from "../utils";

export const RoomForm: React.FC<{
  onSubmit: (roomId: string) => void;
}> = ({ onSubmit }) => {
  const [roomId, setRoomId] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        const msg = isValidRoomCode(roomId);
        if (msg === true) {
          setErrorMessage("");
          onSubmit(roomId);
        } else {
          setErrorMessage(`${msg}`);
        }
      }}
    >
      <Group align="flex-start" gap="sm" mt="md">
        <TextInput
          autoFocus
          error={errorMessage}
          placeholder="Room code"
          value={roomId}
          onChange={(e) => setRoomId(e.currentTarget.value.toUpperCase())}
          maxLength={4}
          w="8rem"
        />
        <Button color="pale-indigo" type="submit">
          Join
        </Button>
        <Group align="center" gap="sm">
          <Text>or</Text>
          <Button
            color="pale-indigo"
            type="button"
            onClick={() => onSubmit("create")}
          >
            Create Room
          </Button>
        </Group>
      </Group>
    </form>
  );
};
