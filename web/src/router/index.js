import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/business',
    name: 'BusinessSystem',
    component: () => import('@/views/business/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/zones',
    name: 'Zones',
    component: () => import('@/views/zone/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/resource-types',
    name: 'ResourceTypes',
    component: () => import('@/views/resource/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/instances',
    name: 'Instances',
    component: () => import('@/views/instance/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/data-sinks',
    name: 'DataSinks',
    component: () => import('@/views/datasink/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/collection-tasks',
    name: 'CollectionTasks',
    component: () => import('@/views/task/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/collectors',
    name: 'Collectors',
    component: () => import('@/views/collector/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/user/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/notification-channels',
    name: 'NotificationChannels',
    component: () => import('@/views/notification/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/alert-groups',
    name: 'AlertGroups',
    component: () => import('@/views/alert-group/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/alert-strategies',
    name: 'AlertStrategies',
    component: () => import('@/views/alert-strategy/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/scheduled-tasks',
    name: 'ScheduledTasks',
    component: () => import('@/views/scheduled-task/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/rc-tasks',
    name: 'RCTasks',
    component: () => import('@/views/rc-task/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/metric-definitions',
    name: 'MetricDefinitions',
    component: () => import('@/views/metric/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/credentials',
    name: 'Credentials',
    component: () => import('@/views/credential/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/alert-events',
    name: 'AlertEvents',
    component: () => import('@/views/alert-event/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/zone-coordinators',
    name: 'ZoneCoordinators',
    component: () => import('@/views/coordinator/Index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/task-types',
    name: 'TaskTypes',
    component: () => import('@/views/task-type/Index.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('dcs_token')
  if (to.meta.requiresAuth !== false && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.path === '/login' && token) {
    next('/')
  } else {
    next()
  }
})

export default router
