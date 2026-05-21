import { Suspense } from "react";
import {
    CheckCircle2,
    Clock,
    ListTodo,
    LayoutDashboard,
} from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

// --- Skeleton components ---

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

// --- Placeholder content (replaces real data fetch until API is wired) ---

const placeholderStats = [
    { title: "Total Tasks", value: 0, description: "Across all projects", icon: <LayoutDashboard className="size-4" /> },
    { title: "To Do", value: 0, description: "Waiting to be started", icon: <ListTodo className="size-4" /> },
    { title: "In Progress", value: 0, description: "Currently being worked on", icon: <Clock className="size-4" /> },
    { title: "Done", value: 0, description: "Completed tasks", icon: <CheckCircle2 className="size-4" /> },
];

function DashboardStats() {
    return (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {placeholderStats.map((stat) => (
                <StatCard key={stat.title} {...stat} />
            ))}
        </div>
    );
}

function ProjectTasksTable() {
    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">Tasks by Project</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-sm text-muted-foreground">
                    No projects yet. Create your first project to see task counts here.
                </div>
            </CardContent>
        </Card>
    );
}

function UpcomingTasks() {
    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">Upcoming Due Dates</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-sm text-muted-foreground">
                    No upcoming tasks. Tasks with due dates will appear here.
                </div>
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
                <Suspense
                    fallback={
                        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                            {Array.from({ length: 4 }).map((_, i) => (
                                <StatCardSkeleton key={i} />
                            ))}
                        </div>
                    }
                >
                    <DashboardStats />
                </Suspense>

                <div className="grid gap-6 lg:grid-cols-2">
                    <Suspense fallback={<ProjectsTableSkeleton />}>
                        <ProjectTasksTable />
                    </Suspense>

                    <Suspense fallback={<UpcomingTasksSkeleton />}>
                        <UpcomingTasks />
                    </Suspense>
                </div>
            </main>
        </>
    );
}
