import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import { toast } from "sonner";
import { config } from "./config";
import { tokenStorage } from "./token";
import type { ApiResponse, TokenPair } from "./types";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function extractMessage(err: AxiosError<ApiResponse>): string {
  return (
    err.response?.data?.error ??
    err.response?.data?.message ??
    err.message ??
    "request failed"
  );
}

// Unauthenticated client for login, register — toasts errors centrally
export const publicClient: AxiosInstance = axios.create({
  baseURL: config.api.baseUrl,
  headers: { "Content-Type": "application/json" },
});

publicClient.interceptors.response.use(
  (res) => res,
  (err: AxiosError<ApiResponse>) => {
    const message = extractMessage(err);
    toast.error(message);
    return Promise.reject(new ApiError(err.response?.status ?? 0, message));
  }
);

// Authenticated client — attaches bearer token, retries once after token refresh
export const apiClient: AxiosInstance = axios.create({
  baseURL: config.api.baseUrl,
  headers: { "Content-Type": "application/json" },
});

apiClient.interceptors.request.use((req) => {
  const token = tokenStorage.getAccess();
  if (token) req.headers.Authorization = `Bearer ${token}`;
  return req;
});

let isRefreshing = false;
let refreshQueue: Array<(token: string | null) => void> = [];

function processQueue(token: string | null) {
  refreshQueue.forEach((cb) => cb(token));
  refreshQueue = [];
}

apiClient.interceptors.response.use(
  (res) => res,
  async (err: AxiosError<ApiResponse>) => {
    const original = err.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (err.response?.status === 401 && !original?._retry) {
      if (isRefreshing) {
        return new Promise<unknown>((resolve, reject) => {
          refreshQueue.push((token) => {
            if (!token) return reject(err);
            original.headers.Authorization = `Bearer ${token}`;
            original._retry = true;
            resolve(apiClient(original));
          });
        });
      }

      original._retry = true;
      isRefreshing = true;

      const refreshToken = tokenStorage.getRefresh();
      if (!refreshToken) {
        isRefreshing = false;
        processQueue(null);
        tokenStorage.clear();
        if (typeof window !== "undefined") window.location.href = "/login";
        return Promise.reject(new ApiError(401, "session expired"));
      }

      try {
        // Use raw axios to avoid publicClient's error interceptor toasting a silent refresh failure
        const res = await axios.post<ApiResponse<TokenPair>>(
          `${config.api.baseUrl}/auth/refresh`,
          { refresh_token: refreshToken },
          { headers: { "Content-Type": "application/json" } }
        );
        const tokens = res.data.data;
        tokenStorage.set(tokens.access_token, tokens.refresh_token);
        processQueue(tokens.access_token);
        isRefreshing = false;
        original.headers.Authorization = `Bearer ${tokens.access_token}`;
        return apiClient(original);
      } catch {
        isRefreshing = false;
        processQueue(null);
        tokenStorage.clear();
        if (typeof window !== "undefined") window.location.href = "/login";
        return Promise.reject(new ApiError(401, "session expired"));
      }
    }

    const message = extractMessage(err);
    toast.error(message);
    return Promise.reject(new ApiError(err.response?.status ?? 0, message));
  }
);

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const method = options.method ?? "GET";
  const data = options.body ? JSON.parse(options.body as string) : undefined;
  const res = await apiClient.request<ApiResponse<T>>({ url: path, method, data });
  return res.data.data;
}
