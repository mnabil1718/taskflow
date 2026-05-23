"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import {
    useReactTable,
    getCoreRowModel,
    getSortedRowModel,
    createColumnHelper,
    type SortingState,
    type RowSelectionState,
} from "@tanstack/react-table";
import { ArrowLeft, ClipboardList, Trash2 } from "lucide-react";

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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { DataTable } from "@/components/data-table";
import { DataTableToolbar } from "@/components/data-table-toolbar";
import { PaginationControls } from "@/components/pagination-controls";
import { CreateTaskDialog } from "@/components/tasks/create-task-dialog";
import { TaskRowActions } from "@/components/tasks/task-row-actions";
import { useProject, useProjectMembers } from "@/hooks/use-projects";
import { useTasks, useBulkDeleteTasks } from "@/hooks/use-tasks";
import { useDebounced } from "@/hooks/use-debounced";
import { formatDate } from "@/lib/date-utils";
import type {
    ProjectMember,
    Task,
    TaskPriority,
    TaskStatus,
} from "@/lib/types";

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

function buildColumns(members: ProjectMember[], projectId: string) {
    const memberMap = new Map(members.map((m) => [m.user_id, m.name]));
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
        columnHelper.accessor("assignee_id", {
            header: "Assignee",
            enableSorting: false,
            cell: (info) => {
                const aid = info.getValue();
                return (
                    <span className="text-muted-foreground">
                        {aid ? (memberMap.get(aid) ?? "Unknown") : "—"}
                    </span>
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
                    projectId={projectId}
                    members={members}
                />
            ),
        }),
    ];
}

export default function ProjectDetailPage() {
    const { id } = useParams<{ id: string }>();

    const [page, setPage] = useState(1);
    const [sorting, setSorting] = useState<SortingState>([]);
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
    const [searchInput, setSearchInput] = useState("");
    const [statusFilter, setStatusFilter] = useState<TaskStatus | "">("");
    const [priorityFilter, setPriorityFilter] = useState<TaskPriority | "">("");
    const [confirmOpen, setConfirmOpen] = useState(false);

    const search = useDebounced(searchInput, 300);

    const sortCol = sorting[0];
    const filter = {
        page,
        limit: PAGE_SIZE,
        search: search || undefined,
        status: statusFilter || undefined,
        priority: priorityFilter || undefined,
        sort_by: sortCol ? sortableColumns[sortCol.id] : undefined,
        sort_order: sortCol ? (sortCol.desc ? "desc" : "asc") as "asc" | "desc" : undefined,
    };

    const { data: project, isLoading: projectLoading } = useProject(id);
    const { data: members = [] } = useProjectMembers(id);
    const { data, isLoading: tasksLoading, isFetching } = useTasks(id, filter);
    const bulkDelete = useBulkDeleteTasks();

    const tasks = data?.items ?? [];
    const total = data?.total ?? 0;
    const totalPages = data?.total_pages ?? 0;

    const columns = buildColumns(members, id);

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

    const handleStatusChange = (val: string | null) => {
        setStatusFilter(!val || val === "all" ? "" : val as TaskStatus);
        setPage(1);
    };

    const handlePriorityChange = (val: string | null) => {
        setPriorityFilter(!val || val === "all" ? "" : val as TaskPriority);
        setPage(1);
    };

    return (
        <>
            <AppNavbar title={projectLoading ? "Loading…" : (project?.name ?? "Project")} />
            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="-ml-2"
                        render={<Link href="/projects" />}
                    >
                        <ArrowLeft className="size-4" />
                        Projects
                    </Button>
                </div>

                <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1">
                        <h2 className="text-lg font-semibold">
                            {project?.name ?? "—"}
                        </h2>
                        {project?.description && (
                            <p className="text-sm text-muted-foreground max-w-prose">
                                {project.description}
                            </p>
                        )}
                        {project?.deadline && (
                            <p className="text-xs text-muted-foreground">
                                Deadline: {formatDate(project.deadline)}
                            </p>
                        )}
                    </div>
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
                        <CreateTaskDialog projectId={id} members={members} />
                    </div>
                </div>

                <DataTable
                    table={table}
                    isLoading={tasksLoading}
                    toolbar={
                        <DataTableToolbar
                            searchValue={searchInput}
                            onSearchChange={handleSearchChange}
                            searchPlaceholder="Search tasks…"
                        >
                            <Select value={statusFilter || "all"} onValueChange={handleStatusChange}>
                                <SelectTrigger className="w-36">
                                    <SelectValue placeholder="All statuses" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All statuses</SelectItem>
                                    <SelectItem value="todo">To Do</SelectItem>
                                    <SelectItem value="in_progress">In Progress</SelectItem>
                                    <SelectItem value="done">Done</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select value={priorityFilter || "all"} onValueChange={handlePriorityChange}>
                                <SelectTrigger className="w-36">
                                    <SelectValue placeholder="All priorities" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All priorities</SelectItem>
                                    <SelectItem value="low">Low</SelectItem>
                                    <SelectItem value="medium">Medium</SelectItem>
                                    <SelectItem value="high">High</SelectItem>
                                </SelectContent>
                            </Select>
                        </DataTableToolbar>
                    }
                    empty={
                        <>
                            <ClipboardList className="mb-3 size-10 text-muted-foreground" />
                            <p className="text-sm font-medium">No tasks yet</p>
                            <p className="text-xs text-muted-foreground mt-1">
                                Create the first task for this project.
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
