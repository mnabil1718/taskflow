"use client";

import Link from "next/link";
import { Bell, CheckCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "@/lib/date-utils";
import { buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Notification } from "@/lib/types";

function windowLabel(window?: string): string {
    if (window === "3d") return "in 3 days";
    if (window === "1d") return "in 1 day";
    return "soon";
}

// Human-readable line per notification. Deadline reminders lead with the
// window so the urgency reads first; an assignment names the task.
function notificationMessage(n: Notification): string {
    const title = n.title ?? "Untitled";
    switch (n.type) {
        case "task.assigned":
            return `Assigned to you: ${title}`;
        case "task.deadline_reminder":
            return `Task due ${windowLabel(n.reminder_window)}: ${title}`;
        case "project.deadline_reminder":
            return `Project due ${windowLabel(n.reminder_window)}: ${title}`;
        default:
            return title;
    }
}

// Where clicking a notification takes the user. Task notifications open the
// task detail; project deadline reminders open the project.
function notificationHref(n: Notification): string {
    if (n.task_id) return `/tasks/${n.task_id}`;
    if (n.project_id) return `/projects/${n.project_id}`;
    return "#";
}

interface NotificationTrayProps {
    notifications: Notification[];
    unreadCount: number;
    onMarkAllRead: () => void;
    onMarkRead: (id: string) => void;
}

export function NotificationTray({
    notifications,
    unreadCount,
    onMarkAllRead,
    onMarkRead,
}: NotificationTrayProps) {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger
                className={cn(buttonVariants({ variant: "ghost", size: "icon" }), "relative")}
                aria-label="Notifications"
            >
                <Bell className="size-4" />
                {unreadCount > 0 && (
                    <Badge className="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] leading-none">
                        {unreadCount > 99 ? "99+" : unreadCount}
                    </Badge>
                )}
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" side="bottom" className="!w-80 p-0">
                <div className="flex items-center justify-between border-b px-3 py-2">
                    <span className="text-sm font-semibold">Notifications</span>
                    {unreadCount > 0 && (
                        <button
                            onClick={onMarkAllRead}
                            className={cn(
                                buttonVariants({ variant: "ghost", size: "sm" }),
                                "h-7 gap-1 text-xs"
                            )}
                        >
                            <CheckCheck className="size-3" />
                            Mark all read
                        </button>
                    )}
                </div>

                <div className="max-h-[360px] overflow-y-auto">
                    {notifications.length === 0 ? (
                        <p className="px-3 py-8 text-center text-sm text-muted-foreground">
                            No notifications yet
                        </p>
                    ) : (
                        <ul>
                            {notifications.map((n) => (
                                <li key={n.id}>
                                    <Link
                                        href={notificationHref(n)}
                                        onClick={() => onMarkRead(n.id)}
                                        className={cn(
                                            "flex gap-2 px-3 py-2.5 hover:bg-accent",
                                            !n.read_at && "bg-accent/40"
                                        )}
                                    >
                                        <span
                                            className={cn(
                                                "mt-1.5 size-2 shrink-0 rounded-full",
                                                n.read_at ? "bg-transparent" : "bg-primary"
                                            )}
                                        />
                                        <div className="min-w-0 flex-1">
                                            <p className="truncate text-sm">{notificationMessage(n)}</p>
                                            <p className="text-xs text-muted-foreground">
                                                {formatDistanceToNow(n.created_at)}
                                            </p>
                                        </div>
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
