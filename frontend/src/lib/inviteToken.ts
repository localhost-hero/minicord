export interface RoomToken {
  token: string;
  url: string;
}

/**
 * Requests a short-lived LiveKit access token for a guest who presents a valid
 * invite token. The invite token is passed in the path and the guest identity in
 * the JSON payload.
 */
export async function fetchInviteRoomToken(
  apiBaseUrl: string,
  inviteToken: string,
  identity: string,
): Promise<RoomToken> {
  const response = await fetch(`${apiBaseUrl}/api/invites/${encodeURIComponent(inviteToken)}/token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ identity }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error ?? `invite token request failed with status ${response.status}`);
  }

  return response.json();
}
