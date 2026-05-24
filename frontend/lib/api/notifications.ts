import { apiRequest } from "../api";
import type { NotificationPage } from "../types";

export const notificationsApi = {
  list: (limit = 50): Promise<NotificationPage> =>
    apiRequest(`/notifications?limit=${limit}`),

  markRead: (id: string): Promise<null> =>
    apiRequest(`/notifications/${id}/read`, { method: "POST" }),

  markAllRead: (): Promise<null> =>
    apiRequest("/notifications/read-all", { method: "POST" }),
};
