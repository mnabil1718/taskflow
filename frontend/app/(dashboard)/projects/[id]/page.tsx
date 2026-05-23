"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import {
    useReactTable,
    getCoreRowModel,
    getSortedRowModel,
    createColumnHelper,
    type SortingState,
    type RowSelectionState,
} from "@tanstack/react-table";
import { ArrowLeft, ClipboardList, FileText, Settings, Trash2, Users } from "lucide-react";

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
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { DataTable } from "@/components/data-table";
import { DataTableToolbar } from "@/components/data-table-toolbar";
import { FilterMultiSelect } from "@/components/filter-multi-select";
import { PaginationControls } from "@/components/pagination-controls";
import { ProjectMembersPanel } from "@/components/projects/project-members-panel";
import { ProjectMetadataForm } from "@/components/projects/project-metadata-form";
import { CreateTaskDialog } from "@/components/tasks/create-task-dialog";
import { TaskRowActions } from "@/components/tasks/task-row-actions";
import { useDeleteProject, useProject, useProjectMembers } from "@/hooks/use-projects";
import { useTasks, useBulkDeleteTasks } from "@/hooks/use-tasks";
import { useDebounced } from "@/hooks/use-debounced";
import { useAuth } from "@/lib/auth-context";
import { formatDate } from "@/lib/date-utils";
import { cn } from "@/lib/utils";
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
            cell: (info) => (
                <Link
                    href={`/tasks/${info.row.original.id}`}
                    className="text-sm font-medium hover:underline"
                >
                    {info.getValue()}
                </Link>
            ),
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
                const { assignee_name, assignee_email } = info.row.original;
                if (!info.getValue()) {
                    return <span className="text-muted-foreground">—</span>;
                }
                return (
                    <div className="flex flex-col">
                        <span className="text-sm">{assignee_name ?? "Unknown"}</span>
                        {assignee_email && (
                            <span className="text-xs text-muted-foreground">{assignee_email}</span>
                        )}
                    </div>
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

type SettingsSection = "details" | "members";

const settingsNav: { id: SettingsSection; label: string; icon: React.ElementType }[] = [
    { id: "details", label: "Project details", icon: FileText },
    { id: "members", label: "Members", icon: Users },
];

export default function ProjectDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { user } = useAuth();
    const router = useRouter();

    const [page, setPage] = useState(1);
    const [sorting, setSorting] = useState<SortingState>([]);
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
    const [searchInput, setSearchInput] = useState("");
    const [statusFilter, setStatusFilter] = useState<TaskStatus[]>([]);
    const [priorityFilter, setPriorityFilter] = useState<TaskPriority[]>([]);
    const [confirmOpen, setConfirmOpen] = useState(false);
    const [settingsSection, setSettingsSection] = useState<SettingsSection>("details");
    const [deleteProjectOpen, setDeleteProjectOpen] = useState(false);

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

    const { data: project, isLoading: projectLoading } = useProject(id);
    const { data: members = [] } = useProjectMembers(id);
    const { data, isLoading: tasksLoading, isFetching } = useTasks(id, filter);
    const bulkDelete = useBulkDeleteTasks();
    const deleteProject = useDeleteProject();

    const handleDeleteProject = async () => {
        await deleteProject.mutateAsync(id);
        router.push("/projects");
    };

    const tasks = data?.items ?? [];
    const total = data?.total ?? 0;
    const totalPages = data?.total_pages ?? 0;

    const isOwner = !!user && !!project && user.id === project.owner_id;

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

    const handleStatusChange = (vals: string[]) => {
        setStatusFilter(vals as TaskStatus[]);
        setPage(1);
    };

    const handlePriorityChange = (vals: string[]) => {
        setPriorityFilter(vals as TaskPriority[]);
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

                {/* Project header */}
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

                {/* Tabs */}
                <Tabs defaultValue="tasks">
                    <TabsList>
                        <TabsTrigger value="tasks">
                            <ClipboardList className="size-4" />
                            Tasks
                        </TabsTrigger>
                        {isOwner && (
                            <TabsTrigger value="settings">
                                <Settings className="size-4" />
                                Settings
                            </TabsTrigger>
                        )}
                    </TabsList>

                    {/* Tasks tab */}
                    <TabsContent value="tasks" className="pt-4 space-y-4">
                        <div className="flex items-center justify-end gap-2">
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

                        <DataTable
                            table={table}
                            isLoading={tasksLoading}
                            toolbar={
                                <DataTableToolbar
                                    searchValue={searchInput}
                                    onSearchChange={handleSearchChange}
                                    searchPlaceholder="Search by title or assignee…"
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
                    </TabsContent>

                    {/* Settings tab (owner only) */}
                    {isOwner && (
                        <TabsContent value="settings" className="pt-6">
                            <div className="flex flex-col gap-6 md:flex-row md:gap-8">
                                {/* Nav — horizontal scrollable on mobile, vertical sidebar on md+ */}
                                <nav className="shrink-0 md:w-52">
                                    <ul className="flex gap-1 overflow-x-auto pb-1 md:flex-col md:overflow-visible md:pb-0 md:space-y-0.5">
                                        {settingsNav.map(({ id: sid, label, icon: Icon }) => (
                                            <li key={sid} className="shrink-0">
                                                <button
                                                    type="button"
                                                    onClick={() => setSettingsSection(sid)}
                                                    className={cn(
                                                        "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors text-left whitespace-nowrap",
                                                        settingsSection === sid
                                                            ? "bg-accent font-medium text-accent-foreground"
                                                            : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                                                    )}
                                                >
                                                    <Icon className="size-4 shrink-0" />
                                                    {label}
                                                </button>
                                            </li>
                                        ))}
                                        <li className="shrink-0 md:mt-2">
                                            <button
                                                type="button"
                                                onClick={() => setDeleteProjectOpen(true)}
                                                className="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-destructive transition-colors text-left whitespace-nowrap hover:bg-destructive/10"
                                            >
                                                <Trash2 className="size-4 shrink-0" />
                                                Delete project
                                            </button>
                                        </li>
                                    </ul>
                                </nav>

                                <Separator className="md:hidden" />
                                <Separator orientation="vertical" className="hidden md:block h-auto self-stretch" />

                                {/* Content area */}
                                <div className="flex-1 min-w-0 space-y-4">
                                    {settingsSection === "details" && (
                                        <>
                                            <div>
                                                <h3 className="text-base font-semibold">Project details</h3>
                                                <p className="text-sm text-muted-foreground">
                                                    Update the project name, description, status, and deadline.
                                                </p>
                                            </div>
                                            <Separator />
                                            {project && <ProjectMetadataForm project={project} />}
                                        </>
                                    )}

                                    {settingsSection === "members" && (
                                        <>
                                            <div>
                                                <h3 className="text-base font-semibold">Members</h3>
                                                <p className="text-sm text-muted-foreground">
                                                    Add or remove project members.
                                                </p>
                                            </div>
                                            <Separator />
                                            <ProjectMembersPanel
                                                projectId={id}
                                                currentUserId={user?.id ?? ""}
                                            />
                                        </>
                                    )}

                                </div>
                            </div>
                        </TabsContent>
                    )}
                </Tabs>
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

            <AlertDialog open={deleteProjectOpen} onOpenChange={setDeleteProjectOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Delete &ldquo;{project?.name}&rdquo;?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This will permanently delete the project and all its tasks.
                            This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={deleteProject.isPending}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={(e) => {
                                e.preventDefault();
                                handleDeleteProject();
                            }}
                            disabled={deleteProject.isPending}
                            className="bg-destructive text-white hover:bg-destructive/90"
                        >
                            {deleteProject.isPending ? "Deleting…" : "Delete project"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
