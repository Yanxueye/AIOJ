<template>
  <div class="home-page page-container">
    <div class="hero-section">
      <div class="hero-content">
        <h1>Terminal<span class="highlight">OJ</span></h1>
        <p class="hero-desc">在线算法评测系统 · AI 辅助训练 · 提升编程能力</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" round @click="$router.push('/problems')">
            <el-icon><Document /></el-icon>开始刷题
          </el-button>
          <el-button size="large" round @click="$router.push('/ai')">
            <el-icon><MagicStick /></el-icon>AI 训练
          </el-button>
        </div>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-num">50+</span>
            <span class="stat-label">题目</span>
          </div>
          <div class="stat-item">
            <span class="stat-num">10</span>
            <span class="stat-label">算法分类</span>
          </div>
          <div class="stat-item">
            <span class="stat-num">4</span>
            <span class="stat-label">编程语言</span>
          </div>
        </div>
      </div>
    </div>

    <div class="home-grid">
      <div class="home-main">
        <AnnouncementBoard />

        <div class="card quick-entry">
          <div class="card-title">快捷入口</div>
          <div class="entry-grid">
            <div class="entry-card" @click="$router.push('/problems')">
              <el-icon :size="32" color="#409eff"><Document /></el-icon>
              <span>题目列表</span>
            </div>
            <div class="entry-card" @click="$router.push('/study-plans')">
              <el-icon :size="32" color="#8b5cf6"><Collection /></el-icon>
              <span>学习计划</span>
            </div>
            <div class="entry-card" @click="$router.push('/status')">
              <el-icon :size="32" color="#67c23a"><DataAnalysis /></el-icon>
              <span>评测状态</span>
            </div>
            <div class="entry-card" @click="$router.push('/profile')">
              <el-icon :size="32" color="#e6a23c"><User /></el-icon>
              <span>个人中心</span>
            </div>
            <div class="entry-card" @click="$router.push('/ai')">
              <el-icon :size="32" color="#ec4899"><MagicStick /></el-icon>
              <span>AI 训练</span>
            </div>
          </div>
        </div>
      </div>

      <div class="home-sidebar">
        <div class="card" v-if="userStore.isLoggedIn">
          <div class="card-title">我的状态</div>
          <div class="user-brief">
            <el-avatar :size="48" style="background: var(--accent-blue)">
              {{ userStore.username.charAt(0).toUpperCase() }}
            </el-avatar>
            <div>
              <div class="user-name">{{ userStore.username }}</div>
              <div class="user-rating">Rating: {{ userStore.userInfo?.rating || '--' }}</div>
            </div>
          </div>
          <el-divider />
          <div class="brief-stats">
            <div><span>已解决</span><strong>{{ userStore.userInfo?.solvedCount || 0 }}</strong></div>
            <div><span>总提交</span><strong>{{ userStore.userInfo?.totalSubmissions || 0 }}</strong></div>
            <div><span>通过率</span><strong>{{ userStore.userInfo?.acceptRate || 0 }}%</strong></div>
          </div>
        </div>
        <div class="card" v-else>
          <div class="card-title">欢迎</div>
          <p style="font-size: 14px; color: var(--text-secondary); margin-bottom: 16px">
            登录后可查看个人做题统计和提交记录
          </p>
          <el-button type="primary" style="width: 100%" @click="$router.push('/login')">立即登录</el-button>
        </div>

        <div class="card">
          <div class="card-title">每日一题</div>
          <div v-if="dailyChallenge" class="daily-card" @click="$router.push(`/problem/${dailyChallenge.problemId}`)">
            <div class="daily-date">{{ dailyChallenge.date }}</div>
            <div class="daily-title">#{{ dailyChallenge.problemId }} {{ dailyChallenge.title }}</div>
            <el-tag :type="diffTagType(dailyChallenge.difficulty)" size="small">{{ dailyChallenge.difficulty }}</el-tag>
          </div>
        </div>

        <div class="card" v-if="userStore.isLoggedIn">
          <div class="card-title">学习打卡</div>
          <div v-if="checkins.length" class="checkin-list">
            <div v-for="item in checkins.slice(0, 5)" :key="item.id" class="checkin-item">
              <span>{{ item.date }}</span>
              <strong>{{ item.count }} 次完成</strong>
            </div>
          </div>
          <el-empty v-else description="今天还没有打卡" :image-size="70" />
        </div>

        <div class="card">
          <div class="card-title">热门题目</div>
          <div class="hot-problems">
            <div
              v-for="p in hotProblems"
              :key="p.id"
              class="hot-item"
              @click="$router.push(`/problem/${p.id}`)"
            >
              <span class="hot-id">#{{ p.id }}</span>
              <span class="hot-name">{{ p.title }}</span>
              <el-tag :type="diffTagType(p.difficulty)" size="small">{{ p.difficulty }}</el-tag>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { problemApi } from '@/api/problem'
import { studyPlanApi } from '@/api/study_plan'
import AnnouncementBoard from '@/components/AnnouncementBoard.vue'

const userStore = useUserStore()
const hotProblems = ref([])
const dailyChallenge = ref(null)
const checkins = ref([])

onMounted(async () => {
  const [problemRes, dailyRes, checkinRes] = await Promise.all([
    problemApi.getList({ page: 1, pageSize: 6 }),
    studyPlanApi.getDailyChallenge(),
    userStore.isLoggedIn ? studyPlanApi.getCheckins() : Promise.resolve({ data: { items: [] } })
  ])
  hotProblems.value = problemRes.data.list
  dailyChallenge.value = dailyRes.data
  checkins.value = checkinRes.data.items || []
})

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
}
</script>

<style scoped>
.hero-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: var(--radius-lg);
  padding: 48px 40px;
  margin-bottom: 28px;
  color: #fff;
}
.hero-content h1 {
  font-size: 40px;
  font-weight: 800;
  margin-bottom: 12px;
}
.highlight {
  color: #ffd700;
}
.hero-desc {
  font-size: 16px;
  opacity: 0.9;
  margin-bottom: 24px;
}
.hero-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 32px;
}
.hero-stats {
  display: flex;
  gap: 48px;
}
.stat-item {
  display: flex;
  flex-direction: column;
}
.stat-num {
  font-size: 28px;
  font-weight: 700;
}
.stat-label {
  font-size: 13px;
  opacity: 0.8;
}

.home-grid {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 24px;
}
.home-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.home-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.entry-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.entry-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 24px 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
  font-weight: 500;
}
.entry-card:hover {
  border-color: var(--accent-blue);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.user-brief {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user-name {
  font-size: 16px;
  font-weight: 600;
}
.user-rating {
  font-size: 13px;
  color: var(--text-muted);
}
.brief-stats {
  display: flex;
  justify-content: space-around;
  text-align: center;
}
.brief-stats div {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.brief-stats span {
  font-size: 12px;
  color: var(--text-muted);
}
.brief-stats strong {
  font-size: 18px;
  color: var(--accent-blue);
}

.hot-problems {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.daily-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
}
.daily-date {
  font-size: 12px;
  color: var(--text-muted);
}
.daily-title {
  font-size: 15px;
  font-weight: 700;
}
.checkin-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.checkin-item {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  font-size: 13px;
}
.hot-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.2s;
  font-size: 13px;
}
.hot-item:hover {
  background: #f5f7fa;
}
.hot-id {
  color: var(--text-muted);
  font-family: monospace;
  min-width: 50px;
}
.hot-name {
  flex: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 960px) {
  .home-grid {
    grid-template-columns: 1fr;
  }
  .entry-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
