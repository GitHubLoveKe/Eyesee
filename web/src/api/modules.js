import api from './index'

export const businessSystemApi = {
  list: (params) => api.get('/business-systems', { params }),
  get: (id) => api.get(`/business-systems/${id}`),
  create: (data) => api.post('/business-systems', data),
  update: (id, data) => api.put(`/business-systems/${id}`, data),
  delete: (id) => api.delete(`/business-systems/${id}`),
  listMembers: (id) => api.get(`/business-systems/${id}/members`),
  addMember: (id, data) => api.post(`/business-systems/${id}/members`, data),
  removeMember: (id, userId) => api.delete(`/business-systems/${id}/members/${userId}`),
}

export const zoneApi = {
  list: (params) => api.get('/zones', { params }),
  get: (id) => api.get(`/zones/${id}`),
  create: (data) => api.post('/zones', data),
  update: (id, data) => api.put(`/zones/${id}`, data),
  delete: (id) => api.delete(`/zones/${id}`),
  listSegments: (id) => api.get(`/zones/${id}/segments`),
  createSegment: (id, data) => api.post(`/zones/${id}/segments`, data),
  deleteSegment: (id, segmentId) => api.delete(`/zones/${id}/segments/${segmentId}`),
}

export const resourceTypeApi = {
  tree: () => api.get('/resource-types/tree'),
  list: (params) => api.get('/resource-types', { params }),
  get: (id) => api.get(`/resource-types/${id}`),
  create: (data) => api.post('/resource-types', data),
  update: (id, data) => api.put(`/resource-types/${id}`, data),
  delete: (id) => api.delete(`/resource-types/${id}`),
}

export const instanceApi = {
  list: (params) => api.get('/instances', { params }),
  get: (id) => api.get(`/instances/${id}`),
  create: (data) => api.post('/instances', data),
  update: (id, data) => api.put(`/instances/${id}`, data),
  delete: (id) => api.delete(`/instances/${id}`),
}

export const dataSinkApi = {
  list: (params) => api.get('/data-sinks', { params }),
  get: (id) => api.get(`/data-sinks/${id}`),
  create: (data) => api.post('/data-sinks', data),
  update: (id, data) => api.put(`/data-sinks/${id}`, data),
  delete: (id) => api.delete(`/data-sinks/${id}`),
}

export const collectionTaskApi = {
  list: (params) => api.get('/collection-tasks', { params }),
  get: (id) => api.get(`/collection-tasks/${id}`),
  create: (data) => api.post('/collection-tasks', data),
  update: (id, data) => api.put(`/collection-tasks/${id}`, data),
  delete: (id) => api.delete(`/collection-tasks/${id}`),
  enable: (id) => api.post(`/collection-tasks/${id}/enable`),
  disable: (id) => api.post(`/collection-tasks/${id}/disable`),
}

export const collectorApi = {
  list: (params) => api.get('/collector-nodes', { params }),
  get: (id) => api.get(`/collector-nodes/${id}`),
  create: (data) => api.post('/collector-nodes', data),
  update: (id, data) => api.put(`/collector-nodes/${id}`, data),
  delete: (id) => api.delete(`/collector-nodes/${id}`),
  heartbeat: (id) => api.post(`/collector-nodes/${id}/heartbeat`),
}

export const statusApi = {
  global: () => api.get('/status/global'),
  zoneOverview: (zoneId) => api.get(`/status/zones/${zoneId}`),
}
