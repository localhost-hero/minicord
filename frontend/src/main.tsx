import { StrictMode, useState, type FormEvent } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";
import { useCallConnection } from "./hooks/useCallConnection";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

const stateLabels: Record<string, string> = {
  idle: "Enter a room to join",
  connecting: "Connecting to the room",
  connected: "Connected",
  reconnecting: "Reconnecting",
  failed: "Connection failed",
  ended: "Call ended",
};

function App() {
  const call = useCallConnection(API_BASE_URL);
  const [room, setRoom] = useState("");
  const [identity, setIdentity] = useState("");

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    void call.join(room, identity);
  };

  const canJoin = call.state === "idle" || call.state === "failed" || call.state === "ended";

  return (
    <main className="app-shell">
      <section className="welcome-panel" aria-labelledby="app-title">
        <p className="eyebrow">Self-hosted voice</p>
        <h1 id="app-title">Minicord</h1>

        {canJoin ? (
          <form className="join-form" onSubmit={handleSubmit}>
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
              Participants: {call.participants.length > 0 ? call.participants.join(", ") : "just you"}
            </p>
            <div className="call-controls">
              <button
                type="button"
                onClick={() => void call.toggleMute()}
                disabled={call.state !== "connected"}
                aria-pressed={call.isMuted}
              >
                {call.isMuted ? "Unmute" : "Mute"}
              </button>
              <button type="button" onClick={() => void call.leave()}>
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
    </main>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
