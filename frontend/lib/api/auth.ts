import { config } from "../config";
import { apiRequest, ApiError } from "../api";
import { tokenStorage } from "../token";
import type { ApiResponse, TokenPair, LoginRequest, RegisterRequest } from "../types";

const BASE_URL = config.api.baseUrl;

async function postPublic<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const json: ApiResponse<T> = await res.json();
  if (!res.ok) throw new ApiError(res.status, json.error ?? json.message ?? "request failed");
  return json.data;
}

export const authApi = {
  register: async (data: RegisterRequest): Promise<TokenPair> => {
    const tokens = await postPublic<TokenPair>("/auth/register", data);
    tokenStorage.set(tokens.access_token, tokens.refresh_token);
    return tokens;
  },

  login: async (data: LoginRequest): Promise<TokenPair> => {
    const tokens = await postPublic<TokenPair>("/auth/login", data);
    tokenStorage.set(tokens.access_token, tokens.refresh_token);
    return tokens;
  },

  logout: async (): Promise<void> => {
    const refreshToken = tokenStorage.getRefresh();
    if (refreshToken) {
      await apiRequest("/auth/logout", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      }).catch(() => {});
    }
    tokenStorage.clear();
  },
};
