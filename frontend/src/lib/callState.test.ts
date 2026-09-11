import { describe, expect, it } from "vitest";
import { reduceCallState } from "./callState";

describe("reduceCallState", () => {
  it("moves from idle to connecting on JOIN", () => {
    expect(reduceCallState("idle", { type: "JOIN" })).toBe("connecting");
  });

  it("moves from connecting to connected on CONNECTED", () => {
    expect(reduceCallState("connecting", { type: "CONNECTED" })).toBe("connected");
  });

  it("moves from connected to reconnecting on CONNECTION_LOST", () => {
    expect(reduceCallState("connected", { type: "CONNECTION_LOST" })).toBe("reconnecting");
  });

  it("moves from reconnecting back to connected on RECONNECTED", () => {
    expect(reduceCallState("reconnecting", { type: "RECONNECTED" })).toBe("connected");
  });

  it("moves from connected to ended on LEAVE", () => {
    expect(reduceCallState("connected", { type: "LEAVE" })).toBe("ended");
  });

  it("moves to failed on ERROR from connecting", () => {
    expect(reduceCallState("connecting", { type: "ERROR" })).toBe("failed");
  });

  it("allows rejoining from failed and ended", () => {
    expect(reduceCallState("failed", { type: "JOIN" })).toBe("connecting");
    expect(reduceCallState("ended", { type: "JOIN" })).toBe("connecting");
  });

  it("rejects invalid transitions", () => {
    expect(() => reduceCallState("idle", { type: "CONNECTED" })).toThrow();
    expect(() => reduceCallState("connected", { type: "JOIN" })).toThrow();
  });
});
