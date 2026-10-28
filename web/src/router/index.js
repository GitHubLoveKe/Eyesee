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
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
