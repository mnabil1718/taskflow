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
import { FolderOpen, Trash2 } from "lucide-react";

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
import { useProjects, useBulkDeleteProjects } from "@/hooks/use-projects";
import { CreateProjectDialog } from "@/components/projects/create-project-dialog";
import { ProjectRowActions } from "@/components/projects/project-row-actions";
import { PaginationControls } from "@/components/pagination-controls";
import { formatDate } from "@/lib/date-utils";
import type { Project, ProjectStatus } from "@/lib/types";

const PAGE_SIZE = 10;

// --- Column definitions ---

const columnHelper = createColumnHelper<Project>();

const statusVariant: Record<ProjectStatus, "default" | "secondary"> = {
    active: "default",
    archived: "secondary",
};

const columns = [
    columnHelper.display({
        id: "select",
        meta: { headerClassName: "w-[40px]", cellClassName: "w-[40px]" },
        header: ({ table }) => (
            <Checkbox
                checked={table.getIsAllPageRowsSelected()}
                indeterminate={table.getIsSomePageRowsSelected()}
                onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
                aria-label="Select all on page"
            />
        ),
        cell: ({ row }) => (
            <Checkbox
                checked={row.getIsSelected()}
                onCheckedChange={(value) => row.toggleSelected(!!value)}
                aria-label={`Select ${row.original.name}`}
            />
        ),
        enableSorting: false,
    }),
    columnHelper.accessor("name", {
        header: "Name",
        cell: (info) => (
            <Link
                href={`/projects/${info.row.original.id}`}
                className="hover:underline"
            >
                {info.getValue()}
            </Link>
        ),
    }),
    columnHelper.accessor("description", {
        header: "Description",
        enableSorting: false,
        cell: (info) => (
            <span className="max-w-xs truncate text-muted-foreground">
                {info.getValue() ?? "—"}
            </span>
        ),
    }),
    columnHelper.accessor("status", {
        header: "Status",
        cell: (info) => {
            const status = info.getValue();
            return (
                <Badge variant={statusVariant[status]} className="capitalize">
                    {status}
                </Badge>
            );
        },
    }),
    columnHelper.accessor("deadline", {
        header: "Deadline",
        cell: (info) => {
            const v = info.getValue();
            return (
                <span className="text-muted-foreground">
                    {v ? formatDate(v) : "—"}
                </span>
            );
        },
    }),
    columnHelper.accessor("created_at", {
        header: "Created",
        cell: (info) => (
            <span className="text-muted-foreground">
                {formatDate(info.getValue())}
            </span>
        ),
    }),
    columnHelper.display({
        id: "actions",
        header: "Actions",
        cell: ({ row }) => <ProjectRowActions project={row.original} />,
        enableSorting: false,
    }),
];

// --- Page ---

export default function ProjectsPage() {
    const [page, setPage] = useState(1);
    const [sorting, setSorting] = useState<SortingState>([]);
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
    const [confirmOpen, setConfirmOpen] = useState(false);

    const { data, isLoading, isFetching } = useProjects({ page, limit: PAGE_SIZE });
    const bulkDelete = useBulkDeleteProjects();

    const projects = data?.items ?? [];
    const total = data?.total ?? 0;
    const totalPages = data?.total_pages ?? 0;

    const table = useReactTable({
        data: projects,
        columns,
        state: { sorting, rowSelection },
        onSortingChange: setSorting,
        onRowSelectionChange: setRowSelection,
        getRowId: (row) => row.id,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
        enableRowSelection: true,
    });

    const selectedIds = Object.keys(rowSelection);
    const selectedCount = selectedIds.length;

    const handleConfirmDelete = async () => {
        await bulkDelete.mutateAsync(selectedIds);
        setRowSelection({});
        setConfirmOpen(false);
    };

    return (
        <>
            <AppNavbar title="Projects" />
            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-lg font-semibold">Your Projects</h2>
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
                        <CreateProjectDialog />
                    </div>
                </div>

                <DataTable
                    table={table}
                    isLoading={isLoading}
                    empty={
                        <>
                            <FolderOpen className="mb-3 size-10 text-muted-foreground" />
                            <p className="text-sm font-medium">No projects yet</p>
                            <p className="text-xs text-muted-foreground mt-1">
                                Projects you create or join will appear here.
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
                        <AlertDialogTitle>Delete {selectedCount} project{selectedCount === 1 ? "" : "s"}?</AlertDialogTitle>
                        <AlertDialogDescription>
                            Only projects you own will be deleted. Others in the selection will be skipped.
                            This action is reversible by an administrator.
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
