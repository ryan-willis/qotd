export const HOST = import.meta.env.PROD
  ? window.location.host
  : "localhost:9075";

const WS_PROTOCOL = window.location.protocol.includes("https") ? "wss" : "ws";

export const SOCKET_URL = `${WS_PROTOCOL}://${HOST}/ws`;

export const API_URL = `${window.location.protocol}//${HOST}`;
