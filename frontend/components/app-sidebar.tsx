"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
    LayoutDashboard,
    FolderKanban,
    CheckSquare,
    Trash2,
    LogOut,
    ChevronsUpDown,
} from "lucide-react";

import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    sidebarMenuButtonVariants,
    useSidebar,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuth } from "@/lib/auth-context";

const navItems = [
    { title: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
    { title: "Projects", href: "/projects", icon: FolderKanban },
    { title: "Tasks", href: "/tasks", icon: CheckSquare },
    { title: "Trash", href: "/trash", icon: Trash2 },
];

export function AppSidebar() {
    const pathname = usePathname();
    const { user, logout } = useAuth();
    const { isMobile, setOpenMobile } = useSidebar();

    const handleNavigate = () => {
        if (isMobile) setOpenMobile(false);
    };

    const initials = user?.email
        ? user.email.slice(0, 2).toUpperCase()
        : "TF";

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" render={<Link href="/dashboard" onClick={handleNavigate} />}>
                            <span className="font-mono text-base font-bold tracking-tighter">TF</span>
                            <span className="font-mono text-base font-bold tracking-tighter">TaskFlow</span>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupLabel>Navigation</SidebarGroupLabel>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {navItems.map((item) => (
                                <SidebarMenuItem
                                    key={item.href}
                                >
                                    <SidebarMenuButton
                                        render={<Link href={item.href} onClick={handleNavigate} />}
                                        isActive={pathname === item.href || pathname.startsWith(item.href + "/")}
                                        tooltip={item.title}
                                        className={cn(item.title === "Trash" ? "text-destructive!" : "")}
                                    >
                                        <item.icon />
                                        <span>{item.title}</span>
                                    </SidebarMenuButton>
                                </SidebarMenuItem>
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>

            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <DropdownMenu>
                            <DropdownMenuTrigger
                                className={cn(sidebarMenuButtonVariants({ size: "lg" }))}
                            >
                                <Avatar className="size-8 rounded-lg">
                                    <AvatarFallback className="rounded-lg text-xs">
                                        {initials}
                                    </AvatarFallback>
                                </Avatar>
                                <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                                    <span className="truncate font-medium">
                                        {user?.email ?? "user"}
                                    </span>
                                </div>
                                <ChevronsUpDown className="ml-auto size-4 group-data-[collapsible=icon]:hidden" />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent side="top" className="min-w-56">
                                <DropdownMenuItem onClick={logout}>
                                    <LogOut />
                                    Log out
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
        </Sidebar>
    );
}
