import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    redirect: '/projects',
  },
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('../views/ProjectList.vue'),
  },
  {
    path: '/projects/:id',
    name: 'ProjectDashboard',
    component: () => import('../views/ProjectDashboard.vue'),
  },
  {
    path: '/projects/:projectId/branches/:branchId',
    name: 'BranchView',
    component: () => import('../views/BranchView.vue'),
  },
  {
    path: '/merge-requests',
    name: 'MergeRequestList',
    component: () => import('../views/MergeRequestList.vue'),
  },
  {
    path: '/merge-requests/:id',
    name: 'MergeRequestDetail',
    component: () => import('../views/MergeRequestDetail.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router