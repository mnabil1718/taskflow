"use client";

import { useMemo, useState } from "react";
import {
    useReactTable,
    getCoreRowModel,
    createColumnHelper,
    type RowSelectionState,
} from "@tanstack/react-table";
import { CheckSquare, FolderKanban, RotateCcw, Trash2, TrashIcon } from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { DataTable } from "@/components/data-table";
import {
    useEmptyTrash,
    usePurgeTrash,
    useRestoreTrash,
    useTrash,
} from "@/hooks/use-trash";
import { formatDistanceToNow } from "@/lib/date-utils";
import type { BulkTrashRequest, TrashItem, TrashKind } from "@/lib/types";

// Splits a flat selection set into the project_ids / task_ids payload shape
// the backend expects. The trash row id is "<kind>:<uuid>" so we can keep
// react-table's selection state as a single flat map.
function partition(
    items: TrashItem[],
    selectedKeys: string[]
): BulkTrashRequest {
    const ids = new Set(selectedKeys);
    const projectIds: string[] = [];
    const taskIds: string[] = [];
    for (const it of items) {
        if (!ids.has(rowKey(it))) continue;
        if (it.kind === "project") projectIds.push(it.id);
        else taskIds.push(it.id);
    }
    return { project_ids: projectIds, task_ids: taskIds };
}

function rowKey(it: TrashItem): string {
    return `${it.kind}:${it.id}`;
}

const kindBadge: Record<TrashKind, { label: string; icon: React.ElementType }> = {
    project: { label: "Project", icon: FolderKanban },
    task: { label: "Task", icon: CheckSquare },
};

function buildColumns(
    onRestore: (item: TrashItem) => void,
    onPurge: (item: TrashItem) => void,
    isMutating: boolean
) {
    const columnHelper = createColumnHelper<TrashItem>();

    return [
        columnHelper.display({
            id: "select",
            meta: { headerClassName: "w-[40px]", cellClassName: "w-[40px]" },
            header: ({ table }) => (
                <Checkbox
                    checked={table.getIsAllPageRowsSelected()}
                    indeterminate={table.getIsSomePageRowsSelected()}
                    onCheckedChange={(v) => table.toggleAllPageRowsSelected(!!v)}
                    aria-label="Select all"
                />
            ),
            cell: ({ row }) => (
                <Checkbox
                    checked={row.getIsSelected()}
                    onCheckedChange={(v) => row.toggleSelected(!!v)}
                    aria-label={`Select ${row.original.title}`}
                />
            ),
        }),
        columnHelper.accessor("kind", {
            header: "Type",
            cell: (info) => {
                const { label, icon: Icon } = kindBadge[info.getValue()];
                return (
                    <Badge variant="outline" className="gap-1.5 whitespace-nowrap">
                        <Icon className="size-3" />
                        {label}
                    </Badge>
                );
            },
        }),
        columnHelper.accessor("title", {
            header: "Name",
            meta: { cellClassName: "max-w-md" },
            cell: (info) => {
                const item = info.row.original;
                return (
                    <div className="min-w-0">
                        <p className="text-sm font-medium truncate">{item.title}</p>
                        {item.kind === "task" && item.project_name && (
                            <p className="text-xs text-muted-foreground truncate">
                                in {item.project_name}
                            </p>
                        )}
                    </div>
                );
            },
        }),
        columnHelper.accessor("deleted_at", {
            header: "Deleted",
            cell: (info) => (
                <span className="text-muted-foreground whitespace-nowrap text-sm">
                    {formatDistanceToNow(info.getValue())}
                </span>
            ),
        }),
        columnHelper.display({
            id: "actions",
            header: "Actions",
            meta: { headerClassName: "w-[100px]", cellClassName: "w-[100px]" },
            cell: ({ row }) => {
                const item = row.original;
                return (
                    <div className="flex items-center gap-1">
                        <Tooltip>
                            <TooltipTrigger
                                render={
                                    <Button
                                        variant="ghost"
                                        size="icon-sm"
                                        aria-label={`Restore ${item.title}`}
                                        onClick={() => onRestore(item)}
                                        disabled={isMutating}
                                    >
                                        <RotateCcw className="size-4" />
                                    </Button>
                                }
                            />
                            <TooltipContent>Restore</TooltipContent>
                        </Tooltip>
                        <Tooltip>
                            <TooltipTrigger
                                render={
                                    <Button
                                        variant="ghost"
                                        size="icon-sm"
                                        aria-label={`Delete ${item.title} forever`}
                                        onClick={() => onPurge(item)}
                                        disabled={isMutating}
                                    >
                                        <Trash2 className="size-4 text-destructive" />
                                    </Button>
                                }
                            />
                            <TooltipContent>Delete forever</TooltipContent>
                        </Tooltip>
                    </div>
                );
            },
        }),
    ];
}

type ConfirmKind = "purge-selected" | "purge-one" | "empty" | null;

export default function TrashPage() {
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
    const [confirm, setConfirm] = useState<ConfirmKind>(null);
    const [pendingPurge, setPendingPurge] = useState<TrashItem | null>(null);

    const { data: items = [], isLoading } = useTrash();
    const restore = useRestoreTrash();
    const purge = usePurgeTrash();
    const empty = useEmptyTrash();

    const isMutating = restore.isPending || purge.isPending || empty.isPending;

    const handleRestoreOne = async (item: TrashItem) => {
        const payload =
            item.kind === "project"
                ? { project_ids: [item.id], task_ids: [] }
                : { project_ids: [], task_ids: [item.id] };
        await restore.mutateAsync(payload);
    };

    const handlePurgeOne = (item: TrashItem) => {
        setPendingPurge(item);
        setConfirm("purge-one");
    };

    const columns = useMemo(
        () => buildColumns(handleRestoreOne, handlePurgeOne, isMutating),
        // handleRestoreOne / handlePurgeOne are stable for the row's lifetime;
        // we only need to rebuild when the disabled state flips.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [isMutating]
    );

    const table = useReactTable({
        data: items,
        columns,
        state: { rowSelection },
        onRowSelectionChange: setRowSelection,
        getRowId: rowKey,
        getCoreRowModel: getCoreRowModel(),
        enableRowSelection: true,
    });

    const selectedKeys = Object.keys(rowSelection);
    const selectedCount = selectedKeys.length;

    const handleRestoreSelected = async () => {
        const payload = partition(items, selectedKeys);
        await restore.mutateAsync(payload);
        setRowSelection({});
    };

    const handleConfirmPurgeSelected = async () => {
        const payload = partition(items, selectedKeys);
        await purge.mutateAsync(payload);
        setRowSelection({});
        setConfirm(null);
    };

    const handleConfirmPurgeOne = async () => {
        if (!pendingPurge) return;
        const payload =
            pendingPurge.kind === "project"
                ? { project_ids: [pendingPurge.id], task_ids: [] }
                : { project_ids: [], task_ids: [pendingPurge.id] };
        await purge.mutateAsync(payload);
        setPendingPurge(null);
        setConfirm(null);
    };

    const handleConfirmEmpty = async () => {
        await empty.mutateAsync();
        setRowSelection({});
        setConfirm(null);
    };

    return (
        <>
            <AppNavbar title="Trash" />

            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1">
                        <h2 className="text-lg font-semibold">Trash</h2>
                        <p className="text-sm text-muted-foreground">
                            Restore deleted items or remove them permanently. Deleted
                            projects keep their tasks until you purge the project.
                        </p>
                    </div>
                    <Button
                        variant="outline"
                        size="lg"
                        onClick={() => setConfirm("empty")}
                        disabled={items.length === 0 || isMutating}
                        className="px-4! shrink-0"
                    >
                        <TrashIcon className="size-4" />
                        Empty trash
                    </Button>
                </div>

                {selectedCount > 0 && (
                    <div className="flex items-center justify-between rounded-lg border bg-muted/30 px-4 py-2.5">
                        <p className="text-sm">
                            <span className="font-medium">{selectedCount}</span>{" "}
                            item{selectedCount === 1 ? "" : "s"} selected
                        </p>
                        <div className="flex items-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRestoreSelected}
                                disabled={isMutating}
                            >
                                <RotateCcw className="size-4" />
                                Restore
                            </Button>
                            <Button
                                variant="destructive"
                                size="sm"
                                onClick={() => setConfirm("purge-selected")}
                                disabled={isMutating}
                            >
                                <Trash2 className="size-4" />
                                Delete forever
                            </Button>
                        </div>
                    </div>
                )}

                <DataTable
                    table={table}
                    isLoading={isLoading}
                    empty={
                        <>
                            <TrashIcon className="mb-3 size-10 text-muted-foreground" />
                            <p className="text-sm font-medium">Trash is empty</p>
                            <p className="text-xs text-muted-foreground mt-1">
                                Projects and tasks you delete will appear here.
                            </p>
                        </>
                    }
                />
            </main>

            <AlertDialog
                open={confirm === "empty"}
                onOpenChange={(o) => !o && setConfirm(null)}
            >
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Empty the trash?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This will permanently delete every project and task in your
                            trash. This cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={empty.isPending}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleConfirmEmpty();
                            }}
                            disabled={empty.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {empty.isPending ? "Emptying…" : "Empty trash"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            <AlertDialog
                open={confirm === "purge-selected"}
                onOpenChange={(o) => !o && setConfirm(null)}
            >
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>
                            Delete {selectedCount} item{selectedCount === 1 ? "" : "s"} forever?
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            This cannot be undone. Purging a project also deletes all of
                            its tasks.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={purge.isPending}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleConfirmPurgeSelected();
                            }}
                            disabled={purge.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {purge.isPending ? "Deleting…" : "Delete forever"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            <AlertDialog
                open={confirm === "purge-one"}
                onOpenChange={(o) => !o && setConfirm(null)}
            >
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>
                            Delete &ldquo;{pendingPurge?.title}&rdquo; forever?
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            This cannot be undone.
                            {pendingPurge?.kind === "project" &&
                                " All tasks under this project will be deleted too."}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={purge.isPending}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleConfirmPurgeOne();
                            }}
                            disabled={purge.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {purge.isPending ? "Deleting…" : "Delete forever"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
