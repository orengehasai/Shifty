import { BrowserRouter, Routes, Route } from 'react-router-dom';
import AppLayout from './components/Layout/AppLayout';
import DashboardPage from './pages/Dashboard';
import StaffPage from './pages/Staff';
import ShiftRequestPage from './pages/ShiftRequest';
import ConstraintsPage from './pages/Constraints';
import ShiftGeneratePage from './pages/ShiftGenerate';
import ShiftEditPage from './pages/ShiftEdit';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/staffs" element={<StaffPage />} />
          <Route path="/requests" element={<ShiftRequestPage />} />
          <Route path="/constraints" element={<ConstraintsPage />} />
          <Route path="/generate" element={<ShiftGeneratePage />} />
          <Route path="/shifts/:patternId/edit" element={<ShiftEditPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
