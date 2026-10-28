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

export const userApi = {
  list: (params) => api.get('/users', { params }),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
}

export const notificationChannelApi = {
  list: (params) => api.get('/notification-channels', { params }),
  get: (id) => api.get(`/notification-channels/${id}`),
  create: (data) => api.post('/notification-channels', data),
  update: (id, data) => api.put(`/notification-channels/${id}`, data),
  delete: (id) => api.delete(`/notification-channels/${id}`),
  test: (id) => api.post(`/notification-channels/${id}/test`),
}

export const alertGroupApi = {
  list: () => api.get('/alert-groups'),
  get: (id) => api.get(`/alert-groups/${id}`),
  create: (data) => api.post('/alert-groups', data),
  update: (id, data) => api.put(`/alert-groups/${id}`, data),
  delete: (id) => api.delete(`/alert-groups/${id}`),
  listMembers: (id) => api.get(`/alert-groups/${id}/members`),
  addMember: (id, data) => api.post(`/alert-groups/${id}/members`, data),
  removeMember: (id, userId) => api.delete(`/alert-groups/${id}/members/${userId}`),
}

export const alertStrategyApi = {
  list: (params) => api.get('/alert-strategies', { params }),
  get: (id) => api.get(`/alert-strategies/${id}`),
  create: (data) => api.post('/alert-strategies', data),
  update: (id, data) => api.put(`/alert-strategies/${id}`, data),
  delete: (id) => api.delete(`/alert-strategies/${id}`),
  linkInstance: (id, data) => api.post(`/alert-strategies/${id}/instances`, data),
  listInstances: (id) => api.get(`/alert-strategies/${id}/instances`),
  unlinkInstance: (id, instanceId) => api.delete(`/alert-strategies/${id}/instances/${instanceId}`),
}

export const scheduledTaskApi = {
  list: (params) => api.get('/scheduled-tasks', { params }),
  get: (id) => api.get(`/scheduled-tasks/${id}`),
  create: (data) => api.post('/scheduled-tasks', data),
  update: (id, data) => api.put(`/scheduled-tasks/${id}`, data),
  delete: (id) => api.delete(`/scheduled-tasks/${id}`),
  enable: (id) => api.post(`/scheduled-tasks/${id}/enable`),
  disable: (id) => api.post(`/scheduled-tasks/${id}/disable`),
  listExecutions: (id, params) => api.get(`/scheduled-tasks/${id}/executions`, { params }),
}

export const rcTaskApi = {
  list: (params) => api.get('/rc-tasks', { params }),
  get: (id) => api.get(`/rc-tasks/${id}`),
  create: (data) => api.post('/rc-tasks', data),
  update: (id, data) => api.put(`/rc-tasks/${id}`, data),
  delete: (id) => api.delete(`/rc-tasks/${id}`),
  listRules: (id) => api.get(`/rc-tasks/${id}/rules`),
  createRule: (id, data) => api.post(`/rc-tasks/${id}/rules`, data),
  updateRule: (ruleId, data) => api.put(`/rc-tasks/rules/${ruleId}`, data),
  deleteRule: (ruleId) => api.delete(`/rc-tasks/rules/${ruleId}`),
}

export const metricDefinitionApi = {
  list: (params) => api.get('/metric-definitions', { params }),
  get: (id) => api.get(`/metric-definitions/${id}`),
  create: (data) => api.post('/metric-definitions', data),
  update: (id, data) => api.put(`/metric-definitions/${id}`, data),
  delete: (id) => api.delete(`/metric-definitions/${id}`),
  listQueries: (id) => api.get(`/metric-definitions/${id}/queries`),
  createQuery: (id, data) => api.post(`/metric-definitions/${id}/queries`, data),
  deleteQuery: (queryId) => api.delete(`/metric-definitions/queries/${queryId}`),
}

export const credentialApi = {
  list: (params) => api.get('/credentials', { params }),
  get: (id) => api.get(`/credentials/${id}`),
  create: (data) => api.post('/credentials', data),
  update: (id, data) => api.put(`/credentials/${id}`, data),
  delete: (id) => api.delete(`/credentials/${id}`),
}

export const alertEventApi = {
  list: (params) => api.get('/alert-events', { params }),
  get: (id) => api.get(`/alert-events/${id}`),
  create: (data) => api.post('/alert-events', data),
  acknowledge: (id, data) => api.post(`/alert-events/${id}/acknowledge`, data),
  statistics: (params) => api.get('/alert-events/statistics', { params }),
}

export const alertSilenceApi = {
  list: (params) => api.get('/alert-silences', { params }),
  create: (data) => api.post('/alert-silences', data),
  revoke: (id) => api.post(`/alert-silences/${id}/revoke`),
}

export const userAuthApi = {
  changePassword: (data) => api.post('/user-auth/password', data),
  listTargets: (userId) => api.get(`/user-auth/${userId}/notification-targets`),
  addTarget: (userId, data) => api.post(`/user-auth/${userId}/notification-targets`, data),
  removeTarget: (id) => api.delete(`/user-auth/notification-targets/${id}`),
  setPrimary: (userId, id) => api.post(`/user-auth/${userId}/notification-targets/${id}/primary`),
}

export const zoneCoordinatorApi = {
  list: (params) => api.get('/zone-coordinators', { params }),
  get: (id) => api.get(`/zone-coordinators/${id}`),
  create: (data) => api.post('/zone-coordinators', data),
  update: (id, data) => api.put(`/zone-coordinators/${id}`, data),
  delete: (id) => api.delete(`/zone-coordinators/${id}`),
  heartbeat: (id) => api.post(`/zone-coordinators/${id}/heartbeat`),
}

export const sdConfigApi = {
  list: (params) => api.get('/sd-configs', { params }),
  get: (id) => api.get(`/sd-configs/${id}`),
  create: (data) => api.post('/sd-configs', data),
  update: (id, data) => api.put(`/sd-configs/${id}`, data),
  delete: (id) => api.delete(`/sd-configs/${id}`),
}

export const businessOverviewApi = {
  get: (systemId) => api.get(`/business-overview/${systemId}`),
}
