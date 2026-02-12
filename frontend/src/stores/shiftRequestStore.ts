import { create } from 'zustand';
import type { ShiftRequest } from '../types';
import { shiftRequestApi } from '../api/shiftRequestApi';

interface ShiftRequestState {
  requests: ShiftRequest[];
  loading: boolean;
  error: string | null;
  fetchRequests: (yearMonth: string, staffId?: string) => Promise<void>;
  batchCreate: (requests: Array<{
    staff_id: string;
    year_month: string;
    date: string;
    start_time?: string;
    end_time?: string;
    request_type: string;
    note?: string;
  }>) => Promise<void>;
  deleteRequest: (id: string) => Promise<void>;
}

export const useShiftRequestStore = create<ShiftRequestState>((set) => ({
  requests: [],
  loading: false,
  error: null,
  fetchRequests: async (yearMonth, staffId) => {
    set({ loading: true, error: null });
    try {
      const res = await shiftRequestApi.list(yearMonth, staffId);
      set({ requests: res.data.shift_requests || [], loading: false });
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'シフト希望の取得に失敗しました';
      set({ error: msg, loading: false });
    }
  },
  batchCreate: async (requests) => {
    await shiftRequestApi.batchCreate(requests);
  },
  deleteRequest: async (id) => {
    await shiftRequestApi.delete(id);
  },
}));
