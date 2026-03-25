import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Providers } from "@/components/providers";

import { AdminLayout } from "@/layouts/AdminLayout";
import { ChatPage } from "@/pages/ChatPage";
import { GpioPage } from "@/pages/GpioPage";
import { SkillsPage } from "@/pages/SkillsPage";
import { SettingsPage } from "@/pages/SettingsPage";
import { LogsPage } from "@/pages/LogsPage";
import { LoginPage } from "@/pages/LoginPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { isAuthenticated } from "@/lib/auth";
import { useSystemStore } from "@/lib/system-store";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    if (!isAuthenticated()) {
        return <Navigate to="/login" replace />;
    }
    return <>{children}</>;
}

function GpioRoute({ children }: { children: React.ReactNode }) {
    const { is_raspberry_pi, is_loading } = useSystemStore();
    if (is_loading) return null;
    if (!is_raspberry_pi) {
        return <Navigate to="/dashboard" replace />;
    }
    return <>{children}</>;
}

function App() {
    return (
        <Router>
            <Providers>
                <TooltipProvider>
                    <div className="min-h-screen bg-background font-sans antialiased">
                        <Routes>
                            <Route path="/login" element={<LoginPage />} />
                            <Route element={<ProtectedRoute><AdminLayout /></ProtectedRoute>}>
                                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                                <Route path="/dashboard" element={<DashboardPage />} />
                                <Route path="/chat" element={<ChatPage />} />
                                <Route path="/gpio" element={<GpioRoute><GpioPage /></GpioRoute>} />
                                <Route path="/skills" element={<SkillsPage />} />
                                <Route path="/settings" element={<SettingsPage />} />
                                <Route path="/logs" element={<LogsPage />} />
                            </Route>
                        </Routes>
                        <Toaster />
                    </div>
                </TooltipProvider>
            </Providers>
        </Router>
    )
}

export default App
