import apiClient from './client';
import type { Constraint } from '../types';

export const constraintApi = {
  list: (params?: { is_active?: boolean; type?: string; category?: string }) =>
    apiClient.get<{ constraints: Constraint[] }>('/constraints', { params }),
  create: (data: {
    name: string;
    type: string;
    category: string;
    config: Record<string, unknown>;
    priority?: number;
  }) => apiClient.post<Constraint>('/constraints', data),
  update: (id: string, data: Partial<{
    name: string;
    type: string;
    category: string;
    config: Record<string, unknown>;
    is_active: boolean;
    priority: number;
  }>) => apiClient.put<Constraint>(`/constraints/${id}`, data),
  delete: (id: string) => apiClient.delete(`/constraints/${id}`),
};
