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
    Languages,
    ScrollText,
    LayoutDashboard
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useState } from "react";
import { useTranslation, useLanguageStore, Language } from "@/lib/i18n";
import { logout } from "@/lib/auth";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

export function Sidebar() {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();
    const location = useLocation();
    const pathname = location.pathname;
    const [isCollapsed, setIsCollapsed] = useState(false);

    const menuItems = [
        { name: t.dashboard || "Dashboard", href: "/dashboard", icon: LayoutDashboard },
        { name: t.chat, href: "/chat", icon: MessageSquare },
        { name: t.gpio, href: "/gpio", icon: Cpu },
        { name: t.skills, href: "/skills", icon: Package },
        { name: t.settings, href: "/settings", icon: Settings },
        { name: t.logs, href: "/logs", icon: ScrollText },
    ];

    // Mock session for local admin
    const session = {
        user: {
            name: "Admin",
            email: "admin@marubot.local",
            image: ""
        }
    };

    const languages: { code: Language; label: string }[] = [
        { code: 'en', label: 'English' },
        { code: 'ko', label: '한국어' },
        { code: 'ja', label: '日本語' },
    ];

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
                {/* Language Switcher */}
                <div className={cn("px-2", isCollapsed ? "flex justify-center" : "")}>
                    <Select value={language} onValueChange={(v) => setLanguage(v as Language)}>
                        <SelectTrigger className={cn("h-9 border-none bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors", isCollapsed ? "px-0 w-9 justify-center" : "w-full justify-start gap-3")}>
                            <Languages size={18} className="text-slate-500" />
                            {!isCollapsed && <SelectValue className="text-xs" />}
                        </SelectTrigger>
                        <SelectContent>
                            {languages.map((lang) => (
                                <SelectItem key={lang.code} value={lang.code}>
                                    {lang.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

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
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-red-500" onClick={logout}>
                                <LogOut size={14} />
                            </Button>
                        )}
                    </div>
                )}

                <div className={cn(
                    "text-[10px] text-slate-400 px-2",
                    isCollapsed ? "text-center" : "flex justify-between"
                )}>
                    {!isCollapsed && <span>{t.status_ok}</span>}
                    <span>v0.3.12</span>
                </div>
            </div>
        </aside>
    );
}
