<template>
  <header class="navbar">
    <div class="navbar-inner">
      <div class="navbar-left">
        <router-link to="/" class="logo">
          <span class="logo-icon">OJ</span>
          <span class="logo-text">TerminalOJ</span>
        </router-link>
        <nav class="nav-links">
          <router-link to="/" :class="{ active: route.name === 'home' }">
            <el-icon><HomeFilled /></el-icon>首页
          </router-link>
          <router-link to="/problems" :class="{ active: route.name === 'problems' }">
            <el-icon><Document /></el-icon>题库
          </router-link>
          <router-link to="/study-plans" :class="{ active: route.name === 'study-plans' || route.name === 'study-plan-detail' }">
            <el-icon><Collection /></el-icon>学习计划
          </router-link>
          <router-link to="/status" :class="{ active: route.name === 'status' }">
            <el-icon><DataAnalysis /></el-icon>评测
          </router-link>
          <router-link to="/ai" :class="{ active: route.name === 'ai-training' }">
            <el-icon><MagicStick /></el-icon>AI 训练
          </router-link>
        </nav>
      </div>
      <div class="navbar-right">
        <template v-if="userStore.isLoggedIn">
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :src="userStore.userInfo?.avatar || undefined">
                {{ userStore.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ userStore.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item command="my-solutions">
                  <el-icon><Document /></el-icon>我的题解
                </el-dropdown-item>
                <el-dropdown-item v-if="userStore.canManageProblems" command="admin-problem">
                  <el-icon><EditPen /></el-icon>题目管理
                </el-dropdown-item>
                <el-dropdown-item v-if="userStore.isAdmin" command="admin-users">
                  <el-icon><UserFilled /></el-icon>用户角色
                </el-dropdown-item>
                <el-dropdown-item v-if="userStore.isAdmin" command="admin-audit">
                  <el-icon><Document /></el-icon>审计日志
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <router-link to="/login">
            <el-button type="primary" round size="small">登录</el-button>
          </router-link>
          <router-link to="/register" style="margin-left: 8px">
            <el-button round size="small">注册</el-button>
          </router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

function handleCommand(cmd) {
  if (cmd === 'profile') {
    router.push('/profile')
  } else if (cmd === 'my-solutions') {
    router.push('/my/solutions')
  } else if (cmd === 'admin-problem') {
    router.push('/admin/problems/new')
  } else if (cmd === 'admin-users') {
    router.push('/admin/users')
  } else if (cmd === 'admin-audit') {
    router.push('/admin/audit-logs')
  } else if (cmd === 'logout') {
    userStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: #fff;
  border-bottom: 1px solid var(--border-color);
  z-index: 1000;
  box-shadow: var(--shadow-sm);
}
.navbar-inner {
  max-width: 1400px;
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}
.navbar-left {
  display: flex;
  align-items: center;
  gap: 32px;
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 700;
  color: var(--accent-blue);
}
.logo-icon {
  font-size: 18px;
  font-weight: 800;
}
.nav-links {
  display: flex;
  gap: 4px;
}
.nav-links a {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all 0.2s;
}
.nav-links a:hover {
  background: #f0f5ff;
  color: var(--accent-blue);
}
.nav-links a.active {
  background: #ecf5ff;
  color: var(--accent-blue);
}
.navbar-right {
  display: flex;
  align-items: center;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  transition: background 0.2s;
}
.user-info:hover {
  background: #f5f7fa;
}
.username {
  font-size: 14px;
  font-weight: 500;
}
</style>
