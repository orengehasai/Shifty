import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUiStore } from '../../stores/uiStore';
import { dashboardApi } from '../../api/dashboardApi';
import type { DashboardSummary } from '../../types';
import Button from '../../components/Common/Button';
import LoadingSpinner from '../../components/Common/LoadingSpinner';

const styles: Record<string, React.CSSProperties> = {
  heading: { fontSize: 20, fontWeight: 700, marginBottom: 20 },
  cards: { display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16, marginBottom: 24 },
  card: {
    background: '#fff',
    borderRadius: 8,
    padding: 20,
    textAlign: 'center',
    border: '1px solid #E5E7EB',
  },
  cardLabel: { fontSize: 13, color: '#6B7280', marginBottom: 4 },
  cardValue: { fontSize: 24, fontWeight: 700, color: '#111827' },
  heatmap: {
    background: '#fff',
    borderRadius: 8,
    padding: 20,
    border: '1px solid #E5E7EB',
    marginBottom: 24,
  },
  heatmapTitle: { fontSize: 14, fontWeight: 600, marginBottom: 12 },
  grid: { display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4 },
  cell: {
    padding: 8,
    textAlign: 'center',
    borderRadius: 4,
    fontSize: 12,
  },
  actions: { display: 'flex', gap: 12 },
  error: { color: '#EF4444', padding: 20 },
  statusBadge: {
    display: 'inline-block',
    padding: '2px 8px',
    borderRadius: 10,
    fontSize: 12,
    fontWeight: 500,
  },
};

const statusLabels: Record<string, string> = {
  not_started: '未開始',
  requests_submitted: '希望入力済',
  generating: '生成中',
  generated: '生成済み',
  selected: '選択済み',
  finalized: '確定済み',
};

function heatColor(count: number): string {
  if (count === 0) return '#F3F4F6';
  if (count <= 2) return '#C7D2FE';
  if (count <= 4) return '#818CF8';
  return '#4F46E5';
}

export default function DashboardPage() {
  const { yearMonth } = useUiStore();
  const navigate = useNavigate();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    dashboardApi.getSummary(yearMonth)
      .then((res) => setSummary(res.data))
      .catch(() => setError('ダッシュボードデータの取得に失敗しました'))
      .finally(() => setLoading(false));
  }, [yearMonth]);

  if (loading) return <LoadingSpinner />;
  if (error) return <div style={styles.error}>{error}</div>;

  const [year, month] = yearMonth.split('-');

  return (
    <div>
      <h1 style={styles.heading}>{year}年{parseInt(month)}月のシフト状況</h1>
      <div style={styles.cards}>
        <div style={styles.card}>
          <div style={styles.cardLabel}>スタッフ数</div>
          <div style={styles.cardValue}>{summary?.active_staff_count ?? '-'}人</div>
        </div>
        <div style={styles.card}>
          <div style={styles.cardLabel}>希望提出率</div>
          <div style={styles.cardValue}>
            {summary ? `${summary.request_submitted_count}/${summary.active_staff_count}` : '-'}
          </div>
        </div>
        <div style={styles.card}>
          <div style={styles.cardLabel}>シフト状態</div>
          <div style={styles.cardValue}>
            <span style={{
              ...styles.statusBadge,
              background: summary?.shift_status === 'finalized' ? '#D1FAE5' : '#E0E7FF',
              color: summary?.shift_status === 'finalized' ? '#065F46' : '#3730A3',
            }}>
              {statusLabels[summary?.shift_status ?? 'not_started']}
            </span>
          </div>
        </div>
        <div style={styles.card}>
          <div style={styles.cardLabel}>制約数</div>
          <div style={styles.cardValue}>{summary?.constraint_count ?? '-'}件</div>
        </div>
      </div>

      {summary?.daily_staff_counts && summary.daily_staff_counts.length > 0 && (
        <div style={styles.heatmap}>
          <div style={styles.heatmapTitle}>日ごとの出勤予定人数</div>
          <div style={styles.grid}>
            {['月', '火', '水', '木', '金', '土', '日'].map((d) => (
              <div key={d} style={{ ...styles.cell, fontWeight: 600, color: '#6B7280' }}>{d}</div>
            ))}
            {(() => {
              const firstDay = new Date(parseInt(year), parseInt(month) - 1, 1).getDay();
              const offset = firstDay === 0 ? 6 : firstDay - 1;
              const blanks = Array.from({ length: offset }, (_, i) => (
                <div key={`blank-${i}`} style={styles.cell} />
              ));
              const days = summary.daily_staff_counts.map((d) => {
                const day = parseInt(d.date.split('-')[2]);
                return (
                  <div key={d.date} style={{ ...styles.cell, background: heatColor(d.count), color: d.count > 2 ? '#fff' : '#374151' }}>
                    {day}<br />{d.count}人
                  </div>
                );
              });
              return [...blanks, ...days];
            })()}
          </div>
        </div>
      )}

      <div style={styles.actions}>
        <Button onClick={() => navigate('/requests')}>シフト希望を入力</Button>
        <Button onClick={() => navigate('/generate')}>シフトを生成</Button>
        {summary?.shift_status && ['generated', 'selected', 'finalized'].includes(summary.shift_status) && (
          <Button variant="secondary" onClick={() => navigate('/generate')}>PDFを出力</Button>
        )}
      </div>
    </div>
  );
}
