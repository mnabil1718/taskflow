import { apiRequest } from "../api";
import type {
  Task,
  TaskPage,
  TaskFilter,
  BoardView,
  TaskActivityLog,
  BulkDeleteTasksResponse,
  CreateTaskRequest,
  UpdateTaskRequest,
  AssignTaskRequest,
  MoveTaskRequest,
} from "../types";

function toQueryString(filter: TaskFilter): string {
  const params = new URLSearchParams();
  if (filter.status) params.set("status", filter.status);
  if (filter.priority) params.set("priority", filter.priority);
  if (filter.assignee_id) params.set("assignee_id", filter.assignee_id);
  if (filter.search) params.set("q", filter.search);
  if (filter.sort_by) params.set("sort_by", filter.sort_by);
  if (filter.sort_order) params.set("sort_order", filter.sort_order);
  if (filter.page) params.set("page", String(filter.page));
  if (filter.limit) params.set("limit", String(filter.limit));
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export const tasksApi = {
  listAll: (filter: TaskFilter = {}): Promise<TaskPage> =>
    apiRequest(`/tasks${toQueryString(filter)}`),

  list: (projectId: string, filter: TaskFilter = {}): Promise<TaskPage> =>
    apiRequest(`/projects/${projectId}/tasks${toQueryString(filter)}`),

  board: (projectId: string): Promise<BoardView> =>
    apiRequest(`/projects/${projectId}/board`),

  create: (projectId: string, data: CreateTaskRequest): Promise<Task> =>
    apiRequest(`/projects/${projectId}/tasks`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getById: (taskId: string): Promise<Task> =>
    apiRequest(`/tasks/${taskId}`),

  update: (taskId: string, data: UpdateTaskRequest): Promise<Task> =>
    apiRequest(`/tasks/${taskId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (taskId: string): Promise<null> =>
    apiRequest(`/tasks/${taskId}`, { method: "DELETE" }),

  assign: (taskId: string, data: AssignTaskRequest): Promise<Task> =>
    apiRequest(`/tasks/${taskId}/assign`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  move: (taskId: string, data: MoveTaskRequest): Promise<Task> =>
    apiRequest(`/tasks/${taskId}/move`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  getActivityLogs: (taskId: string): Promise<TaskActivityLog[]> =>
    apiRequest(`/tasks/${taskId}/activity`),

  bulkDelete: (ids: string[]): Promise<BulkDeleteTasksResponse> =>
    apiRequest(`/tasks/bulk-delete`, {
      method: "POST",
      body: JSON.stringify({ ids }),
    }),
};
