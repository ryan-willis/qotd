export const isValidRoomCode = (code: string): string | boolean => {
  if (code.length !== 4) return "room code must be 4 characters";
  if (!code.match(/^[ABCDEFGHJKLMNPQRSTUVWXYZ]{4}$/))
    return "invalid characters in room code";
  return true;
};
