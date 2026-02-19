import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from 'sonner';
import { Package, Plus, Terminal, RefreshCw } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation } from "@/lib/i18n";

export function SkillsPage() {
    const t = useTranslation();
    const [skillsOutput, setSkillsOutput] = useState<string>('');
    const [installInput, setInstallInput] = useState('');
    const [confirmState, setConfirmState] = useState<{ open: boolean, title: string, desc: string, action: () => Promise<void> }>({
        open: false, title: '', desc: '', action: async () => { }
    });

    useEffect(() => {
        fetchSkills();
    }, []);

    const fetchSkills = async () => {
        try {
            const res = await fetch('/api/skills');
            if (res.ok) {
                const data = await res.json();
                setSkillsOutput(data.output || t.skills_empty);
            } else {
                setSkillsOutput(`(Demo Mode)\n\nINSTALLED SKILLS:\n-----------------\n✓ weather (sipeed/marubot-skills/weather)\n✓ news (sipeed/marubot-skills/news)\n\nAvailable builtin: calculator, stock`);
            }
        } catch (e) {
            setSkillsOutput(`Connection failed. Error: ${e}`);
        }
    };

    const handleActionRequest = (action: string, skill: string) => {
        if (!skill) return;

        const actionName = action === 'install' ? t.skills_install : t.skills_uninstall;
        setConfirmState({
            open: true,
            title: `${t.skills_title} ${actionName}`,
            desc: `[${skill}] ${actionName}?`,
            action: async () => {
                toast.info(`${skill} ${actionName}...`);
                try {
                    const res = await fetch('/api/skills', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ action, skill }),
                    });
                    if (res.ok) {
                        toast.success(`${skill} ${actionName} OK`);
                        fetchSkills();
                        setInstallInput('');
                    } else {
                        toast.success(`(Demo) ${skill} ${actionName} OK`);
                    }
                } catch (error) {
                    toast.error(`${skill} ${actionName} Failed`);
                } finally {
                    setConfirmState(prev => ({ ...prev, open: false }));
                }
            }
        });
    };

    return (
        <div className="p-6 h-screen flex flex-col space-y-6 overflow-hidden">
            <header className="flex-none flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <Package className="text-emerald-600" /> {t.skills_title}
                    </h1>
                    <p className="text-sm text-slate-500">{t.skills_desc}</p>
                </div>
                <Button variant="outline" size="sm" onClick={fetchSkills}>
                    <RefreshCw className="w-4 h-4 mr-2" /> {t.refresh}
                </Button>
            </header>

            <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden min-h-0">
                <CardHeader className="py-4 px-6 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between flex-none">
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-emerald-500" /> {t.skills_cli_output}
                    </CardTitle>
                    <div className="flex gap-2">
                        <Input
                            value={installInput}
                            onChange={(e) => setInstallInput(e.target.value)}
                            placeholder="GitHub user/repo"
                            className="h-9 w-64 text-sm"
                        />
                        <Button size="sm" onClick={() => handleActionRequest('install', installInput)} className="bg-emerald-600 hover:bg-emerald-700 text-white">
                            <Plus className="w-4 h-4 mr-1" /> {t.skills_install}
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden bg-slate-950 text-emerald-400 font-mono text-xs relative">
                    <ScrollArea className="h-full w-full">
                        <pre className="p-6 whitespace-pre-wrap leading-relaxed">{skillsOutput}</pre>
                    </ScrollArea>
                </CardContent>
                <CardFooter className="p-3 border-t bg-slate-900 text-[10px] text-slate-500 justify-between flex-none">
                    <span>marubot skills list</span>
                    <span>SYSTEM READY</span>
                </CardFooter>
            </Card>

            <ConfirmDialog
                open={confirmState.open}
                onOpenChange={(open: boolean) => setConfirmState(prev => ({ ...prev, open }))}
                title={confirmState.title}
                description={confirmState.desc}
                onConfirm={confirmState.action}
            />
        </div>
    );
}
