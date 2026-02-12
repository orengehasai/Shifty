import { create } from 'zustand';
import type { Constraint } from '../types';
import { constraintApi } from '../api/constraintApi';

interface ConstraintState {
  constraints: Constraint[];
  loading: boolean;
  error: string | null;
  fetchConstraints: () => Promise<void>;
  createConstraint: (data: {
    name: string;
    type: string;
    category: string;
    config: Record<string, unknown>;
    priority?: number;
  }) => Promise<void>;
  updateConstraint: (id: string, data: Partial<Constraint>) => Promise<void>;
  deleteConstraint: (id: string) => Promise<void>;
}

export const useConstraintStore = create<ConstraintState>((set) => ({
  constraints: [],
  loading: false,
  error: null,
  fetchConstraints: async () => {
    set({ loading: true, error: null });
    try {
      const res = await constraintApi.list();
      set({ constraints: res.data.constraints || [], loading: false });
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '制約の取得に失敗しました';
      set({ error: msg, loading: false });
    }
  },
  createConstraint: async (data) => {
    await constraintApi.create(data);
    const res = await constraintApi.list();
    set({ constraints: res.data.constraints || [] });
  },
  updateConstraint: async (id, data) => {
    await constraintApi.update(id, data);
    const res = await constraintApi.list();
    set({ constraints: res.data.constraints || [] });
  },
  deleteConstraint: async (id) => {
    await constraintApi.delete(id);
    const res = await constraintApi.list();
    set({ constraints: res.data.constraints || [] });
  },
}));
