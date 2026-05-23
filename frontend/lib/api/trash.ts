import { apiRequest } from "../api";
import type { BulkTrashRequest, BulkTrashResponse, TrashItem } from "../types";

export const trashApi = {
  list: (): Promise<TrashItem[]> => apiRequest("/trash"),

  restore: (data: BulkTrashRequest): Promise<BulkTrashResponse> =>
    apiRequest("/trash/restore", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  purge: (data: BulkTrashRequest): Promise<BulkTrashResponse> =>
    apiRequest("/trash/purge", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  emptyAll: (): Promise<BulkTrashResponse> =>
    apiRequest("/trash", { method: "DELETE" }),
};
