import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import StaffPage from './index';

// Mock the staffStore
const mockFetchStaffs = vi.fn();
const mockCreateStaff = vi.fn();
const mockUpdateStaff = vi.fn();
const mockDeleteStaff = vi.fn();

vi.mock('../../stores/staffStore', () => ({
  useStaffStore: vi.fn(),
}));

import { useStaffStore } from '../../stores/staffStore';
const mockedUseStaffStore = vi.mocked(useStaffStore);

function setupStore(overrides: Partial<ReturnType<typeof useStaffStore>> = {}) {
  mockedUseStaffStore.mockReturnValue({
    staffs: [],
    loading: false,
    error: null,
    fetchStaffs: mockFetchStaffs,
    createStaff: mockCreateStaff,
    updateStaff: mockUpdateStaff,
    deleteStaff: mockDeleteStaff,
    ...overrides,
  } as ReturnType<typeof useStaffStore>);
}

describe('StaffPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('is_active=true のスタッフには削除ボタンが表示される', () => {
    setupStore({
      staffs: [
        { id: 's1', name: '田中太郎', role: 'kitchen', employment_type: 'full_time', is_active: true, created_at: '', updated_at: '' },
      ],
    });

    render(<StaffPage />);

    expect(screen.getByText('田中太郎')).toBeInTheDocument();
    expect(screen.getByText('有効')).toBeInTheDocument();
    expect(screen.getByText('削除')).toBeInTheDocument();
    expect(screen.queryByText('復活')).not.toBeInTheDocument();
  });

  it('is_active=false のスタッフには復活ボタンが表示される', () => {
    setupStore({
      staffs: [
        { id: 's2', name: '山田花子', role: 'hall', employment_type: 'part_time', is_active: false, created_at: '', updated_at: '' },
      ],
    });

    render(<StaffPage />);

    expect(screen.getByText('山田花子')).toBeInTheDocument();
    expect(screen.getByText('無効')).toBeInTheDocument();
    expect(screen.getByText('復活')).toBeInTheDocument();
    expect(screen.queryByText('削除')).not.toBeInTheDocument();
  });

  it('有効・無効スタッフが混在する場合、それぞれ正しいボタンが表示される', () => {
    setupStore({
      staffs: [
        { id: 's1', name: '田中太郎', role: 'kitchen', employment_type: 'full_time', is_active: true, created_at: '', updated_at: '' },
        { id: 's2', name: '山田花子', role: 'hall', employment_type: 'part_time', is_active: false, created_at: '', updated_at: '' },
      ],
    });

    render(<StaffPage />);

    expect(screen.getByText('削除')).toBeInTheDocument();
    expect(screen.getByText('復活')).toBeInTheDocument();
  });

  it('復活ボタンをクリックすると updateStaff(id, { is_active: true }) が呼ばれる', async () => {
    const user = userEvent.setup();
    setupStore({
      staffs: [
        { id: 's2', name: '山田花子', role: 'hall', employment_type: 'part_time', is_active: false, created_at: '', updated_at: '' },
      ],
    });

    render(<StaffPage />);

    await user.click(screen.getByText('復活'));
    expect(mockUpdateStaff).toHaveBeenCalledWith('s2', { is_active: true });
  });

  it('スタッフ一覧が空の場合、空メッセージが表示される', () => {
    setupStore({ staffs: [] });

    render(<StaffPage />);

    expect(screen.getByText('スタッフが登録されていません')).toBeInTheDocument();
  });

  it('ローディング中はスピナーが表示される', () => {
    setupStore({ loading: true });

    const { container } = render(<StaffPage />);

    // LoadingSpinner renders a div with animation
    expect(container.querySelector('div')).toBeInTheDocument();
    expect(screen.queryByText('スタッフ管理')).not.toBeInTheDocument();
  });
});
