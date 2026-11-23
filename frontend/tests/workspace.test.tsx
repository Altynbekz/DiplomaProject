import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { Workspace } from "../src/app/main";
import { AssessmentValue, ErrorState } from "../src/features/evidence";
import { APIError } from "../src/api/client";
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
test("token stays in memory; empty workspace and lock state are explicit", async () => {
  const fetchMock = vi.fn().mockImplementation(
    async () =>
      new Response(JSON.stringify({ items: [], next_cursor: null }), {
        headers: { "Content-Type": "application/json" },
      }),
  );
  vi.stubGlobal("fetch", fetchMock);
  render(
    <QueryClientProvider
      client={
        new QueryClient({ defaultOptions: { queries: { retry: false } } })
      }
    >
      <MemoryRouter>
        <Workspace />
      </MemoryRouter>
    </QueryClientProvider>,
  );
  const input = screen.getByLabelText("API token");
  fireEvent.change(input, { target: { value: "synthetic-ui-token" } });
  fireEvent.click(screen.getByRole("button", { name: "Unlock" }));
  expect(await screen.findByText(/No reports imported/)).toBeInTheDocument();
  expect(localStorage.length).toBe(0);
  expect(sessionStorage.length).toBe(0);
  expect(fetchMock.mock.calls[0][1].headers.Authorization).toBe(
    "Bearer synthetic-ui-token",
  );
  fireEvent.click(screen.getByRole("button", { name: "Lock workspace" }));
  expect(screen.getByLabelText("API token")).toHaveValue("");
});
test("problem detail and source pointer render as escaped text", () => {
  const message = "<script>window.pwned=1</script>";
  render(
    <ErrorState
      error={
        new APIError({
          type: "urn:diploma:problem:validation_failed",
          title: "Invalid",
          status: 422,
          code: "validation_failed",
          detail: message,
          request_id: "synthetic-request",
          errors: [
            {
              pointer: "/site/0/alerts/0",
              code: "validation_failed",
              message: "Invalid location",
            },
          ],
        })
      }
    />,
  );
  expect(screen.getByRole("alert")).toHaveTextContent(message);
  expect(document.querySelector("script")).toBeNull();
  expect(screen.getByRole("alert")).toHaveTextContent("/site/0/alerts/0");
});

test("assessment display uses the API score and vector, including zero", () => {
  const vector =
    "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:N/VI:N/VA:N/SC:N/SI:N/SA:N";
  const { container, rerender } = render(<AssessmentValue assessment={null} />);
  expect(container).toHaveTextContent("Not assessed");
  const assessment = {
    id: "synthetic-ui-assessment",
    kind: "B" as const,
    status: "complete" as const,
    score: 0,
    vector,
    stale: false,
    created_at: "2026-01-01T00:00:00Z",
  };
  rerender(<AssessmentValue assessment={assessment} />);
  expect(container).toHaveTextContent(`0 · ${vector}`);
  rerender(
    <AssessmentValue
      assessment={{
        ...assessment,
        status: "incomplete",
        score: null,
        vector: null,
      }}
    />,
  );
  expect(container).toHaveTextContent("Incomplete");
});
