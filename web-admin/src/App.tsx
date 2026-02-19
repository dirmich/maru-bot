import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Providers } from "@/components/providers";

import { AdminLayout } from "@/layouts/AdminLayout";
import { ChatPage } from "@/pages/ChatPage";
import { GpioPage } from "@/pages/GpioPage";
import { SkillsPage } from "@/pages/SkillsPage";
import { SettingsPage } from "@/pages/SettingsPage";

function App() {
    return (
        <Router>
            <Providers>
                <TooltipProvider>
                    <div className="min-h-screen bg-background font-sans antialiased">
                        <Routes>
                            <Route element={<AdminLayout />}>
                                <Route path="/" element={<Navigate to="/chat" replace />} />
                                <Route path="/chat" element={<ChatPage />} />
                                <Route path="/gpio" element={<GpioPage />} />
                                <Route path="/skills" element={<SkillsPage />} />
                                <Route path="/settings" element={<SettingsPage />} />
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
