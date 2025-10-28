import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add JWT token to requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Handle 401 errors (unauthorized)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('isAuthenticated');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const login = (username, password) => {
  return api.post('/auth/login', { username, password });
};

export const register = (data) => {
  return api.post('/auth/register', data);
};

export const getProfile = () => {
  return api.get('/profile');
};

export const changePassword = (oldPassword, newPassword) => {
  return api.post('/change-password', {
    old_password: oldPassword,
    new_password: newPassword,
  });
};

// Pipelines API
export const getPipelines = (params = {}) => {
  return api.get('/pipelines', { params });
};

export const getPipeline = (id) => {
  return api.get(`/pipelines/${id}`);
};

export const createPipeline = (data) => {
  return api.post('/pipelines', data);
};

export const updatePipeline = (id, data) => {
  return api.put(`/pipelines/${id}`, data);
};

export const deletePipeline = (id) => {
  return api.delete(`/pipelines/${id}`);
};

export const deleteAllPipelines = () => {
  return api.delete('/pipelines/all');
};

export const importCSV = (file, onUploadProgress) => {
  const formData = new FormData();
  formData.append('file', file);

  return api.post('/pipelines/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress,
  });
};

export const getImportProgress = () => {
  return api.get('/pipelines/import/progress');
};

// DI319 API
export const getDI319Data = (params = {}) => {
  return api.get('/di319', { params });
};

export const deleteDI319All = () => {
  return api.delete('/di319/all');
};

// Stats API
export const getStats = () => {
  return api.get('/stats');
};

export default api;
