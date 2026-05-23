"use client";

import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { BoardFilters } from "@/components/kanban/board-filters";
import { KanbanCard } from "@/components/kanban/kanban-card";
import { cn } from "@/lib/utils";
import type { BoardFilter } from "@/lib/board-filters";
import type { Task, TaskStatus } from "@/lib/types";

interface KanbanColumnProps {
    status: TaskStatus;
    title: string;
    tasks: Task[];
    filter: BoardFilter;
    onFilterChange: (next: BoardFilter) => void;
    /** True when the *effective* sort is "position" — the user is on
     * manual order, so dragging cards around within the column is
     * meaningful. Any other sort key turns the cards into a derived view
     * where reordering by hand wouldn't survive a refetch. */
    sortable: boolean;
}

export function KanbanColumn({
    status,
    title,
    tasks,
    filter,
    onFilterChange,
    sortable,
}: KanbanColumnProps) {
    const { setNodeRef, isOver } = useDroppable({ id: status });

    return (
        <Card
            className={cn(
                "flex flex-col gap-3 px-3 py-3 bg-muted/30 transition-colors",
                isOver && "ring-2 ring-primary/40"
            )}
        >
            <CardHeader className="flex flex-row items-center justify-between space-y-0 p-0">
                <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold">{title}</h3>
                    <span className="text-xs text-muted-foreground tabular-nums">
                        {tasks.length}
                    </span>
                </div>
            </CardHeader>

            <BoardFilters scope="column" value={filter} onChange={onFilterChange} />

            <CardContent
                ref={setNodeRef}
                className={cn(
                    "flex-1 flex flex-col gap-2 p-0 min-h-32",
                    // A bit of room at the bottom of the list so the user
                    // can drop a card "below the last one" without having
                    // to aim for a pixel-perfect strip.
                    "pb-2"
                )}
            >
                <SortableContext
                    items={tasks.map((t) => t.id)}
                    strategy={verticalListSortingStrategy}
                >
                    {tasks.length === 0 ? (
                        <p className="text-xs text-muted-foreground italic px-1">
                            No tasks.
                        </p>
                    ) : (
                        tasks.map((task) => (
                            <KanbanCard key={task.id} task={task} sortable={sortable} />
                        ))
                    )}
                </SortableContext>
            </CardContent>
        </Card>
    );
}
