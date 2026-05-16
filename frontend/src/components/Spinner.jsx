export default function Spinner({ fullPage }) {
  if (fullPage) {
    return (
      <div className="spinner-overlay">
        <div className="spinner" />
      </div>
    );
  }
  return <div className="spinner" />;
}
