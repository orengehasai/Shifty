import { useUiStore } from '../../stores/uiStore';

const styles: Record<string, React.CSSProperties> = {
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 24px',
    height: 56,
    background: '#fff',
    borderBottom: '1px solid #E5E7EB',
  },
  appName: {
    fontSize: 18,
    fontWeight: 700,
    color: '#4F46E5',
  },
  selector: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
  },
  label: {
    fontSize: 14,
    color: '#6B7280',
  },
  input: {
    padding: '6px 12px',
    border: '1px solid #D1D5DB',
    borderRadius: 6,
    fontSize: 14,
  },
};

export default function Header() {
  const { yearMonth, setYearMonth } = useUiStore();
  return (
    <header style={styles.header}>
      <span style={styles.appName}>Shift Manager</span>
      <div style={styles.selector}>
        <span style={styles.label}>対象年月:</span>
        <input
          type="month"
          value={yearMonth}
          onChange={(e) => setYearMonth(e.target.value)}
          style={styles.input}
        />
      </div>
    </header>
  );
}
