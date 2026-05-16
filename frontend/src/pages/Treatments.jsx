import { useState } from 'react';
import { createTreatment } from '../api/treatments';
import { useToast } from '../components/Toast';

const EMPTY = { appointmentId: '', diagnosis: '', prescription: '', notes: '' };

export default function Treatments() {
  const [form, setForm]   = useState(EMPTY);
  const [saving, setSaving] = useState(false);
  const [last, setLast]   = useState(null);
  const toast = useToast();

  async function handleSubmit(e) {
    e.preventDefault();
    setSaving(true);
    try {
      const payload = { ...form, appointmentId: Number(form.appointmentId) };
      const { data } = await createTreatment(payload);
      setLast(data);
      setForm(EMPTY);
      toast('Treatment recorded', 'success');
    } catch (err) {
      toast(err.response?.data?.error || 'Failed to record treatment', 'error');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>Record Treatment</h2>
      </div>

      <div className="card" style={{ maxWidth: 560 }}>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Appointment ID</label>
            <input type="number" value={form.appointmentId}
              onChange={(e) => setForm({ ...form, appointmentId: e.target.value })} required />
          </div>
          <div className="form-group">
            <label>Diagnosis</label>
            <input type="text" value={form.diagnosis}
              onChange={(e) => setForm({ ...form, diagnosis: e.target.value })} required />
          </div>
          <div className="form-group">
            <label>Prescription</label>
            <input type="text" value={form.prescription}
              onChange={(e) => setForm({ ...form, prescription: e.target.value })} required />
          </div>
          <div className="form-group">
            <label>Notes</label>
            <textarea rows={3} value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })} />
          </div>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? 'Saving…' : 'Record Treatment'}
            </button>
          </div>
        </form>
      </div>

      {last && (
        <div className="card" style={{ maxWidth: 560, marginTop: '1.5rem' }}>
          <h4>Last recorded treatment</h4>
          <p><strong>ID:</strong> {last.id}</p>
          <p><strong>Appointment:</strong> {last.appointmentId}</p>
          <p><strong>Diagnosis:</strong> {last.diagnosis}</p>
          <p><strong>Prescription:</strong> {last.prescription}</p>
          {last.notes && <p><strong>Notes:</strong> {last.notes}</p>}
        </div>
      )}
    </div>
  );
}
