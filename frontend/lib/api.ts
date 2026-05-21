import { config } from "./config";
import { tokenStorage } from "./token";
import type { ApiResponse, TokenPair } from "./types";

const BASE_URL = config.api.baseUrl;

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function tryRefresh(): Promise<string | null> {
  const refreshToken = tokenStorage.getRefresh();
  if (!refreshToken) return null;

  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!res.ok) {
    tokenStorage.clear();
    return null;
  }

  const body: ApiResponse<TokenPair> = await res.json();
  tokenStorage.set(body.data.access_token, body.data.refresh_token);
  return body.data.access_token;
}

function buildHeaders(token: string | null, extra?: HeadersInit): Headers {
  const headers = new Headers({ "Content-Type": "application/json", ...extra });
  if (token) headers.set("Authorization", `Bearer ${token}`);
  return headers;
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  let token = tokenStorage.getAccess();
  let res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: buildHeaders(token, options.headers),
  });

  if (res.status === 401) {
    token = await tryRefresh();
    if (!token) {
      if (typeof window !== "undefined") window.location.href = "/login";
      throw new ApiError(401, "session expired");
    }
    res = await fetch(`${BASE_URL}${path}`, {
      ...options,
      headers: buildHeaders(token, options.headers),
    });
  }

  const body: ApiResponse<T> = await res.json();

  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? body.message ?? "request failed");
  }

  return body.data;
}
