import { useState, useEffect, useCallback } from 'react';
import { getAppointments, createAppointment, updateAppointmentStatus } from '../api/appointments';
import { createTreatment, getReport } from '../api/treatments';
import { getPatients } from '../api/patients';
import { getUsers } from '../api/users';
import Modal from '../components/Modal';
import SearchableDropdown from '../components/SearchableDropdown';
import Spinner from '../components/Spinner';
import { useToast } from '../components/Toast';
import { useAuth } from '../context/AuthContext';

const STATUS_CLASS = { pending: 'warning', completed: 'success', cancelled: 'danger' };
const TREAT_EMPTY  = { appointmentId: null, diagnosis: '', prescription: '', notes: '' };

function fmtDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// ── Doctor view ──────────────────────────────────────────────────────────────
function DoctorAppointments() {
  const [appts, setAppts]     = useState([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage]       = useState(1);

  const [treatModal, setTreatModal] = useState(false);
  const [treatForm, setTreatForm]   = useState(TREAT_EMPTY);
  const [saving, setSaving]         = useState(false);

  const [reportModal, setReportModal]     = useState(false);
  const [reportData, setReportData]       = useState(null);
  const [reportLoading, setReportLoading] = useState(false);

  const toast = useToast();

  useEffect(() => {
    setLoading(true);
    getAppointments(page)
      .then(({ data }) => setAppts(data.appointments ?? data))
      .catch(() => toast('Failed to load appointments', 'error'))
      .finally(() => setLoading(false));
  }, [page]);

  function openTreat(appt) {
    setTreatForm({ ...TREAT_EMPTY, appointmentId: appt.id });
    setTreatModal(true);
  }

  async function openReport(appt) {
    setReportData(null);
    setReportModal(true);
    setReportLoading(true);
    try {
      const { data } = await getReport(appt.id);
      setReportData(data);
    } catch {
      toast('No report found — add a treatment first', 'warning');
      setReportModal(false);
    } finally {
      setReportLoading(false);
    }
  }

  async function handleTreat(e) {
    e.preventDefault();
    setSaving(true);
    try {
      await createTreatment({
        appointmentId: treatForm.appointmentId,
        diagnosis:     treatForm.diagnosis,
        prescription:  treatForm.prescription,
        notes:         treatForm.notes || null,
      });
      toast('Treatment saved', 'success');
      setTreatModal(false);
      setTreatForm(TREAT_EMPTY);
      // Refetch so the row status updates to 'completed'
      const { data } = await getAppointments(page);
      setAppts(data.appointments ?? data);
    } catch (err) {
      console.error('handleTreat:', err);
      toast(err.response?.data?.message || 'Failed to save treatment', 'error');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page">
      <div className="page-header"><h2>My Appointments</h2></div>

      {loading ? <Spinner /> : (
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Patient</th><th>Date &amp; Time</th>
                <th>Reason</th><th>Status</th><th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {appts.map((a) => (
                <tr key={a.id}>
                  <td>{a.patientName || `Patient #${a.patientId}`}</td>
                  <td>{fmtDate(a.scheduledAt)}</td>
                  <td>{a.reason}</td>
                  <td>
                    <span className={`badge badge-${STATUS_CLASS[a.status] ?? 'info'}`}>
                      {a.status}
                    </span>
                  </td>
                  <td className="action-cell">
                    {a.status === 'pending' && (
                      <button className="btn btn-sm btn-primary" onClick={() => openTreat(a)}>
                        Add Treatment
                      </button>
                    )}
                    <button className="btn btn-sm btn-outline" onClick={() => openReport(a)}>
                      View Report
                    </button>
                  </td>
                </tr>
              ))}
              {appts.length === 0 && (
                <tr><td colSpan={5} className="empty">No appointments found.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <div className="pagination">
        <button className="btn btn-sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>Previous</button>
        <span>Page {page}</span>
        <button className="btn btn-sm" disabled={appts.length < 50} onClick={() => setPage(p => p + 1)}>Next</button>
      </div>

      {treatModal && (
        <Modal title="Add Treatment" onClose={() => setTreatModal(false)}>
          <form onSubmit={handleTreat}>
            <div className="form-group">
              <label>Diagnosis</label>
              <textarea rows={3} value={treatForm.diagnosis}
                onChange={(e) => setTreatForm({ ...treatForm, diagnosis: e.target.value })} required />
            </div>
            <div className="form-group">
              <label>Prescription</label>
              <textarea rows={3} value={treatForm.prescription}
                onChange={(e) => setTreatForm({ ...treatForm, prescription: e.target.value })} />
            </div>
            <div className="form-group">
              <label>Notes (optional)</label>
              <textarea rows={2} value={treatForm.notes}
                onChange={(e) => setTreatForm({ ...treatForm, notes: e.target.value })} />
            </div>
            <div className="form-actions">
              <button type="button" className="btn btn-outline" onClick={() => setTreatModal(false)}>Cancel</button>
              <button type="submit" className="btn btn-primary" disabled={saving}>
                {saving ? 'Saving…' : 'Save Treatment'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {reportModal && (
        <Modal title="Appointment Report" onClose={() => setReportModal(false)}>
          {reportLoading ? <Spinner /> : reportData && (
            <dl className="detail-list">
              <dt>Diagnosis</dt>      <dd>{reportData.diagnosis}</dd>
              <dt>Prescription</dt>  <dd>{reportData.prescription || '—'}</dd>
              {reportData.notes && (<><dt>Notes</dt><dd>{reportData.notes}</dd></>)}
              <dt>Generated</dt>     <dd>{fmtDate(reportData.generatedAt)}</dd>
            </dl>
          )}
          <div className="form-actions" style={{ marginTop: '16px' }}>
            <button className="btn btn-outline" onClick={() => setReportModal(false)}>Close</button>
          </div>
        </Modal>
      )}
    </div>
  );
}

// ── Admin / Registrar view ────────────────────────────────────────────────────
function AdminAppointments() {
  const [appts, setAppts]         = useState([]);
  const [loading, setLoading]     = useState(true);
  const [page, setPage]           = useState(1);
  const [showModal, setShowModal] = useState(false);
  const [saving, setSaving]       = useState(false);

  // ID state (sent to API)
  const [patientId, setPatientId]     = useState(null);
  const [doctorId, setDoctorId]       = useState(null);
  // Label state (shown in dropdowns)
  const [patientLabel, setPatientLabel] = useState('');
  const [doctorLabel, setDoctorLabel]   = useState('');
  // Other fields
  const [scheduledAt, setScheduledAt] = useState('');
  const [reason, setReason]           = useState('');

  // Cached lists for dropdown search (loaded once per modal open)
  const [allPatients, setAllPatients] = useState([]);
  const [allDoctors, setAllDoctors]   = useState([]);

  const toast = useToast();

  useEffect(() => {
    setLoading(true);
    getAppointments(page)
      .then(({ data }) => setAppts(data.appointments ?? data))
      .catch(() => toast('Failed to load appointments', 'error'))
      .finally(() => setLoading(false));
  }, [page]);

  // Load patients + doctors when modal opens
  useEffect(() => {
    if (!showModal) return;
    getPatients(1, 100)
      .then(({ data }) => setAllPatients(data.patients ?? data))
      .catch(() => toast('Failed to load patient list', 'error'));
    getUsers('doctor')
      .then(({ data }) => setAllDoctors(Array.isArray(data) ? data : []))
      .catch(() => toast('Failed to load doctor list', 'error'));
  }, [showModal]);

  function resetModal() {
    setPatientId(null); setPatientLabel('');
    setDoctorId(null);  setDoctorLabel('');
    setScheduledAt(''); setReason('');
  }

  function openModal() { resetModal(); setShowModal(true); }
  function closeModal() { setShowModal(false); }

  // Client-side search functions passed to SearchableDropdown
  const searchPatients = useCallback(async (q) => {
    const lower = q.toLowerCase();
    return allPatients
      .filter((p) =>
        `${p.firstName} ${p.lastName}`.toLowerCase().includes(lower) ||
        p.email.toLowerCase().includes(lower)
      )
      .map((p) => ({ id: p.id, label: `${p.firstName} ${p.lastName} — ${p.email}` }));
  }, [allPatients]);

  const searchDoctors = useCallback(async (q) => {
    const lower = q.toLowerCase();
    return allDoctors
      .filter((d) =>
        `${d.firstName} ${d.lastName}`.toLowerCase().includes(lower) ||
        d.email.toLowerCase().includes(lower)
      )
      .map((d) => ({ id: d.id, label: `${d.firstName} ${d.lastName} — ${d.email}` }));
  }, [allDoctors]);

  async function handleCancel(appt) {
    if (!window.confirm(`Cancel appointment for ${appt.patientName || 'this patient'}?`)) return;
    try {
      await updateAppointmentStatus(appt.id, 'cancelled');
      const { data: listData } = await getAppointments(page);
      setAppts(listData.appointments ?? listData);
      toast('Appointment cancelled', 'success');
    } catch (err) {
      console.error('handleCancel:', err);
      toast(err.response?.data?.message || 'Failed to cancel appointment', 'error');
    }
  }

  async function handleCreate(e) {
    e.preventDefault();
    if (!patientId) { toast('Please select a patient', 'error'); return; }
    if (!doctorId)  { toast('Please select a doctor', 'error');  return; }
    setSaving(true);
    try {
      await createAppointment({
        patientId,
        doctorId,
        scheduledAt: new Date(scheduledAt).toISOString(),
        reason,
      });
      const { data: listData } = await getAppointments(page);
      setAppts(listData.appointments ?? listData);
      closeModal();
      toast('Appointment scheduled', 'success');
    } catch (err) {
      console.error('createAppointment:', err);
      toast(err.response?.data?.message || 'Failed to schedule appointment', 'error');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>Appointments</h2>
        <button className="btn btn-primary" onClick={openModal}>+ Schedule</button>
      </div>

      {loading ? <Spinner /> : (
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Patient</th><th>Scheduled</th><th>Reason</th><th>Status</th><th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {appts.map((a) => (
                <tr key={a.id}>
                  <td>{a.patientName || `Patient #${a.patientId}`}</td>
                  <td>{fmtDate(a.scheduledAt)}</td>
                  <td>{a.reason}</td>
                  <td>
                    <span className={`badge badge-${STATUS_CLASS[a.status] ?? 'info'}`}>
                      {a.status}
                    </span>
                  </td>
                  <td className="action-cell">
                    {a.status === 'pending' && (
                      <button className="btn btn-sm btn-danger" onClick={() => handleCancel(a)}>
                        Cancel
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {appts.length === 0 && (
                <tr><td colSpan={5} className="empty">No appointments found.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <div className="pagination">
        <button className="btn btn-sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>Previous</button>
        <span>Page {page}</span>
        <button className="btn btn-sm" disabled={appts.length < 50} onClick={() => setPage(p => p + 1)}>Next</button>
      </div>

      {showModal && (
        <Modal title="Schedule Appointment" onClose={closeModal}>
          <form onSubmit={handleCreate}>
            <SearchableDropdown
              label="Patient"
              placeholder="Type name or email to search…"
              onSearch={searchPatients}
              onSelect={(id, label) => { setPatientId(id); setPatientLabel(label); }}
              displayValue={patientLabel}
              required
            />
            <SearchableDropdown
              label="Doctor"
              placeholder="Type name to search…"
              onSearch={searchDoctors}
              onSelect={(id, label) => { setDoctorId(id); setDoctorLabel(label); }}
              displayValue={doctorLabel}
              required
            />
            <div className="form-group">
              <label>Scheduled At</label>
              <input type="datetime-local" value={scheduledAt}
                onChange={(e) => setScheduledAt(e.target.value)} required />
            </div>
            <div className="form-group">
              <label>Reason</label>
              <input type="text" value={reason}
                onChange={(e) => setReason(e.target.value)} required />
            </div>
            <div className="form-actions">
              <button type="button" className="btn btn-outline" onClick={closeModal}>Cancel</button>
              <button type="submit" className="btn btn-primary" disabled={saving}>
                {saving ? 'Saving…' : 'Schedule'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  );
}

// ── Entry point ───────────────────────────────────────────────────────────────
export default function Appointments() {
  const { user } = useAuth();
  if (!user) return <Spinner fullPage />;
  return user.role === 'doctor' ? <DoctorAppointments /> : <AdminAppointments />;
}
