import { useEffect, useState, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUiStore } from '../../stores/uiStore';
import { useShiftStore } from '../../stores/shiftStore';
import { dashboardApi } from '../../api/dashboardApi';
import type { DashboardSummary, ShiftPattern } from '../../types';
import Button from '../../components/Common/Button';
import LoadingSpinner from '../../components/Common/LoadingSpinner';

const styles: Record<string, React.CSSProperties> = {
  heading: { fontSize: 20, fontWeight: 700, marginBottom: 20 },
  conditionBox: {
    background: '#fff',
    borderRadius: 8,
    padding: 20,
    border: '1px solid #E5E7EB',
    marginBottom: 24,
  },
  conditionTitle: { fontSize: 16, fontWeight: 600, marginBottom: 12 },
  conditionRow: { fontSize: 14, color: '#374151', marginBottom: 6 },
  controls: { display: 'flex', gap: 16, alignItems: 'center', marginBottom: 24 },
  select: { padding: '8px 12px', border: '1px solid #D1D5DB', borderRadius: 6, fontSize: 14 },
  progressBox: {
    background: '#fff',
    borderRadius: 8,
    padding: 20,
    border: '1px solid #E5E7EB',
    marginBottom: 24,
    textAlign: 'center',
  },
  progressBar: {
    height: 20,
    background: '#E5E7EB',
    borderRadius: 10,
    overflow: 'hidden',
    marginTop: 12,
  },
  progressFill: {
    height: '100%',
    background: '#4F46E5',
    borderRadius: 10,
    transition: 'width 0.5s',
  },
  patternsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
    gap: 16,
    marginBottom: 24,
  },
  patternCard: {
    background: '#fff',
    borderRadius: 8,
    padding: 20,
    border: '2px solid #E5E7EB',
    textAlign: 'center',
  },
  patternCardSelected: {
    borderColor: '#4F46E5',
    background: '#EEF2FF',
  },
  score: { fontSize: 28, fontWeight: 700, color: '#4F46E5', marginBottom: 4 },
  violations: { fontSize: 13, color: '#6B7280', marginBottom: 12 },
  reasoning: {
    background: '#F9FAFB',
    borderRadius: 6,
    padding: 16,
    fontSize: 14,
    color: '#374151',
    lineHeight: 1.6,
    textAlign: 'left',
    marginBottom: 16,
  },
  label: { fontSize: 13, fontWeight: 500, color: '#374151' },
  error: { color: '#EF4444', padding: 20 },
};

export default function ShiftGeneratePage() {
  const { yearMonth } = useUiStore();
  const navigate = useNavigate();
  const { patterns, fetchPatterns, startGeneration, pollJobStatus, selectPattern } = useShiftStore();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [patternCount, setPatternCount] = useState(3);
  const [generating, setGenerating] = useState(false);
  const [progress, setProgress] = useState(0);
  const [statusMsg, setStatusMsg] = useState('');
  const [selectedDetail, setSelectedDetail] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [y, m] = yearMonth.split('-');

  useEffect(() => {
    dashboardApi.getSummary(yearMonth).then((r) => setSummary(r.data)).catch(() => {});
    fetchPatterns(yearMonth);
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, [yearMonth, fetchPatterns]);

  const handleGenerate = useCallback(async () => {
    setGenerating(true);
    setProgress(0);
    setStatusMsg('生成を開始しています...');
    setError(null);
    try {
      const jobId = await startGeneration(yearMonth, patternCount);
      let elapsed = 0;
      pollRef.current = setInterval(async () => {
        elapsed += 2;
        try {
          const job = await pollJobStatus(jobId);
          if (job.status === 'processing') {
            const backendProgress = job.progress ?? 0;
            const fakeProgress = Math.min(90, elapsed * 2);
            setProgress(Math.max(backendProgress, fakeProgress));
            setStatusMsg(job.status_message || 'パターン生成中...');
          } else if (job.status === 'completed') {
            setProgress(100);
            setStatusMsg('生成完了');
            if (pollRef.current) clearInterval(pollRef.current);
            await fetchPatterns(yearMonth);
            setGenerating(false);
          } else if (job.status === 'failed') {
            if (pollRef.current) clearInterval(pollRef.current);
            setError(job.error_message || '生成に失敗しました');
            setGenerating(false);
          }
        } catch {
          if (pollRef.current) clearInterval(pollRef.current);
          setError('ステータス確認に失敗しました');
          setGenerating(false);
        }
      }, 2000);
    } catch {
      setError('生成の開始に失敗しました');
      setGenerating(false);
    }
  }, [yearMonth, patternCount, startGeneration, pollJobStatus, fetchPatterns]);

  const handleSelect = async (p: ShiftPattern) => {
    try {
      await selectPattern(p.id);
      navigate(`/shifts/${p.id}/edit`);
    } catch {
      setError('パターンの選択に失敗しました');
    }
  };

  return (
    <div>
      <h1 style={styles.heading}>シフト自動生成 {y}年{parseInt(m)}月</h1>

      <div style={styles.conditionBox}>
        <div style={styles.conditionTitle}>生成条件</div>
        <div style={styles.conditionRow}>スタッフ: {summary?.active_staff_count ?? '-'}名 (希望提出: {summary?.request_submitted_count ?? '-'}/{summary?.active_staff_count ?? '-'})</div>
        <div style={styles.conditionRow}>制約: {summary?.constraint_count ?? '-'}件</div>
      </div>

      <div style={styles.controls}>
        <span style={styles.label}>生成パターン数:</span>
        <select style={styles.select} value={patternCount} onChange={(e) => setPatternCount(Number(e.target.value))}>
          {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
        </select>
        <Button onClick={handleGenerate} disabled={generating}>
          {generating ? '生成中...' : 'シフトを生成する'}
        </Button>
      </div>

      {generating && (
        <div style={styles.progressBox}>
          <div>{statusMsg}</div>
          <div style={styles.progressBar}>
            <div style={{ ...styles.progressFill, width: `${progress}%` }} />
          </div>
          <div style={{ fontSize: 13, color: '#6B7280', marginTop: 8 }}>{progress}%</div>
        </div>
      )}

      {error && <div style={styles.error}>{error}</div>}

      {patterns.length > 0 && (
        <>
          <h2 style={{ fontSize: 16, fontWeight: 600, marginBottom: 12 }}>生成結果</h2>
          <div style={styles.patternsGrid}>
            {patterns.map((p, i) => (
              <div
                key={p.id}
                style={{
                  ...styles.patternCard,
                  ...(selectedDetail === p.id ? styles.patternCardSelected : {}),
                }}
              >
                <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 8 }}>パターン{i + 1}</div>
                <div style={styles.score}>{p.score?.toFixed(1) ?? '-'}</div>
                <div style={styles.violations}>
                  違反: {p.constraint_violations?.length ?? 0}件
                </div>
                <Button variant="secondary" size="sm" onClick={() => setSelectedDetail(selectedDetail === p.id ? null : p.id)} style={{ marginRight: 8, marginBottom: 8 }}>
                  {selectedDetail === p.id ? '閉じる' : '詳細を見る'}
                </Button>
                <Button size="sm" onClick={() => handleSelect(p)}>選択する</Button>
              </div>
            ))}
          </div>

          {selectedDetail && (() => {
            const p = patterns.find((p) => p.id === selectedDetail);
            if (!p) return null;
            return (
              <div style={styles.reasoning}>
                <div style={{ fontWeight: 600, marginBottom: 8 }}>パターン詳細</div>
                <p>{p.reasoning || 'AIの解説はありません'}</p>
                {p.constraint_violations && p.constraint_violations.length > 0 && (
                  <div style={{ marginTop: 12 }}>
                    <div style={{ fontWeight: 600, marginBottom: 4 }}>制約違反:</div>
                    {p.constraint_violations.map((v, i) => (
                      <div key={i} style={{ fontSize: 13, color: v.type === 'hard' ? '#EF4444' : '#D97706' }}>
                        [{v.type}] {v.constraint_name}: {v.message}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
          })()}
        </>
      )}
    </div>
  );
}
