
'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from 'sonner';
import { Package, Plus, Terminal, RefreshCw } from 'lucide-react';
import { useConfirmDialog } from '@/components/ui-custom-dialog';

export default function SkillsPage() {
    const [skills, setSkills] = useState<string>('');
    const confirm = useConfirmDialog();

    useEffect(() => {
        fetchSkills();
    }, []);

    const fetchSkills = async () => {
        const res = await fetch('/api/skills');
        const data = await res.json();
        setSkills(data.output || 'Skills list empty or error');
    };

    const handleSkillAction = async (action: string, skill: string) => {
        const actionKR = action === 'install' ? '설치' : '삭제';
        confirm.show(
            `툴/스킬 ${actionKR}`,
            `[${skill}]을(를) ${actionKR}하시겠습니까?`,
            async () => {
                toast.info(`${skill} ${actionKR} 중...`);
                try {
                    const res = await fetch('/api/skills', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ action, skill }),
                    });
                    const data = await res.json();
                    toast.success(`${skill} ${actionKR} 완료`);
                    fetchSkills();
                } catch (error) {
                    toast.error(`${skill} ${actionKR} 실패`);
                }
            }
        );
    };

    return (
        <div className="p-6 h-full flex flex-col space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <Package className="text-emerald-600" /> 스킬 & 툴 박스
                    </h1>
                    <p className="text-sm text-slate-500">에이전트의 기능을 확장하는 도구를 관리합니다.</p>
                </div>
                <Button variant="outline" size="sm" onClick={fetchSkills}>
                    <RefreshCw className="w-4 h-4 mr-2" /> 새로고침
                </Button>
            </header>

            <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden">
                <CardHeader className="py-4 px-6 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between">
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-emerald-500" /> CLI 출력
                    </CardTitle>
                    <div className="flex gap-2">
                        <Input id="skillInstall" placeholder="GitHub user/repo" className="h-9 w-64 text-sm" />
                        <Button size="sm" onClick={() => {
                            const el = document.getElementById('skillInstall') as HTMLInputElement;
                            if (el.value) handleSkillAction('install', el.value);
                        }} className="bg-emerald-600 hover:bg-emerald-700">
                            <Plus className="w-4 h-4 mr-1" /> 설치
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden bg-slate-950 text-emerald-400 font-mono text-xs">
                    <ScrollArea className="h-full">
                        <pre className="p-6 whitespace-pre-wrap leading-relaxed">{skills}</pre>
                    </ScrollArea>
                </CardContent>
                <CardFooter className="p-3 border-t bg-slate-900 text-[10px] text-slate-500 justify-between">
                    <span>marubot skills list</span>
                    <span>SYSTEM READY</span>
                </CardFooter>
            </Card>
        </div>
    );
}
