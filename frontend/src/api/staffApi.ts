import apiClient from './client';
import type { Staff } from '../types';

export const staffApi = {
  list: (isActive?: boolean) => {
    const params = isActive !== undefined ? { is_active: isActive } : {};
    return apiClient.get<{ staffs: Staff[] }>('/staffs', { params });
  },
  get: (id: string) => apiClient.get<Staff>(`/staffs/${id}`),
  create: (data: { name: string; role: string; employment_type: string }) =>
    apiClient.post<Staff>('/staffs', data),
  update: (id: string, data: { name?: string; role?: string; employment_type?: string; is_active?: boolean }) =>
    apiClient.put<Staff>(`/staffs/${id}`, data),
  delete: (id: string) => apiClient.delete(`/staffs/${id}`),
};
