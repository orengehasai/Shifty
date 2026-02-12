import apiClient from './client';
import type { StaffMonthlySetting } from '../types';

export const monthlySettingsApi = {
  list: (yearMonth: string, staffId?: string) => {
    const params: Record<string, string> = { year_month: yearMonth };
    if (staffId) params.staff_id = staffId;
    return apiClient.get<{ settings: StaffMonthlySetting[] }>('/staff-monthly-settings', { params });
  },
  create: (data: {
    staff_id: string;
    year_month: string;
    min_preferred_hours: number;
    max_preferred_hours: number;
    note?: string;
  }) => apiClient.post<StaffMonthlySetting>('/staff-monthly-settings', data),
  batchCreate: (settings: Array<{
    staff_id: string;
    year_month: string;
    min_preferred_hours: number;
    max_preferred_hours: number;
    note?: string;
  }>) => apiClient.post<{ created_count: number; settings: StaffMonthlySetting[] }>(
    '/staff-monthly-settings/batch', { settings }
  ),
  update: (id: string, data: Partial<{
    min_preferred_hours: number;
    max_preferred_hours: number;
    note: string;
  }>) => apiClient.put<StaffMonthlySetting>(`/staff-monthly-settings/${id}`, data),
  delete: (id: string) => apiClient.delete(`/staff-monthly-settings/${id}`),
};
