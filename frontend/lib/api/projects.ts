import { apiRequest } from "../api";
import type {
  Project,
  ProjectMember,
  ProjectPage,
  ProjectListParams,
  CreateProjectRequest,
  UpdateProjectRequest,
  AddMemberRequest,
  BulkDeleteProjectsResponse,
} from "../types";

export const projectsApi = {
  list: ({ page = 1, limit = 10 }: ProjectListParams = {}): Promise<ProjectPage> => {
    const qs = new URLSearchParams({ page: String(page), limit: String(limit) });
    return apiRequest(`/projects?${qs.toString()}`);
  },

  getById: (id: string): Promise<Project> =>
    apiRequest(`/projects/${id}`),

  create: (data: CreateProjectRequest): Promise<Project> =>
    apiRequest("/projects", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: string, data: UpdateProjectRequest): Promise<Project> =>
    apiRequest(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: string): Promise<null> =>
    apiRequest(`/projects/${id}`, { method: "DELETE" }),

  bulkDelete: (ids: string[]): Promise<BulkDeleteProjectsResponse> =>
    apiRequest(`/projects/bulk-delete`, {
      method: "POST",
      body: JSON.stringify({ ids }),
    }),

  getMembers: (id: string): Promise<ProjectMember[]> =>
    apiRequest(`/projects/${id}/members`),

  addMember: (id: string, data: AddMemberRequest): Promise<ProjectMember> =>
    apiRequest(`/projects/${id}/members`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  removeMember: (id: string, userID: string): Promise<null> =>
    apiRequest(`/projects/${id}/members/${userID}`, { method: "DELETE" }),
};
