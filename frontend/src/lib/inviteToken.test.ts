import { describe, expect, it, vi, afterEach } from "vitest";
import { fetchInviteRoomToken } from "./inviteToken";

describe("fetchInviteRoomToken", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requests a token for a valid invite", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ token: "signed-token", url: "ws://localhost:7880" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchInviteRoomToken("http://backend", "invite-123", "guest-1");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://backend/api/invites/invite-123/token",
      expect.objectContaining({ method: "POST" }),
    );
    expect(result).toEqual({ token: "signed-token", url: "ws://localhost:7880" });
  });

  it("throws the backend error message on a rejected invite", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 410,
      statusText: "Gone",
      json: async () => ({ error: "invite has expired or been revoked" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchInviteRoomToken("http://backend", "invite-123", "guest-1")).rejects.toThrow(
      "invite has expired or been revoked",
    );
  });
});
