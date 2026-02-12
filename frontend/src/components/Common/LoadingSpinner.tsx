const styles: Record<string, React.CSSProperties> = {
  wrapper: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  spinner: {
    width: 32,
    height: 32,
    border: '3px solid #E5E7EB',
    borderTopColor: '#4F46E5',
    borderRadius: '50%',
    animation: 'spin 0.8s linear infinite',
  },
};

export default function LoadingSpinner() {
  return (
    <div style={styles.wrapper}>
      <div style={styles.spinner} />
      <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
    </div>
  );
}
