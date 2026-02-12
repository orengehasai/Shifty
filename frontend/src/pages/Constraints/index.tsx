import { useEffect, useState } from 'react';
import { useConstraintStore } from '../../stores/constraintStore';
import type { Constraint } from '../../types';
import Button from '../../components/Common/Button';
import Modal from '../../components/Common/Modal';
import LoadingSpinner from '../../components/Common/LoadingSpinner';

const styles: Record<string, React.CSSProperties> = {
  header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 },
  heading: { fontSize: 20, fontWeight: 700 },
  section: { marginBottom: 24 },
  sectionTitle: { fontSize: 16, fontWeight: 600, marginBottom: 12, color: '#374151' },
  table: { width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden', border: '1px solid #E5E7EB' },
  th: { padding: '10px 16px', textAlign: 'left', background: '#F9FAFB', fontSize: 13, color: '#6B7280', borderBottom: '1px solid #E5E7EB', fontWeight: 600 },
  td: { padding: '10px 16px', borderBottom: '1px solid #F3F4F6', fontSize: 14 },
  badge: { display: 'inline-block', padding: '2px 8px', borderRadius: 10, fontSize: 12 },
  form: { display: 'flex', flexDirection: 'column', gap: 12 },
  formGroup: { display: 'flex', flexDirection: 'column', gap: 4 },
  label: { fontSize: 13, fontWeight: 500, color: '#374151' },
  input: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  select: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  formActions: { display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 8 },
  error: { color: '#EF4444', padding: 20 },
  empty: { padding: '16px', textAlign: 'center', color: '#9CA3AF', fontSize: 14 },
};

const categoryLabels: Record<string, string> = {
  min_staff: '最低スタッフ数',
  max_staff: '最大スタッフ数',
  max_consecutive_days: '連勤制限',
  monthly_hours: '月間時間',
  fixed_day_off: '固定休み',
  staff_compatibility: 'スタッフ相性',
  rest_hours: '勤務間インターバル',
};

const categories = Object.keys(categoryLabels);

interface FormData {
  name: string;
  type: 'hard' | 'soft';
  category: string;
  priority: number;
  config: Record<string, unknown>;
}

const emptyForm: FormData = { name: '', type: 'hard', category: 'min_staff', priority: 0, config: {} };

function ConfigForm({ category, config, onChange }: { category: string; config: Record<string, unknown>; onChange: (c: Record<string, unknown>) => void }) {
  const inputStyle = styles.input;
  switch (category) {
    case 'min_staff':
    case 'max_staff': {
      const ranges = (config.time_ranges as Array<{ start: string; end: string; min_count?: number; max_count?: number }>) || [{ start: '09:00', end: '17:00', min_count: 1, max_count: 10 }];
      const r = ranges[0] || { start: '09:00', end: '17:00', min_count: 1, max_count: 10 };
      const countKey = category === 'min_staff' ? 'min_count' : 'max_count';
      return (
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
          <div style={styles.formGroup}>
            <label style={styles.label}>開始時刻</label>
            <input type="time" style={inputStyle} value={r.start} onChange={(e) => onChange({ time_ranges: [{ ...r, start: e.target.value }] })} />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>終了時刻</label>
            <input type="time" style={inputStyle} value={r.end} onChange={(e) => onChange({ time_ranges: [{ ...r, end: e.target.value }] })} />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>{category === 'min_staff' ? '最低人数' : '最大人数'}</label>
            <input type="number" style={inputStyle} value={(r as Record<string, unknown>)[countKey] as number || 1} onChange={(e) => onChange({ time_ranges: [{ ...r, [countKey]: Number(e.target.value) }] })} />
          </div>
        </div>
      );
    }
    case 'max_consecutive_days':
      return (
        <div style={styles.formGroup}>
          <label style={styles.label}>最大連勤日数</label>
          <input type="number" style={inputStyle} value={(config.max_days as number) || 5} onChange={(e) => onChange({ max_days: Number(e.target.value) })} />
        </div>
      );
    case 'monthly_hours':
      return (
        <div style={{ display: 'flex', gap: 8 }}>
          <div style={styles.formGroup}>
            <label style={styles.label}>最小時間</label>
            <input type="number" style={inputStyle} value={(config.min_hours as number) || 0} onChange={(e) => onChange({ ...config, min_hours: Number(e.target.value) })} />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>最大時間</label>
            <input type="number" style={inputStyle} value={(config.max_hours as number) || 160} onChange={(e) => onChange({ ...config, max_hours: Number(e.target.value) })} />
          </div>
        </div>
      );
    case 'fixed_day_off':
      return (
        <div style={styles.formGroup}>
          <label style={styles.label}>曜日 (0=日, 1=月, ... 6=土)</label>
          <input type="number" min={0} max={6} style={inputStyle} value={(config.day_of_week as number) ?? 0} onChange={(e) => onChange({ day_of_week: Number(e.target.value) })} />
        </div>
      );
    case 'staff_compatibility': {
      const staffIds = (config.staff_ids as string[]) || ['', ''];
      const rule = (config.rule as string) || 'prefer_together';
      return (
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
          <div style={styles.formGroup}>
            <label style={styles.label}>スタッフ1 ID</label>
            <input style={inputStyle} value={staffIds[0] || ''} onChange={(e) => onChange({ ...config, staff_ids: [e.target.value, staffIds[1] || ''], rule })} />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>スタッフ2 ID</label>
            <input style={inputStyle} value={staffIds[1] || ''} onChange={(e) => onChange({ ...config, staff_ids: [staffIds[0] || '', e.target.value], rule })} />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>ルール</label>
            <select style={inputStyle} value={rule} onChange={(e) => onChange({ ...config, staff_ids: staffIds, rule: e.target.value })}>
              <option value="prefer_together">一緒にする</option>
              <option value="avoid_together">避ける</option>
            </select>
          </div>
        </div>
      );
    }
    case 'rest_hours':
      return (
        <div style={styles.formGroup}>
          <label style={styles.label}>最低休息時間</label>
          <input type="number" style={inputStyle} value={(config.min_hours as number) || 11} onChange={(e) => onChange({ min_hours: Number(e.target.value) })} />
        </div>
      );
    default:
      return null;
  }
}

export default function ConstraintsPage() {
  const { constraints, loading, error, fetchConstraints, createConstraint, deleteConstraint } = useConstraintStore();
  const [modalOpen, setModalOpen] = useState(false);
  const [form, setForm] = useState<FormData>(emptyForm);

  useEffect(() => { fetchConstraints(); }, [fetchConstraints]);

  const hardConstraints = constraints.filter((c) => c.type === 'hard');
  const softConstraints = constraints.filter((c) => c.type === 'soft');

  const handleSave = async () => {
    if (!form.name.trim()) return;
    await createConstraint({ name: form.name, type: form.type, category: form.category, config: form.config, priority: form.priority });
    setModalOpen(false);
    setForm(emptyForm);
  };

  const handleDelete = async (id: string) => {
    if (confirm('この制約を削除しますか?')) await deleteConstraint(id);
  };

  if (loading) return <LoadingSpinner />;
  if (error) return <div style={styles.error}>{error}</div>;

  const renderTable = (items: Constraint[]) => (
    <table style={styles.table}>
      <thead>
        <tr>
          <th style={styles.th}>制約名</th>
          <th style={styles.th}>カテゴリ</th>
          <th style={{ ...styles.th, width: 60 }}>優先度</th>
          <th style={{ ...styles.th, width: 60 }}>状態</th>
          <th style={{ ...styles.th, width: 120 }}>操作</th>
        </tr>
      </thead>
      <tbody>
        {items.map((c) => (
          <tr key={c.id}>
            <td style={styles.td}>{c.name}</td>
            <td style={styles.td}>{categoryLabels[c.category] || c.category}</td>
            <td style={styles.td}>{c.priority}</td>
            <td style={styles.td}>
              <span style={{ ...styles.badge, background: c.is_active ? '#D1FAE5' : '#FEE2E2', color: c.is_active ? '#065F46' : '#991B1B' }}>
                {c.is_active ? '有効' : '無効'}
              </span>
            </td>
            <td style={styles.td}>
              <Button variant="danger" size="sm" onClick={() => handleDelete(c.id)}>削除</Button>
            </td>
          </tr>
        ))}
        {items.length === 0 && (
          <tr><td style={styles.empty} colSpan={5}>制約がありません</td></tr>
        )}
      </tbody>
    </table>
  );

  return (
    <div>
      <div style={styles.header}>
        <h1 style={styles.heading}>制約設定</h1>
        <Button onClick={() => { setForm(emptyForm); setModalOpen(true); }}>+ 制約追加</Button>
      </div>

      <div style={styles.section}>
        <div style={styles.sectionTitle}>ハード制約 (必ず守る)</div>
        {renderTable(hardConstraints)}
      </div>

      <div style={styles.section}>
        <div style={styles.sectionTitle}>ソフト制約 (できるだけ守る)</div>
        {renderTable(softConstraints)}
      </div>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="制約を追加">
        <div style={styles.form}>
          <div style={styles.formGroup}>
            <label style={styles.label}>制約名</label>
            <input style={styles.input} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="例: ランチタイム最低3名" />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>種類</label>
            <select style={styles.select} value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value as 'hard' | 'soft' })}>
              <option value="hard">ハード制約</option>
              <option value="soft">ソフト制約</option>
            </select>
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>カテゴリ</label>
            <select style={styles.select} value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value, config: {} })}>
              {categories.map((c) => <option key={c} value={c}>{categoryLabels[c]}</option>)}
            </select>
          </div>
          {form.type === 'soft' && (
            <div style={styles.formGroup}>
              <label style={styles.label}>優先度 (1-5)</label>
              <input type="number" min={1} max={5} style={styles.input} value={form.priority} onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })} />
            </div>
          )}
          <div style={{ borderTop: '1px solid #E5E7EB', paddingTop: 12, marginTop: 4 }}>
            <div style={{ fontSize: 13, fontWeight: 500, color: '#6B7280', marginBottom: 8 }}>カテゴリに応じた設定</div>
            <ConfigForm category={form.category} config={form.config} onChange={(c) => setForm({ ...form, config: c })} />
          </div>
          <div style={styles.formActions}>
            <Button variant="secondary" onClick={() => setModalOpen(false)}>キャンセル</Button>
            <Button onClick={handleSave}>保存</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
