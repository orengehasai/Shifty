import { create } from 'zustand';
import type { Staff } from '../types';
import { staffApi } from '../api/staffApi';

interface StaffState {
  staffs: Staff[];
  loading: boolean;
  error: string | null;
  fetchStaffs: () => Promise<void>;
  createStaff: (data: { name: string; role: string; employment_type: string }) => Promise<void>;
  updateStaff: (id: string, data: { name?: string; role?: string; employment_type?: string; is_active?: boolean }) => Promise<void>;
  deleteStaff: (id: string) => Promise<void>;
}

export const useStaffStore = create<StaffState>((set) => ({
  staffs: [],
  loading: false,
  error: null,
  fetchStaffs: async () => {
    set({ loading: true, error: null });
    try {
      const res = await staffApi.list();
      set({ staffs: res.data.staffs || [], loading: false });
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'スタッフの取得に失敗しました';
      set({ error: msg, loading: false });
    }
  },
  createStaff: async (data) => {
    await staffApi.create(data);
    const res = await staffApi.list();
    set({ staffs: res.data.staffs || [] });
  },
  updateStaff: async (id, data) => {
    await staffApi.update(id, data);
    const res = await staffApi.list();
    set({ staffs: res.data.staffs || [] });
  },
  deleteStaff: async (id) => {
    await staffApi.delete(id);
    const res = await staffApi.list();
    set({ staffs: res.data.staffs || [] });
  },
}));
