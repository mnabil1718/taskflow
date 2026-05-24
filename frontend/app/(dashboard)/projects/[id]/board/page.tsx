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
import { CreateTaskDialog } from "@/components/tasks/create-task-dialog";
import { useProject, useProjectMembers } from "@/hooks/use-projects";
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
    const { data: members = [] } = useProjectMembers(id);
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

    // "Sortable" here means the column accepts in-column manual reorder.
    // It requires the effective sort to be "position" — any other sort
    // key would override the new lexorank string immediately on the
    // next render. RBAC is uniform across members and the owner for
    // task mutations, so no role gate.
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
        if (activeId === overId) return;

        const from = locate(filtered, activeId);
        if (!from) return;

        // Dropping straight onto a column droppable id sends the card
        // to the END of that column. Otherwise the over-target is
        // another card, and we still need to decide whether to insert
        // before or after it (see rect comparison below).
        const isColumnDrop = COLUMNS.some((c) => c.status === overId);
        const targetStatus = isColumnDrop
            ? (overId as TaskStatus)
            : locate(filtered, overId)?.status;
        if (!targetStatus) return;

        const crossColumn = from.status !== targetStatus;

        // An in-column reorder is only meaningful when the column is on
        // the manual sort — any derived sort key would override the new
        // position string on the next render. Block instead of writing
        // a value that would be visually discarded.
        if (!crossColumn && !isSortableColumn(targetStatus)) return;

        // Work with the target list minus the dragged card so neighbour
        // lookups don't pick up the card we're moving. For cross-column
        // drops the filter is a no-op (active isn't in the target list).
        const targetList = filtered[targetStatus].filter((t) => t.id !== activeId);

        let prev: Task | null;
        let next: Task | null;

        if (isColumnDrop) {
            prev = targetList[targetList.length - 1] ?? null;
            next = null;
        } else {
            const overIdx = targetList.findIndex((t) => t.id === overId);
            if (overIdx === -1) return;

            // Decide insert-before vs insert-after by comparing the
            // dragged card's vertical midpoint to the over-card's
            // midpoint. closestCorners reports the nearest card as
            // `over` whenever the cursor is anywhere near it, so
            // without this check dropping below the last card would
            // always sandwich the new card *above* it — which was the
            // "can't append to the bottom" bug.
            const draggedRect = active.rect.current.translated ?? active.rect.current.initial;
            const overMid = over.rect.top + over.rect.height / 2;
            const draggedMid = draggedRect
                ? draggedRect.top + draggedRect.height / 2
                : overMid;
            const isBelowOver = draggedMid > overMid;

            if (isBelowOver) {
                prev = targetList[overIdx];
                next = targetList[overIdx + 1] ?? null;
            } else {
                prev = targetList[overIdx - 1] ?? null;
                next = targetList[overIdx];
            }
        }

        // No-op guard: same column, same neighbours. Compare against
        // the filtered list with the dragged card removed — that's the
        // logical "slot" the card occupied before the drag.
        if (!crossColumn) {
            const fromList = filtered[from.status].filter((t) => t.id !== activeId);
            const currentPrev = fromList[from.index - 1] ?? null;
            const currentNext = fromList[from.index] ?? null;
            if (currentPrev?.id === prev?.id && currentNext?.id === next?.id) {
                return;
            }
        }

        const newPosition = computeNewPosition(prev, next);
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

                <Card className="flex flex-row items-center gap-3 px-4 py-3">
                    <div className="flex-1 min-w-0">
                        <BoardFilters scope="board" value={boardFilter} onChange={setBoardFilter} />
                    </div>
                    <CreateTaskDialog projectId={id} members={members} />
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
                                projectId={id}
                                members={members}
                            />
                        ))}
                    </div>

                    {/* dropAnimation=null disables the "fly back to source"
                        tween — useMoveTask's optimistic update has already
                        moved the real card to its destination, so the
                        overlay should just disappear at the drop point
                        instead of animating back to where the user picked
                        the card up. */}
                    <DragOverlay dropAnimation={null}>
                        {activeTask ? (
                            <div className="rotate-2 shadow-lg">
                                <KanbanCard task={activeTask} sortable disableDnd />
                            </div>
                        ) : null}
                    </DragOverlay>
                </DndContext>
            </main>
        </>
    );
}
