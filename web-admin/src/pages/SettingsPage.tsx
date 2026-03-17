import { useState, useEffect, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Settings, Cpu, Wrench, ShieldCheck, Languages, CheckCircle2, RefreshCw, Send, MessageSquare, Bell, ExternalLink, Globe, Trash2, Plus } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation, useLanguageStore, Language } from "@/lib/i18n";
import { authenticatedFetch } from "@/lib/auth";

export function SettingsPage() {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();

    const [config, setConfig] = useState<any>(null);
    const [fetchingModels, setFetchingModels] = useState<Record<string, boolean>>({});
    
    // For Dialogs
    const [activeChannel, setActiveChannel] = useState<string | null>(null);
    const [showSaveConfirm, setShowSaveConfirm] = useState(false);
    const [showResetConfirm, setShowResetConfirm] = useState(false);

    // Provider Management
    const [isAddProviderOpen, setIsAddProviderOpen] = useState(false);
    const [addingProvider, setAddingProvider] = useState<{name: string, index?: number} | null>(null);
    const [tempProvider, setTempProvider] = useState<any>({ api_key: '', api_base: '', models: [] });

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await authenticatedFetch('/api/config');
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
            const res = await authenticatedFetch('/api/config', {
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

    const openAddProvider = (name: string, index?: number) => {
        setAddingProvider({ name, index });
        if (name === 'ollama' && index !== undefined) {
             setTempProvider(config.providers.ollama[index]);
        } else if (config.providers[name] && !Array.isArray(config.providers[name])) {
             setTempProvider(config.providers[name]);
        } else {
             setTempProvider({ api_key: '', api_base: name === 'ollama' ? 'http://localhost:11434' : '', models: [] });
        }
        setIsAddProviderOpen(true);
    };

    const confirmAddProvider = () => {
        const newConfig = { ...config };
        if (addingProvider?.name === 'ollama') {
            if (addingProvider.index !== undefined) {
                newConfig.providers.ollama[addingProvider.index] = tempProvider;
            } else {
                newConfig.providers.ollama = [...(newConfig.providers.ollama || []), tempProvider];
            }
        } else if (addingProvider) {
            newConfig.providers[addingProvider.name] = tempProvider;
        }
        setConfig(newConfig);
        setIsAddProviderOpen(false);
        setAddingProvider(null);
    };

    const deleteProvider = (name: string, index?: number) => {
        const newConfig = { ...config };
        if (name === 'ollama') {
            if (Array.isArray(newConfig.providers.ollama) && newConfig.providers.ollama.length > 1) {
                newConfig.providers.ollama = newConfig.providers.ollama.filter((_: any, i: number) => i !== index);
            } else {
                newConfig.providers.ollama = [{ api_key: '', api_base: '', models: [] }];
            }
        } else {
            newConfig.providers[name] = { api_key: '', api_base: '', models: [] };
        }
        setConfig(newConfig);
    };

    const fetchModelsForTemp = async () => {
        if (!tempProvider.api_base && addingProvider?.name === 'ollama') {
            tempProvider.api_base = 'http://localhost:11434';
        }
        
        setFetchingModels(prev => ({ ...prev, temp: true }));
        try {
            const res = await authenticatedFetch('/api/config/fetch-models', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    provider: addingProvider?.name, 
                    api_key: tempProvider.api_key, 
                    api_base: tempProvider.api_base 
                }),
            });

            if (res.ok) {
                const data = await res.json();
                if (data.models && data.models.length > 0) {
                    const newModels = data.models.map((m: string) => ({
                        model: m,
                        api_key: tempProvider.api_key,
                        api_base: tempProvider.api_base,
                        max_tokens: 4096,
                        temperature: 0.7,
                        max_tool_iterations: 10
                    }));
                    setTempProvider({ ...tempProvider, models: newModels });
                    toast.success(`${data.models.length} models fetched`);
                }
            } else {
                const err = await res.text();
                toast.error(`Fetch failed: ${err}`);
            }
        } catch (e) {
            toast.error("Network error");
        } finally {
            setFetchingModels(prev => ({ ...prev, temp: false }));
        }
    };

    const groupedModels = useMemo(() => {
        if (!config?.providers) return [];
        const groups: any[] = [];
        
        Object.entries(config.providers).forEach(([name, prov]: [string, any]) => {
            if (name === 'ollama' && Array.isArray(prov)) {
                prov.forEach((p, idx) => {
                    if (p?.models?.length > 0) {
                        groups.push({
                            provider: `ollama#${idx}`,
                            models: p.models.map((m: any) => m.model)
                        });
                    }
                });
            } else if (prov?.models?.length > 0) {
                groups.push({
                    provider: name,
                    models: prov.models.map((m: any) => m.model)
                });
            }
        });
        return groups;
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
                <CardHeader className="border-b border-slate-800 bg-slate-900/50 flex flex-row items-center justify-between">
                    <div>
                        <CardTitle className="flex items-center gap-2 text-indigo-400">
                            <Wrench className="w-5 h-5" /> {t.settings_providers_title}
                        </CardTitle>
                        <CardDescription className="text-slate-400">{t.settings_providers_desc}</CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Select onValueChange={(v) => openAddProvider(v)}>
                            <SelectTrigger className="w-[180px] bg-indigo-600 border-none h-9 text-xs font-bold">
                                <Plus className="w-4 h-4 mr-2" />
                                <SelectValue placeholder="Add Provider" />
                            </SelectTrigger>
                            <SelectContent>
                                {['openai', 'anthropic', 'gemini', 'openrouter', 'groq', 'zhipu', 'vllm'].map(p => (
                                    <SelectItem key={p} value={p} className="uppercase font-bold">{p}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <Button size="sm" onClick={() => openAddProvider('ollama')} className="bg-orange-600 hover:bg-orange-700 h-9 text-xs font-bold">
                             + Add Ollama
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="p-8">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        {/* Standard Providers */}
                        {Object.entries(config.providers)
                            .filter(([name]) => name !== 'ollama')
                            .map(([name, prov]: [string, any]) => {
                                if (!prov) return null;
                                const isConfigured = prov.api_key || prov.api_base || (prov.models && prov.models.length > 0);
                                if (!isConfigured) return null;
                                
                                return (
                                    <div key={name} className="group relative p-6 rounded-2xl bg-white/5 border border-white/10 hover:border-indigo-500/50 transition-all">
                                        <div className="flex items-center justify-between mb-4">
                                            <div className="flex items-center gap-3">
                                                <div className="w-2 h-6 bg-indigo-500 rounded-full"></div>
                                                <span className="font-black uppercase text-sm tracking-widest">{name}</span>
                                            </div>
                                            <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-indigo-400" onClick={() => openAddProvider(name)}>
                                                    <Settings className="w-4 h-4" />
                                                </Button>
                                                <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-rose-400" onClick={() => deleteProvider(name)}>
                                                    <Trash2 className="w-4 h-4" />
                                                </Button>
                                            </div>
                                        </div>
                                        <div className="text-xs text-slate-400 space-y-1">
                                            <div className="flex justify-between font-mono">
                                                <span>API Base:</span>
                                                <span className="text-slate-300 truncate max-w-[150px]">{prov.api_base || 'Default'}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span>Models:</span>
                                                <span className="text-indigo-400 font-bold">{prov.models?.length || 0} loaded</span>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}

                        {/* Ollama Providers */}
                        {Array.isArray(config.providers.ollama) && config.providers.ollama.map((prov: any, idx: number) => {
                             if (!prov) return null;
                             const isConfigured = prov.api_base || (prov.models && prov.models.length > 0);
                             if (!isConfigured && config.providers.ollama.length === 1) return null;

                             return (
                                <div key={`ollama-${idx}`} className="group relative p-6 rounded-2xl bg-orange-500/5 border border-orange-500/10 hover:border-orange-500/50 transition-all">
                                    <div className="flex items-center justify-between mb-4">
                                        <div className="flex items-center gap-3">
                                            <div className="w-2 h-6 bg-orange-500 rounded-full"></div>
                                            <span className="font-black uppercase text-sm tracking-widest">Ollama #{idx}</span>
                                        </div>
                                        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-orange-400" onClick={() => openAddProvider('ollama', idx)}>
                                                <Settings className="w-4 h-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-rose-400" onClick={() => deleteProvider('ollama', idx)}>
                                                <Trash2 className="w-4 h-4" />
                                            </Button>
                                        </div>
                                    </div>
                                    <div className="text-xs text-slate-400 space-y-1">
                                        <div className="flex justify-between font-mono">
                                            <span>API Base:</span>
                                            <span className="text-slate-300 truncate max-w-[150px]">{prov.api_base || 'http://localhost:11434'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Models:</span>
                                            <span className="text-orange-400 font-bold">{prov.models?.length || 0} loaded</span>
                                        </div>
                                    </div>
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
                    <div className="flex items-center gap-2 text-slate-400 text-xs font-medium">
                        <ShieldCheck className="w-4 h-4" />
                        {t.settings_security_desc}
                    </div>
                </CardFooter>
            </Card>

            {/* Dialogs */}
            
            {/* Add/Edit Provider Dialog */}
            <Dialog open={isAddProviderOpen} onOpenChange={setIsAddProviderOpen}>
                <DialogContent className="sm:max-w-[500px] border-none shadow-2xl bg-slate-900 text-slate-100 p-0 overflow-hidden">
                    <div className="p-8 space-y-6">
                        <div className="space-y-2">
                            <DialogTitle className="text-2xl font-black uppercase tracking-tight text-indigo-400">
                                {addingProvider?.index !== undefined ? 'Edit' : 'Add'} {addingProvider?.name}
                            </DialogTitle>
                            <DialogDescription className="text-slate-400">
                                Enter the API details for this provider.
                            </DialogDescription>
                        </div>

                        <div className="space-y-4">
                            {addingProvider?.name !== 'ollama' && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">API Key</label>
                                    <Input 
                                        type="password"
                                        className="bg-slate-800 border-slate-700 h-11"
                                        placeholder="sk-..."
                                        value={tempProvider.api_key || ''}
                                        onChange={(e) => setTempProvider({ ...tempProvider, api_key: e.target.value })}
                                    />
                                </div>
                            )}
                            <div className="space-y-2">
                                <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">API Base (Optional)</label>
                                <Input 
                                    className="bg-slate-800 border-slate-700 h-11"
                                    placeholder={addingProvider?.name === 'ollama' ? "http://localhost:11434" : "https://api..."}
                                    value={tempProvider.api_base || ''}
                                    onChange={(e) => setTempProvider({ ...tempProvider, api_base: e.target.value })}
                                />
                            </div>

                            <div className="pt-2">
                                <Button 
                                    className="w-full bg-white/5 hover:bg-white/10 text-indigo-300 font-bold h-11 border border-white/10"
                                    onClick={fetchModelsForTemp}
                                    disabled={fetchingModels.temp}
                                >
                                    {fetchingModels.temp ? <RefreshCw className="w-4 h-4 animate-spin mr-2" /> : <RefreshCw className="w-4 h-4 mr-2" />}
                                    FETCH MODELS
                                </Button>
                            </div>

                            {tempProvider.models?.length > 0 && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black uppercase tracking-widest text-slate-500 ml-1">{tempProvider.models.length} Models Found</label>
                                    <div className="max-h-32 overflow-y-auto p-3 rounded-xl bg-black/20 border border-white/5 space-y-1">
                                        {tempProvider.models.map((m: any) => (
                                            <div key={m.model} className="text-xs font-mono text-slate-300 flex items-center gap-2">
                                                <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                                {m.model}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                    <DialogFooter className="bg-slate-800/50 p-6">
                        <Button variant="ghost" onClick={() => setIsAddProviderOpen(false)} className="text-slate-400 hover:text-white">Cancel</Button>
                        <Button onClick={confirmAddProvider} className="bg-indigo-600 hover:bg-indigo-700 font-black px-8 h-11">
                            {addingProvider?.index !== undefined ? 'UPDATE' : 'REGISTER'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Channel Edit Dialog */}
            <Dialog open={!!activeChannel} onOpenChange={(open) => !open && setActiveChannel(null)}>
                {activeChannel && (
                    <DialogContent className="sm:max-w-[450px] border-none shadow-2xl p-0 overflow-hidden rounded-3xl">
                        <div className={`p-8 ${config.channels[activeChannel].enabled ? 'bg-green-600' : 'bg-slate-900'} text-white space-y-2`}>
                            <DialogTitle className="text-3xl font-black uppercase tracking-tighter flex items-center gap-3">
                                {channelIcons[activeChannel] || <MessageSquare className="w-8 h-8" />}
                                {activeChannel}
                            </DialogTitle>
                            <DialogDescription className="text-white/70">
                                Configure how this channel interacts with MaruBot.
                            </DialogDescription>
                        </div>
                        
                        <div className="p-8 space-y-6 bg-white dark:bg-slate-900">
                            <div className="flex items-center justify-between p-4 rounded-2xl bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700">
                                <div className="space-y-0.5">
                                    <label className="font-bold text-sm tracking-tight capitalize">Enable Channel</label>
                                    <p className="text-xs text-slate-500">봇 활성화 여부를 결정합니다.</p>
                                </div>
                                <input 
                                    type="checkbox" 
                                    className="w-10 h-6 bg-slate-200 rounded-full appearance-none checked:bg-green-500 transition-all cursor-pointer relative after:content-[''] after:absolute after:top-1 after:left-1 after:w-4 after:h-4 after:bg-white after:rounded-full after:transition-all checked:after:left-5 shadow-inner"
                                    checked={config.channels[activeChannel].enabled} 
                                    onChange={(e) => updateConfig(['channels', activeChannel, 'enabled'], e.target.checked)}
                                />
                            </div>

                            {activeChannel === 'telegram' && (
                                <div className="space-y-4 animate-in slide-in-from-top-2">
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_token}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                            placeholder="123456:ABC-DEF..."
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">ALLOWED USERS (Optional)</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].allow_from?.join(', ') || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'allow_from'], (e.target.value.split(',') as any).map((s: string) => s.trim()).filter((s: string) => s))}
                                            placeholder="12345678, 98765432"
                                        />
                                    </div>
                                </div>
                            )}

                            {activeChannel === 'discord' && (
                                <div className="space-y-4 animate-in slide-in-from-top-2">
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">BOT TOKEN</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                            type="password"
                                        />
                                    </div>
                                </div>
                            )}

                            {/* Add other channel types as needed */}
                        </div>

                        <DialogFooter className="p-6 bg-slate-50 dark:bg-slate-800/50 border-t gap-2">
                            <Button variant="outline" onClick={() => setActiveChannel(null)} className="rounded-xl font-bold">Close</Button>
                            <Button onClick={() => { handleSaveConfig(); setActiveChannel(null); }} className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-black px-8">SAVE & APPLY</Button>
                        </DialogFooter>
                    </DialogContent>
                )}
            </Dialog>

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
                title="설정 초기화"
                description="모든 설정을 서버의 현재 값으로 되돌리시겠습니까?"
                onConfirm={fetchConfig}
            />
        </div>
    );
}

