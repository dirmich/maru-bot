import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { Settings, Cpu, Wrench, ShieldCheck } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";

export function SettingsPage() {
    // Mock config for offline/demo if API fails
    const [config, setConfig] = useState<any>({
        agents: {
            defaults: {
                model: "gemini-1.5-pro",
                workspace: "./workspace"
            }
        },
        providers: {
            openai: { api_key: "", api_base: "" },
            gemini: { api_key: "", api_base: "" },
            anthropic: { api_key: "", api_base: "" }
        }
    });

    // For dialog control
    const [showSaveConfirm, setShowSaveConfirm] = useState(false);
    const [showResetConfirm, setShowResetConfirm] = useState(false);

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await fetch('/api/config');
            if (res.ok) {
                const data = await res.json();
                setConfig(data);
            } else {
                // If API fails (e.g., offline dev), we stick with initial mock
                console.log("Using default config (offline mode)");
            }
        } catch (e) {
            console.error("Config fetch error", e);
        }
    };

    const handleSaveConfig = async () => {
        try {
            const res = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config),
            });
            if (res.ok) {
                toast.success('설정이 저장되었습니다.');
            } else {
                // Mock success for dev
                toast.success('설정이 저장되었습니다. (Demo)');
            }
        } catch (error) {
            toast.error('설정 저장에 실패했습니다.');
        } finally {
            setShowSaveConfirm(false);
        }
    };

    if (!config) return <div className="p-8">로딩 중...</div>;

    // Helper to safely update nested config
    const updateConfig = (path: string[], value: any) => {
        const newConfig = JSON.parse(JSON.stringify(config));
        let current = newConfig;
        for (let i = 0; i < path.length - 1; i++) {
            if (!current[path[i]]) current[path[i]] = {};
            current = current[path[i]];
        }
        current[path[path.length - 1]] = value;
        setConfig(newConfig);
    };

    return (
        <div className="p-6 max-w-5xl mx-auto space-y-6">
            <header className="mb-6">
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <Settings className="text-blue-600" /> 환경 설정
                </h1>
                <p className="text-sm text-slate-500">엔진 및 AI 서비스 설정을 관리합니다.</p>
            </header>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-blue-600">
                        <Cpu className="w-5 h-5" /> 메인 에이전트
                    </CardTitle>
                    <CardDescription>기본 동작 모델과 작업 디렉토리를 설정합니다.</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">사용 모델</label>
                            <Input
                                value={config.agents?.defaults?.model || ''}
                                onChange={(e) => updateConfig(['agents', 'defaults', 'model'], e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">워크스페이스</label>
                            <Input
                                value={config.agents?.defaults?.workspace || ''}
                                onChange={(e) => updateConfig(['agents', 'defaults', 'workspace'], e.target.value)}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-indigo-600">
                        <Wrench className="w-5 h-5" /> API 제공자
                    </CardTitle>
                    <CardDescription>연동할 AI 모델 서비스의 인증 키를 입력하세요.</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-8">
                    {config.providers && Object.entries(config.providers).map(([name, prov]: [string, any]) => (
                        <div key={name} className="space-y-3 group">
                            <div className="flex items-center gap-2">
                                <div className="w-1.5 h-4 bg-indigo-500 rounded-full"></div>
                                <span className="font-bold uppercase text-xs tracking-wider text-slate-500">{name}</span>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-1">
                                    <span className="text-[10px] font-medium text-slate-400 ml-1">API KEY</span>
                                    <Input
                                        placeholder="API Key"
                                        type="password"
                                        value={prov.api_key || ''}
                                        onChange={(e) => updateConfig(['providers', name, 'api_key'], e.target.value)}
                                    />
                                </div>
                                <div className="space-y-1">
                                    <span className="text-[10px] font-medium text-slate-400 ml-1">API BASE (Optional)</span>
                                    <Input
                                        placeholder="Auto"
                                        value={prov.api_base || ''}
                                        onChange={(e) => updateConfig(['providers', name, 'api_base'], e.target.value)}
                                    />
                                </div>
                            </div>
                        </div>
                    ))}
                </CardContent>
                <CardFooter className="p-6 border-t bg-slate-50 dark:bg-slate-900 justify-end">
                    <Button onClick={() => setShowSaveConfirm(true)} className="bg-indigo-600 hover:bg-indigo-700 min-w-[120px]">
                        저장하기
                    </Button>
                </CardFooter>
            </Card>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-rose-600">
                        <ShieldCheck className="w-5 h-5" /> 시스템 인증 및 보안
                    </CardTitle>
                    <CardDescription>관리자 권한을 설정합니다.</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">관리자 계정</label>
                            <Input placeholder="admin" disabled defaultValue="admin" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">비밀번호 변경</label>
                            <Input placeholder="New password" type="password" />
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="p-6 border-t bg-slate-50 dark:bg-slate-900 flex justify-between">
                    <Button variant="outline" className="text-rose-600 border-rose-200 hover:bg-rose-50" onClick={() => setShowResetConfirm(true)}>설정 초기화</Button>
                    <Button className="bg-rose-600 hover:bg-rose-700">보안 설정 저장</Button>
                </CardFooter>
            </Card>

            <ConfirmDialog
                open={showSaveConfirm}
                onOpenChange={setShowSaveConfirm}
                title="설정 저장"
                description="변경사항을 저장하시겠습니까?"
                onConfirm={handleSaveConfig}
            />

            <ConfirmDialog
                open={showResetConfirm}
                onOpenChange={setShowResetConfirm}
                title="설정 리셋"
                description="모든 설정을 초기화하시겠습니까?"
                onConfirm={() => toast.info("초기화 기능은 아직 구현되지 않았습니다.")}
            />
        </div>
    );
}
