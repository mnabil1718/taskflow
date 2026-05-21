"use client";

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
import {
    type NotificationItem,
    type NotificationEventType,
} from "@/hooks/use-sse-notifications";

const EVENT_LABELS: Record<NotificationEventType, string> = {
    "task.created": "Task created",
    "task.updated": "Task updated",
    "task.deleted": "Task deleted",
    "task.assigned": "Task assigned",
    "task.moved": "Task moved",
    "task.deadline_reminder": "Deadline reminder",
};

function notificationMessage(n: NotificationItem): string {
    if (n.type === "task.deadline_reminder" && n.reminderWindow) {
        return `Due in ${n.reminderWindow}: ${n.taskTitle ?? "a task"}`;
    }
    const label = EVENT_LABELS[n.type];
    return n.taskTitle ? `${label}: ${n.taskTitle}` : label;
}

interface NotificationTrayProps {
    notifications: NotificationItem[];
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
                                <li
                                    key={n.id}
                                    onClick={() => onMarkRead(n.id)}
                                    className={cn(
                                        "flex cursor-pointer gap-2 px-3 py-2.5 hover:bg-accent",
                                        !n.read && "bg-accent/40"
                                    )}
                                >
                                    <span
                                        className={cn(
                                            "mt-1.5 size-2 shrink-0 rounded-full",
                                            n.read ? "bg-transparent" : "bg-primary"
                                        )}
                                    />
                                    <div className="min-w-0 flex-1">
                                        <p className="truncate text-sm">{notificationMessage(n)}</p>
                                        <p className="text-xs text-muted-foreground">
                                            {formatDistanceToNow(n.timestamp)}
                                        </p>
                                    </div>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
