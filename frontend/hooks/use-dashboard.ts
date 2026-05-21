import { useQuery } from "@tanstack/react-query";
import { dashboardApi } from "@/lib/api/dashboard";

export const dashboardKeys = {
    projectTaskCounts: ["dashboard", "project-task-counts"] as const,
    upcomingTasks: ["dashboard", "upcoming-tasks"] as const,
};

export function useProjectTaskCounts() {
    return useQuery({
        queryKey: dashboardKeys.projectTaskCounts,
        queryFn: dashboardApi.projectTaskCounts,
        staleTime: 60 * 1000,
    });
}

export function useUpcomingTasks() {
    return useQuery({
        queryKey: dashboardKeys.upcomingTasks,
        queryFn: dashboardApi.upcomingTasks,
        staleTime: 60 * 1000,
    });
}
