import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页 - TerminalOJ' }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录 - TerminalOJ', guest: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册 - TerminalOJ', guest: true }
  },
  {
    path: '/problems',
    name: 'problems',
    component: () => import('@/views/ProblemList.vue'),
    meta: { title: '题目列表 - TerminalOJ' }
  },
  {
    path: '/problem/:id',
    name: 'problem-detail',
    component: () => import('@/views/ProblemDetail.vue'),
    meta: { title: '题目详情 - TerminalOJ', auth: true }
  },
  {
    path: '/status',
    name: 'status',
    component: () => import('@/views/SubmissionStatus.vue'),
    meta: { title: '评测状态 - TerminalOJ', auth: true }
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/Profile.vue'),
    meta: { title: '个人中心 - TerminalOJ', auth: true }
  },
  {
    path: '/ai',
    name: 'ai-training',
    component: () => import('@/views/AITraining.vue'),
    meta: { title: 'AI 训练 - TerminalOJ', auth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title || 'TerminalOJ'
  const userStore = useUserStore()

  if (to.meta.auth && !userStore.isLoggedIn) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if (to.meta.guest && userStore.isLoggedIn) {
    next({ name: 'home' })
  } else {
    next()
  }
})

export default router
