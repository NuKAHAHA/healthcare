import { useState, useEffect, useMemo } from 'react';
import { getPatients } from '../api/patients';
import { getAuditLogs } from '../api/audit';
import Modal from '../components/Modal';
import Spinner from '../components/Spinner';
import { useToast } from '../components/Toast';

function fmtDate(iso) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

function fmtDateTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export default function AuditLogs() {
  const [patients, setPatients]           = useState([]);
  const [patientsLoading, setPatientsLoading] = useState(true);
  const [search, setSearch]               = useState('');
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [logs, setLogs]                   = useState([]);
  const [logsLoading, setLogsLoading]     = useState(false);
  const [detailLog, setDetailLog]         = useState(null);
  const toast = useToast();

  useEffect(() => {
    getPatients(1, 100)
      .then(({ data }) => setPatients(data.patients ?? data))
      .catch(() => toast('Failed to load patients', 'error'))
      .finally(() => setPatientsLoading(false));
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return patients;
    return patients.filter((p) =>
      `${p.firstName} ${p.lastName}`.toLowerCase().includes(q) ||
      p.email.toLowerCase().includes(q)
    );
  }, [search, patients]);

  async function selectPatient(p) {
    setSelectedPatient(p);
    setLogs([]);
    setLogsLoading(true);
    try {
      const { data } = await getAuditLogs(p.id);
      setLogs(Array.isArray(data) ? data : (data.logs ?? []));
    } catch {
      toast('Failed to load audit logs', 'error');
    } finally {
      setLogsLoading(false);
    }
  }

  if (patientsLoading) return <Spinner fullPage />;

  return (
    <div className="page">
      <div className="page-header"><h2>Audit Logs</h2></div>

      {/* Search */}
      <div className="toolbar">
        <input
          className="search-input"
          placeholder="Search patient by name or email…"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setSelectedPatient(null); setLogs([]); }}
        />
        {selectedPatient && (
          <button className="btn btn-outline btn-sm" onClick={() => { setSelectedPatient(null); setLogs([]); }}>
            ← All patients
          </button>
        )}
      </div>

      {/* Patient cards or audit log table */}
      {!selectedPatient ? (
        <>
          {filtered.length === 0 && search && (
            <p className="page-subtitle" style={{ marginTop: '16px' }}>No patients found.</p>
          )}
          {!search && (
            <p className="page-subtitle" style={{ marginTop: '16px' }}>Search for a patient to view their audit history.</p>
          )}
          <div className="patient-card-grid">
            {filtered.map((p) => (
              <button
                key={p.id}
                className="patient-card"
                onClick={() => selectPatient(p)}
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
            <h3>Audit history: {selectedPatient.firstName} {selectedPatient.lastName}</h3>
          </div>

          {logsLoading ? <Spinner /> : (
            <div className="table-wrapper" style={{ marginTop: '16px' }}>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Action</th><th>Resource</th><th>Status</th><th>Details</th><th>Timestamp</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((l) => (
                    <tr
                      key={l.id}
                      className="clickable-row"
                      onClick={() => setDetailLog(l)}
                      style={{ cursor: 'pointer' }}
                    >
                      <td>{l.action}</td>
                      <td>{l.resource}</td>
                      <td>
                        <span className={`badge badge-${l.status === 'success' ? 'success' : 'danger'}`}>
                          {l.status}
                        </span>
                      </td>
                      <td>{l.details}</td>
                      <td>{fmtDateTime(l.createdAt)}</td>
                    </tr>
                  ))}
                  {logs.length === 0 && (
                    <tr><td colSpan={5} className="empty">No audit logs found for this patient.</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {/* Detail modal */}
      {detailLog && (
        <Modal title="Audit Event Details" onClose={() => setDetailLog(null)}>
          <dl className="detail-list">
            <dt>Action</dt>    <dd>{detailLog.action}</dd>
            <dt>Resource</dt>  <dd>{detailLog.resource}</dd>
            <dt>Status</dt>    <dd>{detailLog.status}</dd>
            <dt>Details</dt>   <dd>{detailLog.details}</dd>
            <dt>Timestamp</dt> <dd>{fmtDateTime(detailLog.createdAt)}</dd>
          </dl>
          <div className="form-actions" style={{ marginTop: '16px' }}>
            <button className="btn btn-outline" onClick={() => setDetailLog(null)}>Close</button>
          </div>
        </Modal>
      )}
    </div>
  );
}
