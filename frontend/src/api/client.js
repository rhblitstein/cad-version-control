import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Projects
export const projectsAPI = {
  list: () => api.get('/projects'),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
}

// Branches
export const branchesAPI = {
  list: (projectId) => api.get(`/projects/${projectId}/branches`),
  get: (id) => api.get(`/branches/${id}`),
  create: (projectId, data) => api.post(`/projects/${projectId}/branches`, data),
}

// Commits
export const commitsAPI = {
  list: (branchId, params) => api.get(`/branches/${branchId}/commits`, { params }),
  get: (id) => api.get(`/commits/${id}`),
  create: (projectId, formData) => api.post(`/projects/${projectId}/commits`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
}

// Files
export const filesAPI = {
  download: (versionId) => api.get(`/file-versions/${versionId}/download`, {
    responseType: 'blob',
  }),
  versions: (fileId) => api.get(`/files/${fileId}/versions`),
}

// Merge Requests
export const mergeRequestsAPI = {
  list: (params) => api.get('/merge-requests', { params }),
  get: (id) => api.get(`/merge-requests/${id}`),
  create: (data) => api.post('/merge-requests', data),
  approve: (id, data) => api.post(`/merge-requests/${id}/approve`, data),
  merge: (id) => api.post(`/merge-requests/${id}/merge`),
  addComment: (id, data) => api.post(`/merge-requests/${id}/comments`, data),
  getComments: (id) => api.get(`/merge-requests/${id}/comments`),
  getConflicts: (id) => api.get(`/merge-requests/${id}/conflicts`),
}

// Conflicts
export const conflictsAPI = {
  resolve: (id, data) => api.post(`/conflicts/${id}/resolve`, data),
  getDiff: (id) => api.get(`/conflicts/${id}/diff`),
}

export default api