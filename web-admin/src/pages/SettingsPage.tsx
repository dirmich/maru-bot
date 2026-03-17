import { useState, useEffect, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Settings, Cpu, Wrench, ShieldCheck, Languages, CheckCircle2, RefreshCw, Send, MessageSquare, Bell, ExternalLink, Globe } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation, useLanguageStore, Language } from "@/lib/i18n";

export function SettingsPage() {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();

    const [config, setConfig] = useState<any>(null);
    const [fetchingModels, setFetchingModels] = useState<Record<string, boolean>>({});
    
    // For Dialogs
    const [activeChannel, setActiveChannel] = useState<string | null>(null);
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
            }
        } catch (e) {
            console.error("Config fetch error", e);
        }
    };

    const handleSaveConfig = async (silent = false) => {
        try {
            const updatedConfig = { ...config, language: language };
            const res = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(updatedConfig),
            });
            if (res.ok && !silent) {
                toast.success(t.settings_save_success);
            }
        } catch (error) {
            if (!silent) toast.error('Save failed');
        } finally {
            setShowSaveConfirm(false);
        }
    };

    const fetchModels = async (provider: string) => {
        const provData = config.providers[provider];
        // For models[0] approach or generic api_key
        const apiKey = provData.models?.[0]?.api_key || provData.api_key;
        const apiBase = provData.models?.[0]?.api_base || provData.api_base;

        if (!apiKey) {
            toast.error("API Key is required to fetch models");
            return;
        }

        setFetchingModels(prev => ({ ...prev, [provider]: true }));
        try {
            const res = await fetch('/api/config/fetch-models', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ provider, api_key: apiKey, api_base: apiBase }),
            });

            if (res.ok) {
                const data = await res.json();
                if (data.models && data.models.length > 0) {
                    // Update the model list for this provider
                    const newConfig = { ...config };
                    // If doesn't have models array, initialize it or update specific fields
                    // Here we assume the structure supports models array
                    const existingModels = newConfig.providers[provider].models || [];
                    const newModels = data.models.map((m: string) => {
                        const existing = existingModels.find((em: any) => em.model === m);
                        return existing || {
                            model: m,
                            api_key: apiKey,
                            api_base: apiBase,
                            max_tokens: 4096,
                            temperature: 0.7,
                            max_tool_iterations: 10
                        };
                    });
                    newConfig.providers[provider].models = newModels;
                    setConfig(newConfig);
                    toast.success(`${data.models.length} models fetched for ${provider}`);
                }
            } else {
                const err = await res.text();
                toast.error(`Fetch failed: ${err}`);
            }
        } catch (e) {
            toast.error("Network error while fetching models");
        } finally {
            setFetchingModels(prev => ({ ...prev, [provider]: false }));
        }
    };

    const updateConfig = (path: (string | number)[], value: any) => {
        const newConfig = JSON.parse(JSON.stringify(config));
        let current = newConfig;
        for (let i = 0; i < path.length - 1; i++) {
            if (!current[path[i]]) current[path[i]] = {};
            current = current[path[i]];
        }
        current[path[path.length - 1]] = value;
        setConfig(newConfig);
    };

    // Group models for эージェント selection
    const groupedModels = useMemo(() => {
        if (!config?.providers) return [];
        return Object.entries(config.providers).map(([name, prov]: [string, any]) => {
            const models = prov.models || [];
            return {
                provider: name,
                models: models.map((m: any) => m.model)
            };
        }).filter(g => g.models.length > 0);
    }, [config]);

    if (!config) return <div className="p-12 flex justify-center"><RefreshCw className="animate-spin text-blue-500" /></div>;

    const channelIcons: Record<string, any> = {
        telegram: <Send className="w-5 h-5" />,
        discord: <MessageSquare className="w-5 h-5" />, // Discord icon as fallback
        whatsapp: <MessageSquare className="w-5 h-5" />,
        feishu: <Bell className="w-5 h-5" />,
        webhook: <Globe className="w-5 h-5" />
    };

    return (
        <div className="p-6 max-w-6xl mx-auto space-y-10 animate-in fade-in duration-500">
            <header className="flex justify-between items-end border-b pb-6">
                <div>
                    <h1 className="text-3xl font-black tracking-tight flex items-center gap-3">
                        <div className="p-2 bg-blue-600 rounded-lg shadow-lg shadow-blue-200 dark:shadow-none">
                            <Settings className="text-white w-6 h-6" />
                        </div>
                        {t.settings_title}
                    </h1>
                    <p className="text-slate-500 mt-2 font-medium">{t.settings_desc}</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={fetchConfig} className="gap-2">
                        <RefreshCw className="w-4 h-4" /> Reset UI
                    </Button>
                    <Button onClick={() => setShowSaveConfirm(true)} size="sm" className="bg-blue-600 hover:bg-blue-700 shadow-md">
                        {t.settings_save_btn}
                    </Button>
                </div>
            </header>

            {/* 에이전트 및 기본 설정 */}
            <section className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <Card className="lg:col-span-2 border-none shadow-xl bg-gradient-to-br from-white to-slate-50/50 dark:from-slate-900 dark:to-slate-800/50 overflow-hidden">
                    <CardHeader className="border-b bg-white/50 dark:bg-black/20 backdrop-blur-sm">
                        <CardTitle className="flex items-center gap-2 text-blue-600">
                            <Cpu className="w-5 h-5" /> {t.settings_agent_title}
                        </CardTitle>
                        <CardDescription>{t.settings_agent_desc}</CardDescription>
                    </CardHeader>
                    <CardContent className="p-8 space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="space-y-3">
                                <label className="text-sm font-bold text-slate-600 dark:text-slate-400 uppercase tracking-tight">{t.settings_model}</label>
                                <Select 
                                    value={config.agents?.defaults?.model || ''} 
                                    onValueChange={(v) => {
                                        // Find provider for this model
                                        const group = groupedModels.find(g => g.models.includes(v));
                                        if (group) {
                                            updateConfig(['agents', 'defaults', 'provider'], group.provider);
                                        }
                                        updateConfig(['agents', 'defaults', 'model'], v);
                                    }}
                                >
                                    <SelectTrigger className="h-11 shadow-inner bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
                                        <SelectValue placeholder="Select a model" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {groupedModels.map((group) => (
                                            <SelectGroup key={group.provider}>
                                                <SelectLabel className="uppercase text-[10px] font-black text-slate-400 tracking-widest px-2 py-1.5">{group.provider}</SelectLabel>
                                                {group.models.map((m: string) => (
                                                    <SelectItem key={m} value={m} className="font-medium">{m}</SelectItem>
                                                ))}
                                            </SelectGroup>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-3">
                                <label className="text-sm font-bold text-slate-600 dark:text-slate-400 uppercase tracking-tight">{t.settings_workspace}</label>
                                <Input
                                    className="h-11 bg-white dark:bg-slate-800 shadow-inner"
                                    value={config.agents?.defaults?.workspace || ''}
                                    onChange={(e) => updateConfig(['agents', 'defaults', 'workspace'], e.target.value)}
                                />
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-xl bg-indigo-600 text-white overflow-hidden">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Languages className="w-5 h-5" /> System Language
                        </CardTitle>
                        <CardDescription className="text-indigo-100/70">UI 및 응답 언어를 선택하세요.</CardDescription>
                    </CardHeader>
                    <CardContent className="p-6">
                        <div className="grid grid-cols-1 gap-4">
                        {['en', 'ko', 'ja'].map((lang) => (
                            <button
                                key={lang}
                                onClick={() => setLanguage(lang as Language)}
                                className={`p-4 rounded-xl flex items-center justify-between transition-all font-bold ${
                                    language === lang 
                                    ? 'bg-white text-indigo-700 shadow-lg scale-105' 
                                    : 'bg-indigo-500/30 hover:bg-indigo-500/50 text-white'
                                }`}
                            >
                                {lang === 'en' ? 'English' : lang === 'ko' ? '한국어' : '日本語'}
                                {language === lang && <CheckCircle2 className="w-5 h-5" />}
                            </button>
                        ))}
                        </div>
                    </CardContent>
                </Card>
            </section>

            {/* AI 프로바이더 설정 */}
            <Card className="border-none shadow-2xl overflow-hidden bg-slate-900 text-slate-100">
                <CardHeader className="border-b border-slate-800 bg-slate-900/50">
                    <CardTitle className="flex items-center gap-2 text-indigo-400">
                        <Wrench className="w-5 h-5" /> {t.settings_providers_title}
                    </CardTitle>
                    <CardDescription className="text-slate-400">{t.settings_providers_desc}</CardDescription>
                </CardHeader>
                <CardContent className="p-8">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
                        {config.providers && Object.entries(config.providers).map(([name, prov]: [string, any]) => {
                            const hasModels = Array.isArray(prov.models) && prov.models.length > 0;
                            const apiKey = hasModels ? (prov.models[0].api_key || '') : (prov.api_key || '');
                            const apiBase = hasModels ? (prov.models[0].api_base || '') : (prov.api_base || '');
                            
                            return (
                                <div key={name} className="space-y-5 p-6 rounded-2xl bg-white/5 border border-white/10 relative overflow-hidden group">
                                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                                        <Wrench className="w-16 h-16 rotate-12" />
                                    </div>
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-3">
                                            <div className="w-2 h-6 bg-indigo-500 rounded-full"></div>
                                            <span className="font-black uppercase text-sm tracking-widest">{name}</span>
                                        </div>
                                        <Button 
                                            variant="ghost" 
                                            size="sm" 
                                            disabled={fetchingModels[name]}
                                            onClick={() => fetchModels(name)}
                                            className="h-8 text-[10px] font-bold uppercase tracking-widest bg-white/5 hover:bg-white/10 text-indigo-300"
                                        >
                                            {fetchingModels[name] ? <RefreshCw className="w-3 h-3 animate-spin mr-2" /> : <RefreshCw className="w-3 h-3 mr-2" />}
                                            모델 가져오기
                                        </Button>
                                    </div>
                                    <div className="grid grid-cols-1 gap-4">
                                        <div className="space-y-1.5">
                                            <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">{t.settings_api_key}</span>
                                            <Input
                                                className="bg-slate-800 border-slate-700 text-slate-200 h-10"
                                                placeholder="API Key"
                                                type="password"
                                                value={apiKey}
                                                onChange={(e) => {
                                                    if (hasModels) {
                                                        const newModels = prov.models.map((m: any) => ({ ...m, api_key: e.target.value }));
                                                        updateConfig(['providers', name, 'models'], newModels);
                                                    } else {
                                                        updateConfig(['providers', name, 'api_key'], e.target.value);
                                                    }
                                                }}
                                            />
                                        </div>
                                        <div className="space-y-1.5">
                                            <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">{t.settings_api_base}</span>
                                            <Input
                                                className="bg-slate-800 border-slate-700 text-slate-200 h-10"
                                                placeholder="Auto"
                                                value={apiBase}
                                                onChange={(e) => {
                                                    if (hasModels) {
                                                        const newModels = prov.models.map((m: any) => ({ ...m, api_base: e.target.value }));
                                                        updateConfig(['providers', name, 'models'], newModels);
                                                    } else {
                                                        updateConfig(['providers', name, 'api_base'], e.target.value);
                                                    }
                                                }}
                                            />
                                        </div>
                                    </div>
                                    {hasModels && (
                                        <div className="pt-2">
                                            <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">{prov.models.length} Models Loaded</span>
                                            <div className="mt-2 flex flex-wrap gap-1.5">
                                                {prov.models.slice(0, 5).map((m: any) => (
                                                    <span key={m.model} className="px-2 py-0.5 bg-indigo-500/20 text-indigo-300 rounded text-[10px] font-bold border border-indigo-500/30">
                                                        {m.model}
                                                    </span>
                                                ))}
                                                {prov.models.length > 5 && <span className="text-[10px] text-slate-500 font-bold">+{prov.models.length - 5} more</span>}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </CardContent>
            </Card>

            {/* 채널 설정 - 카드 그리드 형태 */}
            <section className="space-y-6">
                <div className="flex items-center gap-3">
                    <h2 className="text-2xl font-black tracking-tight">{t.settings_channels_title}</h2>
                    <div className="h-px flex-1 bg-slate-200 dark:bg-slate-800"></div>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {config.channels && Object.entries(config.channels).map(([name, ch]: [string, any]) => (
                        <Card 
                            key={name} 
                            className={`relative cursor-pointer transition-all duration-300 transform hover:-translate-y-1 hover:shadow-2xl group border-2 ${
                                ch.enabled 
                                ? 'border-green-500/50 bg-green-50/10 dark:bg-green-950/10' 
                                : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900'
                            }`}
                            onClick={() => setActiveChannel(name)}
                        >
                            <CardContent className="p-6 flex flex-col items-center text-center space-y-4">
                                <div className={`p-4 rounded-2xl ${ch.enabled ? 'bg-green-500 text-white shadow-lg shadow-green-200 dark:shadow-none' : 'bg-slate-100 dark:bg-slate-800 text-slate-400'}`}>
                                    {channelIcons[name] || <MessageSquare className="w-8 h-8" />}
                                </div>
                                <div className="space-y-1">
                                    <h3 className="font-black uppercase tracking-widest text-sm">{name}</h3>
                                    <div className="flex items-center gap-2 justify-center">
                                        <div className={`w-2 h-2 rounded-full ${ch.enabled ? 'bg-green-500 animate-pulse' : 'bg-slate-300'}`}></div>
                                        <span className={`text-[10px] font-bold uppercase ${ch.enabled ? 'text-green-600' : 'text-slate-400'}`}>
                                            {ch.enabled ? 'ACTIVE' : 'INACTIVE'}
                                        </span>
                                    </div>
                                </div>
                            </CardContent>
                            <div className="absolute top-4 right-4 pb-2">
                                <ExternalLink className="w-4 h-4 text-slate-300 group-hover:text-blue-500 transition-colors" />
                            </div>
                        </Card>
                    ))}
                </div>
            </section>

            {/* 보안 설정 */}
            <Card className="border-none shadow-xl bg-slate-50 dark:bg-slate-900/50">
                <CardHeader className="p-8">
                    <CardTitle className="flex items-center gap-2 text-rose-600">
                        <ShieldCheck className="w-5 h-5" /> {t.settings_security_title}
                    </CardTitle>
                    <CardDescription>{t.settings_security_desc}</CardDescription>
                </CardHeader>
                <CardContent className="px-8 pb-8 pt-0">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        <div className="space-y-2">
                            <label className="text-xs font-black text-slate-500 uppercase tracking-widest ml-1">{t.settings_admin_account}</label>
                            <Input className="h-11 shadow-inner bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800" placeholder="admin" disabled defaultValue="admin" />
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-black text-slate-500 uppercase tracking-widest ml-1">{t.settings_change_password}</label>
                            <Input 
                                className="h-11 shadow-inner bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800" 
                                placeholder="New password" 
                                type="password" 
                                value={config.admin_password || ''}
                                onChange={(e) => updateConfig(['admin_password'], e.target.value)}
                            />
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="p-6 border-t bg-white/50 dark:bg-slate-900/80 backdrop-blur-sm flex justify-between">
                    <Button variant="outline" className="text-rose-600 border-rose-200 hover:bg-rose-50 font-bold" onClick={() => setShowResetConfirm(true)}>{t.settings_reset}</Button>
                    <Button onClick={() => setShowSaveConfirm(true)} className="bg-rose-600 hover:bg-rose-700 shadow-md transition-all px-8 font-bold">보안 설정 저장</Button>
                </CardFooter>
            </Card>

            {/* 채널 설정 다이얼로그 */}
            <Dialog open={activeChannel !== null} onOpenChange={(open) => !open && setActiveChannel(null)}>
                <DialogContent className="sm:max-w-[500px] border-none shadow-2xl overflow-hidden p-0 bg-white dark:bg-slate-950">
                    {activeChannel && (
                        <>
                        <div className={`p-8 pb-6 flex items-center justify-between bg-gradient-to-r ${config.channels[activeChannel].enabled ? 'from-green-600 to-emerald-600' : 'from-slate-700 to-slate-800'} text-white`}>
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-white/20 rounded-xl">
                                    {channelIcons[activeChannel] || <MessageSquare className="w-6 h-6" />}
                                </div>
                                <div className="space-y-0.5">
                                    <DialogTitle className="uppercase font-black tracking-widest text-lg">{activeChannel}</DialogTitle>
                                    <DialogDescription className="text-white/70 font-medium">채널 상세 설정을 업데이트하세요.</DialogDescription>
                                </div>
                            </div>
                            <div className="flex items-center gap-2 bg-black/20 p-2 rounded-full px-4 border border-white/10">
                                <span className="text-[10px] font-black tracking-widest uppercase">Enabled</span>
                                <input
                                    type="checkbox"
                                    checked={config.channels[activeChannel].enabled || false}
                                    onChange={(e) => updateConfig(['channels', activeChannel, 'enabled'], e.target.checked)}
                                    className="w-5 h-5 rounded border-none text-green-500 focus:ring-offset-0 focus:ring-0"
                                />
                            </div>
                        </div>
                        <div className="p-8 space-y-6">
                            {config.channels[activeChannel].hasOwnProperty('token') && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_token}</label>
                                    <Input
                                        className="h-12 bg-slate-50 dark:bg-slate-900 shadow-inner"
                                        placeholder="Token"
                                        type="password"
                                        value={config.channels[activeChannel].token || ''}
                                        onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                    />
                                </div>
                            )}
                            {config.channels[activeChannel].hasOwnProperty('allow_from') && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
                                    <Input
                                        className="h-12 bg-slate-50 dark:bg-slate-900 shadow-inner"
                                        placeholder="12345, 67890"
                                        value={Array.isArray(config.channels[activeChannel].allow_from) ? config.channels[activeChannel].allow_from.join(', ') : (config.channels[activeChannel].allow_from || '')}
                                        onChange={(e) => {
                                            const val = e.target.value.split(',').map(s => s.trim()).filter(s => s !== '');
                                            updateConfig(['channels', activeChannel, 'allow_from'], val);
                                        }}
                                    />
                                    <p className="text-[10px] text-slate-400 font-medium ml-1">* 허용할 사용자 ID를 쉼표로 구분하여 입력하세요 (비워두면 모두 허용).</p>
                                </div>
                            )}
                            {activeChannel === 'whatsapp' && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Bridge URL</label>
                                    <Input
                                        className="h-12 bg-slate-50 dark:bg-slate-900 shadow-inner"
                                        value={config.channels[activeChannel].bridge_url || ''}
                                        onChange={(e) => updateConfig(['channels', activeChannel, 'bridge_url'], e.target.value)}
                                        placeholder="ws://localhost:3001"
                                    />
                                </div>
                            )}
                            {activeChannel === 'feishu' && (
                                <>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">App ID</label>
                                        <Input
                                            className="h-12 bg-slate-50 dark:bg-slate-900 shadow-inner"
                                            value={config.channels[activeChannel].app_id || ''}
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'app_id'], e.target.value)}
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">App Secret</label>
                                        <Input
                                            className="h-12 bg-slate-50 dark:bg-slate-900 shadow-inner"
                                            type="password"
                                            value={config.channels[activeChannel].app_secret || ''}
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'app_secret'], e.target.value)}
                                        />
                                    </div>
                                </>
                            )}
                        </div>
                        <DialogFooter className="p-8 pt-0 flex justify-end gap-3">
                            <Button variant="ghost" onClick={() => setActiveChannel(null)} className="font-bold">Close</Button>
                            <Button 
                                onClick={() => {
                                    handleSaveConfig(true);
                                    setActiveChannel(null);
                                }} 
                                className={`${config.channels[activeChannel].enabled ? 'bg-green-600 hover:bg-green-700' : 'bg-slate-800 hover:bg-slate-900'} px-8 font-black uppercase tracking-widest text-xs`}
                            >
                                Apply Changes
                            </Button>
                        </DialogFooter>
                        </>
                    )}
                </DialogContent>
            </Dialog>

            <ConfirmDialog
                open={showSaveConfirm}
                onOpenChange={setShowSaveConfirm}
                title={t.settings_save_confirm_title}
                description={t.settings_save_confirm_desc}
                onConfirm={() => handleSaveConfig()}
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
