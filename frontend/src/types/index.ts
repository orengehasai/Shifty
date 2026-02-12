export interface Staff {
  id: string;
  name: string;
  role: 'kitchen' | 'hall' | 'both';
  employment_type: 'full_time' | 'part_time';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface StaffMonthlySetting {
  id: string;
  staff_id: string;
  staff_name: string;
  year_month: string;
  min_preferred_hours: number;
  max_preferred_hours: number;
  note: string;
  created_at: string;
  updated_at: string;
}

export interface ShiftRequest {
  id: string;
  staff_id: string;
  staff_name: string;
  year_month: string;
  date: string;
  start_time: string;
  end_time: string;
  request_type: 'available' | 'unavailable' | 'preferred';
  note: string;
  created_at: string;
  updated_at: string;
}

export interface ConstraintViolation {
  constraint_name: string;
  type: string;
  message: string;
}

export interface Constraint {
  id: string;
  name: string;
  type: 'hard' | 'soft';
  category: 'min_staff' | 'max_staff' | 'max_consecutive_days' | 'monthly_hours' | 'fixed_day_off' | 'staff_compatibility' | 'rest_hours';
  config: Record<string, unknown>;
  is_active: boolean;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface ShiftEntry {
  id: string;
  pattern_id: string;
  staff_id: string;
  staff_name: string;
  date: string;
  start_time: string;
  end_time: string;
  break_minutes: number;
  is_manual_edit: boolean;
  created_at: string;
  updated_at: string;
}

export interface ShiftPattern {
  id: string;
  year_month: string;
  status: 'draft' | 'selected' | 'finalized';
  reasoning: string;
  score: number;
  constraint_violations: ConstraintViolation[];
  summary?: {
    total_entries: number;
    staff_hours: Record<string, number>;
  };
  entries?: ShiftEntry[];
  created_at: string;
}

export interface GenerationJob {
  id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  year_month: string;
  pattern_count: number;
  progress: number;
  status_message: string | null;
  started_at: string;
  completed_at: string | null;
  error_message: string | null;
  created_at: string;
}

export interface DailyStaffCount {
  date: string;
  count: number;
}

export interface DashboardSummary {
  year_month: string;
  staff_count: number;
  active_staff_count: number;
  request_submitted_count: number;
  monthly_settings_count: number;
  shift_status: 'not_started' | 'requests_submitted' | 'generating' | 'generated' | 'selected' | 'finalized';
  constraint_count: number;
  daily_staff_counts: DailyStaffCount[];
}
