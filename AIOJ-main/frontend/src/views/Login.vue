<template>
  <div class="auth-page mesh-bg">
    <div class="auth-ambient" />
    <div class="character-decoration character-decoration--auth" />
    <div class="auth-card">
      <div class="auth-header">
        <div class="auth-logo">
          <svg width="36" height="36" viewBox="0 0 26 26" fill="none">
            <rect x="2" y="2" width="22" height="22" rx="6" fill="url(#authGrad)" />
            <path d="M8 13l3 3 7-7" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <defs>
              <linearGradient id="authGrad" x1="2" y1="2" x2="24" y2="24">
                <stop offset="0%" stop-color="#52c41a"/>
                <stop offset="100%" stop-color="#389e0d"/>
              </linearGradient>
            </defs>
          </svg>
        </div>
        <h2>欢迎回来</h2>
        <p>登录你的 TerminalOJ 账号</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="handleLogin">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            round
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
      <div class="auth-footer">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref(null)
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' }
  ]
}

async function handleLogin() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login(form)
    ElMessage.success('登录成功')
    router.push(route.query.redirect || '/')
  } catch (err) {
    ElMessage.error(err.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  position: relative;
  overflow: hidden;
}

.auth-ambient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 70% 50% at 70% 30%, rgba(82,196,26,0.08) 0%, transparent 60%),
    radial-gradient(ellipse 60% 70% at 25% 70%, rgba(232,168,56,0.06) 0%, transparent 50%),
    radial-gradient(ellipse 40% 40% at 50% 50%, rgba(138,92,196,0.04) 0%, transparent 50%);
  pointer-events: none;
}

.auth-card {
  width: 420px;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(28px) saturate(1.6);
  -webkit-backdrop-filter: blur(28px) saturate(1.6);
  border-radius: var(--radius-xl);
  padding: 44px 40px;
  box-shadow: var(--shadow-xl), 0 0 0 1px var(--glass-border);
  position: relative;
  z-index: 1;
  border: 1px solid rgba(255,255,255,0.5);
}

.auth-header {
  text-align: center;
  margin-bottom: 32px;
}

.auth-logo {
  margin-bottom: 18px;
  display: flex;
  justify-content: center;
}

.auth-header h2 {
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}

.auth-header p {
  color: var(--text-muted);
  font-size: 14px;
}

.auth-footer {
  text-align: center;
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 20px;
}

.auth-footer a {
  color: var(--accent-primary);
  font-weight: 600;
}

.auth-footer a:hover {
  text-decoration: underline;
}
</style>
