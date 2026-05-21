"use client";

import {
    flexRender,
    type Table as ReactTable,
} from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";

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
import { cn } from "@/lib/utils";

// Column-level styling hooks. Consumers attach these via columnDef.meta so the
// generic table can react to per-column layout decisions (width, alignment)
// without hard-coding column ids.
declare module "@tanstack/react-table" {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    interface ColumnMeta<TData extends unknown, TValue> {
        headerClassName?: string;
        cellClassName?: string;
    }
}

interface DataTableProps<TData> {
    table: ReactTable<TData>;
    isLoading?: boolean;
    empty?: React.ReactNode;
    skeletonRows?: number;
}

function SortIcon({ isSorted }: { isSorted: false | "asc" | "desc" }) {
    if (isSorted === "asc") return <ArrowUp className="ml-1 size-3.5" />;
    if (isSorted === "desc") return <ArrowDown className="ml-1 size-3.5" />;
    return <ArrowUpDown className="ml-1 size-3.5 opacity-50" />;
}

export function DataTable<TData>({
    table,
    isLoading,
    empty,
    skeletonRows = 5,
}: DataTableProps<TData>) {
    const columns = table.getAllLeafColumns();
    const rows = table.getRowModel().rows;

    if (isLoading) {
        return (
            <Card>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                {columns.map((col) => (
                                    <TableHead
                                        key={col.id}
                                        className={col.columnDef.meta?.headerClassName}
                                    >
                                        <Skeleton className="h-4 w-20" />
                                    </TableHead>
                                ))}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {Array.from({ length: skeletonRows }).map((_, i) => (
                                <TableRow key={i}>
                                    {columns.map((col) => (
                                        <TableCell
                                            key={col.id}
                                            className={col.columnDef.meta?.cellClassName}
                                        >
                                            <Skeleton className="h-4 w-full max-w-32" />
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

    if (rows.length === 0) {
        return (
            <Card>
                <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                    {empty}
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
                                    const headerClassName = header.column.columnDef.meta?.headerClassName;
                                    return (
                                        <TableHead key={header.id} className={cn(headerClassName)}>
                                            {header.isPlaceholder
                                                ? null
                                                : canSort ? (
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
                                                    <span className="text-[0.8rem] font-medium">
                                                        {flexRender(header.column.columnDef.header, header.getContext())}
                                                    </span>
                                                )}
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {rows.map((row) => (
                            <TableRow
                                key={row.id}
                                data-state={row.getIsSelected() ? "selected" : undefined}
                            >
                                {row.getVisibleCells().map((cell) => (
                                    <TableCell
                                        key={cell.id}
                                        className={cn(cell.column.columnDef.meta?.cellClassName)}
                                    >
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
