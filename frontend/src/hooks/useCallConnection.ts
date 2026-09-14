import { useCallback, useRef, useState } from "react";
import { Room, RoomEvent, Track, LocalAudioTrack, type RemoteParticipant, type RemoteTrack } from "livekit-client";
import { KrispNoiseFilter, isKrispNoiseFilterSupported } from "@livekit/krisp-noise-filter";
import { fetchRoomToken } from "../lib/roomToken";
import { reduceCallState, type CallState } from "../lib/callState";

export interface CallConnection {
  state: CallState;
  error: string | null;
  participants: string[];
  isMuted: boolean;
  isNoiseSuppressionEnabled: boolean;
  isNoiseSuppressionSupported: boolean;
  join: (room: string, identity: string) => Promise<void>;
  leave: () => Promise<void>;
  toggleMute: () => Promise<void>;
  toggleNoiseSuppression: () => Promise<void>;
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
  const [isNoiseSuppressionEnabled, setIsNoiseSuppressionEnabled] = useState(false);
  const isNoiseSuppressionSupported = isKrispNoiseFilterSupported();
  const roomRef = useRef<Room | null>(null);
  const krispProcessorRef = useRef<ReturnType<typeof KrispNoiseFilter> | null>(null);
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
        const room = new Room();
        roomRef.current = room;
        intentionalLeaveRef.current = false;

        room.on(RoomEvent.ParticipantConnected, () => syncParticipants(room));
        room.on(RoomEvent.ParticipantDisconnected, () => syncParticipants(room));
        room.on(RoomEvent.TrackSubscribed, (track: RemoteTrack) => {
          const element = track.attach();
          element.style.display = "none";
          document.body.appendChild(element);
        });
        room.on(RoomEvent.TrackUnsubscribed, (track: RemoteTrack) => {
          track.detach().forEach((element) => element.remove());
        });
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
        await room.startAudio();
        await room.localParticipant.setMicrophoneEnabled(true);
        if (isNoiseSuppressionSupported) {
          try {
            const krispProcessor = KrispNoiseFilter();
            krispProcessorRef.current = krispProcessor;
            const trackPub = room.localParticipant.getTrackPublication(Track.Source.Microphone);
            const localTrack = trackPub?.track as LocalAudioTrack | undefined;
            if (localTrack) {
              const audioCtxClass =
                window.AudioContext ||
                (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
              if (audioCtxClass) {
                localTrack.setAudioContext(new audioCtxClass());
              }
              await localTrack.setProcessor(krispProcessor);
              setIsNoiseSuppressionEnabled(true);
            }
          } catch (cause) {
            console.warn("Failed to enable noise processor:", cause);
            setIsNoiseSuppressionEnabled(false);
          }
        } else {
          setIsNoiseSuppressionEnabled(false);
        }
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
    krispProcessorRef.current = null;
    setIsNoiseSuppressionEnabled(false);
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

  const toggleNoiseSuppression = useCallback(async () => {
    const room = roomRef.current;
    if (!room) {
      return;
    }
    const trackPub = room.localParticipant.getTrackPublication(Track.Source.Microphone);
    const localTrack = trackPub?.track as LocalAudioTrack | undefined;
    if (!localTrack) {
      return;
    }

    if (isNoiseSuppressionEnabled) {
      try {
        await localTrack.stopProcessor();
        krispProcessorRef.current = null;
        setIsNoiseSuppressionEnabled(false);
      } catch (cause) {
        console.warn("Failed to stop noise processor:", cause);
      }
    } else {
      try {
        const krispProcessor = KrispNoiseFilter();
        krispProcessorRef.current = krispProcessor;
        const hasAudioCtx = (localTrack as unknown as { audioContext?: AudioContext }).audioContext;
        if (!hasAudioCtx) {
          const audioCtxClass =
            window.AudioContext ||
            (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
          if (audioCtxClass) {
            localTrack.setAudioContext(new audioCtxClass());
          }
        }
        await localTrack.setProcessor(krispProcessor);
        setIsNoiseSuppressionEnabled(true);
      } catch (cause) {
        console.warn("Failed to enable noise processor:", cause);
      }
    }
  }, [isNoiseSuppressionEnabled]);

  return {
    state,
    error,
    participants,
    isMuted,
    isNoiseSuppressionEnabled,
    isNoiseSuppressionSupported,
    join,
    leave,
    toggleMute,
    toggleNoiseSuppression,
  };
}
