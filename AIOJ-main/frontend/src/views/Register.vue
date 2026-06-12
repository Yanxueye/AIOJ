<template>
  <div class="auth-page mesh-bg">
    <div class="auth-ambient" />
    <div class="character-decoration character-decoration--auth" />
    <div class="auth-card">
      <div class="auth-header">
        <div class="auth-logo">
          <svg width="36" height="36" viewBox="0 0 26 26" fill="none">
            <rect x="2" y="2" width="22" height="22" rx="6" fill="url(#regGrad)" />
            <path d="M8 13l3 3 7-7" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <defs>
              <linearGradient id="regGrad" x1="2" y1="2" x2="24" y2="24">
                <stop offset="0%" stop-color="#52c41a"/>
                <stop offset="100%" stop-color="#389e0d"/>
              </linearGradient>
            </defs>
          </svg>
        </div>
        <h2>创建账号</h2>
        <p>加入 TerminalOJ，开始你的算法之旅</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="handleRegister">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" prefix-icon="Message" size="large" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleRegister"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            round
            @click="handleRegister"
          >
            注 册
          </el-button>
        </el-form-item>
      </el-form>
      <div class="auth-footer">
        已有账号？<router-link to="/login">去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref(null)
const loading = ref(false)
const form = reactive({ username: '', email: '', password: '', confirmPassword: '' })

const validateConfirm = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度 3-20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' }
  ]
}

async function handleRegister() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.register(form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (err) {
    ElMessage.error(err.message || '注册失败')
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
