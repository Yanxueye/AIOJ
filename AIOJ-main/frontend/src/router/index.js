import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  { path: '/', name: 'home', component: () => import('@/views/Home.vue'), meta: { title: '首页 - TerminalOJ' } },
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录 - TerminalOJ', guest: true } },
  { path: '/register', name: 'register', component: () => import('@/views/Register.vue'), meta: { title: '注册 - TerminalOJ', guest: true } },
  { path: '/problems', name: 'problems', component: () => import('@/views/ProblemList.vue'), meta: { title: '题目列表 - TerminalOJ' } },
  { path: '/study-plans', name: 'study-plans', component: () => import('@/views/StudyPlanList.vue'), meta: { title: '学习计划 - TerminalOJ' } },
  { path: '/study-plans/:id', name: 'study-plan-detail', component: () => import('@/views/StudyPlanDetail.vue'), meta: { title: '学习计划详情 - TerminalOJ' } },
  { path: '/problem/:id', name: 'problem-detail', component: () => import('@/views/ProblemDetail.vue'), meta: { title: '题目详情 - TerminalOJ', auth: true } },
  { path: '/my/solutions/new', name: 'my-solution-new', component: () => import('@/views/MySolutionCreate.vue'), meta: { title: '新建题解 - TerminalOJ', auth: true } },
  { path: '/my/solutions/:id/edit', name: 'my-solution-edit', component: () => import('@/views/MySolutionEdit.vue'), meta: { title: '编辑题解 - TerminalOJ', auth: true } },
  { path: '/status', name: 'status', component: () => import('@/views/SubmissionStatus.vue'), meta: { title: '评测状态 - TerminalOJ', auth: true } },
  { path: '/profile', name: 'profile', component: () => import('@/views/Profile.vue'), meta: { title: '个人中心 - TerminalOJ', auth: true } },
  { path: '/knowledge', name: 'knowledge-graph', component: () => import('@/views/KnowledgeGraph.vue'), meta: { title: '知识图谱 - TerminalOJ' } },
  { path: '/ai', name: 'ai-training', component: () => import('@/views/AITraining.vue'), meta: { title: 'AI 训练 - TerminalOJ', auth: true } },
  { path: '/admin/problems/new', name: 'admin-problem-new', component: () => import('@/views/AdminProblemCreate.vue'), meta: { title: '新增题目 - TerminalOJ', auth: true, admin: true } },
  { path: '/admin/problems/:id/edit', name: 'admin-problem-edit', component: () => import('@/views/AdminProblemEdit.vue'), meta: { title: '编辑题目 - TerminalOJ', auth: true, admin: true } },
  { path: '/admin/users', name: 'admin-users', component: () => import('@/views/AdminUsers.vue'), meta: { title: '用户角色管理 - TerminalOJ', auth: true, admin: true } },
  { path: '/admin/audit-logs', name: 'admin-audit-logs', component: () => import('@/views/AdminAuditLogs.vue'), meta: { title: '审计日志 - TerminalOJ', auth: true, admin: true } }
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
  } else if (to.meta.admin && !userStore.isAdmin) {
    next({ name: 'home' })
  } else if (to.meta.guest && userStore.isLoggedIn) {
    next({ name: 'home' })
  } else {
    next()
  }
})

export default router
