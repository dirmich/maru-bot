import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
    Cpu,
    HardDrive,
    MemoryStick,
    Clock,
    Activity,
    Info
} from 'lucide-react';
import { useTranslation } from "@/lib/i18n";

interface SystemStats {
    uptime?: number;
    memory?: {
        total: number;
        free: number;
        available: number;
        used: number;
    };
    disk?: {
        total: number;
        free: number;
        used: number;
    };
    cpu?: {
        load1: number;
        load5: number;
        load15: number;
    };
    version: string;
    os?: string;
}

export function DashboardPage() {
    const t = useTranslation();
    const [stats, setStats] = useState<SystemStats | null>(null);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const res = await fetch('/api/system/stats');
                if (res.ok) {
                    const data = await res.json();
                    setStats(data);
                }
            } catch (e) {
                console.error("Failed to fetch stats", e);
            }
        };

        fetchStats();
        const interval = setInterval(fetchStats, 10000);
        return () => clearInterval(interval);
    }, []);

    const formatBytes = (bytes: number) => {
        if (!bytes) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatUptime = (seconds: number) => {
        if (!seconds) return '0s';
        const d = Math.floor(seconds / (3600 * 24));
        const h = Math.floor((seconds % (3600 * 24)) / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        return `${d}d ${h}h ${m}m`;
    };

    return (
        <div className="p-6 space-y-6">
            <header>
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <Activity className="text-blue-600" /> {t.dashboard}
                </h1>
                <p className="text-sm text-slate-500">System status and resource monitoring</p>
            </header>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {/* Uptime */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Uptime</CardTitle>
                        <Clock className="w-4 h-4 text-slate-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats ? formatUptime(stats.uptime || 0) : '--'}</div>
                    </CardContent>
                </Card>

                {/* CPU Load */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">CPU Load (1m)</CardTitle>
                        <Cpu className="w-4 h-4 text-slate-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.cpu?.load1?.toFixed(2) || '--'}</div>
                    </CardContent>
                </Card>

                {/* Memory */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Memory</CardTitle>
                        <MemoryStick className="w-4 h-4 text-slate-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {stats?.memory ? `${((stats.memory.used / stats.memory.total) * 100).toFixed(1)}%` : '--'}
                        </div>
                        <p className="text-xs text-slate-500">
                            {stats?.memory ? `${formatBytes(stats.memory.used)} / ${formatBytes(stats.memory.total)}` : ''}
                        </p>
                    </CardContent>
                </Card>

                {/* Disk */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Disk</CardTitle>
                        <HardDrive className="w-4 h-4 text-slate-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {stats?.disk ? `${((stats.disk.used / stats.disk.total) * 100).toFixed(1)}%` : '--'}
                        </div>
                        <p className="text-xs text-slate-500">
                            {stats?.disk ? `${formatBytes(stats.disk.used)} / ${formatBytes(stats.disk.total)}` : ''}
                        </p>
                    </CardContent>
                </Card>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg flex items-center gap-2">
                            <Info className="w-5 h-5 text-blue-500" /> System Information
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2">
                        <div className="flex justify-between border-b py-2">
                            <span className="text-slate-500">OS</span>
                            <span className="font-medium">{stats?.os || '--'}</span>
                        </div>
                        <div className="flex justify-between border-b py-2">
                            <span className="text-slate-500">Version</span>
                            <span className="font-medium">v{stats?.version || '0.3.10'}</span>
                        </div>
                        <div className="flex justify-between py-2">
                            <span className="text-slate-500">Status</span>
                            <span className="text-green-500 font-medium font-bold">ONLINE</span>
                        </div>
                    </CardContent>
                </Card>

                {/* Placeholder for more info */}
                <Card className="bg-slate-50 dark:bg-slate-800/20 border-dashed">
                    <CardContent className="flex flex-col items-center justify-center h-full py-10 text-slate-400">
                        <Activity className="w-10 h-10 mb-2 opacity-20" />
                        <p className="text-sm">More metrics coming soon...</p>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
