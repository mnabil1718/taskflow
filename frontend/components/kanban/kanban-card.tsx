"use client";

import Link from "next/link";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Calendar, GripVertical, User } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { formatDate } from "@/lib/date-utils";
import { cn } from "@/lib/utils";
import type { Task, TaskPriority } from "@/lib/types";

const priorityBadge: Record<TaskPriority, { label: string; variant: "default" | "secondary" | "destructive" }> = {
    low: { label: "Low", variant: "secondary" },
    medium: { label: "Medium", variant: "default" },
    high: { label: "High", variant: "destructive" },
};

function initialsOf(name: string | undefined | null): string {
    if (!name) return "";
    return name
        .split(/\s+/)
        .map((p) => p[0])
        .join("")
        .slice(0, 2)
        .toUpperCase();
}

interface KanbanCardProps {
    task: Task;
    /** Drag affordance: when false the grip-handle cursor and hover hint
     * are hidden because in-column reorder is meaningless for this card
     * (column is on a derived sort, or the viewer isn't the owner). The
     * card is still draggable via dnd-kit so cross-column status drops
     * keep working for everyone — the drop handler is the final arbiter
     * of whether a write actually happens. */
    sortable: boolean;
    /** When true the sortable hook is fully disabled. Used by the
     * DragOverlay clone so it doesn't try to claim its own sortable id
     * alongside the real card during the drag. */
    disableDnd?: boolean;
}

export function KanbanCard({ task, sortable, disableDnd }: KanbanCardProps) {
    const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
        useSortable({ id: task.id, disabled: !!disableDnd });

    const style = disableDnd
        ? undefined
        : {
              transform: CSS.Transform.toString(transform),
              transition,
          };

    const priority = priorityBadge[task.priority];
    const initials = initialsOf(task.assignee_name);

    // Outer wrapper is a plain <div> so dnd-kit's setNodeRef attaches to
    // a real DOM node. The shared Card / CardContent components are
    // React 18 function components without forwardRef, so passing
    // setNodeRef directly to <Card> would silently no-op and the drag
    // wouldn't start at all. The wrapper takes the listeners too; the
    // Card inside is just the visual.
    //
    // Drag handle layout: the wrapper carries listeners across the
    // whole card surface, but the title <Link> is wrapped in an inline
    // element that stops the pointerdown from reaching the wrapper.
    // That makes the link only as wide as its text — anywhere else on
    // the card (whitespace to its right, the meta row, the grip area)
    // initiates a drag. Click navigates to /tasks/:id; drag reorders.
    return (
        <div
            ref={setNodeRef}
            style={style}
            {...attributes}
            {...(disableDnd ? {} : listeners)}
            className={cn(
                "touch-none",
                isDragging && "opacity-50",
                !disableDnd && (sortable ? "cursor-grab active:cursor-grabbing" : "cursor-grab")
            )}
        >
            <Card size="sm" className="group/card gap-2 px-3 py-2.5">
                <div className="flex items-start gap-2 min-w-0">
                    {/* The inline wrapper stops the drag from claiming
                        pointerdown on the title; the link itself is
                        max-w-full + truncate so a long title can't push
                        the grip icon off-screen. */}
                    <span
                        className="inline-flex max-w-full min-w-0"
                        onPointerDown={(e) => e.stopPropagation()}
                    >
                        <Link
                            href={`/tasks/${task.id}`}
                            className="text-sm font-medium leading-snug hover:underline truncate"
                        >
                            {task.title}
                        </Link>
                    </span>
                    {sortable && (
                        <GripVertical
                            className="ml-auto size-3.5 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover/card:opacity-60"
                            aria-hidden
                        />
                    )}
                </div>

                <div className="flex items-center justify-between gap-2 pt-1">
                    <Badge variant={priority.variant} className="capitalize text-[0.7rem] px-1.5 py-0">
                        {priority.label}
                    </Badge>

                    <div className="flex items-center gap-2 text-xs text-muted-foreground min-w-0">
                        {task.due_date && (
                            <span className="flex items-center gap-1 whitespace-nowrap" title={`Due ${formatDate(task.due_date)}`}>
                                <Calendar className="size-3" />
                                {formatDate(task.due_date)}
                            </span>
                        )}
                        {task.assignee_id && initials ? (
                            <Avatar className="size-5" title={task.assignee_name ?? "Assignee"}>
                                <AvatarFallback className="text-[0.55rem]">
                                    {initials}
                                </AvatarFallback>
                            </Avatar>
                        ) : (
                            <span
                                className="flex items-center gap-1 text-muted-foreground/70"
                                title="Unassigned"
                            >
                                <User className="size-3" />
                            </span>
                        )}
                    </div>
                </div>
            </Card>
        </div>
    );
}
