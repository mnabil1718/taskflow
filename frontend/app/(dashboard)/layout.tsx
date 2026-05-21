import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <SidebarProvider>
            <AppSidebar />
            <div className="flex flex-1 flex-col overflow-hidden">
                {children}
                <footer className="border-t px-6 py-3 text-xs text-muted-foreground">
                    © {new Date().getFullYear()} TaskFlow
                </footer>
            </div>
        </SidebarProvider>
    );
}
