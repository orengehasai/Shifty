-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- staffs
CREATE TABLE staffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL,
    employment_type VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- staff_monthly_settings
CREATE TABLE staff_monthly_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    year_month VARCHAR(7) NOT NULL,
    min_preferred_hours INTEGER NOT NULL,
    max_preferred_hours INTEGER NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_staff_monthly_settings_unique ON staff_monthly_settings(staff_id, year_month);

-- shift_requests
CREATE TABLE shift_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    year_month VARCHAR(7) NOT NULL,
    date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    request_type VARCHAR(20) NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shift_requests_staff_month ON shift_requests(staff_id, year_month);
CREATE INDEX idx_shift_requests_year_month ON shift_requests(year_month);

-- constraints
CREATE TABLE constraints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    type VARCHAR(10) NOT NULL,
    category VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_constraints_active ON constraints(is_active);
CREATE INDEX idx_constraints_category ON constraints(category);

-- shift_patterns
CREATE TABLE shift_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year_month VARCHAR(7) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    reasoning TEXT,
    score DECIMAL(5,2),
    constraint_violations JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shift_patterns_year_month ON shift_patterns(year_month);
CREATE INDEX idx_shift_patterns_status ON shift_patterns(status);

-- shift_entries
CREATE TABLE shift_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_id UUID NOT NULL REFERENCES shift_patterns(id) ON DELETE CASCADE,
    staff_id UUID NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    break_minutes INTEGER NOT NULL DEFAULT 0,
    is_manual_edit BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shift_entries_pattern ON shift_entries(pattern_id);
CREATE INDEX idx_shift_entries_staff_date ON shift_entries(staff_id, date);

-- generation_jobs
CREATE TABLE generation_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year_month VARCHAR(7) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pattern_count INTEGER NOT NULL DEFAULT 3,
    progress INTEGER NOT NULL DEFAULT 0,
    status_message TEXT,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generation_jobs_status ON generation_jobs(status);
CREATE INDEX idx_generation_jobs_year_month ON generation_jobs(year_month);
