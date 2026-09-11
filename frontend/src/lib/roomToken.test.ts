import { describe, expect, it, vi, afterEach } from "vitest";
import { fetchRoomToken } from "./roomToken";

describe("fetchRoomToken", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requests a token for the encoded room and identity", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ token: "signed-token", url: "ws://localhost:7880" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchRoomToken("http://backend", "team standup", "user-1");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://backend/api/rooms/team%20standup/token",
      expect.objectContaining({ method: "POST" }),
    );
    expect(result).toEqual({ token: "signed-token", url: "ws://localhost:7880" });
  });

  it("throws the backend error message on a non-ok response", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      statusText: "Bad Request",
      json: async () => ({ error: "identity must be between 3 and 64 characters" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchRoomToken("http://backend", "room", "ab")).rejects.toThrow(
      "identity must be between 3 and 64 characters",
    );
  });
});
