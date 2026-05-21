"use client";

import {
    CheckCircle2,
    Clock,
    ListTodo,
    LayoutDashboard,
} from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useProjectTaskCounts, useUpcomingTasks } from "@/hooks/use-dashboard";
import type { ProjectTaskCounts, UpcomingTask, TaskPriority } from "@/lib/types";

// --- Skeletons ---

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

function ProjectsTableSkeleton() {
    return (
        <Card>
            <CardHeader>
                <Skeleton className="h-5 w-40" />
            </CardHeader>
            <CardContent className="space-y-3">
                {Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-4">
                        <Skeleton className="h-4 flex-1" />
                        <Skeleton className="h-4 w-10" />
                        <Skeleton className="h-4 w-10" />
                        <Skeleton className="h-4 w-10" />
                        <Skeleton className="h-4 w-10" />
                    </div>
                ))}
            </CardContent>
        </Card>
    );
}

function UpcomingTasksSkeleton() {
    return (
        <Card>
            <CardHeader>
                <Skeleton className="h-5 w-36" />
            </CardHeader>
            <CardContent className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-3">
                        <Skeleton className="h-4 w-4 rounded-full" />
                        <Skeleton className="h-4 flex-1" />
                        <Skeleton className="h-5 w-16 rounded-full" />
                        <Skeleton className="h-4 w-20" />
                    </div>
                ))}
            </CardContent>
        </Card>
    );
}

// --- Stat card ---

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

    if (isLoading) return <ProjectsTableSkeleton />;

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">Tasks by Project</CardTitle>
            </CardHeader>
            <CardContent>
                {!data || data.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                        No projects yet. Create your first project to see task counts here.
                    </p>
                ) : (
                    <div className="space-y-2">
                        <div className="grid grid-cols-[1fr_auto_auto_auto_auto] gap-4 pb-1 text-xs font-medium text-muted-foreground">
                            <span>Project</span>
                            <span className="text-center">Todo</span>
                            <span className="text-center">In Progress</span>
                            <span className="text-center">Done</span>
                            <span className="text-center">Total</span>
                        </div>
                        {data.map((row: ProjectTaskCounts) => (
                            <div key={row.project_id} className="grid grid-cols-[1fr_auto_auto_auto_auto] gap-4 text-sm py-1 border-t">
                                <span className="truncate font-medium">{row.project_name}</span>
                                <span className="text-center tabular-nums">{row.todo}</span>
                                <span className="text-center tabular-nums">{row.in_progress}</span>
                                <span className="text-center tabular-nums">{row.done}</span>
                                <span className="text-center tabular-nums font-semibold">{row.total}</span>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
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

    if (isLoading) return <UpcomingTasksSkeleton />;

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">Upcoming Due Dates</CardTitle>
            </CardHeader>
            <CardContent>
                {!data || data.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                        No upcoming tasks. Tasks with due dates will appear here.
                    </p>
                ) : (
                    <div className="space-y-3">
                        {data.map((task: UpcomingTask) => (
                            <div key={task.id} className="flex items-center gap-3 text-sm">
                                <CheckCircle2 className="size-4 shrink-0 text-muted-foreground" />
                                <span className="flex-1 truncate">{task.title}</span>
                                <Badge variant={priorityVariant[task.priority]} className="shrink-0">
                                    {task.priority}
                                </Badge>
                                <span className="shrink-0 text-xs text-muted-foreground">
                                    {new Date(task.due_date).toLocaleDateString()}
                                </span>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
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
