<template>
  <div class="profile-page page-container">
    <div class="page-header">
      <h2>个人中心</h2>
    </div>

    <div v-loading="loading" class="profile-layout">
      <div class="profile-sidebar">
        <div class="card user-card">
          <div class="user-avatar">
            <el-avatar :size="80" style="background: linear-gradient(135deg, #667eea, #764ba2); font-size: 32px">
              {{ profile?.username?.charAt(0).toUpperCase() }}
            </el-avatar>
          </div>
          <h3 class="user-name">{{ profile?.username }}</h3>
          <p class="user-bio">{{ profile?.bio || '这个人很懒，什么也没写' }}</p>
          <el-divider />
          <div class="user-meta">
            <div class="meta-item">
              <el-icon><Message /></el-icon>
              <span>{{ profile?.email }}</span>
            </div>
            <div class="meta-item">
              <el-icon><Calendar /></el-icon>
              <span>注册于 {{ profile?.registeredAt }}</span>
            </div>
          </div>
          <el-divider />
          <el-button type="primary" plain style="width: 100%" @click="editDialogVisible = true">
            <el-icon><Edit /></el-icon>编辑资料
          </el-button>
          <el-button style="width: 100%; margin-top: 8px; margin-left: 0" @click="$router.push('/ai')">
            <el-icon><MagicStick /></el-icon>AI 训练
          </el-button>
          <el-button style="width: 100%; margin-top: 8px; margin-left: 0" @click="$router.push('/my/solutions')">
            <el-icon><Document /></el-icon>我的题解
          </el-button>
        </div>
      </div>

      <div class="profile-main">
        <div class="stats-overview">
          <div class="stat-card card">
            <div class="stat-icon" style="background: #ecf5ff; color: #409eff">
              <el-icon :size="24"><Trophy /></el-icon>
            </div>
            <div>
              <div class="stat-value">{{ profile?.rating || 0 }}</div>
              <div class="stat-label">Rating</div>
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-icon" style="background: #f0f9eb; color: #67c23a">
              <el-icon :size="24"><CircleCheckFilled /></el-icon>
            </div>
            <div>
              <div class="stat-value">{{ profile?.solvedCount || 0 }}</div>
              <div class="stat-label">已解决</div>
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-icon" style="background: #fef0f0; color: #f56c6c">
              <el-icon :size="24"><Upload /></el-icon>
            </div>
            <div>
              <div class="stat-value">{{ profile?.totalSubmissions || 0 }}</div>
              <div class="stat-label">总提交</div>
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-icon" style="background: #fdf6ec; color: #e6a23c">
              <el-icon :size="24"><TrendCharts /></el-icon>
            </div>
            <div>
              <div class="stat-value">{{ profile?.acceptRate || 0 }}%</div>
              <div class="stat-label">通过率</div>
            </div>
          </div>
        </div>

        <StatsCharts
          v-if="profile"
          :difficulty-data="profile.solvedByDifficulty || {}"
          :algorithm-data="profile.solvedByAlgorithm || {}"
        />

        <div class="card section-card">
          <div class="section-title">收藏题目</div>
          <div v-if="profile?.favorites?.length" class="favorite-list">
            <router-link
              v-for="item in profile.favorites"
              :key="item.problemId"
              :to="`/problem/${item.problemId}`"
              class="favorite-item"
            >
              <div class="favorite-title">#{{ item.problemId }} {{ item.title }}</div>
              <div class="favorite-meta">
                <span>{{ item.difficulty }}</span>
                <span>通过率 {{ item.acceptRate }}%</span>
                <span>收藏于 {{ item.favoritedAt }}</span>
              </div>
            </router-link>
          </div>
          <el-empty v-else description="还没有收藏题目" :image-size="80" />
        </div>

        <div class="card section-card">
          <div class="section-title">最近提交</div>
          <div v-if="profile?.recentSubmissions?.length" class="timeline-list">
            <div v-for="item in profile.recentSubmissions" :key="item.submissionId" class="timeline-item">
              <router-link :to="`/problem/${item.problemId}`" class="timeline-problem">
                #{{ item.problemId }} {{ item.problemTitle }}
              </router-link>
              <div class="timeline-meta">
                <span :class="statusClass(item.status)">{{ item.status }}</span>
                <span>{{ item.language }}</span>
                <span>{{ item.createdAt }}</span>
              </div>
            </div>
          </div>
          <el-empty v-else description="还没有提交记录" :image-size="80" />
        </div>
      </div>
    </div>

    <el-dialog v-model="editDialogVisible" title="编辑个人资料" width="480px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item label="个人简介">
          <el-input v-model="editForm.bio" type="textarea" :rows="3" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import StatsCharts from '@/components/StatsCharts.vue'

const userStore = useUserStore()

const loading = ref(true)
const profile = ref(null)
const editDialogVisible = ref(false)
const saving = ref(false)
const editForm = reactive({ username: '', email: '', bio: '' })

onMounted(async () => {
  try {
    profile.value = await userStore.fetchProfile()
    editForm.username = profile.value.username
    editForm.email = profile.value.email
    editForm.bio = profile.value.bio || ''
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  saving.value = true
  try {
    const updated = await userStore.updateProfile({
      email: editForm.email,
      bio: editForm.bio
    })
    profile.value = { ...profile.value, ...updated }
    editDialogVisible.value = false
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

function statusClass(status) {
  const map = {
    Accepted: 'status-accepted',
    'Wrong Answer': 'status-wrong',
    'Compile Error': 'status-ce',
    'Time Limit Exceeded': 'status-tle',
    'Memory Limit Exceeded': 'status-mle',
    'Output Limit Exceeded': 'status-ole',
    'System Error': 'status-system'
  }
  return map[status] || 'status-pending'
}
</script>

<style scoped>
.page-header {
  margin-bottom: 20px;
}
.page-header h2 {
  font-size: 24px;
  font-weight: 700;
}
.profile-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 24px;
}
.user-card {
  text-align: center;
}
.user-avatar {
  margin-bottom: 16px;
}
.user-name {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 8px;
}
.user-bio {
  font-size: 13px;
  color: var(--text-muted);
}
.user-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.profile-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.section-card {
  padding: 20px;
}
.section-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 14px;
}
.stats-overview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: var(--text-muted);
}
.favorite-list,
.timeline-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.favorite-item,
.timeline-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 14px;
  background: #fafbfc;
}
.favorite-item {
  display: block;
}
.favorite-title,
.timeline-problem {
  font-weight: 600;
  color: var(--text-primary);
}
.favorite-meta,
.timeline-meta {
  margin-top: 6px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--text-secondary);
}

@media (max-width: 960px) {
  .profile-layout {
    grid-template-columns: 1fr;
  }
  .stats-overview {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
