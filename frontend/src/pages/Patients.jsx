import { useState, useEffect } from 'react';
import { getPatients, createPatient } from '../api/patients';
import Modal from '../components/Modal';
import Spinner from '../components/Spinner';
import { useToast } from '../components/Toast';
import { useAuth } from '../context/AuthContext';

const EMPTY = {
  firstName: '', lastName: '', dateOfBirth: '',
  gender: 'M', phone: '', email: '', address: '',
};

export default function Patients() {
  const [patients, setPatients]   = useState([]);
  const [loading, setLoading]     = useState(true);
  const [page, setPage]           = useState(1);
  const [showModal, setShowModal] = useState(false);
  const [form, setForm]           = useState(EMPTY);
  const [saving, setSaving]       = useState(false);
  const toast   = useToast();
  const { user } = useAuth();
  const isDoctor = user?.role === 'doctor';

  useEffect(() => {
    setLoading(true);
    getPatients(page)
      .then(({ data }) => setPatients(data.patients ?? data))
      .catch(() => toast('Failed to load patients', 'error'))
      .finally(() => setLoading(false));
  }, [page]);

  async function handleCreate(e) {
    e.preventDefault();
    setSaving(true);
    try {
      await createPatient(form);
      const { data: listData } = await getPatients(page);
      setPatients(listData.patients ?? listData);
      setShowModal(false);
      setForm(EMPTY);
      toast('Patient registered', 'success');
    } catch (err) {
      console.error('createPatient:', err);
      toast(err.response?.data?.message || 'Failed to create patient', 'error');
    } finally {
      setSaving(false);
    }
  }

  const field = (key, label, type = 'text', required = true) => (
    <div className="form-group" key={key}>
      <label>{label}</label>
      <input
        type={type}
        value={form[key]}
        onChange={(e) => setForm({ ...form, [key]: e.target.value })}
        required={required}
      />
    </div>
  );

  return (
    <div className="page">
      <div className="page-header">
        <h2>{isDoctor ? 'My Patients' : 'Patients'}</h2>
        {!isDoctor && (
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Register Patient
          </button>
        )}
      </div>

      {loading ? <Spinner /> : (
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Name</th><th>Date of Birth</th><th>Gender</th>
                <th>Phone</th><th>Email</th>
              </tr>
            </thead>
            <tbody>
              {patients.map((p) => (
                <tr key={p.id}>
                  <td>{p.firstName} {p.lastName}</td>
                  <td>{p.dateOfBirth?.slice(0, 10)}</td>
                  <td>{p.gender === 'M' ? 'Male' : p.gender === 'F' ? 'Female' : 'Other'}</td>
                  <td>{p.phone}</td>
                  <td>{p.email}</td>
                </tr>
              ))}
              {patients.length === 0 && (
                <tr><td colSpan={5} className="empty">No patients found.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <div className="pagination">
        <button className="btn btn-sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
          Previous
        </button>
        <span>Page {page}</span>
        <button className="btn btn-sm" disabled={patients.length < 50} onClick={() => setPage(p => p + 1)}>
          Next
        </button>
      </div>

      {showModal && (
        <Modal title="Register Patient" onClose={() => setShowModal(false)}>
          <form onSubmit={handleCreate}>
            {field('firstName',   'First Name')}
            {field('lastName',    'Last Name')}
            {field('dateOfBirth', 'Date of Birth', 'date')}
            {field('phone',       'Phone')}
            {field('email',       'Email', 'email')}
            {field('address',     'Address', 'text', false)}
            <div className="form-group">
              <label>Gender</label>
              <select value={form.gender} onChange={(e) => setForm({ ...form, gender: e.target.value })}>
                <option value="M">Male</option>
                <option value="F">Female</option>
                <option value="O">Other</option>
              </select>
            </div>
            <div className="form-actions">
              <button type="button" className="btn btn-outline" onClick={() => setShowModal(false)}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary" disabled={saving}>
                {saving ? 'Saving…' : 'Register'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  );
}
