import { create } from 'zustand';
import type { ShiftPattern, GenerationJob } from '../types';
import { shiftApi } from '../api/shiftApi';

interface ShiftState {
  patterns: ShiftPattern[];
  currentPattern: ShiftPattern | null;
  generationJob: GenerationJob | null;
  loading: boolean;
  error: string | null;
  fetchPatterns: (yearMonth: string) => Promise<void>;
  fetchPattern: (id: string) => Promise<void>;
  startGeneration: (yearMonth: string, patternCount: number) => Promise<string>;
  pollJobStatus: (jobId: string) => Promise<GenerationJob>;
  selectPattern: (id: string) => Promise<void>;
  finalizePattern: (id: string) => Promise<void>;
  updateEntry: (entryId: string, data: { start_time: string; end_time: string; break_minutes: number }) => Promise<{ is_valid: boolean; warnings: Array<{ type: string; message: string }> }>;
  deleteEntry: (entryId: string) => Promise<void>;
  createEntry: (data: { pattern_id: string; staff_id: string; date: string; start_time: string; end_time: string; break_minutes: number }) => Promise<void>;
}

export const useShiftStore = create<ShiftState>((set) => ({
  patterns: [],
  currentPattern: null,
  generationJob: null,
  loading: false,
  error: null,
  fetchPatterns: async (yearMonth) => {
    set({ loading: true, error: null });
    try {
      const res = await shiftApi.listPatterns(yearMonth);
      set({ patterns: res.data.patterns || [], loading: false });
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'パターンの取得に失敗しました';
      set({ error: msg, loading: false });
    }
  },
  fetchPattern: async (id) => {
    set({ loading: true, error: null });
    try {
      const res = await shiftApi.getPattern(id);
      set({ currentPattern: res.data.pattern, loading: false });
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'パターン詳細の取得に失敗しました';
      set({ error: msg, loading: false });
    }
  },
  startGeneration: async (yearMonth, patternCount) => {
    set({ loading: true, error: null });
    const res = await shiftApi.generate(yearMonth, patternCount);
    set({ generationJob: { id: res.data.job_id, status: 'pending', year_month: yearMonth, pattern_count: patternCount, progress: 0, status_message: null, started_at: new Date().toISOString(), completed_at: null, error_message: null, created_at: new Date().toISOString() }, loading: false });
    return res.data.job_id;
  },
  pollJobStatus: async (jobId) => {
    const res = await shiftApi.getJobStatus(jobId);
    set({ generationJob: res.data });
    return res.data;
  },
  selectPattern: async (id) => {
    await shiftApi.selectPattern(id);
  },
  finalizePattern: async (id) => {
    await shiftApi.finalizePattern(id);
  },
  updateEntry: async (entryId, data) => {
    const res = await shiftApi.updateEntry(entryId, data);
    return res.data.validation;
  },
  deleteEntry: async (entryId) => {
    await shiftApi.deleteEntry(entryId);
  },
  createEntry: async (data) => {
    await shiftApi.createEntry(data);
  },
}));
