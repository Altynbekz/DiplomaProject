import {
  StrictMode,
  createContext,
  useContext,
  useState,
  type FormEvent,
} from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import {
  QueryClient,
  QueryClientProvider,
  useQueryClient,
} from "@tanstack/react-query";
import { Targets, TargetPage, FindingPage } from "../features/evidence";
import { APIError, request } from "../api/client";
import type { components } from "../api/schema";
import "./style.css";

const Session = createContext("");
export function useToken() {
  return useContext(Session);
}

export function Workspace() {
  const [token, setToken] = useState("");
  const [unlocking, setUnlocking] = useState(false);
  const [unlockError, setUnlockError] = useState("");
  const client = useQueryClient();

  async function unlock(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (unlocking) return;
    const form = event.currentTarget;
    const candidate = String(new FormData(form).get("token") ?? "").trim();
    if (!candidate) {
      setUnlockError("Enter the API token.");
      return;
    }
    setUnlocking(true);
    setUnlockError("");
    try {
      const targets = await request<components["schemas"]["TargetPage"]>(
        candidate,
        "/targets?limit=50&cursor=",
      );
      client.setQueryData(["targets", ""], targets);
      form.reset();
      setToken(candidate);
    } catch (error) {
      setUnlockError(
        error instanceof APIError && error.problem.status === 401
          ? "Token not accepted. Check it and try again."
          : "Could not reach the workspace. Check that the API is running and try again.",
      );
    } finally {
      setUnlocking(false);
    }
  }

  return (
    <>
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>
      <header className="site-header">
        <div className="header-inner">
          <Link className="brand" to="/">
            Vulnerability assessment
          </Link>
          {token && (
            <button
              type="button"
              onClick={() => {
                setToken("");
                setUnlockError("");
                client.clear();
              }}
            >
              Lock workspace
            </button>
          )}
        </div>
      </header>
      <main id="main-content" tabIndex={-1}>
        {!token ? (
          <section className="panel unlock-panel">
            <h1>Unlock workspace</h1>
            <p className="muted">The API token clears on reload or lock.</p>
            <form className="stack" onSubmit={unlock}>
              <label>
                API token
                <input
                  name="token"
                  type="password"
                  required
                  autoComplete="off"
                  spellCheck={false}
                  aria-describedby={unlockError ? "unlock-error" : undefined}
                />
              </label>
              {unlockError && (
                <p id="unlock-error" className="error" role="alert">
                  {unlockError}
                </p>
              )}
              <button className="primary" disabled={unlocking}>
                {unlocking ? "Checking token…" : "Unlock"}
              </button>
            </form>
          </section>
        ) : (
          <Session.Provider value={token}>
            <Routes>
              <Route path="/" element={<Targets />} />
              <Route path="/targets/:id" element={<TargetPage />} />
              <Route path="/findings/:id" element={<FindingPage />} />
              <Route
                path="*"
                element={
                  <>
                    <h1>Page not found</h1>
                    <Link to="/">Return to targets</Link>
                  </>
                }
              />
            </Routes>
          </Session.Provider>
        )}
      </main>
    </>
  );
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
    mutations: { retry: false },
  },
});
const mount = document.getElementById("root");
if (mount)
  createRoot(mount).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Workspace />
        </BrowserRouter>
      </QueryClientProvider>
    </StrictMode>,
  );
