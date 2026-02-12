import apiClient from './client';
import type { DashboardSummary } from '../types';

export const dashboardApi = {
  getSummary: (yearMonth: string) =>
    apiClient.get<DashboardSummary>('/dashboard/summary', { params: { year_month: yearMonth } }),
};
