import { create } from 'zustand'

export type DetailPaneType = 'log' | 'job' | null;

interface DetailPaneState {
    isOpen: boolean;
    type: DetailPaneType;
    payload: any;
}

interface UIState {
    theme: string;
    isSidebarCollapsed: boolean;
    logViewerWidth: number;
    detailPane: DetailPaneState;

    setTheme: (theme: string) => void;
    toggleSidebar: () => void;
    setLogViewerWidth: (width: number) => void;
    setDetailPane: (isOpen: boolean, type?: DetailPaneType, payload?: any) => void;
    closeDetailPane: () => void;
}

export const useUIStore = create<UIState>((set) => ({
    // Default theme
    theme: 'business',
    // Sidebar is open by default
    isSidebarCollapsed: false,
    // Default width for the log viewer
    logViewerWidth: 320,
    // Detail pane is closed by default
    detailPane: {
        isOpen: false,
        type: null,
        payload: null
    },

    setTheme: (theme) => set({ theme }),
    toggleSidebar: () => set((state) => ({ isSidebarCollapsed: !state.isSidebarCollapsed })),
    setLogViewerWidth: (width) => set({ logViewerWidth: Math.max(200, Math.min(width, 800)) }),
    setDetailPane: (isOpen, type = null, payload = null) => set({
        detailPane: { isOpen, type, payload }
    }),
    closeDetailPane: () => set((state) => ({
        detailPane: { ...state.detailPane, isOpen: false }
    }))
}));
