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
    if (!name) return "?";
    return name
        .split(/\s+/)
        .map((p) => p[0])
        .join("")
        .slice(0, 2)
        .toUpperCase();
}

interface KanbanCardProps {
    task: Task;
    /** When false the card is rendered statically — no drag handles, no
     * sortable transform. Used by the column toolbar's non-manual sorts
     * so the card layout stays consistent without inviting a drag the
     * server would just ignore against the sort key. */
    sortable: boolean;
}

export function KanbanCard({ task, sortable }: KanbanCardProps) {
    const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
        useSortable({ id: task.id, disabled: !sortable });

    const style = sortable
        ? {
              transform: CSS.Transform.toString(transform),
              transition,
          }
        : undefined;

    const priority = priorityBadge[task.priority];

    return (
        <Card
            ref={setNodeRef}
            style={style}
            size="sm"
            className={cn(
                "group/card relative gap-2 px-3 py-2.5",
                isDragging && "opacity-50",
                sortable && "cursor-grab active:cursor-grabbing"
            )}
            {...attributes}
            {...(sortable ? listeners : {})}
        >
            {sortable && (
                <GripVertical
                    className="pointer-events-none absolute right-1 top-2 size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover/card:opacity-60"
                    aria-hidden
                />
            )}

            <Link
                href={`/tasks/${task.id}`}
                onPointerDown={(e) => e.stopPropagation()}
                className="text-sm font-medium leading-snug hover:underline pr-4"
            >
                {task.title}
            </Link>

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
                    {task.assignee_id ? (
                        <Avatar className="size-5" title={task.assignee_name ?? "Assignee"}>
                            <AvatarFallback className="text-[0.55rem]">
                                {initialsOf(task.assignee_name)}
                            </AvatarFallback>
                        </Avatar>
                    ) : (
                        <span className="flex items-center gap-1 text-muted-foreground/70">
                            <User className="size-3" />
                        </span>
                    )}
                </div>
            </div>
        </Card>
    );
}
