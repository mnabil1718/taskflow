import { apiRequest, publicClient } from "../api";
import { tokenStorage } from "../token";
import type { ApiResponse, TokenPair, LoginRequest, RegisterRequest } from "../types";

export const authApi = {
  register: async (data: RegisterRequest): Promise<TokenPair> => {
    const res = await publicClient.post<ApiResponse<TokenPair>>("/auth/register", data);
    const tokens = res.data.data;
    tokenStorage.set(tokens.access_token, tokens.refresh_token);
    return tokens;
  },

  login: async (data: LoginRequest): Promise<TokenPair> => {
    const res = await publicClient.post<ApiResponse<TokenPair>>("/auth/login", data);
    const tokens = res.data.data;
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
