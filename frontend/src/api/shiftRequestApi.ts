import apiClient from './client';
import type { ShiftRequest } from '../types';

export const shiftRequestApi = {
  list: (yearMonth: string, staffId?: string) => {
    const params: Record<string, string> = { year_month: yearMonth };
    if (staffId) params.staff_id = staffId;
    return apiClient.get<{ shift_requests: ShiftRequest[] }>('/shift-requests', { params });
  },
  create: (data: {
    staff_id: string;
    year_month: string;
    date: string;
    start_time: string;
    end_time: string;
    request_type: string;
    note?: string;
  }) => apiClient.post<ShiftRequest>('/shift-requests', data),
  batchCreate: (requests: Array<{
    staff_id: string;
    year_month: string;
    date: string;
    start_time?: string;
    end_time?: string;
    request_type: string;
    note?: string;
  }>) => apiClient.post<{ created_count: number; shift_requests: ShiftRequest[] }>(
    '/shift-requests/batch', { requests }
  ),
  update: (id: string, data: Partial<ShiftRequest>) =>
    apiClient.put<ShiftRequest>(`/shift-requests/${id}`, data),
  delete: (id: string) => apiClient.delete(`/shift-requests/${id}`),
};
