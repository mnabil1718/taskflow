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

// Prefix shared by all global task list cache entries — used to invalidate
// the /tasks page without knowing the exact filter params.
const GLOBAL_TASKS_KEY = ["tasks", "all"] as const;

export const taskKeys = {
    all: (projectId: string) => ["projects", projectId, "tasks"] as const,
    list: (projectId: string, filter: TaskFilter) =>
        ["projects", projectId, "tasks", "list", filter] as const,
    allTasks: (filter: TaskFilter) => [...GLOBAL_TASKS_KEY, filter] as const,
};

export function useAllTasks(filter: TaskFilter = {}) {
    return useQuery({
        queryKey: taskKeys.allTasks(filter),
        queryFn: () => tasksApi.listAll(filter),
        staleTime: 30 * 1000,
        placeholderData: keepPreviousData,
    });
}

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
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
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
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
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
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            qc.invalidateQueries({ queryKey: ["trash"] });
            toast.success("Task deleted");
        },
    });
}

export function useBulkDeleteTasks() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (ids: string[]) => tasksApi.bulkDelete(ids),
        onSuccess: ({ deleted_count }) => {
            qc.invalidateQueries({ queryKey: ["projects"] });
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            qc.invalidateQueries({ queryKey: ["trash"] });
            toast.success(`Deleted ${deleted_count} task${deleted_count === 1 ? "" : "s"}`);
        },
    });
}
