import { Component, StrictMode, useEffect, useRef, useState, type ErrorInfo, type FormEvent, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import { Room, RoomEvent, Track, LocalAudioTrack, type RemoteParticipant, type RemoteTrack } from "livekit-client";
import { KrispNoiseFilter, isKrispNoiseFilterSupported } from "@livekit/krisp-noise-filter";
import "./styles.css";
import { useCallConnection } from "./hooks/useCallConnection";
import { fetchInviteRoomToken } from "./lib/inviteToken";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "";

const stateLabels: Record<string, string> = {
  idle: "Enter a room to join",
  connecting: "Connecting to the room",
  connected: "Connected",
  reconnecting: "Reconnecting",
  failed: "Connection failed",
  ended: "Call ended",
};

type ChatMessage = {
  id: number;
  room: string;
  identity: string;
  body: string;
  createdAt: string;
};

class ErrorBoundary extends Component<{ children: ReactNode }, { hasError: boolean; error: Error | null }> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("UI Error caught by ErrorBoundary:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="app-shell">
          <section className="panel-wrap standalone-card">
            <div className="brand-block">
              <p className="eyebrow">Minicord Error</p>
              <h1>Something went wrong</h1>
            </div>
            <div className="card">
              <p className="helper-text error-text">
                {this.state.error?.message ?? "An unexpected UI error occurred."}
              </p>
              <div className="button-row" style={{ marginTop: "16px" }}>
                <button type="button" onClick={() => (window.location.href = "/")}>
                  Reload Page
                </button>
              </div>
            </div>
          </section>
        </main>
      );
    }
    return this.props.children;
  }
}

const formatMessageTime = (dateStr: string) => {
  if (!dateStr) return "";
  try {
    const d = new Date(dateStr);
    return isNaN(d.getTime()) ? "" : d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
};

function App() {
  const call = useCallConnection(API_BASE_URL);
  const [room, setRoom] = useState("");
  const [identity, setIdentity] = useState("");

  const [adminUsername, setAdminUsername] = useState("");
  const [adminPassword, setAdminPassword] = useState("");
  const [adminStatus, setAdminStatus] = useState<string | null>(null);
  const [adminLoggedIn, setAdminLoggedIn] = useState(false);
  const [inviteRoom, setInviteRoom] = useState("");
  const [inviteTTLHours, setInviteTTLHours] = useState("0");
  const [createdInviteToken, setCreatedInviteToken] = useState<string | null>(null);
  const [inviteStatus, setInviteStatus] = useState<string | null>(null);

  const [inviteToken, setInviteToken] = useState("");
  const [guestIdentity, setGuestIdentity] = useState("");
  const [guestRoom, setGuestRoom] = useState<string | null>(null);
  const [guestState, setGuestState] = useState<"idle" | "connecting" | "connected" | "failed" | "ended">("idle");
  const [guestError, setGuestError] = useState<string | null>(null);
  const [guestMessages, setGuestMessages] = useState<ChatMessage[]>([]);
  const [chatDraft, setChatDraft] = useState("");
  const [isGuestNoiseSuppressionEnabled, setIsGuestNoiseSuppressionEnabled] = useState(false);
  const [isGuestMuted, setIsGuestMuted] = useState(false);
  const [guestParticipants, setGuestParticipants] = useState<string[]>([]);
  const isNoiseSuppressionSupported = isKrispNoiseFilterSupported();

  const [magicToken, setMagicToken] = useState<string | null>(null);
  const [magicRoomName, setMagicRoomName] = useState<string | null>(null);
  const [magicLoading, setMagicLoading] = useState(false);
  const [magicError, setMagicError] = useState<string | null>(null);
  const [copyStatus, setCopyStatus] = useState<string | null>(null);

  const guestRoomRef = useRef<Room | null>(null);
  const guestSocketRef = useRef<WebSocket | null>(null);
  const guestKrispProcessorRef = useRef<ReturnType<typeof KrispNoiseFilter> | null>(null);

  useEffect(() => {
    const path = window.location.pathname;
    const searchParams = new URLSearchParams(window.location.search);
    let token: string | null = null;

    if (path.startsWith("/invite/")) {
      token = path.substring("/invite/".length).trim();
    } else if (searchParams.has("invite")) {
      token = searchParams.get("invite")?.trim() || null;
    } else if (searchParams.has("token")) {
      token = searchParams.get("token")?.trim() || null;
    }

    if (token) {
      setMagicToken(token);
      setInviteToken(token);
      setMagicLoading(true);

      fetch(`${API_BASE_URL}/api/invites/${encodeURIComponent(token)}`)
        .then(async (res) => {
          if (!res.ok) {
            const body = await res.json().catch(() => ({ error: res.statusText }));
            throw new Error(body.error ?? "Invite link is invalid or has expired");
          }
          return res.json() as Promise<{ room: string }>;
        })
        .then((data) => {
          setMagicRoomName(data.room);
          setMagicLoading(false);
        })
        .catch((err) => {
          setMagicError(err instanceof Error ? err.message : "Invalid invite link");
          setMagicLoading(false);
        });
    }
  }, []);

  const handleJoinRoom = (event: FormEvent) => {
    event.preventDefault();
    void call.join(room, identity);
  };

  const handleAdminLogin = async (event: FormEvent) => {
    event.preventDefault();
    setAdminStatus(null);

    const response = await fetch(`${API_BASE_URL}/api/admin/login`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: adminUsername, password: adminPassword }),
    });

    if (!response.ok) {
      const body = await response.json().catch(() => ({ error: response.statusText }));
      setAdminStatus(body.error ?? "admin login failed");
      setAdminLoggedIn(false);
      return;
    }

    setAdminStatus("Admin session active");
    setAdminLoggedIn(true);
  };

  const handleAdminLogout = async () => {
    await fetch(`${API_BASE_URL}/api/admin/logout`, {
      method: "POST",
      credentials: "include",
    });
    setAdminLoggedIn(false);
    setAdminStatus("Signed out");
  };

  const handleCreateInvite = async (event: FormEvent) => {
    event.preventDefault();
    setInviteStatus(null);
    setCreatedInviteToken(null);

    const ttlHours = Number(inviteTTLHours);
    if (!Number.isFinite(ttlHours) || ttlHours < 0) {
      setInviteStatus("Expiry must be zero or a positive number of hours");
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/admin/rooms/${encodeURIComponent(inviteRoom.trim())}/invites`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ttlSeconds: Math.round(ttlHours * 60 * 60) }),
      });
      const body = await response.json().catch(() => ({ error: response.statusText }));
      if (!response.ok) {
        throw new Error(body.error ?? "failed to create invite");
      }

      setCreatedInviteToken(body.token);
      setInviteStatus(`Invite created for #${body.room}`);
    } catch (error) {
      setInviteStatus(error instanceof Error ? error.message : "failed to create invite");
    }
  };

  const connectGuestToRoom = async (overrideToken?: string) => {
    const tokenValue = (overrideToken || inviteToken || magicToken || "").trim();
    if (!tokenValue || !guestIdentity.trim()) {
      setGuestError("Provide an invite token and a guest name");
      return;
    }

    setGuestState("connecting");
    setGuestError(null);

    try {
      const inviteResponse = await fetch(`${API_BASE_URL}/api/invites/${encodeURIComponent(tokenValue)}`);
      if (!inviteResponse.ok) {
        const body = await inviteResponse.json().catch(() => ({ error: inviteResponse.statusText }));
        throw new Error(body.error ?? "invite not found");
      }
      const inviteData = (await inviteResponse.json()) as { room: string };
      const roomName = inviteData.room;

      const { token, url } = await fetchInviteRoomToken(API_BASE_URL, tokenValue, guestIdentity.trim());
      const roomInstance = new Room();
      guestRoomRef.current = roomInstance;

      const syncGuestParticipants = (r: Room) => {
        const names = Array.from(r.remoteParticipants.values()).map(
          (participant: RemoteParticipant) => participant.identity,
        );
        setGuestParticipants(names);
      };

      roomInstance.on(RoomEvent.ParticipantConnected, () => syncGuestParticipants(roomInstance));
      roomInstance.on(RoomEvent.ParticipantDisconnected, () => syncGuestParticipants(roomInstance));
      roomInstance.on(RoomEvent.Disconnected, () => {
        setGuestState("ended");
      });
      roomInstance.on(RoomEvent.TrackSubscribed, (track: RemoteTrack) => {
        const element = track.attach();
        element.style.display = "none";
        document.body.appendChild(element);
      });
      roomInstance.on(RoomEvent.TrackUnsubscribed, (track: RemoteTrack) => {
        track.detach().forEach((element) => element.remove());
      });
      roomInstance.on(RoomEvent.Reconnecting, () => setGuestState("connecting"));
      roomInstance.on(RoomEvent.Reconnected, () => setGuestState("connected"));

      await roomInstance.connect(url, token);
      await roomInstance.startAudio();
      await roomInstance.localParticipant.setMicrophoneEnabled(true);
      setIsGuestMuted(false);
      syncGuestParticipants(roomInstance);

      if (isNoiseSuppressionSupported) {
        try {
          const krispProcessor = KrispNoiseFilter();
          guestKrispProcessorRef.current = krispProcessor;
          const trackPub = roomInstance.localParticipant.getTrackPublication(Track.Source.Microphone);
          const localTrack = trackPub?.track as LocalAudioTrack | undefined;
          if (localTrack) {
            const audioCtxClass =
              window.AudioContext ||
              (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
            if (audioCtxClass) {
              localTrack.setAudioContext(new audioCtxClass());
            }
            await localTrack.setProcessor(krispProcessor);
            setIsGuestNoiseSuppressionEnabled(true);
          }
        } catch (cause) {
          console.warn("Failed to enable guest noise processor:", cause);
          setIsGuestNoiseSuppressionEnabled(false);
        }
      } else {
        setIsGuestNoiseSuppressionEnabled(false);
      }
      setGuestRoom(roomName);
      setGuestState("connected");

      const historyResponse = await fetch(
        `${API_BASE_URL}/api/rooms/${encodeURIComponent(roomName)}/messages?token=${encodeURIComponent(tokenValue)}`,
      );
      if (historyResponse.ok) {
        const history = (await historyResponse.json()) as ChatMessage[] | null;
        setGuestMessages(Array.isArray(history) ? history : []);
      }

      const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsHost = window.location.host;
      const socket = new WebSocket(
        `${wsProtocol}//${wsHost}/api/rooms/${encodeURIComponent(roomName)}/chat?token=${encodeURIComponent(tokenValue)}&identity=${encodeURIComponent(guestIdentity.trim())}`,
      );
      guestSocketRef.current = socket;
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as ChatMessage;
          if (message && message.body) {
            setGuestMessages((current) => [...(current ?? []), message]);
          }
        } catch (err) {
          console.warn("Failed to parse websocket chat message:", err);
        }
      };
      socket.onerror = () => setGuestError("Chat socket is unavailable");
    } catch (error) {
      setGuestState("failed");
      setGuestError(error instanceof Error ? error.message : "failed to join guest room");
    }
  };

  const toggleGuestMute = async () => {
    if (!guestRoomRef.current) return;
    const nextMuted = !isGuestMuted;
    await guestRoomRef.current.localParticipant.setMicrophoneEnabled(!nextMuted);
    setIsGuestMuted(nextMuted);
  };

  const disconnectGuestRoom = () => {
    guestSocketRef.current?.close();
    guestSocketRef.current = null;
    void guestRoomRef.current?.disconnect();
    guestRoomRef.current = null;
    guestKrispProcessorRef.current = null;
    setIsGuestNoiseSuppressionEnabled(false);
    setIsGuestMuted(false);
    setGuestParticipants([]);
    setGuestRoom(null);
    setGuestMessages([]);
    setChatDraft("");
    setGuestState("idle");
  };

  const toggleGuestNoiseSuppression = async () => {
    const room = guestRoomRef.current;
    if (!room) {
      return;
    }
    const trackPub = room.localParticipant.getTrackPublication(Track.Source.Microphone);
    const localTrack = trackPub?.track as LocalAudioTrack | undefined;
    if (!localTrack) {
      return;
    }

    if (isGuestNoiseSuppressionEnabled) {
      try {
        await localTrack.stopProcessor();
        guestKrispProcessorRef.current = null;
        setIsGuestNoiseSuppressionEnabled(false);
      } catch (cause) {
        console.warn("Failed to stop guest noise processor:", cause);
      }
    } else {
      try {
        const krispProcessor = KrispNoiseFilter();
        guestKrispProcessorRef.current = krispProcessor;
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
        setIsGuestNoiseSuppressionEnabled(true);
      } catch (cause) {
        console.warn("Failed to enable guest noise processor:", cause);
      }
    }
  };

  const sendChatMessage = () => {
    const body = chatDraft.trim();
    if (!guestRoom || !guestSocketRef.current || guestSocketRef.current.readyState !== WebSocket.OPEN || !body) {
      return;
    }

    guestSocketRef.current.send(JSON.stringify({ body }));
    setChatDraft("");
  };

  const canJoin = call.state === "idle" || call.state === "failed" || call.state === "ended";

  if (magicToken) {
    return (
      <main className="app-shell">
        <section className="panel-wrap standalone-card" aria-labelledby="app-title">
          <div className="brand-block">
            <p className="eyebrow">Minicord Voice & Chat</p>
            <h1 id="app-title">{magicRoomName ? `#${magicRoomName}` : "Magic Invite"}</h1>
          </div>

          {magicLoading && <p className="helper-text">Validating invite link...</p>}

          {magicError && (
            <div className="card">
              <h2>Invalid invite link</h2>
              <p className="helper-text error-text">{magicError}</p>
              <div className="button-row">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => {
                    window.history.pushState({}, "", "/");
                    setMagicToken(null);
                    setMagicError(null);
                  }}
                >
                  Go to home page
                </button>
              </div>
            </div>
          )}

          {!magicLoading && !magicError && magicRoomName && (
            <>
              {!guestRoom ? (
                <div className="card">
                  <h2>Enter your name to join room & chat</h2>
                  <form
                    className="stack-form"
                    onSubmit={(e) => {
                      e.preventDefault();
                      void connectGuestToRoom();
                    }}
                  >
                    <label htmlFor="magic-guest-name">Your name</label>
                    <input
                      id="magic-guest-name"
                      value={guestIdentity}
                      onChange={(e) => setGuestIdentity(e.target.value)}
                      placeholder="e.g. Alex"
                      autoFocus
                      required
                    />

                    <button type="submit" disabled={guestState === "connecting"}>
                      {guestState === "connecting" ? "Connecting…" : "Join Room & Chat"}
                    </button>
                  </form>
                  {guestError && <p className="helper-text error-text">{guestError}</p>}
                </div>
              ) : (
                <div className="card">
                  <div className="call-panel">
                    <p className="participants">
                      Participants: {(guestParticipants ?? []).length > 0 ? (guestParticipants ?? []).join(", ") : "just you"}
                    </p>
                    <div className="button-row">
                      <button type="button" onClick={() => void toggleGuestMute()} aria-pressed={isGuestMuted}>
                        {isGuestMuted ? "Unmute" : "Mute"}
                      </button>
                      {isNoiseSuppressionSupported && (
                        <button
                          type="button"
                          className={isGuestNoiseSuppressionEnabled ? "" : "secondary"}
                          onClick={() => void toggleGuestNoiseSuppression()}
                          aria-pressed={isGuestNoiseSuppressionEnabled}
                        >
                          {isGuestNoiseSuppressionEnabled ? "Noise: On" : "Noise: Off"}
                        </button>
                      )}
                      <button
                        type="button"
                        className="secondary"
                        onClick={() => {
                          disconnectGuestRoom();
                          window.history.pushState({}, "", "/");
                          setMagicToken(null);
                        }}
                      >
                        Leave
                      </button>
                    </div>
                  </div>

                  <div className="chat-panel">
                    <div className="chat-header">
                      <strong>Chat in #{guestRoom}</strong>
                      <span>{guestState === "connected" ? "live" : guestState}</span>
                    </div>

                    <div className="messages" aria-live="polite">
                      {(guestMessages ?? []).length === 0 ? (
                        <p className="empty-state">No messages yet.</p>
                      ) : (
                        (guestMessages ?? []).map((message) => (
                          <div className="message" key={message.id ?? `${message.identity}-${message.createdAt}`}>
                            <div className="message-meta">
                              <span>{message.identity}</span>
                              <time>{formatMessageTime(message.createdAt)}</time>
                            </div>
                            <div>{message.body}</div>
                          </div>
                        ))
                      )}
                    </div>

                    <div className="composer">
                      <input
                        value={chatDraft}
                        onChange={(event) => setChatDraft(event.target.value)}
                        onKeyDown={(event) => {
                          if (event.key === "Enter") {
                            event.preventDefault();
                            sendChatMessage();
                          }
                        }}
                        placeholder="Message the room"
                      />
                      <button type="button" onClick={sendChatMessage}>
                        Send
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </section>
      </main>
    );
  }

  return (
    <main className="app-shell">
      <section className="panel-wrap" aria-labelledby="app-title">
        <div className="brand-block">
          <p className="eyebrow">Self-hosted voice</p>
          <h1 id="app-title">Minicord</h1>
        </div>

        <div className="layout-grid">
          <section className="card admin-card">
            <h2>Admin</h2>
            <form className="stack-form" onSubmit={handleAdminLogin}>
              <label htmlFor="username">Username</label>
              <input
                id="username"
                value={adminUsername}
                onChange={(event) => setAdminUsername(event.target.value)}
                placeholder="admin"
                required
              />

              <label htmlFor="password">Password</label>
              <input
                id="password"
                type="password"
                value={adminPassword}
                onChange={(event) => setAdminPassword(event.target.value)}
                placeholder="••••••••"
                required
              />

              <div className="button-row">
                <button type="submit">Login</button>
                {adminLoggedIn && (
                  <button type="button" className="secondary" onClick={handleAdminLogout}>
                    Logout
                  </button>
                )}
              </div>
            </form>
            {adminStatus && <p className="helper-text">{adminStatus}</p>}
            {adminLoggedIn && (
              <form className="stack-form" onSubmit={(event) => void handleCreateInvite(event)}>
                <h3>Create invite</h3>
                <label htmlFor="invite-room">Room</label>
                <input
                  id="invite-room"
                  value={inviteRoom}
                  onChange={(event) => setInviteRoom(event.target.value)}
                  placeholder="general"
                  required
                />

                <label htmlFor="invite-expiry">Expires after (hours)</label>
                <input
                  id="invite-expiry"
                  type="number"
                  min="0"
                  step="1"
                  value={inviteTTLHours}
                  onChange={(event) => setInviteTTLHours(event.target.value)}
                />

                <button type="submit">Create invite</button>
              </form>
            )}
            {inviteStatus && <p className="helper-text">{inviteStatus}</p>}
            {createdInviteToken && (
              <div className="stack-form">
                <label htmlFor="created-invite-token">Magic invite link</label>
                <div className="composer">
                  <input
                    id="created-invite-token"
                    value={`${window.location.origin}/invite/${createdInviteToken}`}
                    readOnly
                  />
                  <button
                    type="button"
                    onClick={() => {
                      const url = `${window.location.origin}/invite/${createdInviteToken}`;
                      void navigator.clipboard.writeText(url).then(() => {
                        setCopyStatus("Copied!");
                        setTimeout(() => setCopyStatus(null), 2000);
                      });
                    }}
                  >
                    {copyStatus ?? "Copy Link"}
                  </button>
                </div>
              </div>
            )}
          </section>

          <section className="card guest-card">
            <h2>Guest by invite</h2>
            <div className="stack-form">
              <label htmlFor="invite-token">Invite link token</label>
              <input
                id="invite-token"
                value={inviteToken}
                onChange={(event) => setInviteToken(event.target.value)}
                placeholder="paste invite token"
              />

              <label htmlFor="guest-name">Your name</label>
              <input
                id="guest-name"
                value={guestIdentity}
                onChange={(event) => setGuestIdentity(event.target.value)}
                placeholder="guest-1"
              />

              <div className="button-row">
                <button type="button" onClick={() => void connectGuestToRoom()} disabled={guestState === "connecting"}>
                  {guestState === "connecting" ? "Connecting…" : "Join room"}
                </button>
                {guestRoom && (
                  <>
                    {isNoiseSuppressionSupported && (
                      <button
                        type="button"
                        className={isGuestNoiseSuppressionEnabled ? "" : "secondary"}
                        onClick={() => void toggleGuestNoiseSuppression()}
                        aria-pressed={isGuestNoiseSuppressionEnabled}
                      >
                        {isGuestNoiseSuppressionEnabled ? "Noise: On" : "Noise: Off"}
                      </button>
                    )}
                    <button type="button" className="secondary" onClick={disconnectGuestRoom}>
                      Leave
                    </button>
                  </>
                )}
              </div>
            </div>

            {guestError && <p className="helper-text error-text">{guestError}</p>}
            {guestRoom && (
              <div className="chat-panel">
                <div className="chat-header">
                  <strong>#{guestRoom}</strong>
                  <span>{guestState === "connected" ? "live" : guestState}</span>
                </div>

                <div className="messages" aria-live="polite">
                  {(guestMessages ?? []).length === 0 ? (
                    <p className="empty-state">No messages yet.</p>
                  ) : (
                    (guestMessages ?? []).map((message) => (
                      <div className="message" key={message.id ?? `${message.identity}-${message.createdAt}`}>
                        <div className="message-meta">
                          <span>{message.identity}</span>
                          <time>{formatMessageTime(message.createdAt)}</time>
                        </div>
                        <div>{message.body}</div>
                      </div>
                    ))
                  )}
                </div>

                <div className="composer">
                  <input
                    value={chatDraft}
                    onChange={(event) => setChatDraft(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter") {
                        event.preventDefault();
                        sendChatMessage();
                      }
                    }}
                    placeholder="Message the room"
                  />
                  <button type="button" onClick={sendChatMessage}>Send</button>
                </div>
              </div>
            )}
          </section>

          <section className="card voice-card">
            <h2>Voice room</h2>
            {canJoin ? (
              <form className="stack-form" onSubmit={handleJoinRoom}>
                <label htmlFor="room">Room</label>
                <input
                  id="room"
                  value={room}
                  onChange={(event) => setRoom(event.target.value)}
                  placeholder="team-standup"
                  required
                />

                <label htmlFor="identity">Your name</label>
                <input
                  id="identity"
                  value={identity}
                  onChange={(event) => setIdentity(event.target.value)}
                  placeholder="user-1"
                  required
                />

                <button type="submit">Join room</button>
              </form>
            ) : (
              <div className="call-panel">
                <p className="participants">
                  Participants: {(call.participants ?? []).length > 0 ? (call.participants ?? []).join(", ") : "just you"}
                </p>
                <div className="button-row">
                  <button type="button" onClick={() => void call.toggleMute()} aria-pressed={call.isMuted}>
                    {call.isMuted ? "Unmute" : "Mute"}
                  </button>
                  {call.isNoiseSuppressionSupported && (
                    <button
                      type="button"
                      className={call.isNoiseSuppressionEnabled ? "" : "secondary"}
                      onClick={() => void call.toggleNoiseSuppression()}
                      aria-pressed={call.isNoiseSuppressionEnabled}
                    >
                      {call.isNoiseSuppressionEnabled ? "Noise: On" : "Noise: Off"}
                    </button>
                  )}
                  <button type="button" className="secondary" onClick={() => void call.leave()}>
                    Leave
                  </button>
                </div>
              </div>
            )}

            <div className="connection-state" role="status">
              <span className={`status-dot status-dot--${call.state}`} aria-hidden="true" />
              {call.error ?? stateLabels[call.state]}
            </div>
          </section>
        </div>
      </section>
    </main>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </StrictMode>,
);
