import { Sidebar } from "@/components/sidebar";
// Removed Next.js specific imports since we are switching to SPA architecture
import { Outlet } from "react-router-dom";
import { ConfirmDialog } from "@/components/mobile-confirm-dialog"; // You might need to adjust or mock this

// MobileConfirmDialog mock if not exists
const MobileConfirmDialog = () => null;

export function AdminLayout() {
    return (
        <div className="flex h-screen bg-slate-50 dark:bg-slate-950 overflow-hidden">
            <Sidebar />
            <main className="flex-1 overflow-auto relative">
                {/* Global dialogs could be placed here */}
                {/* <AlertDialog /> */}
                <MobileConfirmDialog />
                <Outlet />
            </main>
        </div>
    );
}
