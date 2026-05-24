"use client";

import { SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { NotificationTray } from "@/components/notification-tray";
import { useNotifications } from "@/hooks/use-notifications";
import { useAuth } from "@/lib/auth-context";

interface AppNavbarProps {
    title?: string;
}

export function AppNavbar({ title }: AppNavbarProps) {
    const { user } = useAuth();
    const { notifications, unreadCount, markAllRead, markRead } =
        useNotifications(!!user);

    return (
        <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-3 h-full" />
            {title && <h1 className="text-sm font-semibold">{title}</h1>}

            <div className="ml-auto">
                <NotificationTray
                    notifications={notifications}
                    unreadCount={unreadCount}
                    onMarkAllRead={markAllRead}
                    onMarkRead={markRead}
                />
            </div>
        </header>
    );
}
