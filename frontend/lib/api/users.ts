import { apiRequest } from "../api";
import type { User } from "../types";

export const usersApi = {
  search: (q: string): Promise<User[]> => {
    const qs = new URLSearchParams({ q });
    return apiRequest(`/users/search?${qs.toString()}`);
  },
};
