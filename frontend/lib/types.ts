// Mirrors backend internal/model package

export type ProjectStatus = "active" | "archived";
export type ProjectRole = "owner" | "admin" | "member";
export type TaskStatus = "todo" | "in_progress" | "done";
export type TaskPriority = "low" | "medium" | "high";

export interface ApiResponse<T = null> {
  data: T;
  message: string;
  error: string | null;
}

// Auth
export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

// User
export interface User {
  id: string;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

// Project
export interface Project {
  id: string;
  name: string;
  description?: string;
  status: ProjectStatus;
  deadline?: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectMember {
  project_id: string;
  user_id: string;
  name: string;
  email: string;
  role: ProjectRole;
  joined_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  deadline?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  status?: ProjectStatus;
  deadline?: string;
}

export interface AddMemberRequest {
  user_id: string;
  role: ProjectRole;
}

export interface ProjectPage {
  items: Project[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface ProjectListParams {
  page?: number;
  limit?: number;
}

export interface BulkDeleteProjectsResponse {
  deleted_count: number;
}

// Task
export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  position: string;
  project_id: string;
  assignee_id?: string;
  created_by?: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

export interface TaskPage {
  items: Task[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface TaskFilter {
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string;
  sort_by?: string;
  sort_order?: "asc" | "desc";
  page?: number;
  limit?: number;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  priority: TaskPriority;
  position?: string;
  assignee_id?: string | null;
  due_date?: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string | null;
  due_date?: string;
}

export interface AssignTaskRequest {
  assignee_id: string | null;
}

export interface MoveTaskRequest {
  status: TaskStatus;
  position: string;
}

export interface TaskActivityLog {
  id: string;
  task_id: string;
  changed_by?: string;
  from_status: TaskStatus;
  to_status: TaskStatus;
  created_at: string;
}

export interface BoardView {
  todo: Task[];
  in_progress: Task[];
  done: Task[];
}

// Dashboard
export interface ProjectTaskCounts {
  project_id: string;
  project_name: string;
  todo: number;
  in_progress: number;
  done: number;
  total: number;
}

export interface UpcomingTask {
  id: string;
  title: string;
  status: TaskStatus;
  priority: TaskPriority;
  due_date: string;
  project_id: string;
  project_name: string;
}
