"use client";

import Link from "next/link";
import { useState } from "react";
import {
    useReactTable,
    getCoreRowModel,
    getSortedRowModel,
    createColumnHelper,
    type SortingState,
    type RowSelectionState,
} from "@tanstack/react-table";
import { ClipboardList, Trash2 } from "lucide-react";

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
import { DataTable } from "@/components/data-table";
import { DataTableToolbar } from "@/components/data-table-toolbar";
import { FilterMultiSelect } from "@/components/filter-multi-select";
import { PaginationControls } from "@/components/pagination-controls";
import { CreateTaskGlobalDialog } from "@/components/tasks/create-task-global-dialog";
import { TaskRowActions } from "@/components/tasks/task-row-actions";
import { useProjects } from "@/hooks/use-projects";
import { useAllTasks, useBulkDeleteTasks } from "@/hooks/use-tasks";
import { useDebounced } from "@/hooks/use-debounced";
import { formatDate } from "@/lib/date-utils";
import type { Task, TaskPriority, TaskStatus } from "@/lib/types";

const PAGE_SIZE = 10;

const statusBadge: Record<TaskStatus, { label: string; variant: "default" | "secondary" | "outline" }> = {
    todo: { label: "To Do", variant: "secondary" },
    in_progress: { label: "In Progress", variant: "default" },
    done: { label: "Done", variant: "outline" },
};

const priorityBadge: Record<TaskPriority, { label: string; variant: "default" | "secondary" | "destructive" }> = {
    low: { label: "Low", variant: "secondary" },
    medium: { label: "Medium", variant: "default" },
    high: { label: "High", variant: "destructive" },
};

const sortableColumns: Record<string, string> = {
    title: "title",
    status: "status",
    priority: "priority",
    due_date: "due_date",
    created_at: "created_at",
};

function buildColumns(projectMap: Map<string, string>) {
    const columnHelper = createColumnHelper<Task>();

    return [
        columnHelper.display({
            id: "select",
            meta: { headerClassName: "w-[40px]", cellClassName: "w-[40px]" },
            header: ({ table }) => (
                <Checkbox
                    checked={table.getIsAllPageRowsSelected()}
                    indeterminate={table.getIsSomePageRowsSelected()}
                    onCheckedChange={(v) => table.toggleAllPageRowsSelected(!!v)}
                    aria-label="Select all on page"
                />
            ),
            cell: ({ row }) => (
                <Checkbox
                    checked={row.getIsSelected()}
                    onCheckedChange={(v) => row.toggleSelected(!!v)}
                    aria-label={`Select ${row.original.title}`}
                />
            ),
            enableSorting: false,
        }),
        columnHelper.accessor("title", {
            header: "Title",
            meta: { cellClassName: "max-w-xs", cellInnerClassName: "truncate" },
        }),
        columnHelper.accessor("description", {
            header: "Description",
            enableSorting: false,
            meta: { cellClassName: "max-w-xs", cellInnerClassName: "truncate" },
            cell: (info) => (
                <span className="text-muted-foreground">
                    {info.getValue() ?? "—"}
                </span>
            ),
        }),
        columnHelper.accessor("project_id", {
            header: "Project",
            enableSorting: false,
            cell: (info) => {
                const pid = info.getValue();
                const name = projectMap.get(pid);
                return (
                    <Link
                        href={`/projects/${pid}`}
                        className="text-sm hover:underline truncate max-w-[10rem] block"
                    >
                        {name ?? pid.slice(0, 8) + "…"}
                    </Link>
                );
            },
        }),
        columnHelper.accessor("status", {
            header: "Status",
            cell: (info) => {
                const { label, variant } = statusBadge[info.getValue()];
                return (
                    <Badge variant={variant} className="capitalize whitespace-nowrap">
                        {label}
                    </Badge>
                );
            },
        }),
        columnHelper.accessor("priority", {
            header: "Priority",
            cell: (info) => {
                const { label, variant } = priorityBadge[info.getValue()];
                return (
                    <Badge variant={variant} className="capitalize whitespace-nowrap">
                        {label}
                    </Badge>
                );
            },
        }),
        columnHelper.accessor("due_date", {
            header: "Due date",
            cell: (info) => {
                const v = info.getValue();
                return (
                    <span className="text-muted-foreground whitespace-nowrap">
                        {v ? formatDate(v) : "—"}
                    </span>
                );
            },
        }),
        columnHelper.accessor("created_at", {
            header: "Created",
            cell: (info) => (
                <span className="text-muted-foreground whitespace-nowrap">
                    {formatDate(info.getValue())}
                </span>
            ),
        }),
        columnHelper.display({
            id: "actions",
            header: "Actions",
            enableSorting: false,
            cell: ({ row }) => (
                <TaskRowActions
                    task={row.original}
                    projectId={row.original.project_id}
                />
            ),
        }),
    ];
}

export default function TasksPage() {
    const [page, setPage] = useState(1);
    const [sorting, setSorting] = useState<SortingState>([]);
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
    const [searchInput, setSearchInput] = useState("");
    const [statusFilter, setStatusFilter] = useState<TaskStatus[]>([]);
    const [priorityFilter, setPriorityFilter] = useState<TaskPriority[]>([]);
    const [confirmOpen, setConfirmOpen] = useState(false);

    const search = useDebounced(searchInput, 300);

    const sortCol = sorting[0];
    const filter = {
        page,
        limit: PAGE_SIZE,
        search: search || undefined,
        status: statusFilter.length > 0 ? statusFilter : undefined,
        priority: priorityFilter.length > 0 ? priorityFilter : undefined,
        sort_by: sortCol ? sortableColumns[sortCol.id] : undefined,
        sort_order: sortCol ? (sortCol.desc ? "desc" : "asc") as "asc" | "desc" : undefined,
    };

    const { data, isLoading, isFetching } = useAllTasks(filter);
    const bulkDelete = useBulkDeleteTasks();
    const { data: projectsData } = useProjects({ page: 1, limit: 100 });

    const tasks = data?.items ?? [];
    const total = data?.total ?? 0;
    const totalPages = data?.total_pages ?? 0;

    const projectMap = new Map(
        projectsData?.items.map((p) => [p.id, p.name]) ?? []
    );

    const columns = buildColumns(projectMap);

    const table = useReactTable({
        data: tasks,
        columns,
        state: { sorting, rowSelection },
        onSortingChange: (updater) => {
            setSorting(updater);
            setPage(1);
        },
        onRowSelectionChange: setRowSelection,
        getRowId: (row) => row.id,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
        manualSorting: true,
        enableRowSelection: true,
    });

    const selectedIds = Object.keys(rowSelection);
    const selectedCount = selectedIds.length;

    const handleConfirmDelete = async () => {
        await bulkDelete.mutateAsync(selectedIds);
        setRowSelection({});
        setConfirmOpen(false);
    };

    const handleSearchChange = (value: string) => {
        setSearchInput(value);
        setPage(1);
    };

    const handleStatusChange = (vals: string[]) => {
        setStatusFilter(vals as TaskStatus[]);
        setPage(1);
    };

    const handlePriorityChange = (vals: string[]) => {
        setPriorityFilter(vals as TaskPriority[]);
        setPage(1);
    };

    const toolbar = (
        <DataTableToolbar
            searchValue={searchInput}
            onSearchChange={handleSearchChange}
            searchPlaceholder="Search tasks…"
        >
            <FilterMultiSelect
                placeholder="All statuses"
                options={[
                    { value: "todo", label: "To Do" },
                    { value: "in_progress", label: "In Progress" },
                    { value: "done", label: "Done" },
                ]}
                value={statusFilter}
                onChange={handleStatusChange}
            />
            <FilterMultiSelect
                placeholder="All priorities"
                options={[
                    { value: "low", label: "Low" },
                    { value: "medium", label: "Medium" },
                    { value: "high", label: "High" },
                ]}
                value={priorityFilter}
                onChange={handlePriorityChange}
            />
        </DataTableToolbar>
    );

    return (
        <>
            <AppNavbar title="Tasks" />

            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-lg font-semibold">All Tasks</h2>
                    <div className="flex items-center gap-2">
                        {selectedCount > 0 && (
                            <Button
                                variant="destructive"
                                size="lg"
                                onClick={() => setConfirmOpen(true)}
                                disabled={bulkDelete.isPending}
                                className="px-4!"
                            >
                                <Trash2 className="size-4" />
                                Delete {selectedCount} selected
                            </Button>
                        )}
                        <CreateTaskGlobalDialog />
                    </div>
                </div>

                <DataTable
                    table={table}
                    isLoading={isLoading}
                    toolbar={toolbar}
                    empty={
                        <>
                            <ClipboardList className="mb-3 size-10 text-muted-foreground" />
                            <p className="text-sm font-medium">No tasks found</p>
                            <p className="text-xs text-muted-foreground mt-1">
                                {search || statusFilter.length > 0 || priorityFilter.length > 0
                                    ? "Try adjusting your search or filters."
                                    : "Tasks from your projects will appear here."}
                            </p>
                        </>
                    }
                />

                <PaginationControls
                    page={page}
                    pageSize={PAGE_SIZE}
                    total={total}
                    totalPages={totalPages}
                    onPageChange={setPage}
                    disabled={isFetching}
                />
            </main>
            <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>
                            Delete {selectedCount} task{selectedCount === 1 ? "" : "s"}?
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            Only tasks you created or projects you own will be deleted.
                            Others in the selection will be skipped.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={bulkDelete.isPending}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleConfirmDelete();
                            }}
                            disabled={bulkDelete.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {bulkDelete.isPending ? "Deleting…" : "Delete"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
