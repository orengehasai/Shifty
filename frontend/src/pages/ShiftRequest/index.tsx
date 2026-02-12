import { useEffect, useState, useCallback } from 'react';
import { useUiStore } from '../../stores/uiStore';
import { useStaffStore } from '../../stores/staffStore';
import { useShiftRequestStore } from '../../stores/shiftRequestStore';
import { monthlySettingsApi } from '../../api/monthlySettingsApi';
import type { StaffMonthlySetting, ShiftRequest } from '../../types';
import Button from '../../components/Common/Button';
import LoadingSpinner from '../../components/Common/LoadingSpinner';

const styles: Record<string, React.CSSProperties> = {
  heading: { fontSize: 20, fontWeight: 700, marginBottom: 20 },
  controls: { display: 'flex', gap: 16, alignItems: 'center', marginBottom: 20, flexWrap: 'wrap' },
  label: { fontSize: 13, fontWeight: 500, color: '#374151' },
  select: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  input: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14, width: 80 },
  hoursRow: { display: 'flex', gap: 8, alignItems: 'center', marginBottom: 20 },
  calendar: {
    display: 'grid',
    gridTemplateColumns: 'repeat(7, 1fr)',
    gap: 4,
    background: '#fff',
    borderRadius: 8,
    padding: 16,
    border: '1px solid #E5E7EB',
    marginBottom: 16,
  },
  dayHeader: { textAlign: 'center', fontSize: 13, fontWeight: 600, color: '#6B7280', padding: 8 },
  dayCell: {
    border: '1px solid #E5E7EB',
    borderRadius: 4,
    padding: 8,
    minHeight: 100,
    cursor: 'pointer',
    textAlign: 'center',
    fontSize: 13,
    transition: 'background 0.15s',
  },
  dayNum: { fontWeight: 600, marginBottom: 4 },
  legend: { display: 'flex', gap: 16, marginBottom: 16, fontSize: 13, color: '#6B7280' },
  legendItem: { display: 'flex', alignItems: 'center', gap: 4 },
  dot: { width: 12, height: 12, borderRadius: '50%', display: 'inline-block' },
  timeInput: {
    padding: '2px 4px',
    border: '1px solid #D1D5DB',
    borderRadius: 4,
    fontSize: 11,
    width: '100%',
    marginTop: 2,
  },
  footer: { display: 'flex', justifyContent: 'flex-end', gap: 8 },
  error: { color: '#EF4444', padding: 20 },
};

const requestTypeColors: Record<string, string> = {
  available: '#D1FAE5',
  unavailable: '#FEE2E2',
  preferred: '#FEF3C7',
};
const requestTypeLabels: Record<string, string> = {
  available: '○',
  unavailable: 'x',
  preferred: '△',
};
const requestTypeCycle = ['available', 'unavailable', 'preferred'] as const;

interface DayData {
  date: string;
  request_type: typeof requestTypeCycle[number] | null;
  start_time: string;
  end_time: string;
  existingId?: string;
}

function getDaysInMonth(yearMonth: string): DayData[] {
  const [y, m] = yearMonth.split('-').map(Number);
  const daysInMonth = new Date(y, m, 0).getDate();
  return Array.from({ length: daysInMonth }, (_, i) => {
    const day = i + 1;
    const dateStr = `${yearMonth}-${String(day).padStart(2, '0')}`;
    return { date: dateStr, request_type: null, start_time: '09:00', end_time: '17:00' };
  });
}

export default function ShiftRequestPage() {
  const { yearMonth } = useUiStore();
  const { staffs, fetchStaffs } = useStaffStore();
  const { requests, loading, error, fetchRequests, batchCreate } = useShiftRequestStore();
  const [selectedStaff, setSelectedStaff] = useState('');
  const [days, setDays] = useState<DayData[]>([]);
  const [setting, setSetting] = useState<StaffMonthlySetting | null>(null);
  const [minH, setMinH] = useState(80);
  const [maxH, setMaxH] = useState(120);
  const [saving, setSaving] = useState(false);
  const [hoursEnabled, setHoursEnabled] = useState(false);

  useEffect(() => { fetchStaffs(); }, [fetchStaffs]);
  useEffect(() => {
    if (staffs.length > 0 && !selectedStaff) setSelectedStaff(staffs[0].id);
  }, [staffs, selectedStaff]);

  const loadData = useCallback(async () => {
    if (!selectedStaff) return;
    await fetchRequests(yearMonth, selectedStaff);
    try {
      const res = await monthlySettingsApi.list(yearMonth, selectedStaff);
      const s = (res.data.settings || []).find((s: StaffMonthlySetting) => s.staff_id === selectedStaff);
      if (s) { setSetting(s); setMinH(s.min_preferred_hours); setMaxH(s.max_preferred_hours); setHoursEnabled(true); }
      else { setSetting(null); setHoursEnabled(false); }
    } catch { /* ok */ }
  }, [selectedStaff, yearMonth, fetchRequests]);

  useEffect(() => { loadData(); }, [loadData]);

  useEffect(() => {
    const base = getDaysInMonth(yearMonth);
    requests.forEach((r: ShiftRequest) => {
      const idx = base.findIndex((d) => d.date === r.date);
      if (idx >= 0) {
        base[idx].request_type = r.request_type as typeof requestTypeCycle[number];
        base[idx].start_time = r.start_time || '09:00';
        base[idx].end_time = r.end_time || '17:00';
        base[idx].existingId = r.id;
      }
    });
    setDays(base);
  }, [requests, yearMonth]);

  const toggleDay = (date: string) => {
    setDays((prev) =>
      prev.map((d) => {
        if (d.date !== date) return d;
        const currentIdx = d.request_type ? requestTypeCycle.indexOf(d.request_type) : -1;
        const nextIdx = (currentIdx + 1) % (requestTypeCycle.length + 1);
        return { ...d, request_type: nextIdx < requestTypeCycle.length ? requestTypeCycle[nextIdx] : null };
      })
    );
  };

  const updateTime = (date: string, field: 'start_time' | 'end_time', value: string) => {
    setDays((prev) => prev.map((d) => d.date === date ? { ...d, [field]: value } : d));
  };

  const handleSave = async () => {
    if (!selectedStaff) return;
    setSaving(true);
    try {
      // save monthly settings
      if (hoursEnabled) {
        if (setting) {
          await monthlySettingsApi.update(setting.id, { min_preferred_hours: minH, max_preferred_hours: maxH });
        } else {
          await monthlySettingsApi.create({ staff_id: selectedStaff, year_month: yearMonth, min_preferred_hours: minH, max_preferred_hours: maxH });
        }
      } else if (setting) {
        await monthlySettingsApi.delete(setting.id);
      }
      // batch create shift requests
      const reqs = days.filter((d) => d.request_type).map((d) => ({
        staff_id: selectedStaff,
        year_month: yearMonth,
        date: d.date,
        ...(d.request_type !== 'unavailable' ? { start_time: d.start_time, end_time: d.end_time } : {}),
        request_type: d.request_type!,
      }));
      if (reqs.length > 0) await batchCreate(reqs);
      await loadData();
    } catch { /* TODO */ }
    setSaving(false);
  };

  const [y, m] = yearMonth.split('-').map(Number);
  const firstDow = new Date(y, m - 1, 1).getDay();
  const offset = firstDow === 0 ? 6 : firstDow - 1;

  if (loading) return <LoadingSpinner />;
  if (error) return <div style={styles.error}>{error}</div>;

  return (
    <div>
      <h1 style={styles.heading}>シフト希望入力</h1>
      <div style={styles.controls}>
        <div>
          <span style={styles.label}>スタッフ: </span>
          <select style={styles.select} value={selectedStaff} onChange={(e) => setSelectedStaff(e.target.value)}>
            {staffs.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
        </div>
        <div style={{ fontSize: 14, color: '#6B7280' }}>{y}年{m}月</div>
      </div>

      <div style={styles.hoursRow}>
        <label style={{ display: 'flex', alignItems: 'center', gap: 6, cursor: 'pointer' }}>
          <input type="checkbox" checked={hoursEnabled} onChange={(e) => setHoursEnabled(e.target.checked)} />
          <span style={styles.label}>月間希望時間を設定する</span>
        </label>
        {hoursEnabled && (
          <>
            <input type="number" style={styles.input} value={minH} onChange={(e) => setMinH(Number(e.target.value))} />
            <span>〜</span>
            <input type="number" style={styles.input} value={maxH} onChange={(e) => setMaxH(Number(e.target.value))} />
            <span style={{ fontSize: 13, color: '#6B7280' }}>時間</span>
          </>
        )}
      </div>

      <div style={styles.legend}>
        <span style={styles.legendItem}><span style={{ ...styles.dot, background: '#D1FAE5' }} /> ○ 出勤可</span>
        <span style={styles.legendItem}><span style={{ ...styles.dot, background: '#FEE2E2' }} /> x 不可</span>
        <span style={styles.legendItem}><span style={{ ...styles.dot, background: '#FEF3C7' }} /> △ 希望</span>
        <span style={{ fontSize: 12, color: '#9CA3AF' }}>クリックで切替</span>
      </div>

      <div style={styles.calendar}>
        {['月', '火', '水', '木', '金', '土', '日'].map((d) => (
          <div key={d} style={styles.dayHeader}>{d}</div>
        ))}
        {Array.from({ length: offset }, (_, i) => <div key={`blank-${i}`} />)}
        {days.map((d) => {
          const dayNum = parseInt(d.date.split('-')[2]);
          const bg = d.request_type ? requestTypeColors[d.request_type] : '#fff';
          return (
            <div
              key={d.date}
              style={{ ...styles.dayCell, background: bg }}
              onClick={() => toggleDay(d.date)}
            >
              <div style={styles.dayNum}>{dayNum}</div>
              <div>{d.request_type ? requestTypeLabels[d.request_type] : '-'}</div>
              {d.request_type && d.request_type !== 'unavailable' && (
                <div onClick={(e) => e.stopPropagation()}>
                  <input type="time" style={styles.timeInput} value={d.start_time} onChange={(e) => updateTime(d.date, 'start_time', e.target.value)} />
                  <input type="time" style={styles.timeInput} value={d.end_time} onChange={(e) => updateTime(d.date, 'end_time', e.target.value)} />
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div style={styles.footer}>
        <Button onClick={handleSave} disabled={saving}>{saving ? '保存中...' : '保存'}</Button>
      </div>
    </div>
  );
}
