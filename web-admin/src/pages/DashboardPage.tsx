import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from "@/components/ui/progress";
import { 
    Cpu, 
    HardDrive, 
    Zap, 
    Activity, 
    RefreshCcw, 
    AlertTriangle, 
    CheckCircle,
    Info,
    Database,
    Binary,
    Server,
    Layout,
    Clock
} from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from "@/lib/i18n";
import { authenticatedFetch } from "@/lib/auth";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import { ScrollArea } from '@/components/ui/scroll-area';

interface SystemStats {
    cpu: {
        load1: number;
        load5: number;
        load15: number;
    };
    memory: {
        total: number;
        used: number;
        free: number;
        available: number;
    };
    disk: {
        total: number;
        used: number;
        free: number;
    };
    os: string;
    version: string;
    latest_version: string;
    is_update_available: boolean;
    uptime: number;
    is_ai_configured: boolean;
    is_channel_configured: boolean;
    is_rpi: boolean;
    
    // Detailed stats
    cpu_detail?: {
        model: string;
        cores: number;
        vendor: string;
        mhz: number;
        percent: number[];
    };
    memory_detail?: {
        total: number;
        available: number;
        used: number;
        free: number;
        cached: number;
        percent: number;
    };
    disk_detail?: {
        device: string;
        mountpoint: string;
        fstype: string;
        total: number;
        free: number;
        used: number;
        percent: number;
    }[];
}

export function DashboardPage() {
    const t = useTranslation();
    const [stats, setStats] = useState<SystemStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [activeDetail, setActiveDetail] = useState<'cpu' | 'memory' | 'disk' | null>(null);

    const fetchStats = async () => {
        try {
            const res = await authenticatedFetch('/api/system/stats');
            if (res.ok) {
                const data = await res.json();
                setStats(data);
            }
        } catch (error) {
            console.error('Failed to fetch stats:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchStats();
        const interval = setInterval(fetchStats, 5000);
        return () => clearInterval(interval);
    }, []);

    const formatBytes = (bytes: number) => {
        if (!bytes || bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatUptime = (seconds: number) => {
        if (!seconds) return '0s';
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const mins = Math.floor((seconds % 3600) / 60);
        
        let res = '';
        if (days > 0) res += `${days}d `;
        if (hours > 0) res += `${hours}h `;
        res += `${mins}m`;
        return res;
    };

    const getCpuUsage = () => {
        if (!stats) return 0;
        return Math.round(stats.cpu.load1);
    };

    const getMemoryPercent = () => {
        if (!stats) return 0;
        return Math.round((stats.memory.used / stats.memory.total) * 100);
    };

    const getDiskPercent = () => {
        if (!stats) return 0;
        return Math.round((stats.disk.used / stats.disk.total) * 100);
    };

    return (
        <div className="p-4 md:p-6 space-y-6 bg-slate-50 dark:bg-slate-950 min-h-screen">
            <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-2">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                        <Layout className="text-blue-600" /> {t.dashboard_title || "Dashboard"}
                    </h1>
                    <p className="text-sm text-slate-500">{t.dashboard_desc || "System status and resource monitoring"}</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => {
                            setIsLoading(true);
                            fetchStats();
                        }}
                        className="rounded-xl"
                    >
                        <RefreshCcw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                        {t.refresh || "Refresh"}
                    </Button>
                </div>
            </header>

            {/* Quick Status Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="border-none shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-800">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">{t.os_version || "System OS"}</CardTitle>
                        <Activity className="h-4 w-4 text-blue-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold truncate max-w-full">{stats?.os || '---'}</div>
                        <p className="text-xs text-slate-500 mt-1">MaruBot v{stats?.version}</p>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-800">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">{t.uptime || "Uptime"}</CardTitle>
                        <Clock className="h-4 w-4 text-amber-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats ? formatUptime(stats.uptime) : '---'}</div>
                        <p className="text-xs text-slate-500 mt-1">System running since boot</p>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-800">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">{t.ai_status || "AI Service"}</CardTitle>
                        {stats?.is_ai_configured ? <CheckCircle className="h-4 w-4 text-green-500" /> : <AlertTriangle className="h-4 w-4 text-red-500" />}
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.is_ai_configured ? t.active || "Active" : t.inactive || "Inactive"}</div>
                        <p className="text-xs text-slate-500 mt-1">{stats?.is_ai_configured ? 'Provider connected' : 'Check configuration'}</p>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-800">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">{t.channel_status || "Channels"}</CardTitle>
                        {stats?.is_channel_configured ? <CheckCircle className="h-4 w-4 text-green-500" /> : <AlertTriangle className="h-4 w-4 text-slate-300" />}
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.is_channel_configured ? t.active || "Active" : t.inactive || "Inactive"}</div>
                        <p className="text-xs text-slate-500 mt-1">{stats?.is_channel_configured ? 'Service monitoring' : 'None enabled'}</p>
                    </CardContent>
                </Card>
            </div>

            {/* Performance Cards - Clickable */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card 
                    className="border-none shadow-md ring-1 ring-slate-900/5 hover:ring-blue-500/30 transition-all cursor-pointer group"
                    onClick={() => setActiveDetail('cpu')}
                >
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-base font-semibold flex items-center gap-2">
                            <Cpu className="w-5 h-5 text-blue-600" /> CPU Load
                        </CardTitle>
                        <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100"><Info className="w-4 h-4" /></Button>
                    </CardHeader>
                    <CardContent className="pt-2">
                        <div className="flex items-end justify-between mb-2">
                            <div className="text-3xl font-black">{getCpuUsage()}%</div>
                            <div className="text-xs text-slate-500 px-2 py-1 bg-slate-100 dark:bg-slate-800 rounded-lg">Real-time</div>
                        </div>
                        <Progress value={getCpuUsage()} className="h-3" indicatorClassName="bg-blue-600" />
                        <div className="flex justify-between mt-4 text-[10px] text-slate-400 font-medium uppercase tracking-wider">
                            <span>{stats?.cpu.load5}% (5m)</span>
                            <span>{stats?.cpu.load15}% (15m)</span>
                        </div>
                    </CardContent>
                </Card>

                <Card 
                    className="border-none shadow-md ring-1 ring-slate-900/5 hover:ring-purple-500/30 transition-all cursor-pointer group"
                    onClick={() => setActiveDetail('memory')}
                >
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-base font-semibold flex items-center gap-2">
                            <Binary className="w-5 h-5 text-purple-600" /> Memory
                        </CardTitle>
                        <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100"><Info className="w-4 h-4" /></Button>
                    </CardHeader>
                    <CardContent className="pt-2">
                        <div className="flex items-end justify-between mb-2">
                            <div className="text-3xl font-black">{getMemoryPercent()}%</div>
                            <div className="text-xs text-slate-500 px-2 py-1 bg-slate-100 dark:bg-slate-800 rounded-lg">
                                {stats ? formatBytes(stats.memory.used) : '--'}
                            </div>
                        </div>
                        <Progress value={getMemoryPercent()} className="h-3" indicatorClassName="bg-purple-600" />
                        <div className="flex justify-between mt-4 text-[10px] text-slate-400 font-medium uppercase tracking-wider">
                            <span>Total: {stats ? formatBytes(stats.memory.total) : '--'}</span>
                        </div>
                    </CardContent>
                </Card>

                <Card 
                    className="border-none shadow-md ring-1 ring-slate-900/5 hover:ring-amber-500/30 transition-all cursor-pointer group"
                    onClick={() => setActiveDetail('disk')}
                >
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-base font-semibold flex items-center gap-2">
                            <HardDrive className="w-5 h-5 text-amber-600" /> Disk Storage
                        </CardTitle>
                        <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100"><Info className="w-4 h-4" /></Button>
                    </CardHeader>
                    <CardContent className="pt-2">
                        <div className="flex items-end justify-between mb-2">
                            <div className="text-3xl font-black">{getDiskPercent()}%</div>
                            <div className="text-xs text-slate-500 px-2 py-1 bg-slate-100 dark:bg-slate-800 rounded-lg">
                                {stats ? formatBytes(stats.disk.used) : '--'}
                            </div>
                        </div>
                        <Progress value={getDiskPercent()} className="h-3" indicatorClassName="bg-amber-600" />
                        <div className="flex justify-between mt-4 text-[10px] text-slate-400 font-medium uppercase tracking-wider">
                            <span>Total: {stats ? formatBytes(stats.disk.total) : '--'}</span>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Platform Feature Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                 <Card className="border-none shadow-sm ring-1 ring-slate-900/5 overflow-hidden">
                    <CardHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                        <CardTitle className="text-sm font-bold uppercase tracking-wider text-slate-500 flex items-center gap-2">
                            <Server className="w-4 h-4" /> System Environment
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-0">
                        <div className="divide-y dark:divide-slate-800">
                            {[
                                { label: 'Platform', value: stats?.os || 'Unknown' },
                                { label: 'Raspberry Pi', value: stats?.is_rpi ? 'Yes' : 'No' },
                                { label: 'Architecture', value: stats?.os.includes('Windows') ? 'AMD64/x86' : 'ARM/Linux' },
                                { label: 'Service Version', value: stats?.version || '0.0.0' }
                            ].map((item, i) => (
                                <div key={i} className="px-4 py-3 flex justify-between text-sm">
                                    <span className="text-slate-500">{item.label}</span>
                                    <span className="font-semibold">{item.value}</span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-sm ring-1 ring-slate-900/5 overflow-hidden">
                    <CardHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                        <CardTitle className="text-sm font-bold uppercase tracking-wider text-slate-500 flex items-center gap-2">
                           <Database className="w-4 h-4" /> AI Capability
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-0">
                         <div className="divide-y dark:divide-slate-800">
                             {[
                                { label: 'AI Ready', value: stats?.is_ai_configured ? 'Yes' : 'No' },
                                { label: 'Active Channels', value: stats?.is_channel_configured ? 'Enabled' : 'Disabled' },
                                { label: 'Skill Load', value: 'Dynamic' },
                                { label: 'Local API', value: 'Running' }
                            ].map((item, i) => (
                                <div key={i} className="px-4 py-3 flex justify-between text-sm">
                                    <span className="text-slate-500">{item.label}</span>
                                    <span className="font-semibold">{item.value}</span>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Detail Dialogs */}
            
            {/* CPU Detail */}
            <Dialog open={activeDetail === 'cpu'} onOpenChange={(open) => !open && setActiveDetail(null)}>
                <DialogContent className="sm:max-w-md rounded-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <Cpu className="text-blue-600" /> CPU Detailed Info
                        </DialogTitle>
                        <DialogDescription>
                            Hardware specifications and core utilization.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-2">
                        <div className="bg-slate-50 dark:bg-slate-900 p-3 rounded-xl border">
                             <div className="text-[10px] text-slate-400 font-bold uppercase mb-1">Processor Model</div>
                             <div className="text-sm font-bold">{stats?.cpu_detail?.model || "Standard Processor"}</div>
                             <div className="text-xs text-slate-500 mt-1">{stats?.cpu_detail?.vendor} {stats?.cpu_detail?.mhz}MHz</div>
                        </div>
                        <div className="grid grid-cols-2 gap-3">
                             <div className="p-3 bg-slate-50 dark:bg-slate-900 rounded-xl border text-center">
                                 <div className="text-[10px] text-slate-400 font-bold uppercase">Cores</div>
                                 <div className="text-xl font-bold">{stats?.cpu_detail?.cores || 1}</div>
                             </div>
                             <div className="p-3 bg-slate-50 dark:bg-slate-900 rounded-xl border text-center">
                                 <div className="text-[10px] text-slate-400 font-bold uppercase">Usage</div>
                                 <div className="text-xl font-bold">{getCpuUsage()}%</div>
                             </div>
                        </div>
                        {stats?.cpu_detail?.percent && (
                            <div className="space-y-2">
                                <div className="text-[10px] text-slate-400 font-bold uppercase">Core Distribution</div>
                                <div className="grid grid-cols-2 gap-2">
                                    {stats.cpu_detail.percent.map((p, i) => (
                                        <div key={i} className="space-y-1">
                                            <div className="flex justify-between text-[10px]">
                                                <span>Core {i}</span>
                                                <span>{Math.round(p)}%</span>
                                            </div>
                                            <Progress value={p} className="h-1.5" />
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </DialogContent>
            </Dialog>

            {/* Memory Detail */}
            <Dialog open={activeDetail === 'memory'} onOpenChange={(open) => !open && setActiveDetail(null)}>
                <DialogContent className="sm:max-w-md rounded-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <Binary className="text-purple-600" /> Memory Statistics
                        </DialogTitle>
                        <DialogDescription>
                            Detailed breakdown of temporary and virtual memory.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-2">
                        <div className="flex items-center justify-center p-6">
                             <div className="relative w-32 h-32 flex items-center justify-center">
                                <div className="text-2xl font-black">{getMemoryPercent()}%</div>
                                <svg className="absolute w-full h-full -rotate-90">
                                    <circle cx="64" cy="64" r="58" fill="none" stroke="currentColor" strokeWidth="8" className="text-slate-100 dark:text-slate-800" />
                                    <circle cx="64" cy="64" r="58" fill="none" stroke="currentColor" strokeWidth="8" strokeDasharray={364} strokeDashoffset={364 - (364 * getMemoryPercent() / 100)} className="text-purple-600 transition-all duration-1000" />
                                </svg>
                             </div>
                        </div>
                        <div className="space-y-2">
                            {[
                                { label: 'Total Physical', value: stats ? formatBytes(stats.memory.total) : '0' },
                                { label: 'Used', value: stats ? formatBytes(stats.memory.used) : '0', color: 'text-purple-600' },
                                { label: 'Available', value: stats ? formatBytes(stats.memory.available) : '0', color: 'text-green-600' },
                                { label: 'Cached/Buffers', value: stats?.memory_detail ? formatBytes(stats.memory_detail.cached as number) : '---' }
                            ].map((item, i) => (
                                <div key={i} className="flex justify-between items-center px-2 py-1.5 border-b last:border-0">
                                    <span className="text-sm text-slate-500 font-medium">{item.label}</span>
                                    <span className={`text-sm font-bold ${item.color || ''}`}>{item.value}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Disk Detail */}
            <Dialog open={activeDetail === 'disk'} onOpenChange={(open) => !open && setActiveDetail(null)}>
                <DialogContent className="sm:max-w-lg rounded-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <HardDrive className="text-amber-600" /> Disk Partitions
                        </DialogTitle>
                        <DialogDescription>
                            All mounted storage devices and utilization.
                        </DialogDescription>
                    </DialogHeader>
                    <ScrollArea className="max-h-[300px]">
                        <div className="space-y-3 py-2 pr-4">
                            {stats?.disk_detail?.map((disk, i) => (
                                <div key={i} className="p-3 bg-slate-50 dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800">
                                    <div className="flex justify-between items-start mb-2">
                                        <div>
                                            <div className="text-sm font-black">{disk.mountpoint}</div>
                                            <div className="text-[10px] text-slate-500">{disk.device} ({disk.fstype})</div>
                                        </div>
                                        <div className="text-xs font-bold bg-white dark:bg-black px-2 py-1 rounded-md border shadow-sm">
                                            {Math.round(disk.percent)}%
                                        </div>
                                    </div>
                                    <Progress value={disk.percent} className="h-2 mb-2" indicatorClassName={disk.percent > 90 ? 'bg-red-500' : 'bg-amber-500'} />
                                    <div className="flex justify-between text-[10px] text-slate-400">
                                        <span>Used: {formatBytes(disk.used)}</span>
                                        <span>Free: {formatBytes(disk.free)}</span>
                                        <span>Total: {formatBytes(disk.total)}</span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </ScrollArea>
                </DialogContent>
            </Dialog>
        </div>
    );
}
