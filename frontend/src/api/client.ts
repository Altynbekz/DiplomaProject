import type { components } from "./schema";
export type Target = components["schemas"]["Target"];
export type Finding = components["schemas"]["Finding"];
export type Observation = components["schemas"]["Observation"];
export type ImportResult = components["schemas"]["Import"];
export type Problem = components["schemas"]["Problem"];
export class APIError extends Error {
  constructor(public problem: Problem) {
    super(problem.detail);
  }
}
export async function request<T>(
  token: string,
  path: string,
  init?: RequestInit,
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers: { Authorization: `Bearer ${token}`, ...init?.headers },
    cache: "no-store",
  });
  if (!response.ok) {
    const body: unknown = await response.json().catch(() => null);
    if (body && typeof body === "object" && "detail" in body && "code" in body)
      throw new APIError(body as Problem);
    throw new Error(`Request failed (${response.status})`);
  }
  return response.json() as Promise<T>;
}
