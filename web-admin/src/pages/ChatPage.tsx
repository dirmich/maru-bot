import { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from 'sonner';
import { Send, MessageSquare, Trash2, Calendar, History, Plus } from 'lucide-react';
import { ConfirmDialog } from "@/components/ui-custom-dialog";
import { useTranslation } from "@/lib/i18n";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import { authenticatedFetch } from "@/lib/auth";

// Simple ID generator
const generateId = () => Math.random().toString(36).substring(2, 9);

interface Message {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    timestamp: string; // Store as string for JSON
}

const CHAT_STORAGE_KEY = 'marubot_chat_history';

export function ChatPage() {
    const t = useTranslation();
    const [messages, setMessages] = useState<Message[]>([]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const scrollViewportRef = useRef<HTMLDivElement>(null);
    const bottomRef = useRef<HTMLDivElement>(null);
    const [showClearConfirm, setShowClearConfirm] = useState(false);
    const [isLoaded, setIsLoaded] = useState(false);
    
    // History states
    const [days, setDays] = useState<string[]>([]);
    const [selectedDate, setSelectedDate] = useState<string | null>(null); // null means "Live Current Session"
    const [isHistoryLoading, setIsHistoryLoading] = useState(false);

    // Initial load
    useEffect(() => {
        loadHistoryDays();
        
        const saved = localStorage.getItem(CHAT_STORAGE_KEY);
        if (saved) {
            try {
                setMessages(JSON.parse(saved));
            } catch (e) {
                console.error('Failed to parse chat history:', e);
            }
        } else {
            // Default welcome message
            setMessages([
                {
                    id: '1',
                    role: 'assistant',
                    content: t.chat_welcome,
                    timestamp: new Date().toISOString()
                }
            ]);
        }
        setIsLoaded(true);
    }, []);

    // Save current session to localStorage Only if we are NOT in history mode
    useEffect(() => {
        if (isLoaded && selectedDate === null) {
            localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(messages));
        }
    }, [messages, isLoaded, selectedDate]);

    const loadHistoryDays = async () => {
        try {
            const res = await authenticatedFetch('/api/history/chat/days');
            if (res.ok) {
                const data = await res.json();
                setDays(data.days || []);
            }
        } catch (e) {
            console.error("Failed to load history days", e);
        }
    };

    const loadHistoryForDate = async (date: string) => {
        setIsHistoryLoading(true);
        setSelectedDate(date);
        try {
            const res = await authenticatedFetch(`/api/history/chat/day?date=${date}`);
            if (res.ok) {
                const data = await res.json();
                setMessages(data.messages || []);
                scrollToBottom();
            }
        } catch (e) {
            toast.error("Failed to load history");
            console.error(e);
        } finally {
            setIsHistoryLoading(false);
        }
    };

    const handleNewChat = () => {
        setSelectedDate(null);
        const saved = localStorage.getItem(CHAT_STORAGE_KEY);
        if (saved) {
            setMessages(JSON.parse(saved));
        } else {
            setMessages([
                {
                    id: generateId(),
                    role: 'assistant',
                    content: t.chat_welcome,
                    timestamp: new Date().toISOString()
                }
            ]);
        }
    };

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
            timestamp: new Date().toISOString()
        };

        // If we are viewing history, and we send a message, we should switch to Live session or at least warn
        if (selectedDate !== null) {
            // Just move to live for now as a simple UX
            setSelectedDate(null);
            // Prefix the previous messages if appropriate? No, just start new/resume live.
        }

        setMessages(prev => [...prev, userMsg]);
        setInput('');
        setIsLoading(true);
        scrollToBottom();

        try {
            const res = await authenticatedFetch('/api/chat', {
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
                    timestamp: new Date().toISOString()
                };
                setMessages(prev => [...prev, assistantMsg]);
                loadHistoryDays(); // Refresh days list as server just saved
            }
        } catch (error) {
            console.error('Chat error:', error);
            toast.error(t.chat_send_error);
            setIsLoading(false);
            scrollToBottom();
        } finally {
            setIsLoading(false);
            scrollToBottom();
        }
    };

    const handleClearChat = () => {
        if (selectedDate) {
            // Delete from server
            authenticatedFetch(`/api/history/chat/day?date=${selectedDate}`, { method: 'DELETE' })
                .then(() => {
                    toast.success("History deleted");
                    loadHistoryDays();
                    handleNewChat();
                });
        } else {
            setMessages([]);
            localStorage.removeItem(CHAT_STORAGE_KEY);
            toast.success(t.chat_clear_success);
        }
        setShowClearConfirm(false);
    };

    return (
        <div className="flex h-screen bg-slate-50 dark:bg-slate-950 overflow-hidden">
            {/* Sidebar for History */}
            <aside className="w-64 border-r bg-white dark:bg-slate-900 flex flex-col flex-none hidden md:flex">
                <div className="p-4 border-b">
                    <Button 
                        onClick={handleNewChat}
                        variant={selectedDate === null ? "default" : "outline"} 
                        className="w-full justify-start gap-2 rounded-xl"
                    >
                        <Plus size={18} /> {t.chat_new || "New Chat"}
                    </Button>
                </div>
                <ScrollArea className="flex-1">
                    <div className="p-2 space-y-1">
                        <div className="px-3 py-2 text-[10px] font-black uppercase tracking-widest text-slate-400">
                            Conversation History
                        </div>
                        {days.map(day => (
                            <div 
                                key={day}
                                className={`group relative w-full flex items-center rounded-xl transition-all ${
                                    selectedDate === day 
                                    ? "bg-blue-50 dark:bg-blue-900/20" 
                                    : "hover:bg-slate-50 dark:hover:bg-slate-800"
                                }`}
                            >
                                <button
                                    onClick={() => loadHistoryForDate(day)}
                                    className={`flex-1 flex items-center gap-3 px-3 py-2.5 text-sm text-left ${
                                        selectedDate === day 
                                        ? "text-blue-600 dark:text-blue-400 font-bold" 
                                        : "text-slate-500"
                                    }`}
                                >
                                    <Calendar size={16} className={selectedDate === day ? "text-blue-500" : "text-slate-400"} />
                                    <span className="flex-1 truncate">{day}</span>
                                </button>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setSelectedDate(day);
                                        setShowClearConfirm(true);
                                    }}
                                    className="h-8 w-8 mr-1 opacity-0 group-hover:opacity-100 text-slate-400 hover:text-red-500 transition-opacity"
                                >
                                    <Trash2 size={14} />
                                </Button>
                            </div>
                        ))}
                    </div>
                </ScrollArea>
            </aside>

            <div className="flex-1 flex flex-col min-w-0 p-4 md:p-6 overflow-hidden">
                <header className="mb-4 flex items-center justify-between flex-none">
                    <div>
                        <h1 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                            <MessageSquare className="text-blue-600" /> {t.chat_title}
                        </h1>
                        <p className="text-sm text-slate-500">
                            {selectedDate ? `Viewing History: ${selectedDate}` : t.chat_desc}
                        </p>
                    </div>
                    {/* Mobile History View Button could go here */}
                </header>

                <Card className="flex-1 flex flex-col border-none shadow-lg overflow-hidden min-h-0 ring-1 ring-slate-900/5">
                    <CardHeader className="py-3 px-4 border-b bg-white dark:bg-slate-900 flex flex-row items-center justify-between flex-none">
                        <CardTitle className="text-sm font-medium flex items-center gap-2">
                            <span className={`w-2 h-2 rounded-full ${isLoading || isHistoryLoading ? 'bg-green-500 animate-pulse' : 'bg-slate-300'}`}></span>
                            {selectedDate ? "Archived Chat" : t.chat_live}
                        </CardTitle>
                        <div className="flex gap-2">
                             <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setShowClearConfirm(true)}
                                className="text-slate-500 hover:text-red-500 h-8 w-8 p-0"
                            >
                                <Trash2 className="w-4 h-4" />
                            </Button>
                        </div>
                    </CardHeader>

                    <CardContent className="flex-1 p-0 overflow-hidden bg-slate-50/50 dark:bg-slate-900/50 relative">
                        <ScrollArea className="h-full w-full p-4" ref={scrollViewportRef}>
                            <div className="flex flex-col space-y-4 pb-4">
                                {messages.length === 0 && !isHistoryLoading && (
                                    <div className="h-40 flex flex-col items-center justify-center text-slate-400 py-10">
                                        <History className="w-12 h-12 mb-2 opacity-20" />
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
                                            {m.role === 'assistant' ? (
                                                <div className="markdown-content overflow-x-auto">
                                                    <ReactMarkdown 
                                                        remarkPlugins={[remarkGfm]}
                                                        rehypePlugins={[rehypeRaw]}
                                                    >
                                                        {m.content}
                                                    </ReactMarkdown>
                                                </div>
                                            ) : (
                                                <p className="whitespace-pre-wrap leading-relaxed break-words">{m.content}</p>
                                            )}
                                            <span className={`text-[10px] block mt-1 ${m.role === 'user' ? 'text-blue-100' : 'text-slate-400'}`}>
                                                {new Date(m.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
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
                                placeholder={selectedDate ? "History Mode - Switch to Live to chat" : t.chat_input_placeholder}
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
                    title={selectedDate ? "Delete History?" : t.chat_clear_confirm_title}
                    description={selectedDate ? `Are you sure you want to delete conversation history for ${selectedDate}?` : t.chat_clear_confirm_desc}
                    onConfirm={handleClearChat}
                />
            </div>
        </div>
    );
}
