import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

function App() {
  return (
    <main className="app-shell">
      <section className="welcome-panel" aria-labelledby="app-title">
        <p className="eyebrow">Self-hosted voice</p>
        <h1 id="app-title">Minicord</h1>
        <p className="status">Development environment is ready for the call experience.</p>
        <div className="connection-state" role="status">
          <span className="status-dot" aria-hidden="true" />
          Waiting for the backend and LiveKit connection
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
