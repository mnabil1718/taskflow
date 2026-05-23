"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import {
    ArrowLeft,
    ArrowRight,
    Calendar,
    FolderKanban,
    History,
    Loader2,
    User,
} from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useProject } from "@/hooks/use-projects";
import { useTask, useTaskActivityLogs } from "@/hooks/use-tasks";
import { formatDate, formatDistanceToNow } from "@/lib/date-utils";
import type { TaskPriority, TaskStatus } from "@/lib/types";

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

function initialsOf(name: string | undefined | null): string {
    if (!name) return "?";
    return name
        .split(/\s+/)
        .map((p) => p[0])
        .join("")
        .slice(0, 2)
        .toUpperCase();
}

export default function TaskDetailPage() {
    const { id } = useParams<{ id: string }>();

    const { data: task, isLoading: taskLoading, error } = useTask(id);
    const { data: project } = useProject(task?.project_id ?? "");
    const activity = useTaskActivityLogs(id);

    if (taskLoading) {
        return (
            <>
                <AppNavbar title="Loading…" />
                <main className="flex-1 overflow-y-auto p-6 flex items-center justify-center">
                    <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </main>
            </>
        );
    }

    if (error || !task) {
        return (
            <>
                <AppNavbar title="Task" />
                <main className="flex-1 overflow-y-auto p-6 space-y-4">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="-ml-2"
                        render={<Link href="/tasks" />}
                    >
                        <ArrowLeft className="size-4" />
                        Tasks
                    </Button>
                    <Card>
                        <CardContent className="py-16 text-center text-sm text-muted-foreground">
                            Task not found.
                        </CardContent>
                    </Card>
                </main>
            </>
        );
    }

    const status = statusBadge[task.status];
    const priority = priorityBadge[task.priority];

    // Flatten paginated activity into a single array for rendering. The
    // "Load more" button drives fetchNextPage which appends another page
    // to .pages, so this re-runs and renders the appended entries.
    const activityEntries =
        activity.data?.pages.flatMap((p) => p.items) ?? [];

    return (
        <>
            <AppNavbar title={task.title} />

            <main className="flex-1 overflow-y-auto p-6 space-y-6">
                <Button
                    variant="ghost"
                    size="sm"
                    className="-ml-2"
                    render={<Link href={project ? `/projects/${task.project_id}` : "/tasks"} />}
                >
                    <ArrowLeft className="size-4" />
                    {project ? project.name : "Tasks"}
                </Button>

                <div className="grid gap-6 md:grid-cols-[1fr_280px]">
                    {/* Main column — title, badges, description */}
                    <div className="space-y-6 min-w-0">
                        <div className="space-y-3">
                            <h1 className="text-2xl font-semibold leading-tight break-words">
                                {task.title}
                            </h1>
                            <div className="flex flex-wrap items-center gap-2">
                                <Badge variant={status.variant} className="capitalize">
                                    {status.label}
                                </Badge>
                                <Badge variant={priority.variant} className="capitalize">
                                    {priority.label} priority
                                </Badge>
                            </div>
                        </div>

                        <Separator />

                        <div className="space-y-2">
                            <h2 className="text-sm font-medium text-muted-foreground">Description</h2>
                            {task.description ? (
                                <p className="text-sm whitespace-pre-wrap leading-relaxed">
                                    {task.description}
                                </p>
                            ) : (
                                <p className="text-sm text-muted-foreground italic">
                                    No description.
                                </p>
                            )}
                        </div>

                        <Separator />

                        {/* Activity feed */}
                        <div className="space-y-3">
                            <div className="flex items-center gap-2">
                                <History className="size-4 text-muted-foreground" />
                                <h2 className="text-sm font-medium">Activity</h2>
                            </div>

                            {activity.isLoading ? (
                                <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
                                    <Loader2 className="size-4 animate-spin" />
                                    Loading activity…
                                </div>
                            ) : activityEntries.length === 0 ? (
                                <p className="text-sm text-muted-foreground py-4">
                                    No status changes yet.
                                </p>
                            ) : (
                                <>
                                    <ul className="space-y-3">
                                        {activityEntries.map((entry) => (
                                            <li key={entry.id} className="flex gap-3">
                                                <Avatar className="size-7 shrink-0">
                                                    <AvatarFallback className="text-[0.65rem]">
                                                        {initialsOf(entry.changed_by_name)}
                                                    </AvatarFallback>
                                                </Avatar>
                                                <div className="flex-1 min-w-0 text-sm leading-snug">
                                                    <span className="font-medium">
                                                        {entry.changed_by_name ?? "Someone"}
                                                    </span>{" "}
                                                    moved the task from{" "}
                                                    <Badge variant="outline" className="capitalize mx-0.5">
                                                        {statusBadge[entry.from_status].label}
                                                    </Badge>
                                                    <ArrowRight className="inline size-3 mx-0.5 text-muted-foreground" />
                                                    <Badge variant="outline" className="capitalize mx-0.5">
                                                        {statusBadge[entry.to_status].label}
                                                    </Badge>
                                                    <span className="block text-xs text-muted-foreground mt-0.5">
                                                        {formatDistanceToNow(entry.created_at)}
                                                    </span>
                                                </div>
                                            </li>
                                        ))}
                                    </ul>

                                    {activity.hasNextPage && (
                                        <div className="pt-1">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => activity.fetchNextPage()}
                                                disabled={activity.isFetchingNextPage}
                                            >
                                                {activity.isFetchingNextPage ? (
                                                    <>
                                                        <Loader2 className="size-3.5 animate-spin" />
                                                        Loading…
                                                    </>
                                                ) : (
                                                    "Load older activity"
                                                )}
                                            </Button>
                                        </div>
                                    )}
                                </>
                            )}
                        </div>
                    </div>

                    {/* Sidebar — metadata */}
                    <aside className="space-y-4 md:border-l md:pl-6">
                        <SidebarRow
                            icon={<FolderKanban className="size-4" />}
                            label="Project"
                            value={
                                project ? (
                                    <Link
                                        href={`/projects/${task.project_id}`}
                                        className="text-sm hover:underline"
                                    >
                                        {project.name}
                                    </Link>
                                ) : (
                                    <span className="text-sm text-muted-foreground">
                                        {task.project_id.slice(0, 8)}…
                                    </span>
                                )
                            }
                        />

                        <SidebarRow
                            icon={<User className="size-4" />}
                            label="Assignee"
                            value={
                                task.assignee_id ? (
                                    <div className="flex items-center gap-2 min-w-0">
                                        <Avatar className="size-6 shrink-0">
                                            <AvatarFallback className="text-[0.6rem]">
                                                {initialsOf(task.assignee_name)}
                                            </AvatarFallback>
                                        </Avatar>
                                        <div className="min-w-0">
                                            <p className="text-sm truncate">
                                                {task.assignee_name ?? "Unknown"}
                                            </p>
                                            {task.assignee_email && (
                                                <p className="text-xs text-muted-foreground truncate">
                                                    {task.assignee_email}
                                                </p>
                                            )}
                                        </div>
                                    </div>
                                ) : (
                                    <span className="text-sm text-muted-foreground">Unassigned</span>
                                )
                            }
                        />

                        <SidebarRow
                            icon={<Calendar className="size-4" />}
                            label="Due date"
                            value={
                                <span className="text-sm">
                                    {task.due_date ? formatDate(task.due_date) : "—"}
                                </span>
                            }
                        />

                        <Separator />

                        <div className="space-y-1 text-xs text-muted-foreground">
                            <p>Created {formatDate(task.created_at)}</p>
                            <p>Updated {formatDate(task.updated_at)}</p>
                        </div>
                    </aside>
                </div>
            </main>
        </>
    );
}

function SidebarRow({
    icon,
    label,
    value,
}: {
    icon: React.ReactNode;
    label: string;
    value: React.ReactNode;
}) {
    return (
        <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                {icon}
                <span>{label}</span>
            </div>
            <div>{value}</div>
        </div>
    );
}
