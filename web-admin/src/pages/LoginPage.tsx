import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";

export function LoginPage() {
    const t = useTranslation();
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ password }),
            });

            if (response.ok) {
                localStorage.setItem('marubot_auth', 'true');
                toast.success(t.login_success);
                navigate('/chat');
            } else {
                toast.error(t.login_failed);
            }
        } catch (error) {
            toast.error(t.conn_error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-muted/40 p-4">
            <Card className="w-full max-w-sm shadow-xl border-none ring-1 ring-slate-900/5">
                <CardHeader className="space-y-1 text-center">
                    <CardTitle className="text-2xl font-bold flex items-center justify-center gap-2">
                        <span>🦞 MaruBot</span>
                    </CardTitle>
                    <CardDescription>
                        {t.login_title}
                    </CardDescription>
                </CardHeader>
                <form onSubmit={handleLogin}>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="password">{t.admin_password}</Label>
                            <Input
                                id="password"
                                type="password"
                                autoFocus
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="••••••••"
                                className="bg-slate-50 dark:bg-slate-800"
                            />
                        </div>
                    </CardContent>
                    <CardFooter>
                        <Button className="w-full bg-blue-600 hover:bg-blue-700" type="submit" disabled={loading}>
                            {loading ? t.logging_in : t.login_btn}
                        </Button>
                    </CardFooter>
                </form>
            </Card>
        </div>
    );
};
