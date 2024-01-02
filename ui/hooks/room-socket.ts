import { useCallback, useState } from "react";
import useWebSocket from "react-use-websocket";
import { RoomState } from "../types";
import { SOCKET_URL } from "../config";

const LATENCY_INTERVAL_MS = 3_000;

export const useRoomSocket = ({
  updateRoom,
  onConnect,
}: {
  updateRoom: (room: RoomState) => void;
  onConnect: () => void;
}) => {
  const [appState, setAppState] = useState({
    latency: -1,
    roomId: "",
    connected: false,
    pings: 0,
    answers: {} as Record<string, string>,
  });

  const ws = useWebSocket(SOCKET_URL, {
    onClose() {
      setAppState({ ...appState, connected: false, pings: 0 });
    },
    onMessage(event) {
      if (event.data == "ok") {
        setAppState({ ...appState, connected: true });
        onConnect();
        if (appState.latency > -1) return;
        ws.sendJsonMessage({ do: "latency", stamp: `${Date.now()}` });
        return;
      }
      let msg;
      try {
        msg = JSON.parse(event.data);
      } catch (e) {
        console.error(e);
        return;
      }
      if (msg["latency"]) {
        setAppState({
          ...appState,
          latency: Date.now() - Number(msg["latency"]),
          pings: appState.pings + 1,
        });
        setTimeout(
          () => {
            if (ws.readyState !== WebSocket.OPEN) return;
            ws.sendJsonMessage({ do: "latency", stamp: `${Date.now()}` });
          },
          appState.pings == 0 ? 500 : LATENCY_INTERVAL_MS
        );
        return;
      }
      console.log(msg);
      if (msg["room"]) {
        setAppState({ ...appState, roomId: msg["room"]["id"] });
        updateRoom(msg["room"]);
      }
      if (msg["answers"]) {
        setAppState({ ...appState, answers: msg["answers"] });
      }
    },
    shouldReconnect: () => true,
  });

  const sendMessage = useCallback(
    (message: Record<string, string>) => {
      if (message["do"] === "leave") {
        setAppState({ ...appState, roomId: "" });
      }
      ws.sendJsonMessage(message);
    },
    [ws, appState]
  );

  return {
    appState,
    sendMessage,
  };
};
