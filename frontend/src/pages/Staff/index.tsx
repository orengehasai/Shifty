import { useEffect, useState } from 'react';
import { useStaffStore } from '../../stores/staffStore';
import Button from '../../components/Common/Button';
import Modal from '../../components/Common/Modal';
import LoadingSpinner from '../../components/Common/LoadingSpinner';

const roleLabels: Record<string, string> = { kitchen: 'キッチン', hall: 'ホール', both: '両方' };
const empLabels: Record<string, string> = { full_time: '正社員', part_time: 'パート' };

const styles: Record<string, React.CSSProperties> = {
  header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 },
  heading: { fontSize: 20, fontWeight: 700 },
  table: { width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden', border: '1px solid #E5E7EB' },
  th: { padding: '12px 16px', textAlign: 'left', background: '#F9FAFB', fontSize: 13, color: '#6B7280', borderBottom: '1px solid #E5E7EB', fontWeight: 600 },
  td: { padding: '12px 16px', borderBottom: '1px solid #F3F4F6', fontSize: 14 },
  badge: { display: 'inline-block', padding: '2px 8px', borderRadius: 10, fontSize: 12 },
  form: { display: 'flex', flexDirection: 'column', gap: 12 },
  formGroup: { display: 'flex', flexDirection: 'column', gap: 4 },
  label: { fontSize: 13, fontWeight: 500, color: '#374151' },
  input: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  select: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  formActions: { display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 8 },
  error: { color: '#EF4444', padding: 20 },
};

interface FormData { name: string; role: string; employment_type: string }
const emptyForm: FormData = { name: '', role: 'kitchen', employment_type: 'full_time' };

export default function StaffPage() {
  const { staffs, loading, error, fetchStaffs, createStaff, updateStaff, deleteStaff } = useStaffStore();
  const [modalOpen, setModalOpen] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [form, setForm] = useState<FormData>(emptyForm);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  useEffect(() => { fetchStaffs(); }, [fetchStaffs]);

  const openCreate = () => { setForm(emptyForm); setEditId(null); setModalOpen(true); };
  const openEdit = (s: typeof staffs[0]) => {
    setForm({ name: s.name, role: s.role, employment_type: s.employment_type });
    setEditId(s.id);
    setModalOpen(true);
  };

  const handleSave = async () => {
    if (!form.name.trim()) return;
    if (editId) {
      await updateStaff(editId, form);
    } else {
      await createStaff(form);
    }
    setModalOpen(false);
  };

  const handleDelete = async () => {
    if (deleteTarget) {
      await deleteStaff(deleteTarget);
      setDeleteTarget(null);
    }
  };

  const handleRestore = async (id: string) => {
    await updateStaff(id, { is_active: true });
  };

  if (loading) return <LoadingSpinner />;
  if (error) return <div style={styles.error}>{error}</div>;

  return (
    <div>
      <div style={styles.header}>
        <h1 style={styles.heading}>スタッフ管理</h1>
        <Button onClick={openCreate}>+ 新規追加</Button>
      </div>

      <table style={styles.table}>
        <thead>
          <tr>
            <th style={styles.th}>名前</th>
            <th style={styles.th}>役割</th>
            <th style={styles.th}>雇用形態</th>
            <th style={styles.th}>状態</th>
            <th style={styles.th}>操作</th>
          </tr>
        </thead>
        <tbody>
          {staffs.map((s) => (
            <tr key={s.id}>
              <td style={styles.td}>{s.name}</td>
              <td style={styles.td}>{roleLabels[s.role] || s.role}</td>
              <td style={styles.td}>{empLabels[s.employment_type] || s.employment_type}</td>
              <td style={styles.td}>
                <span style={{ ...styles.badge, background: s.is_active ? '#D1FAE5' : '#FEE2E2', color: s.is_active ? '#065F46' : '#991B1B' }}>
                  {s.is_active ? '有効' : '無効'}
                </span>
              </td>
              <td style={styles.td}>
                <Button variant="secondary" size="sm" onClick={() => openEdit(s)} style={{ marginRight: 8 }}>編集</Button>
                {s.is_active ? (
                  <Button variant="danger" size="sm" onClick={() => setDeleteTarget(s.id)}>削除</Button>
                ) : (
                  <Button size="sm" onClick={() => handleRestore(s.id)}>復活</Button>
                )}
              </td>
            </tr>
          ))}
          {staffs.length === 0 && (
            <tr><td style={{ ...styles.td, textAlign: 'center', color: '#9CA3AF' }} colSpan={5}>スタッフが登録されていません</td></tr>
          )}
        </tbody>
      </table>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editId ? 'スタッフ編集' : 'スタッフ登録'}>
        <div style={styles.form}>
          <div style={styles.formGroup}>
            <label style={styles.label}>名前</label>
            <input style={styles.input} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="名前を入力" />
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>役割</label>
            <select style={styles.select} value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}>
              <option value="kitchen">キッチン</option>
              <option value="hall">ホール</option>
              <option value="both">両方</option>
            </select>
          </div>
          <div style={styles.formGroup}>
            <label style={styles.label}>雇用形態</label>
            <select style={styles.select} value={form.employment_type} onChange={(e) => setForm({ ...form, employment_type: e.target.value })}>
              <option value="full_time">正社員</option>
              <option value="part_time">パート</option>
            </select>
          </div>
          <div style={styles.formActions}>
            <Button variant="secondary" onClick={() => setModalOpen(false)}>キャンセル</Button>
            <Button onClick={handleSave}>保存</Button>
          </div>
        </div>
      </Modal>

      <Modal open={deleteTarget !== null} onClose={() => setDeleteTarget(null)} title="削除確認">
        <p style={{ marginBottom: 16 }}>このスタッフを削除しますか? (論理削除されます)</p>
        <div style={styles.formActions}>
          <Button variant="secondary" onClick={() => setDeleteTarget(null)}>キャンセル</Button>
          <Button variant="danger" onClick={handleDelete}>削除する</Button>
        </div>
      </Modal>
    </div>
  );
}
