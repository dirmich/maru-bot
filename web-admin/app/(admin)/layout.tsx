export const dynamic = "force-dynamic";

import { Sidebar } from "@/components/sidebar";
import { authOptions } from "@/lib/auth";
import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { AlertDialog, ConfirmDialog } from "@/components/ui-custom-dialog";

export default async function AdminLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const session = await getServerSession(authOptions);

    if (!session) {
        redirect("/auth/signin");
    }

    return (
        <div className="flex h-screen bg-slate-50 dark:bg-slate-950 overflow-hidden">
            <Sidebar />
            <main className="flex-1 overflow-auto relative">
                <AlertDialog />
                <ConfirmDialog />
                {children}
            </main>
        </div>
    );
}
