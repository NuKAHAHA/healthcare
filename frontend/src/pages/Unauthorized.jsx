import { useNavigate } from 'react-router-dom';

export default function Unauthorized() {
  const navigate = useNavigate();
  return (
    <div className="error-page">
      <h1>403</h1>
      <p>You do not have permission to access this page.</p>
      <button className="btn btn-primary" onClick={() => navigate('/dashboard')}>
        Back to Dashboard
      </button>
    </div>
  );
}
