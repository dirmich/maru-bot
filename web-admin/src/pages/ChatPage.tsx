import { useState, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from 'sonner';
import { Send, MessageSquare, Trash2 } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation } from "@/lib/i18n";

// Simple ID generator
const generateId = () => Math.random().toString(36).substring(2, 9);

interface Message {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    timestamp: Date;
}

export function ChatPage() {
    const t = useTranslation();
    const [messages, setMessages] = useState<Message[]>([
        {
            id: '1',
            role: 'assistant',
            content: t.chat_welcome,
            timestamp: new Date()
        }
    ]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const scrollViewportRef = useRef<HTMLDivElement>(null);
    const bottomRef = useRef<HTMLDivElement>(null);
    const [showClearConfirm, setShowClearConfirm] = useState(false);

    const scrollToBottom = () => {
        setTimeout(() => {
            bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
        }, 100);
    };

    const handleSendMessage = async (e?: React.FormEvent) => {
        e?.preventDefault();

        if (!input.trim() || isLoading) return;

        const userMsg: Message = {
            id: generateId(),
            role: 'user',
            content: input,
            timestamp: new Date()
        };

        setMessages(prev => [...prev, userMsg]);
        setInput('');
        setIsLoading(true);
        scrollToBottom();

        try {
            const res = await fetch('/api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ message: userMsg.content }),
            });

            if (!res.ok) throw new Error('Network response was not ok');

            const data = await res.json();

            if (data.response) {
                const assistantMsg: Message = {
                    id: generateId(),
                    role: 'assistant',
                    content: data.response,
                    timestamp: new Date()
                };
                setMessages(prev => [...prev, assistantMsg]);
            }
        } catch (error) {
            console.error('Chat error:', error);
            toast.error(t.chat_send_error);

            // Mock response if API fails (for development/demo)
            setTimeout(() => {
                const mockResponse: Message = {
                    id: generateId(),
                    role: 'assistant',
                    content: `현재 서버와 연결할 수 없습니다. 입력하신 내용: "${userMsg.content}"`,
                    timestamp: new Date()
                };
                setMessages(prev => [...prev, mockResponse]);
                setIsLoading(false);
                scrollToBottom();
            }, 1000);
            return;
        } finally {
            setIsLoading(false);
            scrollToBottom();
        }
    };

    const handleClearChat = () => {
        setMessages([]);
        toast.success(t.chat_clear_success);
        setShowClearConfirm(false);
    };

    return (
        <div className="flex flex-col h-screen bg-slate-50 dark:bg-slate-950 p-4 md:p-6 overflow-hidden">
            <header className="mb-4 flex-none">
                <h1 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <MessageSquare className="text-blue-600" /> {t.chat_title}
                </h1>
                <p className="text-sm text-slate-500">{t.chat_desc}</p>
            </header>

            <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden min-h-0 ring-1 ring-slate-900/5">
                <CardHeader className="py-3 px-4 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between flex-none">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        <span className={`w-2 h-2 rounded-full ${isLoading ? 'bg-green-500 animate-pulse' : 'bg-slate-300'}`}></span>
                        {t.chat_live}
                    </CardTitle>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setShowClearConfirm(true)}
                        className="text-slate-500 hover:text-red-500 h-8 w-8 p-0"
                    >
                        <Trash2 className="w-4 h-4" />
                    </Button>
                </CardHeader>

                <CardContent className="flex-1 p-0 overflow-hidden bg-slate-50/50 dark:bg-slate-900/50 relative">
                    <ScrollArea className="h-full w-full p-4" ref={scrollViewportRef}>
                        <div className="flex flex-col space-y-4 pb-4">
                            {messages.length === 0 && (
                                <div className="h-40 flex flex-col items-center justify-center text-slate-400 py-10">
                                    <MessageSquare className="w-12 h-12 mb-2 opacity-20" />
                                    <p>{t.chat_empty_msg}</p>
                                </div>
                            )}

                            {messages.map((m) => (
                                <div
                                    key={m.id}
                                    className={`flex w-full ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}
                                >
                                    <div
                                        className={`max-w-[85%] md:max-w-[75%] rounded-2xl px-4 py-3 text-sm shadow-sm ${m.role === 'user'
                                            ? 'bg-blue-600 text-white rounded-tr-none'
                                            : 'bg-white dark:bg-slate-800 border rounded-tl-none text-slate-800 dark:text-slate-200'
                                            }`}
                                    >
                                        <p className="whitespace-pre-wrap leading-relaxed break-words">{m.content}</p>
                                        <span className={`text-[10px] block mt-1 ${m.role === 'user' ? 'text-blue-100' : 'text-slate-400'}`}>
                                            {m.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                        </span>
                                    </div>
                                </div>
                            ))}

                            {isLoading && (
                                <div className="flex justify-start w-full animate-in fade-in slide-in-from-bottom-2 duration-300">
                                    <div className="bg-white dark:bg-slate-800 border rounded-2xl rounded-tl-none px-4 py-3 text-sm shadow-sm flex items-center gap-2">
                                        <div className="flex gap-1">
                                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
                                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
                                            <span className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-bounce"></span>
                                        </div>
                                        <span className="text-slate-500 text-xs">{t.chat_thinking}</span>
                                    </div>
                                </div>
                            )}
                            <div ref={bottomRef} />
                        </div>
                    </ScrollArea>
                </CardContent>

                <CardFooter className="p-3 md:p-4 border-t bg-white dark:bg-slate-900 flex-none">
                    <form
                        className="w-full flex gap-2"
                        onSubmit={handleSendMessage}
                    >
                        <Input
                            placeholder={t.chat_input_placeholder}
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            disabled={isLoading}
                            className="flex-1 bg-slate-50 dark:bg-slate-800 border-none ring-offset-transparent focus-visible:ring-1 focus-visible:ring-blue-500"
                        />
                        <Button type="submit" disabled={isLoading || !input.trim()} size="icon" className="bg-blue-600 hover:bg-blue-700 shrink-0">
                            <Send className="w-4 h-4" />
                        </Button>
                    </form>
                </CardFooter>
            </Card>

            <ConfirmDialog
                open={showClearConfirm}
                onOpenChange={setShowClearConfirm}
                title={t.chat_clear_confirm_title}
                description={t.chat_clear_confirm_desc}
                onConfirm={handleClearChat}
            />
        </div>
    );
}
