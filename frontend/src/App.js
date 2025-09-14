import React, { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

function App() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [newUser, setNewUser] = useState({ name: '', email: '' });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await axios.get(`${API_BASE_URL}/users`);
      setUsers(response.data);
    } catch (err) {
      setError('Failed to fetch users: ' + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!newUser.name.trim() || !newUser.email.trim()) {
      setError('Name and email are required');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      setSuccess('');
      
      const response = await axios.post(`${API_BASE_URL}/users`, newUser);
      setUsers([...users, response.data]);
      setNewUser({ name: '', email: '' });
      setSuccess('User created successfully!');
    } catch (err) {
      setError('Failed to create user: ' + (err.response?.data?.error || err.message));
    } finally {
      setSubmitting(false);
    }
  };

  const handleInputChange = (e) => {
    setNewUser({
      ...newUser,
      [e.target.name]: e.target.value
    });
  };

  return (
    <div className="container">
      <div className="header">
        <h1>TaskHub - Task Management</h1>
        <p>A simple 2-tier application demonstrating backend and frontend integration</p>
      </div>

      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}

      <div className="card">
        <h2>Add New User</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="name">Name:</label>
            <input
              type="text"
              id="name"
              name="name"
              value={newUser.name}
              onChange={handleInputChange}
              placeholder="Enter user name"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="email">Email:</label>
            <input
              type="email"
              id="email"
              name="email"
              value={newUser.email}
              onChange={handleInputChange}
              placeholder="Enter email address"
              required
            />
          </div>
          <button type="submit" className="btn" disabled={submitting}>
            {submitting ? 'Creating...' : 'Add User'}
          </button>
        </form>
      </div>

      <div className="card">
        <h2>Users ({users.length})</h2>
        {loading ? (
          <div className="loading">Loading users...</div>
        ) : users.length > 0 ? (
          <div>
            {users.map((user) => (
              <div key={user.id} className="user-item">
                <div className="user-info">
                  <div className="user-name">{user.name}</div>
                  <div className="user-email">{user.email}</div>
                </div>
                <div className="user-id">ID: {user.id}</div>
              </div>
            ))}
          </div>
        ) : (
          <div className="loading">No users found</div>
        )}
        <button onClick={fetchUsers} className="btn" style={{marginTop: '20px'}}>
          Refresh
        </button>
      </div>
    </div>
  );
}

export default App;