import { apiRequest } from "../api";
import type {
  Project,
  ProjectMember,
  CreateProjectRequest,
  UpdateProjectRequest,
  AddMemberRequest,
} from "../types";

export const projectsApi = {
  list: (): Promise<Project[]> =>
    apiRequest("/projects"),

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
