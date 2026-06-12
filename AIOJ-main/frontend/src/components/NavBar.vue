<template>
  <header class="navbar">
    <div class="navbar-inner">
      <div class="navbar-left">
        <router-link to="/" class="logo">
          <div class="logo-mark">
            <svg width="26" height="26" viewBox="0 0 26 26" fill="none">
              <rect x="2" y="2" width="22" height="22" rx="6" fill="url(#logoGrad)" />
              <path d="M8 13l3 3 7-7" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
              <defs>
                <linearGradient id="logoGrad" x1="2" y1="2" x2="24" y2="24">
                  <stop offset="0%" stop-color="#52c41a"/>
                  <stop offset="100%" stop-color="#389e0d"/>
                </linearGradient>
              </defs>
            </svg>
          </div>
          <span class="logo-text">Terminal<span class="logo-accent">OJ</span></span>
        </router-link>
        <nav class="nav-links">
          <router-link to="/" :class="{ active: route.name === 'home' }">
            <el-icon><HomeFilled /></el-icon><span>首页</span>
          </router-link>
          <router-link to="/problems" :class="{ active: route.name === 'problems' }">
            <el-icon><Document /></el-icon><span>题库</span>
          </router-link>
          <router-link to="/study-plans" :class="{ active: route.name === 'study-plans' || route.name === 'study-plan-detail' }">
            <el-icon><Collection /></el-icon><span>学习计划</span>
          </router-link>
          <router-link to="/knowledge" :class="{ active: route.name === 'knowledge-graph' }">
            <el-icon><Share /></el-icon><span>知识图谱</span>
          </router-link>
          <router-link to="/status" :class="{ active: route.name === 'status' }">
            <el-icon><DataAnalysis /></el-icon><span>评测</span>
          </router-link>
          <router-link to="/ai" :class="{ active: route.name === 'ai-training' }">
            <el-icon><MagicStick /></el-icon><span>AI 训练</span>
          </router-link>
        </nav>
      </div>
      <div class="navbar-right">
        <template v-if="userStore.isLoggedIn">
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="30" :src="userStore.userInfo?.avatar || undefined" class="user-avatar">
                {{ userStore.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ userStore.username }}</span>
              <el-icon class="dropdown-arrow"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item v-if="userStore.isAdmin" command="admin-problem">
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
            <el-button type="primary" round size="small" class="login-btn">登录</el-button>
          </router-link>
          <router-link to="/register">
            <el-button round size="small" class="register-btn">注册</el-button>
          </router-link>
        </template>
        <el-tooltip :content="theme === 'dark' ? '切换亮色模式' : '切换暗色模式'" placement="bottom">
          <button class="theme-toggle" @click="toggleTheme" :aria-label="theme === 'dark' ? '切换亮色模式' : '切换暗色模式'">
            <transition name="theme-icon" mode="out-in">
              <el-icon v-if="theme === 'dark'" key="sunny"><Sunny /></el-icon>
              <el-icon v-else key="moon"><Moon /></el-icon>
            </transition>
          </button>
        </el-tooltip>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useTheme } from '@/composables/useTheme'
import { Sunny, Moon } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { theme, toggleTheme } = useTheme()

function handleCommand(cmd) {
  if (cmd === 'profile') {
    router.push('/profile')
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
  background: var(--navbar-bg);
  backdrop-filter: blur(16px) saturate(1.4);
  -webkit-backdrop-filter: blur(16px) saturate(1.4);
  border-bottom: 1px solid var(--glass-border);
  z-index: 1000;
  transition: background var(--transition-normal), border-color var(--transition-normal);
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
  gap: 36px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  transition: opacity var(--transition-fast);
}
.logo:hover {
  opacity: 0.85;
}
.logo-mark {
  display: flex;
  align-items: center;
  justify-content: center;
}
.logo-text {
  font-family: var(--font-display);
  font-size: 19px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.03em;
}
.logo-accent {
  color: var(--glass-green);
}

.nav-links {
  display: flex;
  gap: 2px;
}

.nav-links a {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 14px;
  border-radius: var(--radius-sm);
  font-size: 13.5px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
  position: relative;
  white-space: nowrap;
}

.nav-links a:hover {
  background: var(--accent-primary-bg);
  color: var(--accent-primary);
}

.nav-links a.active {
  background: var(--accent-primary-bg);
  color: var(--accent-primary);
  font-weight: 600;
}

.nav-links a.active::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 2px;
  background: var(--accent-primary);
  border-radius: var(--radius-full);
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 10px 4px 4px;
  border-radius: var(--radius-full);
  transition: background var(--transition-fast);
}
.user-info:hover {
  background: var(--bg-hover);
}
.user-avatar {
  background: var(--gradient-amber) !important;
  color: #fff !important;
  font-size: 13px !important;
  font-weight: 700;
}
.username {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-primary);
}
.dropdown-arrow {
  font-size: 12px;
  color: var(--text-muted);
}

.login-btn {
  font-weight: 600;
  padding: 6px 18px;
}
.register-btn {
  font-weight: 500;
  padding: 6px 16px;
  border-color: var(--border-color);
}

.theme-toggle {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 16px;
  transition: all var(--transition-fast);
  margin-left: 4px;
}
.theme-toggle:hover {
  background: var(--accent-primary-bg);
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.theme-icon-enter-active,
.theme-icon-leave-active {
  transition: all 0.2s ease;
}
.theme-icon-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.6);
}
.theme-icon-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.6);
}

@media (max-width: 960px) {
  .nav-links a span {
    display: none;
  }
  .nav-links a {
    padding: 7px 10px;
  }
  .navbar-left {
    gap: 16px;
  }
}
</style>
