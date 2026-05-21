import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { AppFooter } from "@/components/app-footer";

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
                <AppFooter />
            </div>
        </SidebarProvider>
    );
}
