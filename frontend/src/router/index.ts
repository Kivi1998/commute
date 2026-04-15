import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/pages/Home.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'companies',
        name: 'companies',
        component: () => import('@/pages/Companies.vue'),
        meta: { title: '公司管理' },
      },
      {
        path: 'commute',
        name: 'commute',
        component: () => import('@/pages/Commute.vue'),
        meta: { title: '通勤对比' },
      },
      {
        path: 'history',
        name: 'history',
        component: () => import('@/pages/History.vue'),
        meta: { title: '历史记录' },
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/pages/Settings.vue'),
        meta: { title: '设置' },
      },
    ],
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.afterEach((to) => {
  const title = (to.meta.title as string) || ''
  document.title = title ? `${title} - 通勤查询` : '通勤查询'
})
