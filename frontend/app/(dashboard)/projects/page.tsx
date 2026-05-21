"use client";

import Link from "next/link";
import {
    useReactTable,
    getCoreRowModel,
    getSortedRowModel,
    flexRender,
    createColumnHelper,
    type SortingState,
} from "@tanstack/react-table";
import { ArrowUpDown, ArrowUp, ArrowDown, FolderOpen } from "lucide-react";
import { useState } from "react";

import { AppNavbar } from "@/components/app-navbar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { useProjects } from "@/hooks/use-projects";
import type { Project, ProjectStatus } from "@/lib/types";

// --- Column definitions ---

const columnHelper = createColumnHelper<Project>();

const statusVariant: Record<ProjectStatus, "default" | "secondary"> = {
    active: "default",
    archived: "secondary",
};

const columns = [
    columnHelper.accessor("name", {
        header: "Name",
        cell: (info) => (
            <Link
                href={`/projects/${info.row.original.id}`}
                className="font-medium hover:underline"
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
                    {v ? new Date(v).toLocaleDateString() : "—"}
                </span>
            );
        },
    }),
    columnHelper.accessor("created_at", {
        header: "Created",
        cell: (info) => (
            <span className="text-muted-foreground">
                {new Date(info.getValue()).toLocaleDateString()}
            </span>
        ),
    }),
];

// --- Skeleton ---

function ProjectsTableSkeleton() {
    return (
        <Card>
            <CardContent className="p-0">
                <Table>
                    <TableHeader>
                        <TableRow>
                            {["Name", "Description", "Status", "Deadline", "Created"].map((h) => (
                                <TableHead key={h}>
                                    <Skeleton className="h-4 w-20" />
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {Array.from({ length: 5 }).map((_, i) => (
                            <TableRow key={i}>
                                <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                                <TableCell><Skeleton className="h-4 w-48" /></TableCell>
                                <TableCell><Skeleton className="h-5 w-16 rounded-full" /></TableCell>
                                <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                                <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </CardContent>
        </Card>
    );
}

// --- Sort icon helper ---

function SortIcon({ isSorted }: { isSorted: false | "asc" | "desc" }) {
    if (isSorted === "asc") return <ArrowUp className="ml-1 size-3.5" />;
    if (isSorted === "desc") return <ArrowDown className="ml-1 size-3.5" />;
    return <ArrowUpDown className="ml-1 size-3.5 opacity-50" />;
}

// --- Data table ---

function ProjectsDataTable({ data }: { data: Project[] }) {
    const [sorting, setSorting] = useState<SortingState>([]);

    const table = useReactTable({
        data,
        columns,
        state: { sorting },
        onSortingChange: setSorting,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
    });

    if (data.length === 0) {
        return (
            <Card>
                <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                    <FolderOpen className="mb-3 size-10 text-muted-foreground" />
                    <p className="text-sm font-medium">No projects yet</p>
                    <p className="text-xs text-muted-foreground mt-1">
                        Projects you create or join will appear here.
                    </p>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardContent className="p-0">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((hg) => (
                            <TableRow key={hg.id}>
                                {hg.headers.map((header) => {
                                    const canSort = header.column.getCanSort();
                                    return (
                                        <TableHead key={header.id}>
                                            {canSort ? (
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="-ml-3 h-8"
                                                    onClick={header.column.getToggleSortingHandler()}
                                                >
                                                    {flexRender(header.column.columnDef.header, header.getContext())}
                                                    <SortIcon isSorted={header.column.getIsSorted()} />
                                                </Button>
                                            ) : (
                                                flexRender(header.column.columnDef.header, header.getContext())
                                            )}
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows.map((row) => (
                            <TableRow key={row.id}>
                                {row.getVisibleCells().map((cell) => (
                                    <TableCell key={cell.id}>
                                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                    </TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </CardContent>
        </Card>
    );
}

// --- Page ---

export default function ProjectsPage() {
    const { data, isLoading } = useProjects();

    return (
        <>
            <AppNavbar title="Projects" />
            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h2 className="text-lg font-semibold">Your Projects</h2>
                        <p className="text-sm text-muted-foreground">
                            {isLoading ? "Loading…" : `${data?.length ?? 0} project${data?.length === 1 ? "" : "s"}`}
                        </p>
                    </div>
                </div>
                {isLoading ? <ProjectsTableSkeleton /> : <ProjectsDataTable data={data ?? []} />}
            </main>
        </>
    );
}
