import {
  AppShell,
  Button,
  Divider,
  Flex,
  MantineProvider,
  Modal,
  ScrollArea,
  Title,
} from "@mantine/core";
import { Notifications, notifications } from "@mantine/notifications";
import { useEffect, useState } from "react";
import Header from "./components/header";
import { RoomForm } from "./components/room-form";
import { useDisclosure, useLocalStorage } from "@mantine/hooks";
import { generateUniqueId, isValidRoomCode } from "./utils";
import { RoomState } from "./types";
import { useRoomSocket } from "./hooks/room-socket";
import { NameForm } from "./components/name-form";
import { API_URL } from "./config";
import { ParticipantCard } from "./components/participant-card";
import { AnswerCard } from "./components/answer-card";
import { AnswerForm } from "./components/answer-form";
import { QuestionForm } from "./components/question-form";

export const App = () => {
  const [requestedRoomId, setRequestedRoomId] = useState<string | null>(null);
  const [participantId] = useLocalStorage({
    key: "playerId",
    defaultValue: window.localStorage.getItem("playerId") ?? generateUniqueId(),
  });
  const [participantToken] = useLocalStorage({
    key: "playerToken",
    defaultValue:
      window.localStorage.getItem("playerToken") ?? generateUniqueId(),
  });
  useEffect(() => {
    window.localStorage.setItem("playerId", participantId);
    window.localStorage.setItem("playerToken", participantToken);
  }, [participantId, participantToken]);

  useEffect(() => {
    const sp = new URLSearchParams(window.location.search);
    const roomParam = sp.get("room");
    if (roomParam !== null) {
      setRequestedRoomId(roomParam);
    }
  }, [window.location.search]);

  const [formState, setFormState] = useState({ question: "", answer: "" });
  const [playerName, setPlayerName] = useLocalStorage({
    key: "playerName",
    defaultValue: "",
  });
  const [
    nameDialogOpened,
    { toggle: toggleNameDialog, close: closeNameDialog },
  ] = useDisclosure(false);

  useEffect(() => {
    if (playerName) {
      if (requestedRoomId && isValidRoomCode(requestedRoomId)) {
        fetch(`${API_URL}/rooms/${requestedRoomId}`).then((res) => {
          if (res.ok) {
            sendMessage({
              do: "join",
              roomId: requestedRoomId,
              name: playerName,
              participantId,
              participantToken,
            });
            window.history.replaceState(
              {},
              document.title,
              `?room=${requestedRoomId}`
            );
          } else {
            notifications.show({
              title: "Error",
              message: `Room "${requestedRoomId}" not found`,
              color: "red",
              autoClose: 4000,
            });
            setRequestedRoomId(null);
          }
        });
      } else {
        window.history.replaceState(
          {},
          document.title,
          window.location.pathname
        );
      }
    }
  }, [requestedRoomId, playerName]);

  const [roomState, setRoomState] = useState<RoomState | null>(null);

  const { sendMessage, appState } = useRoomSocket({
    updateRoom: (room) => {
      setRoomState(room);
    },
    onConnect: () => {
      if (roomState && roomState.state != "finished") {
        sendMessage({
          do: "join",
          roomId: appState.roomId,
          name: playerName,
          participantId,
          participantToken,
        });
      }
    },
  });

  useEffect(() => {
    if (!roomState) return;
    if (roomState.state === "answered") {
      if (
        Object.keys(appState.answers).length === 0 &&
        roomState.participants.find((p) => p.id === participantId)?.owner
      ) {
        sendMessage({
          do: "get_answers",
          room_id: appState.roomId,
        });
      }
    }
  }, [appState.answers, roomState?.state, participantId, sendMessage]);

  useEffect(() => {
    if (!roomState) return;
    if (roomState.state === "playing") {
      setFormState({ question: "", answer: "" });
    }
    if (roomState.state === "finished") {
      sendMessage({ do: "leave", room_id: appState.roomId });
      setRoomState(null);
    }
  }, [roomState?.state]);

  useEffect(() => {
    const inattentiveEvent = () => {
      if (appState.roomId === "") return;
      sendMessage({
        do: "participant_inattentive",
        participant_id: participantId,
        room_id: appState.roomId,
      });
    };
    const attentiveEvent = () => {
      if (appState.roomId === "") return;
      sendMessage({
        do: "participant_attentive",
        participant_id: participantId,
        room_id: appState.roomId,
      });
    };
    window.addEventListener("blur", inattentiveEvent);
    window.addEventListener("focus", attentiveEvent);

    return () => {
      window.removeEventListener("blur", inattentiveEvent);
      window.removeEventListener("focus", attentiveEvent);
    };
  }, [sendMessage, participantId, appState.roomId]);

  const isOwner = roomState?.participants.find(
    (p) => p.id === participantId
  )?.owner;

  const answerState =
    appState.connected && roomState && roomState.state === "answered";

  const roomFormHandler = (roomId: string) => {
    if (roomId === "create") {
      let status = 0;
      fetch(`${API_URL}/rooms`, { method: "POST" })
        .then((res) => {
          status = res.status;
          return res.json();
        })
        .then((body) => {
          if (status === 429) {
            notifications.show({
              title: "Error",
              message: `You're creating rooms too quickly, try again in a few minutes.`,
              color: "red",
              autoClose: 4000,
            });
          } else if (status === 201) {
            setRequestedRoomId(body.room.id);
          } else {
            notifications.show({
              title: "Error",
              message: `Failed to create room`,
              color: "red",
              autoClose: 4000,
            });
          }
        });
    } else {
      if (requestedRoomId === roomId) {
        setRequestedRoomId(null);
        setTimeout(() => {
          setRequestedRoomId(roomId);
        }, 10);
      } else {
        setRequestedRoomId(roomId);
      }
    }
  };

  return (
    <MantineProvider
      defaultColorScheme="dark"
      theme={{
        colors: {
          "pale-indigo": [
            "#eef3ff",
            "#dee2f2",
            "#bdc2de",
            "#98a0ca",
            "#7a84ba",
            "#6672b0",
            "#5c68ac",
            "#4c5897",
            "#424e88",
            "#364379",
          ],
        },
      }}
    >
      <Notifications />
      <AppShell header={{ height: 60 }} padding="md">
        <Header
          {...appState}
          onDisconnect={() => {
            sendMessage({ do: "leave", room_id: appState.roomId });
            setRoomState(null);
          }}
        />
        {!playerName ? (
          <AppShell.Main>
            Welcome to QOTD! Enter a name to get started:
            <NameForm
              open={true}
              onSubmit={(playerName) => {
                setPlayerName(playerName);
                if (roomState) {
                  sendMessage({
                    do: "update_name",
                    room_id: appState.roomId,
                    name: playerName,
                  });
                }
              }}
              initialName={playerName}
            />
          </AppShell.Main>
        ) : (
          <AppShell.Main>
            <Modal
              title="Edit Name"
              opened={nameDialogOpened}
              onClose={closeNameDialog}
              size="auto"
              withCloseButton
            >
              <NameForm
                open={true}
                onSubmit={(playerName) => {
                  setPlayerName(playerName);
                  closeNameDialog();
                  if (roomState) {
                    sendMessage({
                      do: "update_name",
                      room_id: appState.roomId,
                      name: playerName,
                    });
                  }
                }}
                initialName={playerName}
              />
            </Modal>
            <ScrollArea>
              <Flex direction="row" gap="sm" wrap="nowrap" p="md" pt="0">
                <ParticipantCard
                  name={playerName}
                  onNameChange={() => {
                    toggleNameDialog();
                  }}
                  participant={
                    roomState
                      ? roomState.participants.find(
                          (p) => p.id === participantId
                        )
                      : undefined
                  }
                />
                {roomState && roomState.participants.length > 1 && (
                  <Divider orientation="vertical" />
                )}
                {roomState &&
                  roomState.participants
                    .filter((p) => p.id !== participantId)
                    .map((participant) => (
                      <ParticipantCard
                        key={participant.id}
                        participant={participant}
                      />
                    ))}
              </Flex>
            </ScrollArea>
            <Divider />

            {!roomState && <RoomForm onSubmit={roomFormHandler} />}

            <Flex
              direction="column"
              align="center"
              justify="center"
              p="sm"
              gap="sm"
            >
              {appState.connected &&
                roomState &&
                roomState.state == "playing" &&
                roomState.question != "" && (
                  <div>Question: {roomState.question}</div>
                )}
              {roomState && roomState.state === "waiting" && (
                <div>
                  {isOwner ? (
                    "Click Start to begin!"
                  ) : (
                    <>
                      Waiting for{" "}
                      <strong>
                        {roomState.participants.find((p) => p.owner)?.name ||
                          "owner"}
                      </strong>{" "}
                      to start...
                    </>
                  )}
                </div>
              )}
              {roomState &&
                roomState.state === "playing" &&
                !roomState.question && (
                  <div>
                    {isOwner ? "Ask a question!" : "Waiting for question..."}
                  </div>
                )}

              {answerState && (
                <Flex
                  direction="column"
                  gap="md"
                  justify="center"
                  align="center"
                >
                  <strong>{roomState.question}</strong>
                  {isOwner && "Click on a participant to show their answer!"}
                  {!isOwner && Object.keys(appState.answers).length === 0 ? (
                    "Waiting for answers to be shown..."
                  ) : !isOwner ? (
                    <Title order={3}>Answers</Title>
                  ) : (
                    ""
                  )}
                  <Flex direction="row" wrap="wrap" gap="sm" justify="center">
                    {roomState.participants.map((participant) => (
                      <AnswerCard
                        key={participant.id}
                        participant={participant}
                        answer={appState.answers[participant.id]}
                        showAnswer={
                          isOwner
                            ? () => {
                                sendMessage({
                                  do: "show_answer",
                                  room_id: appState.roomId,
                                  participant_id: participant.id,
                                });
                              }
                            : undefined
                        }
                      />
                    ))}
                  </Flex>
                </Flex>
              )}
              {isOwner &&
                appState.connected &&
                roomState.state !== "playing" && (
                  <>
                    <Button
                      color="pale-indigo"
                      onClick={() => {
                        sendMessage({
                          do: "start",
                          room_id: appState.roomId,
                        });
                        setFormState({ question: "", answer: "" });
                      }}
                    >
                      {roomState.state == "answered"
                        ? "Start New Round"
                        : "Start"}
                    </Button>
                  </>
                )}
              {appState.connected &&
                roomState?.state === "playing" &&
                roomState.question && (
                  <AnswerForm
                    value={formState.answer}
                    onChange={(value) => {
                      setFormState({ ...formState, answer: value });
                    }}
                    onSubmit={() => {
                      sendMessage({
                        do: "answer",
                        room_id: appState.roomId,
                        answer: formState.answer,
                      });
                    }}
                  />
                )}
              {isOwner &&
                appState.connected &&
                roomState?.state === "playing" &&
                !roomState?.question && (
                  <QuestionForm
                    value={formState.question}
                    onChange={(value) => {
                      setFormState({ ...formState, question: value });
                    }}
                    onSubmit={() => {
                      sendMessage({
                        do: "ask_question",
                        room_id: appState.roomId,
                        question: formState.question,
                      });
                    }}
                  />
                )}
              {appState.connected &&
                isOwner &&
                roomState?.state === "playing" &&
                roomState?.question && (
                  <Button
                    color="pale-indigo"
                    onClick={() => {
                      sendMessage({
                        do: "close_answers",
                        room_id: appState.roomId,
                      });
                    }}
                  >
                    Close Answers
                  </Button>
                )}
            </Flex>
          </AppShell.Main>
        )}
      </AppShell>
    </MantineProvider>
  );
};
