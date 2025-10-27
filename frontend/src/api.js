import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

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

// Stats API
export const getStats = () => {
  return api.get('/stats');
};

export default api;
