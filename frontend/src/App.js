import React, { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

function App() {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [newTask, setNewTask] = useState({ 
    title: '', 
    description: '', 
    status: 'pending' 
  });
  const [submitting, setSubmitting] = useState(false);
  const [editingTask, setEditingTask] = useState(null);

  useEffect(() => {
    fetchTasks();
  }, []);

  const fetchTasks = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await axios.get(`${API_BASE_URL}/tasks`);
      setTasks(response.data);
    } catch (err) {
      setError('Failed to fetch tasks: ' + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!newTask.title.trim()) {
      setError('Task title is required');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      setSuccess('');
      
      const response = await axios.post(`${API_BASE_URL}/tasks`, newTask);
      setTasks([response.data, ...tasks]);
      setNewTask({ title: '', description: '', status: 'pending' });
      setSuccess('Task created successfully!');
    } catch (err) {
      setError('Failed to create task: ' + (err.response?.data?.error || err.message));
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async (taskId, updatedTask) => {
    try {
      setError('');
      setSuccess('');
      
      const response = await axios.put(`${API_BASE_URL}/tasks/${taskId}`, updatedTask);
      setTasks(tasks.map(task => 
        task.id === taskId ? response.data : task
      ));
      setEditingTask(null);
      setSuccess('Task updated successfully!');
    } catch (err) {
      setError('Failed to update task: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleDelete = async (taskId) => {
    if (!window.confirm('Are you sure you want to delete this task?')) {
      return;
    }

    try {
      setError('');
      setSuccess('');
      
      await axios.delete(`${API_BASE_URL}/tasks/${taskId}`);
      setTasks(tasks.filter(task => task.id !== taskId));
      setSuccess('Task deleted successfully!');
    } catch (err) {
      setError('Failed to delete task: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleInputChange = (e) => {
    setNewTask({
      ...newTask,
      [e.target.name]: e.target.value
    });
  };

  const handleEditInputChange = (e, taskId) => {
    setEditingTask({
      ...editingTask,
      [e.target.name]: e.target.value
    });
  };

  const startEdit = (task) => {
    setEditingTask({ ...task });
  };

  const cancelEdit = () => {
    setEditingTask(null);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed':
        return '#28a745';
      case 'in_progress':
        return '#ffc107';
      case 'pending':
      default:
        return '#6c757d';
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="container">
      <div className="header">
        <h1>TaskHub - Task Management</h1>
        <p>A modern task management application demonstrating backend and frontend integration</p>
      </div>

      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}

      <div className="card">
        <h2>Add New Task</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="title">Title:</label>
            <input
              type="text"
              id="title"
              name="title"
              value={newTask.title}
              onChange={handleInputChange}
              placeholder="Enter task title"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="description">Description:</label>
            <textarea
              id="description"
              name="description"
              value={newTask.description}
              onChange={handleInputChange}
              placeholder="Enter task description (optional)"
              rows="3"
            />
          </div>
          <div className="form-group">
            <label htmlFor="status">Status:</label>
            <select
              id="status"
              name="status"
              value={newTask.status}
              onChange={handleInputChange}
            >
              <option value="pending">Pending</option>
              <option value="in_progress">In Progress</option>
              <option value="completed">Completed</option>
            </select>
          </div>
          <button type="submit" className="btn" disabled={submitting}>
            {submitting ? 'Creating...' : 'Add Task'}
          </button>
        </form>
      </div>

      <div className="card">
        <h2>Tasks ({tasks.length})</h2>
        {loading ? (
          <div className="loading">Loading tasks...</div>
        ) : tasks.length > 0 ? (
          <div>
            {tasks.map((task) => (
              <div key={task.id} className="task-item">
                {editingTask && editingTask.id === task.id ? (
                  <div className="edit-form">
                    <div className="form-group">
                      <input
                        type="text"
                        name="title"
                        value={editingTask.title}
                        onChange={handleEditInputChange}
                        placeholder="Task title"
                      />
                    </div>
                    <div className="form-group">
                      <textarea
                        name="description"
                        value={editingTask.description}
                        onChange={handleEditInputChange}
                        placeholder="Task description"
                        rows="2"
                      />
                    </div>
                    <div className="form-group">
                      <select
                        name="status"
                        value={editingTask.status}
                        onChange={handleEditInputChange}
                      >
                        <option value="pending">Pending</option>
                        <option value="in_progress">In Progress</option>
                        <option value="completed">Completed</option>
                      </select>
                    </div>
                    <div className="edit-buttons">
                      <button 
                        onClick={() => handleUpdate(task.id, editingTask)}
                        className="btn btn-sm btn-save"
                      >
                        Save
                      </button>
                      <button 
                        onClick={cancelEdit}
                        className="btn btn-sm btn-cancel"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="task-content">
                    <div className="task-header">
                      <div className="task-title">{task.title}</div>
                      <div 
                        className="task-status"
                        style={{ backgroundColor: getStatusColor(task.status) }}
                      >
                        {task.status.replace('_', ' ').toUpperCase()}
                      </div>
                    </div>
                    {task.description && (
                      <div className="task-description">{task.description}</div>
                    )}
                    <div className="task-footer">
                      <div className="task-meta">
                        <span className="task-id">ID: {task.id}</span>
                        <span className="task-date">Created: {formatDate(task.created_at)}</span>
                      </div>
                      <div className="task-actions">
                        <button 
                          onClick={() => startEdit(task)}
                          className="btn btn-sm btn-edit"
                        >
                          Edit
                        </button>
                        <button 
                          onClick={() => handleDelete(task.id)}
                          className="btn btn-sm btn-delete"
                        >
                          Delete
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="loading">No tasks found</div>
        )}
        <button onClick={fetchTasks} className="btn" style={{marginTop: '20px'}}>
          Refresh
        </button>
      </div>
    </div>
  );
}

export default App;