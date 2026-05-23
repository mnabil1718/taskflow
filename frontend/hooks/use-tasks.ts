import {
    useMutation,
    useQuery,
    useQueryClient,
    keepPreviousData,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { tasksApi } from "@/lib/api/tasks";
import type {
    CreateTaskRequest,
    TaskFilter,
    UpdateTaskRequest,
} from "@/lib/types";

export const taskKeys = {
    all: (projectId: string) => ["projects", projectId, "tasks"] as const,
    list: (projectId: string, filter: TaskFilter) =>
        ["projects", projectId, "tasks", "list", filter] as const,
};

export function useTasks(projectId: string, filter: TaskFilter = {}) {
    return useQuery({
        queryKey: taskKeys.list(projectId, filter),
        queryFn: () => tasksApi.list(projectId, filter),
        staleTime: 30 * 1000,
        placeholderData: keepPreviousData,
        enabled: !!projectId,
    });
}

export function useCreateTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateTaskRequest) => tasksApi.create(projectId, data),
        onSuccess: (task) => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            toast.success(`Task "${task.title}" created`);
        },
    });
}

export function useUpdateTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: UpdateTaskRequest }) =>
            tasksApi.update(id, data),
        onSuccess: (task) => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            toast.success(`Task "${task.title}" updated`);
        },
    });
}

export function useDeleteTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (taskId: string) => tasksApi.delete(taskId),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            toast.success("Task deleted");
        },
    });
}
