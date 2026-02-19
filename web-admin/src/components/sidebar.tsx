import { Link, useLocation } from "react-router-dom";
import {
    MessageSquare,
    Settings,
    Package,
    Cpu,
    LogOut,
    User as UserIcon,
    ChevronLeft,
    ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useState } from "react";

const menuItems = [
    { name: "채팅", href: "/chat", icon: MessageSquare },
    { name: "GPIO 설정", href: "/gpio", icon: Cpu },
    { name: "스킬 & 툴", href: "/skills", icon: Package },
    { name: "환경 설정", href: "/settings", icon: Settings },
];

export function Sidebar() {
    const location = useLocation();
    const pathname = location.pathname;
    const [isCollapsed, setIsCollapsed] = useState(false);

    // Mock session for local admin
    const session = {
        user: {
            name: "Admin",
            email: "admin@marubot.local",
            image: ""
        }
    };

    return (
        <aside className={cn(
            "h-screen bg-white dark:bg-slate-900 border-r flex flex-col transition-all duration-300 shadow-xl",
            isCollapsed ? "w-20" : "w-64"
        )}>
            <div className="p-6 flex items-center justify-between">
                {!isCollapsed && (
                    <div className="flex items-center gap-2">
                        <span className="text-2xl">🦞</span>
                        <span className="font-bold text-lg bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                            MaruBot
                        </span>
                    </div>
                )}
                {isCollapsed && <span className="text-2xl mx-auto">🦞</span>}
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setIsCollapsed(!isCollapsed)}
                    className="hidden md:flex ml-auto"
                >
                    {isCollapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
                </Button>
            </div>

            <nav className="flex-1 px-3 space-y-1">
                {menuItems.map((item) => (
                    <Link
                        key={item.href}
                        to={item.href}
                        className={cn(
                            "flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all group",
                            pathname === item.href
                                ? "bg-blue-600 text-white shadow-md shadow-blue-200 dark:shadow-none"
                                : "text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                        )}
                    >
                        <item.icon className={cn(
                            "w-5 h-5",
                            pathname === item.href ? "text-white" : "group-hover:text-blue-500 transition-colors"
                        )} />
                        {!isCollapsed && <span className="font-medium">{item.name}</span>}
                    </Link>
                ))}
            </nav>

            <div className="p-4 border-t space-y-4">
                {session?.user && (
                    <div className={cn(
                        "flex items-center gap-3 p-2 rounded-lg bg-slate-50 dark:bg-slate-800/50",
                        isCollapsed ? "justify-center" : ""
                    )}>
                        <Avatar className="w-8 h-8">
                            <AvatarImage src={session.user.image || ""} />
                            <AvatarFallback><UserIcon size={16} /></AvatarFallback>
                        </Avatar>
                        {!isCollapsed && (
                            <div className="flex-1 min-w-0">
                                <p className="text-xs font-semibold truncate">{session.user.name}</p>
                                <p className="text-[10px] text-slate-400 truncate">{session.user.email}</p>
                            </div>
                        )}
                        {!isCollapsed && (
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-red-500" onClick={() => console.log("Logout")}>
                                <LogOut size={14} />
                            </Button>
                        )}
                    </div>
                )}

                <div className={cn(
                    "text-[10px] text-slate-400 px-2",
                    isCollapsed ? "text-center" : "flex justify-between"
                )}>
                    {!isCollapsed && <span>Engine Status: Active</span>}
                    <span>v0.2.4</span>
                </div>
            </div>
        </aside>
    );
}
