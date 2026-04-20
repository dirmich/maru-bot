import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { useTranslation } from "@/lib/i18n";
import { useSystemStore } from "@/lib/system-store";
import { AlertCircle, CheckCircle2, XCircle } from "lucide-react";
import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

export function SetupNotice() {
    const t = useTranslation();
    const navigate = useNavigate();
    const location = useLocation();
    const { is_ai_configured, is_channel_configured, is_loading } = useSystemStore();
    const [open, setOpen] = useState(false);

    useEffect(() => {
        // Show if loading is done AND (AI or Channel not configured)
        // BUT don't show if already on settings page to avoid annoyance while configuring
        if (!is_loading && (!is_ai_configured || !is_channel_configured)) {
            if (location.pathname !== '/settings') {
                setOpen(true);
            } else {
                setOpen(false);
            }
        } else {
            setOpen(false);
        }
    }, [is_ai_configured, is_channel_configured, is_loading, location.pathname]);

    const handleGoSettings = () => {
        setOpen(false);
        navigate('/settings');
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="sm:max-w-md border-none shadow-2xl">
                <DialogHeader>
                    <div className="mx-auto w-12 h-12 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center mb-4">
                        <AlertCircle className="w-6 h-6 text-amber-600 dark:text-amber-500" />
                    </div>
                    <DialogTitle className="text-center text-xl">{t.setup_notice_title}</DialogTitle>
                    <DialogDescription className="text-center pt-2">
                        {t.setup_notice_desc}
                    </DialogDescription>
                </DialogHeader>
                
                <div className="space-y-3 py-4">
                    <div className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-900/50 border">
                        <span className="text-sm font-medium">{t.setup_notice_ai.replace('{status}', '')}</span>
                        {is_ai_configured ? (
                            <div className="flex items-center gap-1.5 text-green-600 dark:text-green-500">
                                <CheckCircle2 size={16} />
                                <span className="text-xs font-bold uppercase">{t.setup_notice_configured}</span>
                            </div>
                        ) : (
                            <div className="flex items-center gap-1.5 text-rose-600 dark:text-rose-500">
                                <XCircle size={16} />
                                <span className="text-xs font-bold uppercase">{t.setup_notice_not_configured}</span>
                            </div>
                        )}
                    </div>

                    <div className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-900/50 border">
                        <span className="text-sm font-medium">{t.setup_notice_channel.replace('{status}', '')}</span>
                        {is_channel_configured ? (
                            <div className="flex items-center gap-1.5 text-green-600 dark:text-green-500">
                                <CheckCircle2 size={16} />
                                <span className="text-xs font-bold uppercase">{t.setup_notice_configured}</span>
                            </div>
                        ) : (
                            <div className="flex items-center gap-1.5 text-rose-600 dark:text-rose-500">
                                <XCircle size={16} />
                                <span className="text-xs font-bold uppercase">{t.setup_notice_not_configured}</span>
                            </div>
                        )}
                    </div>
                </div>

                <DialogFooter className="sm:justify-center">
                    <Button 
                        type="button" 
                        onClick={handleGoSettings}
                        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-6 rounded-xl shadow-lg shadow-blue-200 dark:shadow-none transition-all"
                    >
                        {t.setup_notice_go_settings}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
