import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Terminal, RefreshCw, ScrollText } from 'lucide-react';
import { useTranslation } from "@/lib/i18n";
import { authenticatedFetch } from "@/lib/auth";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

export function LogsPage() {
    const t = useTranslation();
    const [logs, setLogs] = useState<string>('Loading logs...');
    const [logFiles, setLogFiles] = useState<string[]>([]);
    const [selectedFile, setSelectedFile] = useState<string>('');

    useEffect(() => {
        fetchLogList();
    }, []);

    useEffect(() => {
        fetchLogs();
        const interval = setInterval(fetchLogs, 5000);
        return () => clearInterval(interval);
    }, [selectedFile]);

    const fetchLogList = async () => {
        try {
            const res = await authenticatedFetch('/api/logs/list');
            if (res.ok) {
                const data = await res.json();
                setLogFiles(data.files || []);
                if (data.files && data.files.length > 0 && !selectedFile) {
                    setSelectedFile(data.files[0]);
                }
            }
        } catch (e) {
            console.error('Failed to fetch log list:', e);
        }
    };

    const fetchLogs = async () => {
        try {
            const url = selectedFile ? `/api/logs?file=${selectedFile}` : '/api/logs';
            const res = await authenticatedFetch(url);
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
                <CardHeader className="py-3 px-6 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between flex-none">
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-blue-500" /> 
                        {selectedFile || 'latest.log'}
                    </CardTitle>
                    <div className="flex items-center gap-3">
                        <span className="text-xs text-slate-500 font-normal">Select Log:</span>
                        <Select value={selectedFile} onValueChange={setSelectedFile}>
                            <SelectTrigger className="w-[240px] h-8 text-xs bg-slate-50 dark:bg-slate-800 border-none ring-1 ring-slate-200 dark:ring-slate-700">
                                <SelectValue placeholder="Select log file" />
                            </SelectTrigger>
                            <SelectContent>
                                {logFiles.map(file => (
                                    <SelectItem key={file} value={file} className="text-xs">
                                        {file}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden bg-slate-950 text-slate-300 font-mono text-[11px] relative">
                    <ScrollArea className="h-full w-full">
                        <pre className="p-6 whitespace-pre-wrap leading-relaxed">
                            {logs}
                        </pre>
                    </ScrollArea>
                </CardContent>
                <CardFooter className="p-3 border-t bg-slate-900 justify-between flex-none">
                    <span className="text-blue-400 font-bold text-[10px]">LOG PATH: ~/.marubot/logs/{selectedFile || ''}</span>
                    <span className="text-slate-500 text-[10px]">LIVE UPDATING</span>
                </CardFooter>
            </Card>
        </div>
    );
}
