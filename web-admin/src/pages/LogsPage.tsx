import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Terminal, RefreshCw, ScrollText } from 'lucide-react';
import { useTranslation } from "@/lib/i18n";
import { authenticatedFetch } from "@/lib/auth";

export function LogsPage() {
    const t = useTranslation();
    const [logs, setLogs] = useState<string>('Loading logs...');

    useEffect(() => {
        fetchLogs();
        const interval = setInterval(fetchLogs, 5000); // Poll every 5 seconds
        return () => clearInterval(interval);
    }, []);

    const fetchLogs = async () => {
        try {
            const res = await authenticatedFetch('/api/logs');
            if (res.ok) {
                const data = await res.json();
                setLogs(data.logs || 'No logs available.');
            }
        } catch (e) {
            setLogs(`Failed to fetch logs: ${e}`);
        }
    };

    return (
        <div className="p-6 h-screen flex flex-col space-y-6 overflow-hidden">
            <header className="flex-none flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <ScrollText className="text-blue-600" /> {t.logs}
                    </h1>
                    <p className="text-sm text-slate-500">{t.logs_desc}</p>
                </div>
                <Button variant="outline" size="sm" onClick={fetchLogs}>
                    <RefreshCw className="w-4 h-4 mr-2" /> {t.refresh}
                </Button>
            </header>

            <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden min-h-0">
                <CardHeader className="py-4 px-6 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between flex-none">
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-blue-500" /> dashboard.log
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden bg-slate-950 text-slate-300 font-mono text-[11px] relative">
                    <ScrollArea className="h-full w-full">
                        <pre className="p-6 whitespace-pre-wrap leading-relaxed">
                            {logs}
                        </pre>
                    </ScrollArea>
                </CardContent>
                <CardFooter className="p-3 border-t bg-slate-900 justify-between flex-none">
                    <span className="text-blue-400 font-bold">LOG PATH: ~/.marubot/dashboard.log</span>
                    <span className="text-slate-500 text-[10px]">LIVE UPDATING</span>
                </CardFooter>
            </Card>
        </div>
    );
}
