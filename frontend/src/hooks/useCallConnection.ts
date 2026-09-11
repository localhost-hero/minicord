import { useCallback, useRef, useState } from "react";
import { Room, RoomEvent, type RemoteParticipant } from "livekit-client";
import { fetchRoomToken } from "../lib/roomToken";
import { reduceCallState, type CallState } from "../lib/callState";

export interface CallConnection {
  state: CallState;
  error: string | null;
  participants: string[];
  isMuted: boolean;
  join: (room: string, identity: string) => Promise<void>;
  leave: () => Promise<void>;
  toggleMute: () => Promise<void>;
}

/**
 * Owns the LiveKit room lifecycle so visual components never manage the
 * network connection directly.
 */
export function useCallConnection(apiBaseUrl: string): CallConnection {
  const [state, setState] = useState<CallState>("idle");
  const [error, setError] = useState<string | null>(null);
  const [participants, setParticipants] = useState<string[]>([]);
  const [isMuted, setIsMuted] = useState(false);
  const roomRef = useRef<Room | null>(null);
  const intentionalLeaveRef = useRef(false);

  const syncParticipants = useCallback((room: Room) => {
    const names = Array.from(room.remoteParticipants.values()).map(
      (participant: RemoteParticipant) => participant.identity,
    );
    setParticipants(names);
  }, []);

  const join = useCallback(
    async (roomName: string, identity: string) => {
      if (state !== "idle" && state !== "failed" && state !== "ended") {
        return;
      }

      setState(reduceCallState(state, { type: "JOIN" }));
      setError(null);

      try {
        const { token, url } = await fetchRoomToken(apiBaseUrl, roomName, identity);
        const room = new Room({
          audioCaptureDefaults: {
            noiseSuppression: true,
            echoCancellation: true,
            autoGainControl: true,
          },
        });
        roomRef.current = room;
        intentionalLeaveRef.current = false;

        room.on(RoomEvent.ParticipantConnected, () => syncParticipants(room));
        room.on(RoomEvent.ParticipantDisconnected, () => syncParticipants(room));
        room.on(RoomEvent.Reconnecting, () => {
          setState((current) => {
            try {
              return reduceCallState(current, { type: "CONNECTION_LOST" });
            } catch {
              return current;
            }
          });
        });
        room.on(RoomEvent.Reconnected, () => {
          setState((current) => {
            try {
              return reduceCallState(current, { type: "RECONNECTED" });
            } catch {
              return current;
            }
          });
        });
        room.on(RoomEvent.Disconnected, () => {
          const wasIntentional = intentionalLeaveRef.current;
          intentionalLeaveRef.current = false;
          roomRef.current = null;
          setParticipants([]);
          setState((current) => {
            if (current === "idle" || current === "ended") {
              return current;
            }
            try {
              return reduceCallState(current, { type: wasIntentional ? "LEAVE" : "ERROR" });
            } catch {
              return current;
            }
          });
        });

        await room.connect(url, token);
        await room.localParticipant.setMicrophoneEnabled(true);
        setIsMuted(false);
        syncParticipants(room);
        setState((current) => reduceCallState(current, { type: "CONNECTED" }));
      } catch (cause) {
        roomRef.current = null;
        setError(cause instanceof Error ? cause.message : "failed to join the room");
        setState((current) =>
          current === "connecting" ? reduceCallState(current, { type: "ERROR" }) : current,
        );
      }
    },
    [apiBaseUrl, state, syncParticipants],
  );

  const leave = useCallback(async () => {
    const room = roomRef.current;
    if (room) {
      intentionalLeaveRef.current = true;
      await room.disconnect();
    } else {
      setParticipants([]);
      setState((current) =>
        current === "idle" || current === "ended" ? current : reduceCallState(current, { type: "LEAVE" }),
      );
    }
  }, []);

  const toggleMute = useCallback(async () => {
    const room = roomRef.current;
    if (!room) {
      return;
    }
    const nextMuted = !isMuted;
    await room.localParticipant.setMicrophoneEnabled(!nextMuted);
    setIsMuted(nextMuted);
  }, [isMuted]);

  return { state, error, participants, isMuted, join, leave, toggleMute };
}
