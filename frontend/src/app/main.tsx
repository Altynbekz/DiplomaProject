import { StrictMode, createContext, useContext, useState } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import {
  QueryClient,
  QueryClientProvider,
  useQueryClient,
} from "@tanstack/react-query";
import { Targets, TargetPage, FindingPage } from "../features/evidence";
import "./style.css";
const Session = createContext("");
export function useToken() {
  return useContext(Session);
}
export function Workspace() {
  const [token, setToken] = useState("");
  const client = useQueryClient();
  return (
    <>
      <header>
        <Link to="/">D / Evidence workspace</Link>
        <span>Local lab · offline import</span>
        {token && (
          <button
            onClick={() => {
              setToken("");
              client.clear();
            }}
          >
            Lock workspace
          </button>
        )}
      </header>
      <main>
        <p className="eyebrow">DIPLOMA RESEARCH PROTOTYPE / M1</p>
        <h1>Scanner evidence, kept in context.</h1>
        <p>
          Review persisted observations from an authorized Juice Shop lab.
          Scanning is disabled. Verification and CVSS assessment come later.
        </p>
        {!token ? (
          <section className="panel">
            <h2>Unlock the local workspace</h2>
            <p>
              Your API token stays in memory and clears when you reload or lock
              this page.
            </p>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                const form = e.currentTarget;
                const value = new FormData(form).get("token");
                if (typeof value === "string") setToken(value.trim());
                form.reset();
              }}
            >
              <label>
                API token
                <input
                  name="token"
                  type="password"
                  required
                  autoComplete="off"
                  spellCheck={false}
                />
              </label>
              <button>Unlock</button>
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
                  <p>
                    Page not found. <Link to="/">Return to targets</Link>
                  </p>
                }
              />
            </Routes>
          </Session.Provider>
        )}
      </main>
      <footer>
        Evidence is a scanner assertion. Unreviewed does not mean confirmed.
        CVSS: Not assessed.
      </footer>
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
