"use client";

import Link from "next/link";
import { useState } from "react";
import {
    useReactTable,
    getCoreRowModel,
    getSortedRowModel,
    createColumnHelper,
    type SortingState,
} from "@tanstack/react-table";
import { ClipboardList } from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Badge } from "@/components/ui/badge";
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
import { useProjects } from "@/hooks/use-projects";
import { useAllTasks } from "@/hooks/use-tasks";
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
    ];
}

export default function TasksPage() {
    const [page, setPage] = useState(1);
    const [sorting, setSorting] = useState<SortingState>([]);
    const [searchInput, setSearchInput] = useState("");
    const [statusFilter, setStatusFilter] = useState<TaskStatus | "">("");
    const [priorityFilter, setPriorityFilter] = useState<TaskPriority | "">("");

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

    const { data, isLoading, isFetching } = useAllTasks(filter);
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
        state: { sorting },
        onSortingChange: (updater) => {
            setSorting(updater);
            setPage(1);
        },
        getRowId: (row) => row.id,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
        manualSorting: true,
    });

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

    const toolbar = (
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
    );

    return (
        <>
            <AppNavbar title="Tasks" />
            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-lg font-semibold">All Tasks</h2>
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
                                {search || statusFilter || priorityFilter
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
        </>
    );
}
