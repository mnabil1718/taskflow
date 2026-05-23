import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { trashApi } from "@/lib/api/trash";
import type { BulkTrashRequest } from "@/lib/types";

export const trashKeys = {
    all: ["trash"] as const,
};

export function useTrash() {
    return useQuery({
        queryKey: trashKeys.all,
        queryFn: () => trashApi.list(),
        staleTime: 30 * 1000,
    });
}

// All trash mutations have the same fan-out: refresh the trash list itself,
// the project/task lists (a restore makes items reappear there), and the
// dashboard counts.
function invalidateAfterTrashMutation(qc: ReturnType<typeof useQueryClient>) {
    qc.invalidateQueries({ queryKey: trashKeys.all });
    qc.invalidateQueries({ queryKey: ["projects"] });
    qc.invalidateQueries({ queryKey: ["tasks", "all"] });
    qc.invalidateQueries({ queryKey: ["dashboard"] });
}

export function useRestoreTrash() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: BulkTrashRequest) => trashApi.restore(data),
        onSuccess: (resp) => {
            invalidateAfterTrashMutation(qc);
            const total = (resp.restored_projects ?? 0) + (resp.restored_tasks ?? 0);
            toast.success(`Restored ${total} item${total === 1 ? "" : "s"}`);
        },
    });
}

export function usePurgeTrash() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: BulkTrashRequest) => trashApi.purge(data),
        onSuccess: (resp) => {
            invalidateAfterTrashMutation(qc);
            const total = (resp.purged_projects ?? 0) + (resp.purged_tasks ?? 0);
            toast.success(`Deleted ${total} item${total === 1 ? "" : "s"} permanently`);
        },
    });
}

export function useEmptyTrash() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: () => trashApi.emptyAll(),
        onSuccess: (resp) => {
            invalidateAfterTrashMutation(qc);
            const total = (resp.purged_projects ?? 0) + (resp.purged_tasks ?? 0);
            toast.success(
                total > 0
                    ? `Trash emptied — ${total} item${total === 1 ? "" : "s"} deleted`
                    : "Trash is empty"
            );
        },
    });
}
