"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import {
    DndContext,
    DragOverlay,
    PointerSensor,
    closestCorners,
    useSensor,
    useSensors,
    type DragEndEvent,
    type DragStartEvent,
} from "@dnd-kit/core";
import { ArrowLeft, Loader2 } from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { BoardFilters } from "@/components/kanban/board-filters";
import { KanbanCard } from "@/components/kanban/kanban-card";
import { KanbanColumn } from "@/components/kanban/kanban-column";
import { useProject } from "@/hooks/use-projects";
import { useBoard, useMoveTask } from "@/hooks/use-tasks";
import {
    DEFAULT_BOARD_FILTER,
    applyFilter,
    composeFilters,
    type BoardFilter,
} from "@/lib/board-filters";
import { Lexorank } from "@/lib/lexorank";
import type { Task, TaskStatus } from "@/lib/types";

const COLUMNS: { status: TaskStatus; title: string }[] = [
    { status: "todo", title: "To Do" },
    { status: "in_progress", title: "In Progress" },
    { status: "done", title: "Done" },
];

const lexorank = new Lexorank();

// Returns the (status, index) of a task by id within the bucketed board.
function locate(
    board: Record<TaskStatus, Task[]>,
    id: string
): { status: TaskStatus; index: number } | null {
    for (const status of ["todo", "in_progress", "done"] as TaskStatus[]) {
        const idx = board[status].findIndex((t) => t.id === id);
        if (idx !== -1) return { status, index: idx };
    }
    return null;
}

// computeNewPosition figures out the lexorank string that should land
// `task` between `prev` and `next` in the target column. Either side
// can be null (top or bottom of the column).
function computeNewPosition(prev: Task | null, next: Task | null): string {
    const [pos] = lexorank.insert(prev?.position ?? null, next?.position ?? null);
    return pos;
}

export default function BoardPage() {
    const { id } = useParams<{ id: string }>();

    const { data: project } = useProject(id);
    const { data: board, isLoading } = useBoard(id);
    const moveTask = useMoveTask(id);

    const [boardFilter, setBoardFilter] = useState<BoardFilter>(DEFAULT_BOARD_FILTER);
    const [columnFilters, setColumnFilters] = useState<Record<TaskStatus, BoardFilter>>({
        todo: DEFAULT_BOARD_FILTER,
        in_progress: DEFAULT_BOARD_FILTER,
        done: DEFAULT_BOARD_FILTER,
    });

    // The card being dragged — used to render the DragOverlay so the
    // moving card appears under the cursor instead of leaving its slot.
    const [activeTask, setActiveTask] = useState<Task | null>(null);

    // For each column, compute the effective filter (column overrides
    // board) then run applyFilter to get the rendered list. Memoised so
    // the columns don't recompute on unrelated parent renders.
    const filtered = useMemo(() => {
        const result: Record<TaskStatus, Task[]> = {
            todo: [],
            in_progress: [],
            done: [],
        };
        const buckets: Record<TaskStatus, Task[]> = {
            todo: board?.todo ?? [],
            in_progress: board?.in_progress ?? [],
            done: board?.done ?? [],
        };
        for (const status of Object.keys(result) as TaskStatus[]) {
            const effective = composeFilters(boardFilter, columnFilters[status]);
            result[status] = applyFilter(buckets[status], effective);
        }
        return result;
    }, [board, boardFilter, columnFilters]);

    const isSortableColumn = (status: TaskStatus): boolean =>
        composeFilters(boardFilter, columnFilters[status]).sortBy === "position";

    // Drag-and-drop sensors. A small distance threshold keeps a single
    // click on the card (which navigates to /tasks/:id) from being
    // mistaken for the start of a drag.
    const sensors = useSensors(
        useSensor(PointerSensor, { activationConstraint: { distance: 6 } })
    );

    const handleDragStart = (event: DragStartEvent) => {
        const id = String(event.active.id);
        const all = [
            ...(board?.todo ?? []),
            ...(board?.in_progress ?? []),
            ...(board?.done ?? []),
        ];
        setActiveTask(all.find((t) => t.id === id) ?? null);
    };

    const handleDragEnd = (event: DragEndEvent) => {
        setActiveTask(null);

        const { active, over } = event;
        if (!over || !board) return;

        const activeId = String(active.id);
        const overId = String(over.id);

        const from = locate(filtered, activeId);
        if (!from) return;

        // Dropping straight onto a column droppable id ("todo" / etc.)
        // sends the card to the END of that column. Dropping on another
        // card means "insert before that card in its column".
        const isColumnDrop = (COLUMNS.some((c) => c.status === overId));
        const targetStatus = isColumnDrop
            ? (overId as TaskStatus)
            : locate(filtered, overId)?.status;
        if (!targetStatus) return;

        // Same column reorder is only meaningful when the column is on
        // the manual sort — otherwise the sort would override the new
        // position immediately. Block it instead of writing then losing.
        if (from.status === targetStatus && !isSortableColumn(targetStatus)) {
            return;
        }

        const targetList = filtered[targetStatus];

        let prev: Task | null;
        let next: Task | null;
        if (isColumnDrop) {
            prev = targetList[targetList.length - 1] ?? null;
            // If the user is dropping into the column they already came
            // from, `prev` ends up as the dragged card itself — strip it.
            if (prev?.id === activeId) {
                prev = targetList[targetList.length - 2] ?? null;
            }
            next = null;
        } else {
            const overIdx = targetList.findIndex((t) => t.id === overId);
            if (overIdx === -1) return;
            // Insert *above* the over-card.
            next = targetList[overIdx] ?? null;
            prev = targetList[overIdx - 1] ?? null;
            // If `prev` would be the dragged card itself (dragging down
            // by one slot in the same column), skip it so the new
            // position sits between the two neighbours of its old slot.
            if (prev?.id === activeId) {
                prev = targetList[overIdx - 2] ?? null;
            }
        }
        if (prev?.id === activeId) prev = null;
        if (next?.id === activeId) next = null;

        const newPosition = computeNewPosition(prev, next);

        // No-op guard: same column, same neighbours.
        const fromList = filtered[from.status];
        const currentPrev = fromList[from.index - 1] ?? null;
        const currentNext = fromList[from.index + 1] ?? null;
        if (
            from.status === targetStatus &&
            currentPrev?.id === prev?.id &&
            currentNext?.id === next?.id
        ) {
            return;
        }

        moveTask.mutate({ id: activeId, status: targetStatus, position: newPosition });
    };

    if (isLoading) {
        return (
            <>
                <AppNavbar title="Board" />
                <main className="flex-1 overflow-y-auto p-6 flex items-center justify-center">
                    <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </main>
            </>
        );
    }

    return (
        <>
            <AppNavbar title={project?.name ?? "Board"} />

            <main className="flex-1 overflow-y-auto p-6 space-y-4">
                <Button
                    variant="ghost"
                    size="sm"
                    className="-ml-2"
                    render={<Link href={`/projects/${id}`} />}
                >
                    <ArrowLeft className="size-4" />
                    {project?.name ?? "Project"}
                </Button>

                <Card className="px-4 py-3">
                    <BoardFilters scope="board" value={boardFilter} onChange={setBoardFilter} />
                </Card>

                <DndContext
                    sensors={sensors}
                    collisionDetection={closestCorners}
                    onDragStart={handleDragStart}
                    onDragEnd={handleDragEnd}
                    onDragCancel={() => setActiveTask(null)}
                >
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {COLUMNS.map(({ status, title }) => (
                            <KanbanColumn
                                key={status}
                                status={status}
                                title={title}
                                tasks={filtered[status]}
                                filter={columnFilters[status]}
                                onFilterChange={(next) =>
                                    setColumnFilters((prev) => ({ ...prev, [status]: next }))
                                }
                                sortable={isSortableColumn(status)}
                            />
                        ))}
                    </div>

                    <DragOverlay>
                        {activeTask ? (
                            <KanbanCard task={activeTask} sortable={false} />
                        ) : null}
                    </DragOverlay>
                </DndContext>
            </main>
        </>
    );
}
