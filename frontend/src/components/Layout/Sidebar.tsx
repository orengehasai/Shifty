import { NavLink } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'ダッシュボード' },
  { to: '/staffs', label: 'スタッフ管理' },
  { to: '/requests', label: '希望入力' },
  { to: '/constraints', label: '制約設定' },
  { to: '/generate', label: 'シフト生成' },
];

const styles: Record<string, React.CSSProperties> = {
  sidebar: {
    width: 200,
    minHeight: '100vh',
    background: '#1E1B4B',
    color: '#fff',
    padding: '16px 0',
    flexShrink: 0,
  },
  title: {
    padding: '0 16px 16px',
    fontSize: 13,
    color: '#A5B4FC',
    borderBottom: '1px solid #312E81',
    marginBottom: 8,
  },
  link: {
    display: 'block',
    padding: '10px 16px',
    color: '#C7D2FE',
    textDecoration: 'none',
    fontSize: 14,
    borderLeft: '3px solid transparent',
  },
  activeLink: {
    color: '#fff',
    background: '#312E81',
    borderLeft: '3px solid #818CF8',
  },
};

export default function Sidebar() {
  return (
    <nav style={styles.sidebar}>
      <div style={styles.title}>MENU</div>
      {navItems.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.to === '/'}
          style={({ isActive }) => ({
            ...styles.link,
            ...(isActive ? styles.activeLink : {}),
          })}
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  );
}
