import { useEffect, useState, useMemo, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useShiftStore } from '../../stores/shiftStore';
import type { ShiftEntry, ShiftPattern } from '../../types';
import Button from '../../components/Common/Button';
import LoadingSpinner from '../../components/Common/LoadingSpinner';
import { exportShiftPDF } from '../../components/PDF/ShiftPDFExporter';

const styles: Record<string, React.CSSProperties> = {
  header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 },
  heading: { fontSize: 20, fontWeight: 700 },
  headerActions: { display: 'flex', gap: 8 },
  weekNav: { display: 'flex', gap: 8, alignItems: 'center', justifyContent: 'center', marginBottom: 16 },
  grid: {
    overflowX: 'auto',
    background: '#fff',
    borderRadius: 8,
    border: '1px solid #E5E7EB',
    marginBottom: 24,
  },
  table: { width: '100%', borderCollapse: 'collapse', minWidth: 600 },
  th: { padding: '10px 8px', background: '#F9FAFB', fontSize: 13, color: '#6B7280', borderBottom: '1px solid #E5E7EB', fontWeight: 600, textAlign: 'center', minWidth: 80 },
  thStaff: { padding: '10px 12px', background: '#F9FAFB', fontSize: 13, color: '#6B7280', borderBottom: '1px solid #E5E7EB', fontWeight: 600, textAlign: 'left', minWidth: 80 },
  td: { padding: '8px', borderBottom: '1px solid #F3F4F6', fontSize: 13, textAlign: 'center', cursor: 'pointer' },
  tdManual: { background: '#FEF3C7' },
  editPopover: {
    position: 'absolute',
    background: '#fff',
    border: '1px solid #D1D5DB',
    borderRadius: 6,
    padding: 12,
    zIndex: 100,
    boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
    display: 'flex',
    flexDirection: 'column',
    gap: 8,
  },
  input: { padding: '6px 8px', border: '1px solid #D1D5DB', borderRadius: 4, fontSize: 13, width: 90 },
  summaryBox: { background: '#fff', borderRadius: 8, padding: 20, border: '1px solid #E5E7EB', marginBottom: 24 },
  summaryTitle: { fontSize: 16, fontWeight: 600, marginBottom: 12 },
  summaryTable: { width: '100%', borderCollapse: 'collapse' },
  warning: { background: '#FEF3C7', borderLeft: '4px solid #F59E0B', padding: '12px 16px', borderRadius: 4, fontSize: 14, marginBottom: 8 },
  error: { color: '#EF4444', padding: 20 },
  dayOfWeek: { fontSize: 11, color: '#9CA3AF' },
};

const DAY_LABELS = ['日', '月', '火', '水', '木', '金', '土'];

function getWeeks(yearMonth: string): string[][] {
  const [y, m] = yearMonth.split('-').map(Number);
  const daysInMonth = new Date(y, m, 0).getDate();
  const weeks: string[][] = [];
  let week: string[] = [];
  for (let d = 1; d <= daysInMonth; d++) {
    const date = `${yearMonth}-${String(d).padStart(2, '0')}`;
    const dow = new Date(y, m - 1, d).getDay();
    if (dow === 1 && week.length > 0) {
      weeks.push(week);
      week = [];
    }
    week.push(date);
  }
  if (week.length > 0) weeks.push(week);
  return weeks;
}

interface EditState {
  entryId: string;
  start_time: string;
  end_time: string;
  break_minutes: number;
}

export default function ShiftEditPage() {
  const { patternId } = useParams<{ patternId: string }>();
  const { currentPattern, loading, error, fetchPattern, updateEntry, finalizePattern } = useShiftStore();
  const [weekIdx, setWeekIdx] = useState(0);
  const [editing, setEditing] = useState<EditState | null>(null);
  const [warnings, setWarnings] = useState<Array<{ type: string; message: string }>>([]);
  const [manualEdits, setManualEdits] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (patternId) fetchPattern(patternId);
  }, [patternId, fetchPattern]);

  const pattern: ShiftPattern | null = currentPattern;
  const entries: ShiftEntry[] = pattern?.entries || [];
  const yearMonth = pattern?.year_month || '';
  const weeks = useMemo(() => yearMonth ? getWeeks(yearMonth) : [], [yearMonth]);
  const currentWeek = weeks[weekIdx] || [];

  const staffNames = useMemo(() => {
    const names = new Map<string, string>();
    entries.forEach((e) => names.set(e.staff_id, e.staff_name));
    return Array.from(names.entries());
  }, [entries]);

  const getEntry = useCallback((staffId: string, date: string) => {
    return entries.find((e) => e.staff_id === staffId && e.date === date);
  }, [entries]);

  const handleCellClick = (entry: ShiftEntry | undefined) => {
    if (!entry) return;
    setEditing({
      entryId: entry.id,
      start_time: entry.start_time,
      end_time: entry.end_time,
      break_minutes: entry.break_minutes,
    });
  };

  const handleSaveEdit = async () => {
    if (!editing) return;
    try {
      const validation = await updateEntry(editing.entryId, {
        start_time: editing.start_time,
        end_time: editing.end_time,
        break_minutes: editing.break_minutes,
      });
      setManualEdits((prev) => new Set(prev).add(editing.entryId));
      if (validation.warnings) setWarnings(validation.warnings);
      setEditing(null);
      if (patternId) fetchPattern(patternId);
    } catch {
      setWarnings([{ type: 'error', message: '更新に失敗しました' }]);
    }
  };

  const handleFinalize = async () => {
    if (!patternId) return;
    if (!confirm('このシフトを確定しますか?')) return;
    try {
      await finalizePattern(patternId);
      await fetchPattern(patternId);
    } catch {
      setWarnings([{ type: 'error', message: '確定に失敗しました' }]);
    }
  };

  const handlePDF = () => {
    if (!pattern) return;
    exportShiftPDF(pattern, entries, staffNames);
  };

  // Calculate staff hours
  const staffHours = useMemo(() => {
    const hours: Record<string, number> = {};
    entries.forEach((e) => {
      if (!e.start_time || !e.end_time) return;
      const [sh, sm] = e.start_time.split(':').map(Number);
      const [eh, em] = e.end_time.split(':').map(Number);
      const worked = (eh * 60 + em - sh * 60 - sm - (e.break_minutes || 0)) / 60;
      hours[e.staff_id] = (hours[e.staff_id] || 0) + Math.max(0, worked);
    });
    return hours;
  }, [entries]);

  if (loading) return <LoadingSpinner />;
  if (error) return <div style={styles.error}>{error}</div>;
  if (!pattern) return <div style={styles.error}>パターンが見つかりません</div>;

  const [y, mo] = yearMonth.split('-');

  return (
    <div>
      <div style={styles.header}>
        <div>
          <h1 style={styles.heading}>シフト編集 {y}年{parseInt(mo)}月</h1>
          <span style={{ fontSize: 14, color: '#6B7280' }}>スコア: {pattern.score?.toFixed(1)} / ステータス: {pattern.status}</span>
        </div>
        <div style={styles.headerActions}>
          <Button variant="secondary" onClick={handlePDF}>PDF出力</Button>
          <Button onClick={handleFinalize} disabled={pattern.status === 'finalized'}>
            {pattern.status === 'finalized' ? '確定済み' : '確定'}
          </Button>
        </div>
      </div>

      <div style={styles.grid}>
        <table style={styles.table}>
          <thead>
            <tr>
              <th style={styles.thStaff}>名前</th>
              {currentWeek.map((date) => {
                const d = new Date(date);
                const day = d.getDate();
                const dow = d.getDay();
                return (
                  <th key={date} style={styles.th}>
                    {parseInt(mo)}/{day}<br /><span style={styles.dayOfWeek}>{DAY_LABELS[dow]}</span>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {staffNames.map(([staffId, name]) => (
              <tr key={staffId}>
                <td style={{ ...styles.td, textAlign: 'left', fontWeight: 500, cursor: 'default', paddingLeft: 12 }}>{name}</td>
                {currentWeek.map((date) => {
                  const entry = getEntry(staffId, date);
                  const isManual = entry && (entry.is_manual_edit || manualEdits.has(entry.id));
                  return (
                    <td
                      key={date}
                      style={{ ...styles.td, ...(isManual ? styles.tdManual : {}) }}
                      onClick={() => handleCellClick(entry)}
                    >
                      {entry ? `${entry.start_time}-${entry.end_time}` : '休み'}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div style={styles.weekNav}>
        <Button variant="secondary" size="sm" onClick={() => setWeekIdx(Math.max(0, weekIdx - 1))} disabled={weekIdx === 0}>
          ← 前の週
        </Button>
        <span style={{ fontSize: 14, color: '#6B7280' }}>第{weekIdx + 1}週 / 全{weeks.length}週</span>
        <Button variant="secondary" size="sm" onClick={() => setWeekIdx(Math.min(weeks.length - 1, weekIdx + 1))} disabled={weekIdx >= weeks.length - 1}>
          次の週 →
        </Button>
      </div>

      {editing && (
        <div style={styles.editPopover}>
          <div style={{ fontWeight: 600, fontSize: 14 }}>時間編集</div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input type="time" style={styles.input} value={editing.start_time} onChange={(e) => setEditing({ ...editing, start_time: e.target.value })} />
            <span>-</span>
            <input type="time" style={styles.input} value={editing.end_time} onChange={(e) => setEditing({ ...editing, end_time: e.target.value })} />
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <span style={{ fontSize: 13 }}>休憩:</span>
            <input type="number" style={{ ...styles.input, width: 60 }} value={editing.break_minutes} onChange={(e) => setEditing({ ...editing, break_minutes: Number(e.target.value) })} />
            <span style={{ fontSize: 13 }}>分</span>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button variant="secondary" size="sm" onClick={() => setEditing(null)}>キャンセル</Button>
            <Button size="sm" onClick={handleSaveEdit}>保存</Button>
          </div>
        </div>
      )}

      <div style={styles.summaryBox}>
        <div style={styles.summaryTitle}>サマリー</div>
        <table style={styles.summaryTable}>
          <thead>
            <tr>
              <th style={{ ...styles.th, textAlign: 'left' }}>名前</th>
              <th style={styles.th}>月間H</th>
              <th style={styles.th}>差分</th>
            </tr>
          </thead>
          <tbody>
            {staffNames.map(([staffId, name]) => {
              const actual = staffHours[staffId] || 0;
              const preferred = pattern.summary?.staff_hours?.[name];
              return (
                <tr key={staffId}>
                  <td style={{ ...styles.td, textAlign: 'left', cursor: 'default' }}>{name}</td>
                  <td style={{ ...styles.td, cursor: 'default' }}>{actual.toFixed(1)}h</td>
                  <td style={{ ...styles.td, cursor: 'default', color: preferred !== undefined ? (actual - preferred >= 0 ? '#10B981' : '#EF4444') : '#6B7280' }}>
                    {preferred !== undefined ? `${(actual - preferred) >= 0 ? '+' : ''}${(actual - preferred).toFixed(1)}h` : '-'}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {warnings.length > 0 && (
        <div>
          {warnings.map((w, i) => (
            <div key={i} style={styles.warning}>{w.message}</div>
          ))}
        </div>
      )}

      {pattern.constraint_violations && pattern.constraint_violations.length > 0 && (
        <div>
          {pattern.constraint_violations.map((v, i) => (
            <div key={i} style={styles.warning}>
              [{v.type}] {v.constraint_name}: {v.message}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
