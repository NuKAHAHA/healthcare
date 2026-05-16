import { useState, useEffect, useMemo } from 'react';
import { getPatients } from '../api/patients';
import { getAppointments } from '../api/appointments';
import { getReport } from '../api/treatments';
import Modal from '../components/Modal';
import Spinner from '../components/Spinner';
import { useToast } from '../components/Toast';

const STATUS_CLASS = { pending: 'warning', completed: 'success', cancelled: 'danger' };

function fmtDate(iso) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

function fmtDateTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export default function Reports() {
  const [patients, setPatients]         = useState([]);
  const [appointments, setAppointments] = useState([]);
  const [loadingData, setLoadingData]   = useState(true);

  const [search, setSearch]                 = useState('');
  const [selectedPatient, setSelectedPatient] = useState(null);

  const [reportModal, setReportModal]     = useState(false);
  const [reportData, setReportData]       = useState(null);
  const [reportLoading, setReportLoading] = useState(false);

  const toast = useToast();

  useEffect(() => {
    Promise.all([
      getPatients(1, 100),
      getAppointments(1, 100),
    ])
      .then(([pRes, aRes]) => {
        setPatients(pRes.data.patients ?? pRes.data);
        setAppointments(aRes.data.appointments ?? aRes.data);
      })
      .catch(() => toast('Failed to load data', 'error'))
      .finally(() => setLoadingData(false));
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return patients;
    return patients.filter((p) =>
      `${p.firstName} ${p.lastName}`.toLowerCase().includes(q) ||
      p.email.toLowerCase().includes(q)
    );
  }, [search, patients]);

  function patientAppts(patientId) {
    return appointments
      .filter((a) => a.patientId === patientId)
      .sort((a, b) => new Date(b.scheduledAt) - new Date(a.scheduledAt));
  }

  async function viewReport(apptId) {
    setReportData(null);
    setReportModal(true);
    setReportLoading(true);
    try {
      const { data } = await getReport(apptId);
      setReportData(data);
    } catch {
      toast('No report found — appointment may have no treatment yet', 'warning');
      setReportModal(false);
    } finally {
      setReportLoading(false);
    }
  }

  if (loadingData) return <Spinner fullPage />;

  return (
    <div className="page">
      <div className="page-header"><h2>Reports</h2></div>

      {/* Search */}
      <div className="toolbar">
        <input
          className="search-input"
          placeholder="Search by patient name or email…"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setSelectedPatient(null); }}
        />
        {selectedPatient && (
          <button className="btn btn-outline btn-sm" onClick={() => setSelectedPatient(null)}>
            ← All patients
          </button>
        )}
      </div>

      {/* Patient list or selected patient appointments */}
      {!selectedPatient ? (
        <>
          {filtered.length === 0 && search && (
            <p className="page-subtitle" style={{ marginTop: '16px' }}>No patients found.</p>
          )}
          {filtered.length === 0 && !search && (
            <p className="page-subtitle" style={{ marginTop: '16px' }}>Start typing to search patients.</p>
          )}
          <div className="patient-card-grid">
            {filtered.map((p) => (
              <button
                key={p.id}
                className="patient-card"
                onClick={() => setSelectedPatient(p)}
              >
                <div className="patient-card-name">{p.firstName} {p.lastName}</div>
                <div className="patient-card-meta">{p.email}</div>
                <div className="patient-card-meta">DOB: {fmtDate(p.dateOfBirth)}</div>
              </button>
            ))}
          </div>
        </>
      ) : (
        <>
          <div className="selected-patient-header">
            <h3>{selectedPatient.firstName} {selectedPatient.lastName}</h3>
            <span className="patient-card-meta">{selectedPatient.email}</span>
          </div>

          {patientAppts(selectedPatient.id).length === 0 ? (
            <p className="page-subtitle" style={{ marginTop: '16px' }}>
              No appointments found for this patient.
            </p>
          ) : (
            <div className="table-wrapper" style={{ marginTop: '16px' }}>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Date &amp; Time</th><th>Reason</th><th>Status</th><th>Report</th>
                  </tr>
                </thead>
                <tbody>
                  {patientAppts(selectedPatient.id).map((a) => (
                    <tr key={a.id}>
                      <td>{fmtDateTime(a.scheduledAt)}</td>
                      <td>{a.reason}</td>
                      <td>
                        <span className={`badge badge-${STATUS_CLASS[a.status] ?? 'info'}`}>
                          {a.status}
                        </span>
                      </td>
                      <td>
                        <button className="btn btn-sm btn-primary" onClick={() => viewReport(a.id)}>
                          View Report
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {reportModal && (
        <Modal title="Treatment Report" onClose={() => setReportModal(false)}>
          {reportLoading ? <Spinner /> : reportData && (
            <>
              <dl className="detail-list">
                <dt>Patient</dt>     <dd>{selectedPatient?.firstName} {selectedPatient?.lastName}</dd>
                <dt>Diagnosis</dt>   <dd>{reportData.diagnosis}</dd>
                <dt>Prescription</dt><dd>{reportData.prescription || '—'}</dd>
                {reportData.notes && (<><dt>Notes</dt><dd>{reportData.notes}</dd></>)}
                <dt>Generated</dt>   <dd>{fmtDateTime(reportData.generatedAt)}</dd>
              </dl>
              <div className="form-actions" style={{ marginTop: '16px' }}>
                <button className="btn btn-outline" onClick={() => setReportModal(false)}>Close</button>
              </div>
            </>
          )}
        </Modal>
      )}
    </div>
  );
}
