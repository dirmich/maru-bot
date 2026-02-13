'use client';

import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from 'sonner';
import { Send, MessageSquare, Trash2 } from 'lucide-react';
import { useConfirmDialog } from '@/components/ui-custom-dialog';

export default function ChatPage() {
    const [messages, setMessages] = useState<any[]>([]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);
    const confirm = useConfirmDialog();

    useEffect(() => {
        fetchMessages();
    }, []);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const fetchMessages = async () => {
        const res = await fetch('/api/chat');
        const data = await res.json();
        setMessages(data);
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
        <div className="flex flex-col h-full bg-slate-50 dark:bg-slate-950 p-6">
            <header className="mb-6 flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                        <MessageSquare className="text-blue-600" /> AI 어시스턴트
                    </h1>
                    <p className="text-sm text-slate-500">에이전트와 실시간으로 대화하세요.</p>
                </div>
            </header>

            <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden min-h-0">
                <CardHeader className="py-3 px-4 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between">
                    <CardTitle className="text-sm font-medium">실시간 대화</CardTitle>
                    <Button variant="ghost" size="sm" onClick={handleClearChat} className="text-slate-500 hover:text-red-500 h-8 w-8 p-0">
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
        </div>
    );
}
