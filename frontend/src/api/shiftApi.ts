import apiClient from './client';
import type { GenerationJob, ShiftPattern, ShiftEntry } from '../types';

export const shiftApi = {
  generate: (yearMonth: string, patternCount: number) =>
    apiClient.post<{ job_id: string; status: string; message: string }>(
      '/shifts/generate', { year_month: yearMonth, pattern_count: patternCount }
    ),
  getJobStatus: (jobId: string) =>
    apiClient.get<GenerationJob>(`/shifts/generate/${jobId}`),
  listPatterns: (yearMonth: string) =>
    apiClient.get<{ patterns: ShiftPattern[] }>('/shifts/patterns', { params: { year_month: yearMonth } }),
  getPattern: (id: string) =>
    apiClient.get<{ pattern: ShiftPattern }>(`/shifts/patterns/${id}`),
  selectPattern: (id: string) =>
    apiClient.put<{ pattern: ShiftPattern }>(`/shifts/patterns/${id}/select`),
  finalizePattern: (id: string) =>
    apiClient.put<{ pattern: ShiftPattern }>(`/shifts/patterns/${id}/finalize`),
  updateEntry: (id: string, data: { start_time: string; end_time: string; break_minutes: number }) =>
    apiClient.put<{ entry: ShiftEntry; validation: { is_valid: boolean; warnings: Array<{ type: string; message: string }> } }>(
      `/shifts/entries/${id}`, data
    ),
  createEntry: (data: {
    pattern_id: string;
    staff_id: string;
    date: string;
    start_time: string;
    end_time: string;
    break_minutes: number;
  }) => apiClient.post<ShiftEntry>('/shifts/entries', data),
  deleteEntry: (id: string) => apiClient.delete(`/shifts/entries/${id}`),
};
