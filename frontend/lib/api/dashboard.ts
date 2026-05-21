import { apiRequest } from "../api";
import type { ProjectTaskCounts, UpcomingTask } from "../types";

export const dashboardApi = {
  projectTaskCounts: (): Promise<ProjectTaskCounts[]> =>
    apiRequest("/dashboard/project-task-counts"),

  upcomingTasks: (): Promise<UpcomingTask[]> =>
    apiRequest("/dashboard/upcoming-tasks"),
};
