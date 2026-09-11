export type CallState =
  | "idle"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "failed"
  | "ended";

export type CallEvent =
  | { type: "JOIN" }
  | { type: "CONNECTED" }
  | { type: "CONNECTION_LOST" }
  | { type: "RECONNECTED" }
  | { type: "ERROR" }
  | { type: "LEAVE" };

const allowedTransitions: Record<CallState, Partial<Record<CallEvent["type"], CallState>>> = {
  idle: { JOIN: "connecting" },
  connecting: { CONNECTED: "connected", ERROR: "failed", LEAVE: "idle" },
  connected: { CONNECTION_LOST: "reconnecting", ERROR: "failed", LEAVE: "ended" },
  reconnecting: { RECONNECTED: "connected", ERROR: "failed", LEAVE: "ended" },
  failed: { JOIN: "connecting", LEAVE: "idle" },
  ended: { JOIN: "connecting" },
};

/**
 * Applies a call event to the current state, rejecting transitions that are not
 * explicitly allowed by the call lifecycle state machine.
 */
export function reduceCallState(state: CallState, event: CallEvent): CallState {
  const nextState = allowedTransitions[state][event.type];
  if (!nextState) {
    throw new Error(`invalid call state transition: ${event.type} from ${state}`);
  }
  return nextState;
}
