import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import Header from './Header';

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    minHeight: '100vh',
    background: '#F9FAFB',
  },
  main: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  content: {
    flex: 1,
    padding: 24,
  },
};

export default function AppLayout() {
  return (
    <div style={styles.container}>
      <Sidebar />
      <div style={styles.main}>
        <Header />
        <main style={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
