import { SetupNotice } from "@/components/setup-notice";
import { Sidebar } from "@/components/sidebar";
import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";
import { useState } from "react";
import { Outlet } from "react-router-dom";

export function AdminLayout() {
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    return (
        <div className="flex h-screen bg-slate-50 dark:bg-slate-950 overflow-hidden relative">
            {/* Desktop Sidebar */}
            <Sidebar className="hidden lg:flex" />

            {/* Mobile Sidebar Overlay */}
            {isMobileMenuOpen && (
                <div 
                    className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-40 lg:hidden animate-in fade-in duration-300"
                    onClick={() => setIsMobileMenuOpen(false)}
                />
            )}
            
            <div className={cn(
                "fixed inset-y-0 left-0 z-50 lg:hidden transition-transform duration-300 transform",
                isMobileMenuOpen ? "translate-x-0" : "-translate-x-full"
            )}>
                <Sidebar className="w-72" onClose={() => setIsMobileMenuOpen(false)} />
                <Button 
                    variant="ghost" 
                    size="icon" 
                    className="absolute top-4 -right-12 text-white hover:bg-white/20"
                    onClick={() => setIsMobileMenuOpen(false)}
                >
                    <X size={24} />
                </Button>
            </div>

            <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
                {/* Mobile Top Header */}
                <header className="flex lg:hidden items-center justify-between px-6 py-4 bg-white dark:bg-slate-900 border-b">
                    <div className="flex items-center gap-2">
                        <span className="text-2xl">🦞</span>
                        <span className="font-bold text-lg bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent italic tracking-tighter">
                            MaruBot
                        </span>
                    </div>
                    <Button 
                        variant="ghost" 
                        size="icon" 
                        onClick={() => setIsMobileMenuOpen(true)}
                        className="text-slate-500"
                    >
                        <Menu size={24} />
                    </Button>
                </header>

                <div className="flex-1 overflow-auto p-0 md:p-4">
                    <SetupNotice />
                    <Outlet />
                </div>
            </main>
        </div>
    );
}

// Helper for class names since it might not be imported
function cn(...inputs: any[]) {
    return inputs.filter(Boolean).join(" ");
}
