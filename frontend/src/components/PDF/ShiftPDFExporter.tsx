import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import type { ShiftPattern, ShiftEntry } from '../../types';

export function exportShiftPDF(
  pattern: ShiftPattern,
  entries: ShiftEntry[],
  staffNames: [string, string][]
) {
  const doc = new jsPDF({ orientation: 'landscape' });
  const [y, m] = pattern.year_month.split('-');
  const daysInMonth = new Date(parseInt(y), parseInt(m), 0).getDate();

  // Header
  doc.setFontSize(16);
  doc.text(`Shift Schedule - ${y}/${parseInt(m)}`, 14, 15);
  doc.setFontSize(10);
  doc.text(`Score: ${pattern.score?.toFixed(1) ?? '-'} | Status: ${pattern.status}`, 14, 22);

  // Build date headers
  const dateHeaders = Array.from({ length: daysInMonth }, (_, i) => String(i + 1));

  // Build rows
  const rows = staffNames.map(([staffId, name]) => {
    const cells = Array.from({ length: daysInMonth }, (_, i) => {
      const date = `${pattern.year_month}-${String(i + 1).padStart(2, '0')}`;
      const entry = entries.find((e) => e.staff_id === staffId && e.date === date);
      return entry ? `${entry.start_time}-${entry.end_time}` : '-';
    });

    // Total hours
    let total = 0;
    entries.forEach((e) => {
      if (e.staff_id !== staffId || !e.start_time || !e.end_time) return;
      const [sh, sm] = e.start_time.split(':').map(Number);
      const [eh, em] = e.end_time.split(':').map(Number);
      total += Math.max(0, (eh * 60 + em - sh * 60 - sm - (e.break_minutes || 0)) / 60);
    });

    return [name, ...cells, `${total.toFixed(1)}h`];
  });

  autoTable(doc, {
    head: [['Name', ...dateHeaders, 'Total']],
    body: rows,
    startY: 28,
    styles: { fontSize: 6, cellPadding: 1.5 },
    headStyles: { fillColor: [79, 70, 229], fontSize: 6, cellPadding: 1.5 },
    columnStyles: {
      0: { cellWidth: 22 },
    },
    theme: 'grid',
  });

  doc.save(`shift_${pattern.year_month}.pdf`);
}
