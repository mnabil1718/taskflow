"use client";

import Link from "next/link";
import {
    ArrowRight,
    CheckCircle2,
    Clock,
    ListTodo,
    LayoutDashboard,
} from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { useProjectTaskCounts, useUpcomingTasks } from "@/hooks/use-dashboard";
import { formatDate } from "@/lib/date-utils";
import { cn } from "@/lib/utils";
import type { ProjectTaskCounts, UpcomingTask, TaskPriority } from "@/lib/types";

// Dashboard surfaces a snapshot — not a full list. These caps keep the
// cards readable; users can drill into the linked pages for the rest.
const PROJECTS_ON_DASHBOARD = 5;
const UPCOMING_TASKS_ON_DASHBOARD = 5;

// --- Generic dashboard primitives ---

// StatCard renders one of the metric tiles in the top row. Kept inline so
// new tiles can be added by tweaking the stats array below.
interface StatCardProps {
    title: string;
    value: number;
    description: string;
    icon: React.ReactNode;
}

function StatCard({ title, value, description, icon }: StatCardProps) {
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{title}</CardTitle>
                <span className="text-muted-foreground">{icon}</span>
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{value}</div>
                <p className="text-xs text-muted-foreground">{description}</p>
            </CardContent>
        </Card>
    );
}

function StatCardSkeleton() {
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-4 rounded" />
            </CardHeader>
            <CardContent>
                <Skeleton className="h-8 w-16 mb-1" />
                <Skeleton className="h-3 w-28" />
            </CardContent>
        </Card>
    );
}

// DashboardListCard is the shared wrapper for the two snapshot cards on
// the bottom row: title + "View all →" link in the header, then either an
// empty-state message or the caller's children. Pulled out so both
// ProjectTasksTable and UpcomingTasks render through one source of truth
// for spacing and the View-all affordance.
interface DashboardListCardProps {
    title: string;
    viewAllHref: string;
    totalCount: number;
    shownCount: number;
    emptyMessage: string;
    children: React.ReactNode;
    /** Padding overrides for CardContent. Set "flush" so a child <Table>
     * can run edge-to-edge while text-only cards keep the default inset. */
    contentVariant?: "default" | "flush";
}

function DashboardListCard({
    title,
    viewAllHref,
    totalCount,
    shownCount,
    emptyMessage,
    children,
    contentVariant = "default",
}: DashboardListCardProps) {
    const hidden = Math.max(0, totalCount - shownCount);
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 px-6 pt-2">
                <CardTitle className="text-base">{title}</CardTitle>
                {totalCount > 0 && (
                    <Link
                        href={viewAllHref}
                        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground hover:underline"
                    >
                        {hidden > 0 ? `View all (${totalCount})` : "View all"}
                        <ArrowRight className="size-3" />
                    </Link>
                )}
            </CardHeader>
            <CardContent
                className={cn(
                    contentVariant === "flush" ? "px-0 pb-0" : "px-6 pb-4"
                )}
            >
                {totalCount === 0 ? (
                    <p className={cn(
                        "text-sm text-muted-foreground",
                        contentVariant === "flush" && "px-6 pb-6"
                    )}>
                        {emptyMessage}
                    </p>
                ) : (
                    children
                )}
            </CardContent>
        </Card>
    );
}

function ListCardSkeleton({ rows = 5 }: { rows?: number }) {
    return (
        <Card>
            <CardHeader className="px-6 pt-2">
                <Skeleton className="h-5 w-40" />
            </CardHeader>
            <CardContent className="space-y-3 px-6 pb-4">
                {Array.from({ length: rows }).map((_, i) => (
                    <div key={i} className="flex items-center gap-3">
                        <Skeleton className="h-4 flex-1" />
                        <Skeleton className="h-5 w-16 rounded-full" />
                        <Skeleton className="h-4 w-20" />
                    </div>
                ))}
            </CardContent>
        </Card>
    );
}

// --- Dashboard stats section ---

function DashboardStats() {
    const { data, isLoading } = useProjectTaskCounts();

    const totals = (data ?? []).reduce(
        (acc, p) => ({
            total: acc.total + p.total,
            todo: acc.todo + p.todo,
            in_progress: acc.in_progress + p.in_progress,
            done: acc.done + p.done,
        }),
        { total: 0, todo: 0, in_progress: 0, done: 0 }
    );

    const stats = [
        { title: "Total Tasks", value: totals.total, description: "Across all projects", icon: <LayoutDashboard className="size-4" /> },
        { title: "To Do", value: totals.todo, description: "Waiting to be started", icon: <ListTodo className="size-4" /> },
        { title: "In Progress", value: totals.in_progress, description: "Currently being worked on", icon: <Clock className="size-4" /> },
        { title: "Done", value: totals.done, description: "Completed tasks", icon: <CheckCircle2 className="size-4" /> },
    ];

    if (isLoading) {
        return (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                {Array.from({ length: 4 }).map((_, i) => <StatCardSkeleton key={i} />)}
            </div>
        );
    }

    return (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
                <StatCard key={stat.title} {...stat} />
            ))}
        </div>
    );
}

// --- Tasks by project table ---

function ProjectTasksTable() {
    const { data, isLoading } = useProjectTaskCounts();

    if (isLoading) return <ListCardSkeleton />;

    const all = data ?? [];
    // Show busiest projects first so the dashboard prioritises what
    // actually needs attention. Fall back to project name for stable
    // ordering when totals tie.
    const sorted = [...all].sort(
        (a, b) => b.total - a.total || a.project_name.localeCompare(b.project_name)
    );
    const rows = sorted.slice(0, PROJECTS_ON_DASHBOARD);

    return (
        <DashboardListCard
            title="Tasks by Project"
            viewAllHref="/projects"
            totalCount={all.length}
            shownCount={rows.length}
            emptyMessage="No projects yet. Create your first project to see task counts here."
            contentVariant="flush"
        >
            {/* A real <table> shares column widths across header and rows so
                the count cells line up with their column labels (the old
                per-row CSS Grid sized "auto" columns independently per row,
                which was the alignment bug). */}
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="px-6">Project</TableHead>
                        <TableHead className="px-4 text-right w-20">Todo</TableHead>
                        <TableHead className="px-4 text-right w-28">In Progress</TableHead>
                        <TableHead className="px-4 text-right w-20">Done</TableHead>
                        <TableHead className="px-6 text-right w-20">Total</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {rows.map((row: ProjectTaskCounts) => (
                        <TableRow key={row.project_id}>
                            <TableCell className="px-6 py-3 font-medium">
                                <Link
                                    href={`/projects/${row.project_id}`}
                                    className="truncate hover:underline"
                                >
                                    {row.project_name}
                                </Link>
                            </TableCell>
                            <TableCell className="px-4 py-3 text-right tabular-nums">{row.todo}</TableCell>
                            <TableCell className="px-4 py-3 text-right tabular-nums">{row.in_progress}</TableCell>
                            <TableCell className="px-4 py-3 text-right tabular-nums">{row.done}</TableCell>
                            <TableCell className="px-6 py-3 text-right tabular-nums font-semibold">{row.total}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </DashboardListCard>
    );
}

// --- Upcoming due dates ---

const priorityVariant: Record<TaskPriority, "default" | "secondary" | "destructive"> = {
    low: "secondary",
    medium: "default",
    high: "destructive",
};

function UpcomingTasks() {
    const { data, isLoading } = useUpcomingTasks();

    if (isLoading) return <ListCardSkeleton />;

    const all = data ?? [];
    // The backend already sorts by due_date ASC, so a head-slice keeps
    // the most-imminent items on the dashboard.
    const rows = all.slice(0, UPCOMING_TASKS_ON_DASHBOARD);

    return (
        <DashboardListCard
            title="Upcoming Due Dates"
            viewAllHref="/tasks"
            totalCount={all.length}
            shownCount={rows.length}
            emptyMessage="No upcoming tasks. Tasks with due dates will appear here."
        >
            <div className="space-y-3">
                {rows.map((task: UpcomingTask) => (
                    <div key={task.id} className="flex items-center gap-3 text-sm">
                        <CheckCircle2 className="size-4 shrink-0 text-muted-foreground" />
                        <Link
                            href={`/tasks/${task.id}`}
                            className="flex-1 truncate hover:underline"
                        >
                            {task.title}
                        </Link>
                        <Badge variant={priorityVariant[task.priority]} className="shrink-0 capitalize">
                            {task.priority}
                        </Badge>
                        <span className="shrink-0 text-xs text-muted-foreground tabular-nums">
                            {formatDate(task.due_date)}
                        </span>
                    </div>
                ))}
            </div>
        </DashboardListCard>
    );
}

// --- Page ---

export default function DashboardPage() {
    return (
        <>
            <AppNavbar title="Dashboard" />
            <main className="flex-1 overflow-y-auto p-6 space-y-6">
                <DashboardStats />
                <div className="grid gap-6 lg:grid-cols-2">
                    <ProjectTasksTable />
                    <UpcomingTasks />
                </div>
            </main>
        </>
    );
}
