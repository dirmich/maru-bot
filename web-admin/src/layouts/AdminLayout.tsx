import { SetupNotice } from "@/components/setup-notice";
import { Sidebar } from "@/components/sidebar";
import { Outlet } from "react-router-dom";

// MobileConfirmDialog mock if not exists
const MobileConfirmDialog = () => null;

export function AdminLayout() {
    return (
        <div className="flex h-screen bg-slate-50 dark:bg-slate-950 overflow-hidden">
            <Sidebar />
            <main className="flex-1 overflow-auto relative">
                <SetupNotice />
                {/* Global dialogs could be placed here */}
                {/* <AlertDialog /> */}
                <MobileConfirmDialog />
                <Outlet />
            </main>
        </div>
    );
}
