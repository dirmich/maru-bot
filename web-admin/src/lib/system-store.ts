import { create } from 'zustand';

interface SystemState {
    version: string;
    latest_version: string;
    is_update_available: boolean;
    is_raspberry_pi: boolean;
    is_ai_configured: boolean;
    is_channel_configured: boolean;
    is_loading: boolean;
    setSystemInfo: (info: Partial<SystemState>) => void;
    fetchSystemInfo: () => Promise<void>;
}

export const useSystemStore = create<SystemState>((set) => ({
    version: "v0.0.0",
    latest_version: "",
    is_update_available: false,
    is_raspberry_pi: false,
    is_ai_configured: true, // Default to true to avoid flash of notice
    is_channel_configured: true,
    is_loading: true,
    setSystemInfo: (info) => set((state) => ({ ...state, ...info })),
    fetchSystemInfo: async () => {
        try {
            const resp = await fetch("/api/system/stats");
            if (resp.ok) {
                const data = await resp.json();
                set({
                    version: data.version || "v0.0.0",
                    latest_version: data.latest_version || "",
                    is_update_available: !!data.is_update_available,
                    is_raspberry_pi: !!data.is_raspberry_pi,
                    is_ai_configured: !!data.is_ai_configured,
                    is_channel_configured: !!data.is_channel_configured,
                    is_loading: false
                });
            }
        } catch (error) {
            console.error("Failed to fetch system info", error);
            set({ is_loading: false });
        }
    }
}));
