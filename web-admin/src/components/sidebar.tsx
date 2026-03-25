import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { logout } from "@/lib/auth";
import { Language, useLanguageStore, useTranslation } from "@/lib/i18n";
import { useSystemStore } from "@/lib/system-store";
import { cn } from "@/lib/utils";
import {
    ArrowUpCircle,
    ChevronLeft,
    ChevronRight,
    Cpu,
    Languages,
    LayoutDashboard,
    Loader2,
    LogOut,
    MessageSquare,
    Package,
    ScrollText,
    Settings,
    User as UserIcon
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { toast } from "sonner";

export function Sidebar({ className, onClose }: { className?: string; onClose?: () => void }) {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();
    const { 
        version, 
        is_update_available, 
        is_raspberry_pi, 
        latest_version, 
        fetchSystemInfo 
    } = useSystemStore();
    
    const location = useLocation();
    const pathname = location.pathname;
    const [isCollapsed, setIsCollapsed] = useState(false);
    const [isUpgrading, setIsUpgrading] = useState(false);
    const [confirmUpgradeOpen, setConfirmUpgradeOpen] = useState(false);

    useEffect(() => {
        fetchSystemInfo();
        const interval = setInterval(fetchSystemInfo, 600000); 
        return () => clearInterval(interval);
    }, [fetchSystemInfo]);

    const handleUpgrade = async () => {
        setIsUpgrading(true);
        try {
            const resp = await fetch("/api/upgrade", { method: "POST" });
            if (resp.ok) {
                toast.success(t.upgrading);
                setTimeout(() => {
                    window.location.reload();
                }, 10000);
            } else {
                toast.error("Upgrade failed to start");
                setIsUpgrading(false);
            }
        } catch (error) {
            toast.error("Network error during upgrade");
            setIsUpgrading(false);
        }
    };

    const menuItems = [
        { name: t.dashboard || "Dashboard", href: "/dashboard", icon: LayoutDashboard },
        { name: t.chat, href: "/chat", icon: MessageSquare },
        ...(is_raspberry_pi ? [{ name: t.gpio, href: "/gpio", icon: Cpu }] : []),
        { name: t.skills, href: "/skills", icon: Package },
        { name: t.settings, href: "/settings", icon: Settings },
        { name: t.logs, href: "/logs", icon: ScrollText },
    ];

    const session = {
        user: { name: "Admin", email: "admin@marubot.local", image: "" }
    };

    const languages: { code: Language; label: string }[] = [
        { code: 'en', label: 'English' },
        { code: 'ko', label: '한국어' },
        { code: 'ja', label: '日本語' },
    ];

    return (
        <aside className={cn(
            "h-screen bg-white dark:bg-slate-900 border-r flex flex-col transition-all duration-300 shadow-xl z-50",
            isCollapsed ? "w-20" : "w-64",
            className
        )}>
            <div className="p-6 flex items-center justify-between">
                {!isCollapsed && (
                    <div className="flex items-center gap-2">
                        <span className="text-2xl">🦞</span>
                        <span className="font-bold text-lg bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent italic tracking-tighter">
                            MaruBot
                        </span>
                    </div>
                )}
                {isCollapsed && <span className="text-2xl mx-auto">🦞</span>}
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setIsCollapsed(!isCollapsed)}
                    className="hidden lg:flex ml-auto text-slate-400 hover:text-blue-500 rounded-full"
                >
                    {isCollapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
                </Button>
            </div>

            <nav className="flex-1 px-3 space-y-1 overflow-y-auto">
                {menuItems.map((item) => (
                    <Link
                        key={item.href}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                            "flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all group",
                            pathname === item.href
                                ? "bg-blue-600 text-white shadow-lg shadow-blue-200 dark:shadow-none"
                                : "text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                        )}
                    >
                        <item.icon className={cn(
                            "w-5 h-5",
                            pathname === item.href ? "text-white" : "group-hover:text-blue-500 transition-colors"
                        )} />
                        {!isCollapsed && <span className="font-bold text-sm">{item.name}</span>}
                    </Link>
                ))}
            </nav>

            <div className="p-4 border-t space-y-4">
                {is_update_available && (
                    <div className={cn("px-2", isCollapsed ? "flex justify-center" : "")}>
                        <Button
                            variant="secondary"
                            size={isCollapsed ? "icon" : "sm"}
                            disabled={isUpgrading}
                            onClick={() => setConfirmUpgradeOpen(true)}
                            className={cn(
                                "w-full bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 border-none rounded-xl",
                                isCollapsed ? "h-9 w-9 p-0" : "h-9 justify-start gap-3 px-3"
                            )}
                        >
                            {isUpgrading ? (
                                <Loader2 size={18} className="animate-spin" />
                            ) : (
                                <ArrowUpCircle size={18} className="animate-bounce" />
                            )}
                            {!isCollapsed && <span className="text-xs font-black uppercase tracking-tight">{t.upgrade_button}</span>}
                        </Button>
                    </div>
                )}

                <div className={cn("px-2", isCollapsed ? "flex justify-center" : "")}>
                    <Select value={language} onValueChange={(v) => { setLanguage(v as Language); onClose?.(); }}>
                        <SelectTrigger className={cn("h-9 border-none bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors", isCollapsed ? "px-0 w-9 justify-center" : "w-full justify-start gap-3")}>
                            <Languages size={18} className="text-slate-500 font-black" />
                            {!isCollapsed && <SelectValue className="text-xs font-bold" />}
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
                        "flex items-center gap-3 p-2 rounded-2xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-700",
                        isCollapsed ? "justify-center" : ""
                    )}>
                        <Avatar className="w-8 h-8 ring-2 ring-white dark:ring-slate-700">
                            <AvatarImage src={session.user.image || ""} />
                            <AvatarFallback><UserIcon size={16} /></AvatarFallback>
                        </Avatar>
                        {!isCollapsed && (
                            <div className="flex-1 min-w-0">
                                <p className="text-[10px] font-black truncate text-slate-900 dark:text-slate-100 leading-tight">{session.user.name}</p>
                                <p className="text-[9px] text-slate-400 truncate leading-tight font-medium">{session.user.email}</p>
                            </div>
                        )}
                        {!isCollapsed && (
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-full" onClick={logout}>
                                <LogOut size={14} />
                            </Button>
                        )}
                    </div>
                )}

                <div className={cn(
                    "text-[9px] text-slate-400 px-2 font-black uppercase tracking-widest",
                    isCollapsed ? "text-center" : "flex justify-between"
                )}>
                    {!isCollapsed && <span>{t.status_ok}</span>}
                    <span className={cn(is_update_available && "text-indigo-500")}>
                        {version}
                    </span>
                </div>
            </div>

            <ConfirmDialog
                open={confirmUpgradeOpen}
                onOpenChange={setConfirmUpgradeOpen}
                title={t.upgrade_confirm}
                description={t.upgrade_available + ": " + latest_version}
                onConfirm={handleUpgrade}
            />
        </aside>
    );
}
