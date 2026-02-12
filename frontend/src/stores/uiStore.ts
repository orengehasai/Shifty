import { create } from 'zustand';

const now = new Date();
const defaultYearMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;

interface UiState {
  yearMonth: string;
  setYearMonth: (ym: string) => void;
  sidebarOpen: boolean;
  toggleSidebar: () => void;
}

export const useUiStore = create<UiState>((set) => ({
  yearMonth: defaultYearMonth,
  setYearMonth: (ym) => set({ yearMonth: ym }),
  sidebarOpen: true,
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
}));
