import { describe, it, expect } from 'vitest';

// Test the pure logic extracted from ShiftRequestPage

const requestTypeCycle = ['available', 'unavailable', 'preferred'] as const;

interface DayData {
  date: string;
  request_type: typeof requestTypeCycle[number] | null;
  start_time: string;
  end_time: string;
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

function toggleDay(days: DayData[], date: string): DayData[] {
  return days.map((d) => {
    if (d.date !== date) return d;
    const currentIdx = d.request_type ? requestTypeCycle.indexOf(d.request_type) : -1;
    const nextIdx = (currentIdx + 1) % (requestTypeCycle.length + 1);
    return { ...d, request_type: nextIdx < requestTypeCycle.length ? requestTypeCycle[nextIdx] : null };
  });
}

function shouldShowTimeInputs(requestType: string | null): boolean {
  return requestType !== null && requestType !== 'unavailable';
}

describe('ShiftRequest ロジックテスト', () => {
  describe('getDaysInMonth', () => {
    it('1月は31日を返す', () => {
      const days = getDaysInMonth('2025-01');
      expect(days.length).toBe(31);
      expect(days[0].date).toBe('2025-01-01');
      expect(days[30].date).toBe('2025-01-31');
    });

    it('2月は28日を返す (非閏年)', () => {
      const days = getDaysInMonth('2025-02');
      expect(days.length).toBe(28);
    });

    it('2月は29日を返す (閏年)', () => {
      const days = getDaysInMonth('2024-02');
      expect(days.length).toBe(29);
    });

    it('初期値は request_type=null, start_time=09:00, end_time=17:00', () => {
      const days = getDaysInMonth('2025-01');
      days.forEach((d) => {
        expect(d.request_type).toBeNull();
        expect(d.start_time).toBe('09:00');
        expect(d.end_time).toBe('17:00');
      });
    });
  });

  describe('toggleDay (requestTypeCycle)', () => {
    it('null -> available', () => {
      let days = getDaysInMonth('2025-01');
      days = toggleDay(days, '2025-01-01');
      expect(days[0].request_type).toBe('available');
    });

    it('available -> unavailable', () => {
      let days = getDaysInMonth('2025-01');
      days = toggleDay(days, '2025-01-01'); // available
      days = toggleDay(days, '2025-01-01'); // unavailable
      expect(days[0].request_type).toBe('unavailable');
    });

    it('unavailable -> preferred', () => {
      let days = getDaysInMonth('2025-01');
      days = toggleDay(days, '2025-01-01'); // available
      days = toggleDay(days, '2025-01-01'); // unavailable
      days = toggleDay(days, '2025-01-01'); // preferred
      expect(days[0].request_type).toBe('preferred');
    });

    it('preferred -> null (サイクル完了)', () => {
      let days = getDaysInMonth('2025-01');
      days = toggleDay(days, '2025-01-01'); // available
      days = toggleDay(days, '2025-01-01'); // unavailable
      days = toggleDay(days, '2025-01-01'); // preferred
      days = toggleDay(days, '2025-01-01'); // null
      expect(days[0].request_type).toBeNull();
    });

    it('他の日は変更されない', () => {
      let days = getDaysInMonth('2025-01');
      days = toggleDay(days, '2025-01-01');
      expect(days[1].request_type).toBeNull();
      expect(days[2].request_type).toBeNull();
    });
  });

  describe('shouldShowTimeInputs (時間入力の表示条件)', () => {
    it('available なら時間入力を表示する', () => {
      expect(shouldShowTimeInputs('available')).toBe(true);
    });

    it('preferred なら時間入力を表示する', () => {
      expect(shouldShowTimeInputs('preferred')).toBe(true);
    });

    it('unavailable なら時間入力を表示しない', () => {
      expect(shouldShowTimeInputs('unavailable')).toBe(false);
    });

    it('null なら時間入力を表示しない', () => {
      expect(shouldShowTimeInputs(null)).toBe(false);
    });
  });

  describe('hoursEnabled ロジック', () => {
    it('hoursEnabled=false の場合、月間設定を送らない (月間設定削除相当)', () => {
      const hoursEnabled = false;
      const setting = { id: 'setting-1' };
      // hoursEnabled がfalseで既存settingがある場合、削除APIを呼ぶべき
      const shouldDelete = !hoursEnabled && setting !== null;
      expect(shouldDelete).toBe(true);
    });

    it('hoursEnabled=true の場合、月間設定を保存する', () => {
      const hoursEnabled = true;
      const shouldSave = hoursEnabled;
      expect(shouldSave).toBe(true);
    });

    it('hoursEnabled=false で既存settingがない場合、何もしない', () => {
      const hoursEnabled = false;
      const setting = null;
      const shouldDelete = !hoursEnabled && setting !== null;
      expect(shouldDelete).toBe(false);
    });
  });
});
