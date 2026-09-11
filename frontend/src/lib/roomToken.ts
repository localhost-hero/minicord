export interface RoomToken {
  token: string;
  url: string;
}

/**
 * Requests a short-lived LiveKit access token from the backend for the given
 * room and participant identity.
 */
export async function fetchRoomToken(
  apiBaseUrl: string,
  room: string,
  identity: string,
): Promise<RoomToken> {
  const response = await fetch(`${apiBaseUrl}/api/rooms/${encodeURIComponent(room)}/token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ identity }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error ?? `token request failed with status ${response.status}`);
  }

  return response.json();
}
