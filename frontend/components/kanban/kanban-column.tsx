"use client";

import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { Card } from "@/components/ui/card";
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
    /** True when the *effective* sort is "position" AND the viewer is
     * allowed to reorder. The cards still drag-render so cross-column
     * drops keep working, but in-column reorders are blocked at drop
     * time when this is false. */
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
        <Card className="flex flex-col gap-3 px-3 py-3 bg-muted/30">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold">{title}</h3>
                    <span className="text-xs text-muted-foreground tabular-nums">
                        {tasks.length}
                    </span>
                </div>
            </div>

            <BoardFilters scope="column" value={filter} onChange={onFilterChange} />

            {/* The droppable ref must land on a real DOM node — wrapping
                a div instead of using <CardContent> directly because the
                shared CardContent function component doesn't forward
                refs (React 18). The min-height keeps empty columns
                accepting drops on their entire visible area. */}
            <div
                ref={setNodeRef}
                className={cn(
                    "flex-1 flex flex-col gap-2 min-h-32 pb-2 rounded-md transition-colors",
                    isOver && "bg-primary/5 outline outline-2 outline-primary/40 outline-offset-2"
                )}
            >
                <SortableContext
                    items={tasks.map((t) => t.id)}
                    strategy={verticalListSortingStrategy}
                >
                    {tasks.length === 0 ? (
                        <p className="text-xs text-muted-foreground italic px-1 py-2">
                            {isOver ? "Drop here" : "No tasks."}
                        </p>
                    ) : (
                        tasks.map((task) => (
                            <KanbanCard key={task.id} task={task} sortable={sortable} />
                        ))
                    )}
                </SortableContext>
            </div>
        </Card>
    );
}
