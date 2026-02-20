import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from 'sonner';
import { Settings, Cpu, Wrench, ShieldCheck, Languages } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation, useLanguageStore, Language } from "@/lib/i18n";

export function SettingsPage() {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();

    // Mock config for offline/demo if API fails
    const [config, setConfig] = useState<any>({
        language: "ko",
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
                // Also sync i18n store if needed, but localstorage is primary for browser
                if (data.language && data.language !== language) {
                    // console.log("Backend language choice:", data.language);
                }
            } else {
                console.log("Using default config (offline mode)");
            }
        } catch (e) {
            console.error("Config fetch error", e);
        }
    };

    const handleSaveConfig = async () => {
        try {
            // Include current UI language in config to save to backend
            const updatedConfig = { ...config, language: language };

            const res = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(updatedConfig),
            });
            if (res.ok) {
                toast.success(t.settings_save_success);
            } else {
                toast.success(t.settings_save_success + ' (Demo)');
            }
        } catch (error) {
            toast.error('Network Error');
        } finally {
            setShowSaveConfirm(false);
        }
    };

    if (!config) return <div className="p-8">{t.loading}</div>;

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
                    <Settings className="text-blue-600" /> {t.settings_title}
                </h1>
                <p className="text-sm text-slate-500">{t.settings_desc}</p>
            </header>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-blue-600">
                        <Cpu className="w-5 h-5" /> {t.settings_agent_title}
                    </CardTitle>
                    <CardDescription>{t.settings_agent_desc}</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t.settings_model}</label>
                            <Input
                                value={config.agents?.defaults?.model || ''}
                                onChange={(e) => updateConfig(['agents', 'defaults', 'model'], e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t.settings_workspace}</label>
                            <Input
                                value={config.agents?.defaults?.workspace || ''}
                                onChange={(e) => updateConfig(['agents', 'defaults', 'workspace'], e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
                                <Languages className="w-4 h-4" /> System Language
                            </label>
                            <Select value={language} onValueChange={(v) => setLanguage(v as Language)}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="en">English</SelectItem>
                                    <SelectItem value="ko">한국어</SelectItem>
                                    <SelectItem value="ja">日本語</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-indigo-600">
                        <Wrench className="w-5 h-5" /> {t.settings_providers_title}
                    </CardTitle>
                    <CardDescription>{t.settings_providers_desc}</CardDescription>
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
                                    <span className="text-[10px] font-medium text-slate-400 ml-1">{t.settings_api_key}</span>
                                    <Input
                                        placeholder="API Key"
                                        type="password"
                                        value={prov.api_key || ''}
                                        onChange={(e) => updateConfig(['providers', name, 'api_key'], e.target.value)}
                                    />
                                </div>
                                <div className="space-y-1">
                                    <span className="text-[10px] font-medium text-slate-400 ml-1">{t.settings_api_base}</span>
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
                        {t.settings_save_btn}
                    </Button>
                </CardFooter>
            </Card>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-green-600">
                        <Languages className="w-5 h-5" /> {t.settings_channels_title}
                    </CardTitle>
                    <CardDescription>{t.settings_channels_desc}</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-8">
                    {config.channels && Object.entries(config.channels).map(([name, ch]: [string, any]) => (
                        <div key={name} className="space-y-3 p-4 rounded-xl bg-slate-50/50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <span className="font-bold uppercase text-sm tracking-tight">{name}</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <span className="text-xs text-slate-500">{t.settings_channel_enabled}</span>
                                    <input
                                        type="checkbox"
                                        checked={ch.enabled || false}
                                        onChange={(e) => updateConfig(['channels', name, 'enabled'], e.target.checked)}
                                        className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {ch.hasOwnProperty('token') && (
                                    <div className="space-y-1">
                                        <span className="text-[10px] font-medium text-slate-400 ml-1">{t.settings_channel_token}</span>
                                        <Input
                                            placeholder="Token"
                                            type="password"
                                            value={ch.token || ''}
                                            onChange={(e) => updateConfig(['channels', name, 'token'], e.target.value)}
                                        />
                                    </div>
                                )}
                                {ch.hasOwnProperty('allow_from') && (
                                    <div className="space-y-1">
                                        <span className="text-[10px] font-medium text-slate-400 ml-1">{t.settings_channel_allow_from}</span>
                                        <Input
                                            placeholder="12345, 67890"
                                            value={Array.isArray(ch.allow_from) ? ch.allow_from.join(', ') : (ch.allow_from || '')}
                                            onChange={(e) => {
                                                const val = e.target.value.split(',').map(s => s.trim()).filter(s => s !== '');
                                                updateConfig(['channels', name, 'allow_from'], val);
                                            }}
                                        />
                                    </div>
                                )}
                                {name === 'whatsapp' && (
                                    <div className="space-y-1">
                                        <span className="text-[10px] font-medium text-slate-400 ml-1">Bridge URL</span>
                                        <Input
                                            value={ch.bridge_url || ''}
                                            onChange={(e) => updateConfig(['channels', name, 'bridge_url'], e.target.value)}
                                        />
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </CardContent>
            </Card>

            <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                    <CardTitle className="flex items-center gap-2 text-rose-600">
                        <ShieldCheck className="w-5 h-5" /> {t.settings_security_title}
                    </CardTitle>
                    <CardDescription>{t.settings_security_desc}</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t.settings_admin_account}</label>
                            <Input placeholder="admin" disabled defaultValue="admin" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t.settings_change_password}</label>
                            <Input placeholder="New password" type="password" />
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="p-6 border-t bg-slate-50 dark:bg-slate-900 flex justify-between">
                    <Button variant="outline" className="text-rose-600 border-rose-200 hover:bg-rose-50" onClick={() => setShowResetConfirm(true)}>{t.settings_reset}</Button>
                    <Button className="bg-rose-600 hover:bg-rose-700">보안 설정 저장</Button>
                </CardFooter>
            </Card>

            <ConfirmDialog
                open={showSaveConfirm}
                onOpenChange={setShowSaveConfirm}
                title={t.settings_save_confirm_title}
                description={t.settings_save_confirm_desc}
                onConfirm={handleSaveConfig}
            />

            <ConfirmDialog
                open={showResetConfirm}
                onOpenChange={setShowResetConfirm}
                title={t.settings_reset_confirm_title}
                description={t.settings_reset_confirm_desc}
                onConfirm={() => toast.info(t.settings_reset_not_impl)}
            />
        </div>
    );
}
