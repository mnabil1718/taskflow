import {
    useMutation,
    useQuery,
    useQueryClient,
    useInfiniteQuery,
    keepPreviousData,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { tasksApi } from "@/lib/api/tasks";
import { compareLexorank } from "@/lib/lexorank";
import type {
    BoardView,
    CreateTaskRequest,
    Task,
    TaskFilter,
    TaskStatus,
    UpdateTaskRequest,
} from "@/lib/types";

const ACTIVITY_PAGE_SIZE = 10;

// Prefix shared by all global task list cache entries — used to invalidate
// the /tasks page without knowing the exact filter params.
const GLOBAL_TASKS_KEY = ["tasks", "all"] as const;

export const taskKeys = {
    all: (projectId: string) => ["projects", projectId, "tasks"] as const,
    list: (projectId: string, filter: TaskFilter) =>
        ["projects", projectId, "tasks", "list", filter] as const,
    allTasks: (filter: TaskFilter) => [...GLOBAL_TASKS_KEY, filter] as const,
    board: (projectId: string) => ["projects", projectId, "board"] as const,
};

export function useBoard(projectId: string) {
    return useQuery({
        queryKey: taskKeys.board(projectId),
        queryFn: () => tasksApi.board(projectId),
        staleTime: 30 * 1000,
        enabled: !!projectId,
    });
}

// applyMoveToBoardView produces the next BoardView snapshot for a task
// move: remove from its current bucket, update status+position, insert
// into the target bucket in lexorank order. Exported as a pure function
// so it can be reused for optimistic updates without the hook closure.
function applyMoveToBoardView(
    board: BoardView,
    move: { id: string; status: TaskStatus; position: string }
): BoardView {
    let moved: Task | null = null;
    const without = (list: Task[]): Task[] => {
        const idx = list.findIndex((t) => t.id === move.id);
        if (idx === -1) return list;
        moved = list[idx];
        return [...list.slice(0, idx), ...list.slice(idx + 1)];
    };
    const next: BoardView = {
        todo: without(board.todo),
        in_progress: without(board.in_progress),
        done: without(board.done),
    };
    if (!moved) return board;
    const updated: Task = { ...(moved as Task), status: move.status, position: move.position };
    const targetList = [...next[move.status], updated].sort(
        (a, b) => compareLexorank(a.position, b.position)
    );
    return { ...next, [move.status]: targetList };
}

// useMoveTask persists a drag-and-drop. The optimistic onMutate writes
// the new (status, position) into the board cache immediately so the
// real card is already at its destination by the time the DragOverlay
// disappears — no "snap back to origin, then jump to new spot" between
// the drop and the refetch. If the server rejects (e.g. permission or
// race), the previous snapshot is restored and the next refetch yields
// the truth either way.
export function useMoveTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, status, position }: { id: string; status: TaskStatus; position: string }) =>
            tasksApi.move(id, { status, position }),
        onMutate: async (vars) => {
            const boardKey = taskKeys.board(projectId);
            await qc.cancelQueries({ queryKey: boardKey });
            const previous = qc.getQueryData<BoardView>(boardKey);
            if (previous) {
                qc.setQueryData(boardKey, applyMoveToBoardView(previous, vars));
            }
            return { previous };
        },
        onError: (_err, _vars, ctx) => {
            if (ctx?.previous) {
                qc.setQueryData(taskKeys.board(projectId), ctx.previous);
            }
        },
        onSettled: (task) => {
            qc.invalidateQueries({ queryKey: taskKeys.board(projectId) });
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            if (task) {
                qc.invalidateQueries({ queryKey: ["tasks", "detail", task.id] });
                qc.invalidateQueries({ queryKey: ["tasks", task.id, "activity"] });
            }
        },
    });
}

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

export function useTask(taskId: string) {
    return useQuery({
        queryKey: ["tasks", "detail", taskId] as const,
        queryFn: () => tasksApi.getById(taskId),
        staleTime: 30 * 1000,
        enabled: !!taskId,
    });
}

// useTaskActivityLogs paginates via useInfiniteQuery. Each fetched page
// has its oldest entry's created_at used as the cursor for the next
// page; the backend's has_more flag tells us when to stop. The page
// array is exposed as a flat list so the detail page can render it with
// a simple .map() while still calling fetchNextPage from a "Load more"
// button.
export function useTaskActivityLogs(taskId: string) {
    return useInfiniteQuery({
        queryKey: ["tasks", taskId, "activity"] as const,
        queryFn: ({ pageParam }) =>
            tasksApi.getActivityLogs(taskId, {
                limit: ACTIVITY_PAGE_SIZE,
                before: pageParam ?? undefined,
            }),
        initialPageParam: undefined as string | undefined,
        getNextPageParam: (lastPage) => {
            if (!lastPage.has_more || lastPage.items.length === 0) return undefined;
            return lastPage.items[lastPage.items.length - 1].created_at;
        },
        enabled: !!taskId,
        staleTime: 30 * 1000,
    });
}

export function useCreateTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateTaskRequest) => tasksApi.create(projectId, data),
        onSuccess: (task) => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: taskKeys.board(projectId) });
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
            qc.invalidateQueries({ queryKey: taskKeys.board(projectId) });
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            // The detail page reads /tasks/:id and /tasks/:id/activity
            // through its own cache keys — invalidate both so a status
            // change made from the edit dialog renders without a refresh.
            qc.invalidateQueries({ queryKey: ["tasks", "detail", task.id] });
            qc.invalidateQueries({ queryKey: ["tasks", task.id, "activity"] });
            toast.success(`Task "${task.title}" updated`);
        },
    });
}

// useUpdateTaskStatus calls the member-allowed PATCH /tasks/:id/status
// endpoint — both project owners and regular members can use this, and the
// backend rejects any non-status field by virtue of the endpoint's request
// shape (it only accepts { status }). Use this whenever you want to move
// a task between columns without opening the full edit dialog.
export function useUpdateTaskStatus(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, status }: { id: string; status: TaskStatus }) =>
            tasksApi.updateStatus(id, { status }),
        onSuccess: (task) => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: taskKeys.board(projectId) });
            qc.invalidateQueries({ queryKey: GLOBAL_TASKS_KEY });
            qc.invalidateQueries({ queryKey: ["dashboard"] });
            qc.invalidateQueries({ queryKey: ["tasks", "detail", task.id] });
            qc.invalidateQueries({ queryKey: ["tasks", task.id, "activity"] });
            toast.success(`Status changed to ${task.status.replace("_", " ")}`);
        },
    });
}

export function useDeleteTask(projectId: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (taskId: string) => tasksApi.delete(taskId),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: taskKeys.all(projectId) });
            qc.invalidateQueries({ queryKey: taskKeys.board(projectId) });
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
