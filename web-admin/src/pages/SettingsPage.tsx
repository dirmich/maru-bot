import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { authenticatedFetch } from "@/lib/auth";
import { Language, useLanguageStore, useTranslation } from "@/lib/i18n";
import { Cpu, ExternalLink, Globe, HelpCircle, Languages, MessageSquare, Monitor, Moon, Plus, RefreshCw, Send, Settings, ShieldCheck, Sun, Trash2, Wrench } from 'lucide-react';
import { useTheme } from "next-themes";
import { cn } from "@/lib/utils";
import { useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

export function SettingsPage() {
    const t = useTranslation();
    const { language, setLanguage } = useLanguageStore();

    const [config, setConfig] = useState<any>(null);
    const [fetchingModels, setFetchingModels] = useState<Record<string, boolean>>({});
    
    // For Dialogs
    const [activeChannel, setActiveChannel] = useState<string | null>(null);
    const [showSaveConfirm, setShowSaveConfirm] = useState(false);
    const [showResetConfirm, setShowResetConfirm] = useState(false);
    const [showResultDialog, setShowResultDialog] = useState(false);
    const [saveResult, setSaveResult] = useState<{success: boolean, message?: string} | null>(null);

    // Provider Management
    const [isAddProviderOpen, setIsAddProviderOpen] = useState(false);
    const [addingProvider, setAddingProvider] = useState<{name: string, index?: number} | null>(null);
    const [tempProvider, setTempProvider] = useState<any>({ api_key: '', api_base: '', models: [] });

    // For Theme
    const { theme, setTheme } = useTheme();

    // For "How to Get" Dialog
    const [showHowToGet, setShowHowToGet] = useState(false);
    const [helpChannel, setHelpChannel] = useState<string | null>(null);


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
                setSaveResult({ success: true });
                setShowResultDialog(true);
                toast.success(t.settings_save_success);
            } else if (!silent) {
                const errText = await res.text();
                setSaveResult({ success: false, message: errText || 'Unknown error' });
                setShowResultDialog(true);
                toast.error(`Save status: ${res.status}`);
            }
        } catch (error) {
            if (!silent) {
                setSaveResult({ success: false, message: 'Network or internal error' });
                setShowResultDialog(true);
                toast.error('Save failed');
            }
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

    const providerLabel = (provider: string) => provider.startsWith('ollama#') ? provider.replace('#', ' #') : provider;
    const makeModelValue = (provider: string, model: string) => `${provider}::${model}`;
    const parseModelValue = (value: string) => {
        if (!value.includes('::')) {
            return { provider: '', model: value };
        }
        const [provider, ...modelParts] = value.split('::');
        const model = modelParts.join('::');
        return { provider, model };
    };

    const resetAgentProviderRefs = (cfg: any, predicate: (provider: string) => boolean, mapper?: (provider: string) => string) => {
        const currentProvider = cfg.agents?.defaults?.provider || '';
        if (predicate(currentProvider)) {
            const mapped = mapper ? mapper(currentProvider) : '';
            if (mapped) {
                cfg.agents.defaults.provider = mapped;
            } else {
                cfg.agents.defaults.provider = '';
                cfg.agents.defaults.model = '';
            }
        }

        cfg.agents.defaults.fallback_models = (cfg.agents?.defaults?.fallback_models || [])
            .map((entry: string) => {
                const parsed = parseModelValue(entry);
                if (!predicate(parsed.provider)) return entry;
                const mapped = mapper ? mapper(parsed.provider) : '';
                return mapped ? makeModelValue(mapped, parsed.model) : '';
            })
            .filter(Boolean);
    };

    const toggleProviderEnabled = (name: string, enabled: boolean, index?: number) => {
        const newConfig = JSON.parse(JSON.stringify(config));
        if (name === 'ollama' && index !== undefined) {
            if (!newConfig.providers.ollama[index]) return;
            newConfig.providers.ollama[index].enabled = enabled;
        } else if (newConfig.providers[name]) {
            newConfig.providers[name].enabled = enabled;
        }

        if (!enabled) {
            const providerRef = index !== undefined ? `${name}#${index}` : name;
            newConfig.agents.defaults.fallback_models = (newConfig.agents.defaults.fallback_models || [])
                .filter((entry: string) => !entry.startsWith(`${providerRef}::`));
        }

        setConfig(newConfig);
    };

    const openAddProvider = (name: string, index?: number) => {
        setAddingProvider({ name, index });
        if (name === 'ollama' && index !== undefined) {
             setTempProvider(config.providers.ollama[index]);
        } else if (config.providers[name] && !Array.isArray(config.providers[name])) {
             setTempProvider(config.providers[name]);
        } else {
             setTempProvider({
                enabled: true,
                api_key: '',
                api_base: name === 'ollama'
                    ? 'http://localhost:11434'
                    : name === 'llamacpp'
                        ? 'http://localhost:8080/v1'
                        : '',
                models: []
            });
        }
        setIsAddProviderOpen(true);
    };

    const confirmAddProvider = () => {
        const newConfig = JSON.parse(JSON.stringify(config));
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
        const newConfig = JSON.parse(JSON.stringify(config));
        if (name === 'ollama') {
            if (Array.isArray(newConfig.providers.ollama) && newConfig.providers.ollama.length > 1) {
                newConfig.providers.ollama = newConfig.providers.ollama.filter((_: any, i: number) => i !== index);
                if (index !== undefined) {
                    resetAgentProviderRefs(
                        newConfig,
                        (provider) => provider === `ollama#${index}` || /^ollama#\d+$/.test(provider) && Number(provider.split('#')[1]) > index,
                        (provider) => {
                            if (provider === `ollama#${index}`) return '';
                            const currentIndex = Number(provider.split('#')[1]);
                            return `ollama#${currentIndex - 1}`;
                        }
                    );
                }
            } else {
                newConfig.providers.ollama = [{ enabled: false, api_key: '', api_base: '', models: [] }];
                resetAgentProviderRefs(newConfig, (provider) => provider.startsWith('ollama'));
            }
        } else {
            newConfig.providers[name] = { enabled: false, api_key: '', api_base: '', models: [] };
            resetAgentProviderRefs(newConfig, (provider) => provider === name);
        }
        setConfig(newConfig);
    };

    const fetchModelsForTemp = async () => {
        if (!tempProvider.api_base && addingProvider?.name === 'ollama') {
            tempProvider.api_base = 'http://localhost:11434';
        }
        if (!tempProvider.api_base && addingProvider?.name === 'llamacpp') {
            tempProvider.api_base = 'http://localhost:8080/v1';
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
                            label: providerLabel(`ollama#${idx}`),
                            enabled: p.enabled !== false,
                            models: p.models.map((m: any) => ({
                                name: m.model,
                                value: makeModelValue(`ollama#${idx}`, m.model),
                            }))
                        });
                    }
                });
            } else if (prov?.models?.length > 0) {
                groups.push({
                    provider: name,
                    label: providerLabel(name),
                    enabled: prov.enabled !== false,
                    models: prov.models.map((m: any) => ({
                        name: m.model,
                        value: makeModelValue(name, m.model),
                    }))
                });
            }
        });
        return groups;
    }, [config]);

    const modelOptions = useMemo(
        () => groupedModels.flatMap((group) => group.models.map((model: any) => ({
            ...model,
            provider: group.provider,
        }))),
        [groupedModels]
    );

    const currentModelValue = useMemo(() => {
        const provider = config?.agents?.defaults?.provider || '';
        const model = config?.agents?.defaults?.model || '';
        if (!model) return '';

        if (provider) {
            const explicitValue = makeModelValue(provider, model);
            if (modelOptions.some((option: any) => option.value === explicitValue)) {
                return explicitValue;
            }
        }

        const legacyMatches = modelOptions.filter((option: any) => option.name === model);
        if (legacyMatches.length === 1) {
            return legacyMatches[0].value;
        }

        return provider ? makeModelValue(provider, model) : '';
    }, [config, modelOptions]);

    const fallbackOptions = useMemo(() => {
        const primaryValue = currentModelValue;
        return groupedModels
            .filter((group) => group.enabled)
            .flatMap((group) =>
                group.models
                    .filter((model: any) => model.value !== primaryValue)
                    .map((model: any) => ({
                        provider: group.provider,
                        providerLabel: group.label,
                        name: model.name,
                        value: model.value,
                    }))
            );
    }, [groupedModels, currentModelValue]);

    const toggleFallbackModel = (value: string) => {
        const selected = new Set(config.agents?.defaults?.fallback_models || []);
        if (selected.has(value)) {
            selected.delete(value);
        } else {
            selected.add(value);
        }
        updateConfig(['agents', 'defaults', 'fallback_models'], Array.from(selected));
    };

    const setPrimaryModel = (value: string) => {
        const parsed = parseModelValue(value);
        setConfig((prev: any) => {
            const newConfig = JSON.parse(JSON.stringify(prev));
            newConfig.agents.defaults.provider = parsed.provider;
            newConfig.agents.defaults.model = parsed.model;
            newConfig.agents.defaults.fallback_models = (newConfig.agents.defaults.fallback_models || [])
                .filter((entry: string) => entry !== value && entry !== parsed.model);
            return newConfig;
        });
    };

    if (!config) return <div className="p-12 flex justify-center"><RefreshCw className="animate-spin text-blue-500" /></div>;

    const channelIcons: Record<string, any> = {
        telegram: <Send className="w-5 h-5" />,
        discord: <MessageSquare className="w-5 h-5" />, 
        slack: <MessageSquare className="w-5 h-5" />,
        whatsapp: <MessageSquare className="w-5 h-5" />,
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
                                {['openai', 'anthropic', 'gemini', 'openrouter', 'groq', 'zhipu', 'vllm', 'llamacpp'].map(p => (
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
                                            <div className="flex items-center gap-3">
                                                <div className="flex items-center gap-2 text-[10px] uppercase font-black tracking-widest text-slate-400">
                                                    <span>{prov.enabled !== false ? 'Enabled' : 'Disabled'}</span>
                                                    <Switch
                                                        checked={prov.enabled !== false}
                                                        onCheckedChange={(checked) => toggleProviderEnabled(name, checked)}
                                                    />
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
                                        <div className="flex items-center gap-3">
                                            <div className="flex items-center gap-2 text-[10px] uppercase font-black tracking-widest text-slate-400">
                                                <span>{prov.enabled !== false ? 'Enabled' : 'Disabled'}</span>
                                                <Switch
                                                    checked={prov.enabled !== false}
                                                    onCheckedChange={(checked) => toggleProviderEnabled('ollama', checked, idx)}
                                                />
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
                                    value={currentModelValue} 
                                    onValueChange={setPrimaryModel}
                                >
                                    <SelectTrigger className="h-11 shadow-inner bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
                                        <SelectValue placeholder="Select a model" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {groupedModels.map((group) => (
                                            <SelectGroup key={group.provider}>
                                                <SelectLabel className="uppercase text-[10px] font-black text-slate-400 tracking-widest px-2 py-1.5">{group.label}</SelectLabel>
                                                {group.models.map((m: any) => (
                                                    <SelectItem key={m.value} value={m.value} className="font-medium">{m.name}</SelectItem>
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

                        <div className="pt-6 border-t border-slate-100 dark:border-slate-800 space-y-4">
                            <div className="space-y-1">
                                <label className="text-sm font-bold text-slate-600 dark:text-slate-400 uppercase tracking-tight">Fallback Models</label>
                                <p className="text-xs text-slate-500 dark:text-slate-400">
                                    Enabled provider만 fallback 대상으로 사용합니다. 저장 형식은 `provider::model`입니다.
                                </p>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                {fallbackOptions.map((option) => {
                                    const selected = (config.agents?.defaults?.fallback_models || []).includes(option.value);
                                    return (
                                        <button
                                            key={option.value}
                                            type="button"
                                            onClick={() => toggleFallbackModel(option.value)}
                                            className={`rounded-2xl border p-4 text-left transition-all ${
                                                selected
                                                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/20 shadow-md'
                                                    : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900'
                                            }`}
                                        >
                                            <div className="flex items-center justify-between gap-3">
                                                <div>
                                                    <div className="font-bold text-sm text-slate-900 dark:text-slate-100">{option.name}</div>
                                                    <div className="text-[10px] uppercase tracking-widest text-slate-500">{option.providerLabel}</div>
                                                    <div className="text-[11px] text-slate-400 mt-1 font-mono">{option.value}</div>
                                                </div>
                                                <Switch
                                                    checked={selected}
                                                    onCheckedChange={() => toggleFallbackModel(option.value)}
                                                    onClick={(e) => e.stopPropagation()}
                                                    aria-label={`Toggle fallback model ${option.value}`}
                                                />
                                            </div>
                                        </button>
                                    );
                                })}
                            </div>
                        </div>

                        {/* Theme Selection */}
                        <div className="pt-6 border-t border-slate-100 dark:border-slate-800 space-y-4">
                            <label className="text-sm font-bold text-slate-600 dark:text-slate-400 uppercase tracking-tight">{t.settings_theme_title}</label>
                            <div className="grid grid-cols-3 gap-3">
                                {[
                                    { id: 'light', icon: <Sun className="w-4 h-4" />, label: t.settings_theme_light },
                                    { id: 'dark', icon: <Moon className="w-4 h-4" />, label: t.settings_theme_dark },
                                    { id: 'system', icon: <Monitor className="w-4 h-4" />, label: t.settings_theme_system }
                                ].map((item) => (
                                    <button
                                        key={item.id}
                                        onClick={() => setTheme(item.id)}
                                        className={`flex flex-col items-center justify-center p-3 rounded-xl border-2 transition-all gap-2 ${
                                            theme === item.id 
                                            ? 'border-blue-500 bg-blue-50/50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 font-bold' 
                                            : 'border-slate-100 dark:border-slate-800 hover:border-slate-200 dark:hover:border-slate-700 text-slate-500'
                                        }`}
                                    >
                                        {item.icon}
                                        <span className="text-[10px] uppercase tracking-wider">{item.label}</span>
                                    </button>
                                ))}
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none shadow-xl bg-indigo-600 text-white overflow-hidden">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Languages className="w-5 h-5" /> {t.settings_system_language}
                        </CardTitle>
                        <CardDescription className="text-indigo-100/70">{t.settings_system_language_desc}</CardDescription>
                    </CardHeader>
                    <CardContent className="p-6">
                        <div className="space-y-4">
                            <Select 
                                value={language} 
                                onValueChange={(v) => setLanguage(v as Language)}
                            >
                                <SelectTrigger className="h-14 bg-white/20 border-white/20 text-white font-bold text-lg rounded-2xl">
                                    <SelectValue placeholder="Select Language" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="en">English</SelectItem>
                                    <SelectItem value="ko">한국어</SelectItem>
                                    <SelectItem value="ja">日本語</SelectItem>
                                </SelectContent>
                            </Select>
                            
                            <div className="p-4 bg-white/10 rounded-2xl border border-white/10 text-sm">
                                <p className="opacity-80">{t.settings_current_language}:</p>
                                <p className="text-xl font-black mt-1">
                                    {language === 'en' ? 'English' : language === 'ko' ? '한국어' : '日本語'}
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </section>

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

            {/* Bottom Save Action */}
            <Card className="border-none shadow-2xl bg-gradient-to-r from-blue-600 to-indigo-700 text-white overflow-hidden">
                <CardContent className="p-8 flex flex-col md:flex-row items-center justify-between gap-6">
                    <div className="text-center md:text-left space-y-1">
                        <h3 className="text-2xl font-black tracking-tight">{t.settings_save_confirm_title}</h3>
                        <p className="text-blue-100/70 font-medium">{t.settings_save_bottom_desc}</p>
                    </div>
                    <div className="flex gap-3">
                        <Button 
                            variant="outline" 
                            size="lg" 
                            onClick={() => setShowResetConfirm(true)}
                            className="bg-white/10 border-white/20 hover:bg-white/20 text-white font-bold h-14 px-8 rounded-2xl"
                        >
                            {t.refresh}
                        </Button>
                        <Button 
                            size="lg" 
                            onClick={() => setShowSaveConfirm(true)}
                            className="bg-white text-blue-600 hover:bg-blue-50 font-black h-14 px-12 rounded-2xl shadow-xl shadow-blue-900/20"
                        >
                            {t.settings_save_btn}
                        </Button>
                    </div>
                </CardContent>
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
                            <div className="flex items-center justify-between rounded-2xl bg-slate-100 dark:bg-slate-800/60 p-4">
                                <div>
                                    <div className="text-sm font-black uppercase tracking-widest text-slate-700 dark:text-slate-200">Provider Enabled</div>
                                    <div className="text-xs text-slate-500">Disabled provider는 fallback 대상에서 제외됩니다.</div>
                                </div>
                                <Switch
                                    checked={tempProvider.enabled !== false}
                                    onCheckedChange={(checked) => setTempProvider({ ...tempProvider, enabled: checked })}
                                />
                            </div>
                            {addingProvider?.name !== 'ollama' && addingProvider?.name !== 'llamacpp' && (
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
                                    <p className="text-xs text-slate-500">{t.settings_channel_enable_desc}</p>
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
                                        <div className="flex justify-between items-end px-1">
                                            <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none">{t.settings_channel_token}</label>
                                            <Button 
                                                variant="link" 
                                                size="sm" 
                                                className="h-auto p-0 text-[10px] font-bold text-blue-500 hover:text-blue-600 flex items-center gap-1"
                                                onClick={() => { setHelpChannel(activeChannel); setShowHowToGet(true); }}
                                            >
                                                <HelpCircle className="w-3 h-3" /> {t.settings_how_to_get}
                                            </Button>
                                        </div>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                            placeholder="123456:ABC-DEF..."
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
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
                                        <div className="flex justify-between items-end px-1">
                                            <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_token}</label>
                                            <Button 
                                                variant="link" 
                                                size="sm" 
                                                className="h-auto p-0 text-[10px] font-bold text-blue-500 hover:text-blue-600 flex items-center gap-1"
                                                onClick={() => { setHelpChannel(activeChannel); setShowHowToGet(true); }}
                                            >
                                                <HelpCircle className="w-3 h-3" /> {t.settings_how_to_get}
                                            </Button>
                                        </div>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].allow_from?.join(', ') || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'allow_from'], (e.target.value.split(',') as any).map((s: string) => s.trim()).filter((s: string) => s))}
                                            placeholder="UserID1, UserID2"
                                        />
                                    </div>
                                </div>
                            )}

                            {activeChannel === 'slack' && (
                                <div className="space-y-4 animate-in slide-in-from-top-2">
                                    <div className="space-y-2">
                                        <div className="flex justify-between items-end px-1">
                                            <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_token}</label>
                                            <Button 
                                                variant="link" 
                                                size="sm" 
                                                className="h-auto p-0 text-[10px] font-bold text-blue-500 hover:text-blue-600 flex items-center gap-1"
                                                onClick={() => { setHelpChannel(activeChannel); setShowHowToGet(true); }}
                                            >
                                                <HelpCircle className="w-3 h-3" /> {t.settings_how_to_get}
                                            </Button>
                                        </div>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'token'], e.target.value)}
                                            placeholder="xoxb-..."
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_app_token}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].app_token || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'app_token'], e.target.value)}
                                            placeholder="xapp-..."
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].allow_from?.join(', ') || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'allow_from'], (e.target.value.split(',') as any).map((s: string) => s.trim()).filter((s: string) => s))}
                                            placeholder="UserID1, UserID2"
                                        />
                                    </div>
                                </div>
                            )}

                            {activeChannel === 'whatsapp' && (
                                <div className="space-y-4 animate-in slide-in-from-top-2">
                                    <div className="space-y-2">
                                        <div className="flex justify-between items-end px-1">
                                            <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_bridge_url}</label>
                                            <Button 
                                                variant="link" 
                                                size="sm" 
                                                className="h-auto p-0 text-[10px] font-bold text-blue-500 hover:text-blue-600 flex items-center gap-1"
                                                onClick={() => { setHelpChannel(activeChannel); setShowHowToGet(true); }}
                                            >
                                                <HelpCircle className="w-3 h-3" /> {t.settings_how_to_get}
                                            </Button>
                                        </div>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].bridge_url || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'bridge_url'], e.target.value)}
                                            placeholder="ws://localhost:3001"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_api_key}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].api_key || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'api_key'], e.target.value)}
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].allow_from?.join(', ') || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'allow_from'], (e.target.value.split(',') as any).map((s: string) => s.trim()).filter((s: string) => s))}
                                            placeholder="UserID1, UserID2"
                                        />
                                    </div>
                                </div>
                            )}

                            {activeChannel === 'webhook' && (
                                <div className="space-y-4 animate-in slide-in-from-top-2">
                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <div className="flex items-center justify-between">
                                                <label className="text-sm font-bold">{t.settings_channel_port}</label>
                                                <span className="text-[10px] text-slate-400 font-medium italic">{t.settings_channel_port_desc}</span>
                                            </div>
                                            <Input
                                                type="number"
                                                className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none"
                                                value={config.channels.webhook.port || 0}
                                                onChange={(e) => updateConfig(['channels', 'webhook', 'port'], parseInt(e.target.value))}
                                                placeholder="8080"
                                            />
                                        </div>
                                        <div className="space-y-2">
                                            <div className="flex justify-between items-end px-1">
                                                <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_path}</label>
                                                <Button 
                                                    variant="link" 
                                                    size="sm" 
                                                    className="h-auto p-0 text-[10px] font-bold text-blue-500 hover:text-blue-600 flex items-center gap-1"
                                                    onClick={() => { setHelpChannel(activeChannel); setShowHowToGet(true); }}
                                                >
                                                    <HelpCircle className="w-3 h-3" /> {t.settings_how_to_get}
                                                </Button>
                                            </div>
                                            <Input 
                                                className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                                value={config.channels[activeChannel].path || ''} 
                                                onChange={(e) => updateConfig(['channels', activeChannel, 'path'], e.target.value)}
                                                placeholder="/api/channels/webhook"
                                            />
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_secret}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].secret || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'secret'], e.target.value)}
                                            placeholder={t.settings_api_key}
                                            type="password"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">{t.settings_channel_allow_from}</label>
                                        <Input 
                                            className="h-12 shadow-inner bg-slate-50 dark:bg-slate-800 border-none" 
                                            value={config.channels[activeChannel].allow_from?.join(', ') || ''} 
                                            onChange={(e) => updateConfig(['channels', activeChannel, 'allow_from'], (e.target.value.split(',') as any).map((s: string) => s.trim()).filter((s: string) => s))}
                                            placeholder="IP Address or UserID"
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
                onConfirm={() => handleSaveConfig()}
            />

            <ConfirmDialog 
                open={showResetConfirm} 
                onOpenChange={setShowResetConfirm}
                title={t.settings_reset_confirm_title}
                description={t.settings_reset_confirm_desc}
                onConfirm={fetchConfig}
            />

            {/* Save Result Dialog */}
            <Dialog open={showResultDialog} onOpenChange={setShowResultDialog}>
                <DialogContent className="sm:max-w-md border-none shadow-2xl rounded-3xl p-0 overflow-hidden">
                    <div className={cn(
                        "p-8 flex flex-col items-center text-center gap-4",
                        saveResult?.success ? "bg-emerald-50 dark:bg-emerald-900/20" : "bg-red-50 dark:bg-red-900/20"
                    )}>
                        <div className={cn(
                            "w-20 h-20 rounded-full flex items-center justify-center mb-2 animate-in zoom-in duration-500",
                            saveResult?.success ? "bg-emerald-100 dark:bg-emerald-900 text-emerald-600 shadow-lg shadow-emerald-200" : "bg-red-100 dark:bg-red-900 text-red-600 shadow-lg shadow-red-200"
                        )}>
                            {saveResult?.success ? <ShieldCheck size={40} /> : <Trash2 size={40} />}
                        </div>
                        <div className="space-y-2">
                            <DialogTitle className="text-3xl font-black tracking-tight">
                                {saveResult?.success ? t.settings_save_success_title : t.settings_save_error_title}
                            </DialogTitle>
                            <DialogDescription className="text-slate-600 dark:text-slate-400 font-bold text-lg leading-tight uppercase tracking-tighter italic">
                                {saveResult?.success ? t.settings_save_success_desc : saveResult?.message}
                            </DialogDescription>
                        </div>
                    </div>
                    <DialogFooter className="p-6 bg-white dark:bg-slate-900 flex justify-center sm:justify-center border-t border-slate-100 dark:border-slate-800">
                        <Button 
                            className={cn(
                                "w-full sm:w-48 rounded-2xl font-black h-14 text-lg shadow-xl uppercase tracking-widest",
                                saveResult?.success ? "bg-emerald-600 hover:bg-emerald-700 shadow-emerald-200" : "bg-red-600 hover:bg-red-700 shadow-red-200"
                            )}
                            onClick={() => setShowResultDialog(false)}
                        >
                            {t.confirm}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* How to get Token Help Dialog */}
            <Dialog open={showHowToGet} onOpenChange={setShowHowToGet}>
                <DialogContent className="sm:max-w-[450px] border-none shadow-2xl p-0 overflow-hidden rounded-3xl">
                    <div className="p-8 bg-blue-600 text-white space-y-2">
                        <DialogTitle className="text-2xl font-black uppercase tracking-tight flex items-center gap-2">
                            <HelpCircle className="w-6 h-6" /> {helpChannel} Token Guide
                        </DialogTitle>
                        <DialogDescription className="text-blue-100">
                             Follow these steps to generate your bot token.
                        </DialogDescription>
                    </div>
                    <div className="p-8 space-y-6 bg-white dark:bg-slate-900">
                        <div className="text-sm text-slate-600 dark:text-slate-300 whitespace-pre-wrap leading-relaxed bg-slate-50 dark:bg-slate-800 p-6 rounded-2xl border border-slate-100 dark:border-slate-700 italic">
                             {helpChannel === 'telegram' && t.settings_how_to_get_telegram}
                             {helpChannel === 'discord' && t.settings_how_to_get_discord}
                             {helpChannel === 'slack' && t.settings_how_to_get_slack}
                             {helpChannel === 'whatsapp' && t.settings_how_to_get_whatsapp}
                             {helpChannel === 'webhook' && t.settings_how_to_get_webhook}
                        </div>
                        
                        <div className="flex gap-2">
                            {helpChannel === 'telegram' && (
                                <Button className="w-full bg-blue-500 hover:bg-blue-600 text-white font-bold h-12 rounded-xl" asChild>
                                    <a href="https://t.me/botfather" target="_blank" rel="noopener noreferrer">
                                        Open @BotFather <ExternalLink className="w-4 h-4 ml-2" />
                                    </a>
                                </Button>
                            )}
                            {helpChannel === 'discord' && (
                                <Button className="w-full bg-[#5865F2] hover:bg-[#4752c4] text-white font-bold h-12 rounded-xl" asChild>
                                    <a href="https://discord.com/developers/applications" target="_blank" rel="noopener noreferrer">
                                        Open Developer Portal <ExternalLink className="w-4 h-4 ml-2" />
                                    </a>
                                </Button>
                            )}
                            {helpChannel === 'feishu' && (
                                <Button className="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold h-12 rounded-xl" asChild>
                                    <a href="https://open.feishu.cn/app" target="_blank" rel="noopener noreferrer">
                                        Open Developer Console <ExternalLink className="w-4 h-4 ml-2" />
                                    </a>
                                </Button>
                            )}
                        </div>
                    </div>
                    <DialogFooter className="p-6 bg-slate-50 dark:bg-slate-800/50 border-t">
                        <Button variant="outline" onClick={() => setShowHowToGet(false)} className="w-full rounded-xl font-bold">Got it!</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

