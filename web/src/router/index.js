import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/Index.vue'),
  },
  {
    path: '/business',
    name: 'BusinessSystem',
    component: () => import('@/views/business/Index.vue'),
  },
  {
    path: '/zones',
    name: 'Zones',
    component: () => import('@/views/zone/Index.vue'),
  },
  {
    path: '/resource-types',
    name: 'ResourceTypes',
    component: () => import('@/views/resource/Index.vue'),
  },
  {
    path: '/instances',
    name: 'Instances',
    component: () => import('@/views/instance/Index.vue'),
  },
  {
    path: '/data-sinks',
    name: 'DataSinks',
    component: () => import('@/views/datasink/Index.vue'),
  },
  {
    path: '/collection-tasks',
    name: 'CollectionTasks',
    component: () => import('@/views/task/Index.vue'),
  },
  {
    path: '/collectors',
    name: 'Collectors',
    component: () => import('@/views/collector/Index.vue'),
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/user/Index.vue'),
  },
  {
    path: '/notification-channels',
    name: 'NotificationChannels',
    component: () => import('@/views/notification/Index.vue'),
  },
  {
    path: '/alert-groups',
    name: 'AlertGroups',
    component: () => import('@/views/alert-group/Index.vue'),
  },
  {
    path: '/alert-strategies',
    name: 'AlertStrategies',
    component: () => import('@/views/alert-strategy/Index.vue'),
  },
  {
    path: '/scheduled-tasks',
    name: 'ScheduledTasks',
    component: () => import('@/views/scheduled-task/Index.vue'),
  },
  {
    path: '/rc-tasks',
    name: 'RCTasks',
    component: () => import('@/views/rc-task/Index.vue'),
  },
  {
    path: '/metric-definitions',
    name: 'MetricDefinitions',
    component: () => import('@/views/metric/Index.vue'),
  },
  {
    path: '/credentials',
    name: 'Credentials',
    component: () => import('@/views/credential/Index.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
