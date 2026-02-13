'use client';

import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Switch } from '@/components/ui/switch';
import { Toaster } from '@/components/ui/sonner';
import { toast } from 'sonner';
import { Send, Settings, Package, MessageSquare, RefreshCw, Trash2, Cpu, Wrench, Terminal, Plus, Trash } from 'lucide-react';
import { AlertDialog, useAlertDialog, ConfirmDialog, useConfirmDialog } from '@/components/ui-custom-dialog';

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState('chat');
  const [messages, setMessages] = useState<any[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [config, setConfig] = useState<any>(null);
  const [skills, setSkills] = useState<string>('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const alert = useAlertDialog();
  const confirm = useConfirmDialog();

  useEffect(() => {
    fetchMessages();
    fetchConfig();
    fetchSkills();
  }, []);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const fetchMessages = async () => {
    try {
      const res = await fetch('/api/chat');
      const data = await res.json();
      setMessages(data);
    } catch (e) {
      console.error(e);
    }
  };

  const fetchConfig = async () => {
    const res = await fetch('/api/config');
    if (res.ok) {
      const data = await res.json();
      setConfig(data);
    }
  };

  const fetchSkills = async () => {
    const res = await fetch('/api/skills');
    const data = await res.json();
    setSkills(data.output || 'Skills list empty or error');
  };

  const handleSendMessage = async () => {
    if (!input.trim() || isLoading) return;

    const userMsg = { role: 'user', content: input };
    setMessages([...messages, userMsg]);
    setInput('');
    setIsLoading(true);

    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: input }),
      });
      const data = await res.json();
      if (data.response) {
        setMessages(prev => [...prev, { role: 'assistant', content: data.response }]);
      }
    } catch (error) {
      toast.error('메시지 전송에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveConfig = () => {
    confirm.show(
      "설정 저장",
      "변경사항을 저장하시겠습니까?",
      async () => {
        try {
          const res = await fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config),
          });
          if (res.ok) {
            toast.success('설정이 저장되었습니다.');
          }
        } catch (error) {
          toast.error('설정 저장에 실패했습니다.');
        }
      }
    );
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

  const handleClearChat = () => {
    confirm.show(
      "채팅 내역 삭제",
      "모든 채팅 내역이 삭제됩니다. 계속하시겠습니까?",
      async () => {
        setMessages([]);
        toast.success('채팅 내역이 초기화되었습니다.');
      }
    );
  };

  return (
    <div className="flex flex-col h-screen bg-slate-50 dark:bg-slate-950 font-sans">
      <AlertDialog />
      <ConfirmDialog />
      <header className="border-b bg-white dark:bg-slate-900 px-6 py-4 flex justify-between items-center shadow-sm">
        <div className="flex items-center gap-2">
          <div className="bg-blue-600 p-1.5 rounded-lg">
            <span className="text-xl">🦞</span>
          </div>
          <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
            MaruBot Admin
          </h1>
        </div>
        <div className="flex items-center gap-4">
          <Button variant="outline" size="sm" onClick={() => { fetchConfig(); fetchSkills(); }}>
            <RefreshCw className="w-4 h-4 mr-2" />
            새로고침
          </Button>
        </div>
      </header>

      <main className="flex-1 overflow-hidden p-6 max-w-6xl mx-auto w-full">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
          <TabsList className="grid grid-cols-3 mb-6 bg-slate-200/50 dark:bg-slate-800/50 p-1">
            <TabsTrigger value="chat" className="flex items-center gap-2 data-[state=active]:bg-white data-[state=active]:shadow-sm">
              <MessageSquare className="w-4 h-4" />
              채팅
            </TabsTrigger>
            <TabsTrigger value="settings" className="flex items-center gap-2 data-[state=active]:bg-white data-[state=active]:shadow-sm">
              <Settings className="w-4 h-4" />
              설정
            </TabsTrigger>
            <TabsTrigger value="skills" className="flex items-center gap-2 data-[state=active]:bg-white data-[state=active]:shadow-sm">
              <Package className="w-4 h-4" />
              스킬 & 툴
            </TabsTrigger>
          </TabsList>

          <TabsContent value="chat" className="flex-1 overflow-hidden mt-0">
            <Card className="h-full flex flex-col border-none shadow-lg overflow-hidden">
              <CardHeader className="py-3 px-4 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-medium">에이전트 채팅</CardTitle>
                </div>
                <Button variant="ghost" size="sm" onClick={handleClearChat} className="text-slate-500 hover:text-red-500">
                  <Trash2 className="w-4 h-4" />
                </Button>
              </CardHeader>
              <CardContent className="flex-1 p-0 overflow-hidden bg-slate-50/50 dark:bg-slate-900/50">
                <ScrollArea className="h-full p-4" viewportRef={scrollRef}>
                  <div className="space-y-4 pb-4">
                    {messages.length === 0 && (
                      <div className="h-full flex flex-col items-center justify-center text-slate-400 py-20">
                        <MessageSquare className="w-12 h-12 mb-2 opacity-20" />
                        <p>메시지를 입력하여 대화를 시작하세요.</p>
                      </div>
                    )}
                    {messages.map((m, i) => (
                      <div
                        key={i}
                        className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}
                      >
                        <div
                          className={`max-w-[80%] rounded-2xl px-4 py-2 text-sm shadow-sm ${m.role === 'user'
                              ? 'bg-blue-600 text-white rounded-tr-none'
                              : 'bg-white dark:bg-slate-800 border rounded-tl-none text-slate-800 dark:text-slate-200'
                            }`}
                        >
                          <p className="whitespace-pre-wrap leading-relaxed">{m.content}</p>
                        </div>
                      </div>
                    ))}
                    {isLoading && (
                      <div className="flex justify-start">
                        <div className="bg-white dark:bg-slate-800 border rounded-2xl rounded-tl-none px-4 py-2 text-sm shadow-sm flex items-center gap-2">
                          <div className="flex gap-1">
                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce"></span>
                          </div>
                          <span className="text-slate-500">생각 중...</span>
                        </div>
                      </div>
                    )}
                  </div>
                </ScrollArea>
              </CardContent>
              <CardFooter className="p-4 border-t bg-white dark:bg-slate-900">
                <form
                  className="w-full flex gap-2"
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSendMessage();
                  }}
                >
                  <Input
                    placeholder="메시지를 입력하세요..."
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    className="flex-1 bg-slate-50 dark:bg-slate-800 border-none ring-offset-transparent focus-visible:ring-1 focus-visible:ring-blue-500"
                  />
                  <Button type="submit" disabled={isLoading || !input.trim()} className="bg-blue-600 hover:bg-blue-700">
                    <Send className="w-4 h-4" />
                  </Button>
                </form>
              </CardFooter>
            </Card>
          </TabsContent>

          <TabsContent value="settings" className="flex-1 overflow-y-auto mt-0 pr-2">
            <div className="grid gap-6 pb-6">
              <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                  <CardTitle className="flex items-center gap-2 text-blue-600">
                    <Cpu className="w-5 h-5" />
                    기본 설정
                  </CardTitle>
                  <CardDescription>메인 에이전트 설정 및 모델을 관리합니다.</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">기본 모델</label>
                      <Input
                        value={config?.agents?.defaults?.model || ''}
                        placeholder="예: gemini-1.5-flash"
                        onChange={(e) => setConfig({
                          ...config,
                          agents: {
                            ...config.agents,
                            defaults: { ...config.agents.defaults, model: e.target.value }
                          }
                        })}
                      />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-slate-700 dark:text-slate-300">워크스페이스 경로</label>
                      <Input
                        value={config?.agents?.defaults?.workspace || ''}
                        onChange={(e) => setConfig({
                          ...config,
                          agents: {
                            ...config.agents,
                            defaults: { ...config.agents.defaults, workspace: e.target.value }
                          }
                        })}
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="border-none shadow-md overflow-hidden">
                <CardHeader className="bg-white dark:bg-slate-900 border-b">
                  <CardTitle className="flex items-center gap-2 text-indigo-600">
                    <Wrench className="w-5 h-5" />
                    API 제공자
                  </CardTitle>
                  <CardDescription>각 AI 서비스의 API 키를 설정합니다.</CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-6">
                  {config?.providers && Object.entries(config.providers).map(([name, prov]: [string, any]) => (
                    <div key={name} className="space-y-3 group">
                      <div className="flex items-center gap-2">
                        <div className="w-1.5 h-4 bg-indigo-500 rounded-full"></div>
                        <span className="font-bold uppercase text-xs tracking-wider text-slate-500">{name}</span>
                      </div>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-1">
                          <span className="text-[10px] font-medium text-slate-400 ml-1">API KEY</span>
                          <Input
                            placeholder="API Key 입력"
                            type="password"
                            value={prov.api_key || ''}
                            className="bg-slate-50 dark:bg-slate-800' border-slate-200"
                            onChange={(e) => {
                              const newProv = { ...prov, api_key: e.target.value };
                              setConfig({
                                ...config,
                                providers: { ...config.providers, [name]: newProv }
                              });
                            }}
                          />
                        </div>
                        <div className="space-y-1">
                          <span className="text-[10px] font-medium text-slate-400 ml-1">API BASE (URL)</span>
                          <Input
                            placeholder="기본값 사용 (비워둠)"
                            value={prov.api_base || ''}
                            className="bg-slate-50 dark:bg-slate-800 border-slate-200"
                            onChange={(e) => {
                              const newProv = { ...prov, api_base: e.target.value };
                              setConfig({
                                ...config,
                                providers: { ...config.providers, [name]: newProv }
                              });
                            }}
                          />
                        </div>
                      </div>
                    </div>
                  ))}
                </CardContent>
                <CardFooter className="p-4 border-t bg-slate-50 dark:bg-slate-900 justify-end">
                  <Button onClick={handleSaveConfig} className="bg-indigo-600 hover:bg-indigo-700">설정 저장하기</Button>
                </CardFooter>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="skills" className="flex-1 overflow-hidden mt-0">
            < Card className="h-full flex flex-col border-none shadow-lg overflow-hidden">
              <CardHeader className="py-4 px-6 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-emerald-500" />
                    설치된 스킬 및 도구
                  </CardTitle>
                </div>
                <div className="flex gap-2">
                  <Input id="skillInstall" placeholder="사용자/저장소" className="h-9 w-48 text-sm" />
                  <Button size="sm" onClick={() => {
                    const el = document.getElementById('skillInstall') as HTMLInputElement;
                    if (el.value) handleSkillAction('install', el.value);
                  }} className="bg-emerald-600 hover:bg-emerald-700">
                    <Plus className="w-4 h-4 mr-1" />
                    설치
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="flex-1 p-0 overflow-hidden bg-slate-950 text-emerald-400 font-mono text-xs">
                <ScrollArea className="h-full">
                  <pre className="p-6 whitespace-pre-wrap leading-relaxed">{skills}</pre>
                </ScrollArea>
              </CardContent>
              <CardFooter className="p-3 border-t bg-slate-900 text-[10px] text-slate-500 justify-between">
                <span>System: maruminibot skills list output</span>
                <span className="flex items-center gap-1">
                  <RefreshCw className="w-3 h-3" />
                  실시간 연동 중
                </span>
              </CardFooter>
            </Card>
          </TabsContent>
        </Tabs>
      </main>
      <footer className="px-6 py-3 border-t bg-white dark:bg-slate-900 text-[10px] text-slate-400 flex justify-between">
        <p>© 2026 MaruBot Engine v1.0.0</p>
        <p>Running on Raspberry Pi Mode</p>
      </footer>
      <Toaster position="top-right" richColors />
    </div>
  );
}
