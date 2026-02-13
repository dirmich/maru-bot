
'use client';

import { useState } from "react";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { Settings, Lock, Mail, Globe } from "lucide-react";
import { useRouter } from "next/navigation";

export default function SetupPage() {
    const [formData, setFormData] = useState({
        adminGmail: "",
        clientId: "",
        clientSecret: "",
        nextAuthSecret: "maru-secret-" + Math.random().toString(36).substring(7),
    });
    const [isLoading, setIsLoading] = useState(false);
    const router = useRouter();

    const handleSave = async () => {
        if (!formData.adminGmail || !formData.clientId || !formData.clientSecret) {
            toast.error("모든 필드를 입력해주세요.");
            return;
        }

        setIsLoading(true);
        try {
            const res = await fetch("/api/setup", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(formData),
            });

            if (res.ok) {
                toast.success("설정이 저장되었습니다. 서버가 재시작될 수 있습니다.");
                setTimeout(() => router.push("/"), 2000);
            } else {
                toast.error("저장에 실패했습니다.");
            }
        } catch (e) {
            toast.error("오류 발생");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950 p-6">
            <Card className="w-full max-w-lg shadow-2xl border-none">
                <CardHeader className="space-y-1 text-center bg-gradient-to-b from-blue-600 to-indigo-700 text-white rounded-t-xl py-8">
                    <div className="w-16 h-16 bg-white/20 rounded-full flex items-center justify-center mx-auto mb-4 backdrop-blur-sm">
                        <Settings className="w-8 h-8" />
                    </div>
                    <CardTitle className="text-2xl font-bold">MaruBot 시스템 초기 설정</CardTitle>
                    <CardDescription className="text-blue-100">
                        Google SSO 연동 및 관리자 설정을 완료해주세요.
                    </CardDescription>
                </CardHeader>
                <CardContent className="p-8 space-y-6">
                    <div className="space-y-4">
                        <div className="space-y-2">
                            <Label className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
                                <Mail className="w-4 h-4" /> 관리자 Gmail 아이디
                            </Label>
                            <Input
                                placeholder="admin@gmail.com"
                                value={formData.adminGmail}
                                onChange={(e) => setFormData({ ...formData, adminGmail: e.target.value })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
                                <Globe className="w-4 h-4" /> Google Client ID
                            </Label>
                            <Input
                                placeholder="GCP 콘솔에서 발급받은 ID"
                                value={formData.clientId}
                                onChange={(e) => setFormData({ ...formData, clientId: e.target.value })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
                                <Lock className="w-4 h-4" /> Google Client Secret
                            </Label>
                            <Input
                                type="password"
                                placeholder="GCP 콘솔에서 발급받은 Secret"
                                value={formData.clientSecret}
                                onChange={(e) => setFormData({ ...formData, clientSecret: e.target.value })}
                            />
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="p-8 border-t bg-slate-50 dark:bg-slate-900 rounded-b-xl">
                    <Button
                        className="w-full h-12 bg-blue-600 hover:bg-blue-700 text-lg font-bold"
                        onClick={handleSave}
                        disabled={isLoading}
                    >
                        {isLoading ? "저장 중..." : "시스템 시작하기"}
                    </Button>
                </CardFooter>
            </Card>
        </div>
    );
}
