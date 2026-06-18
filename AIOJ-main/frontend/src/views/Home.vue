<template>
  <div class="home-page page-container mesh-bg">
    <!-- Character Decoration -->
    <div class="character-decoration" />

    <!-- Hero Section -->
    <div class="hero-section">
      <div class="hero-bg-pattern" />
      <div class="hero-glass-orb hero-orb-1" />
      <div class="hero-glass-orb hero-orb-2" />
      <div class="hero-content">
        <div class="hero-badge">
          <el-icon><MagicStick /></el-icon>
          AI 驱动的算法训练平台
        </div>
        <h1 class="hero-title">
          代码即诗意，<span class="hero-highlight">算法即远方</span>
        </h1>
        <p class="hero-desc">
          在线评测 · AI 智能辅助 · 知识图谱 · 个性化路径
        </p>
        <div class="hero-actions">
          <el-button type="primary" size="large" round class="hero-btn-primary" @click="$router.push('/problems')">
            <el-icon><Document /></el-icon>开始刷题
          </el-button>
          <el-button size="large" round class="hero-btn-secondary" @click="$router.push('/ai')">
            <el-icon><MagicStick /></el-icon>AI 对话
          </el-button>
        </div>
        <div class="hero-stats">
          <div class="stat-pill">
            <span class="stat-num">{{ problemCount }}</span>
            <span class="stat-label">精选题目</span>
          </div>
          <div class="stat-divider" />
          <div class="stat-pill">
            <span class="stat-num">10</span>
            <span class="stat-label">算法分类</span>
          </div>
          <div class="stat-divider" />
          <div class="stat-pill">
            <span class="stat-num">3</span>
            <span class="stat-label">编程语言</span>
          </div>
        </div>
      </div>
    </div>

    <div class="home-grid">
      <div class="home-main">
        <AnnouncementBoard />

        <!-- Quick Entry -->
        <div class="card quick-entry">
          <div class="card-title">快捷入口</div>
          <div class="entry-grid">
            <div class="entry-card" @click="$router.push('/problems')">
              <div class="entry-icon" style="background: var(--accent-primary-bg); color: var(--accent-primary)">
                <el-icon :size="24"><Document /></el-icon>
              </div>
              <span class="entry-label">题目列表</span>
            </div>
            <div class="entry-card" @click="$router.push('/study-plans')">
              <div class="entry-icon" style="background: var(--accent-purple-bg); color: var(--accent-purple)">
                <el-icon :size="24"><Collection /></el-icon>
              </div>
              <span class="entry-label">学习计划</span>
            </div>
            <div class="entry-card" @click="$router.push('/status')">
              <div class="entry-icon" style="background: var(--accent-green-bg); color: var(--accent-green)">
                <el-icon :size="24"><DataAnalysis /></el-icon>
              </div>
              <span class="entry-label">评测状态</span>
            </div>
            <div class="entry-card" @click="$router.push('/profile')">
              <div class="entry-icon" style="background: var(--accent-orange-bg); color: var(--accent-orange)">
                <el-icon :size="24"><User /></el-icon>
              </div>
              <span class="entry-label">个人中心</span>
            </div>
            <div class="entry-card" @click="$router.push('/knowledge')">
              <div class="entry-icon" style="background: var(--accent-blue-bg); color: var(--accent-blue)">
                <el-icon :size="24"><Share /></el-icon>
              </div>
              <span class="entry-label">知识图谱</span>
            </div>
            <div class="entry-card" @click="$router.push('/ai')">
              <div class="entry-icon" style="background: var(--accent-gold-bg); color: var(--accent-gold)">
                <el-icon :size="24"><MagicStick /></el-icon>
              </div>
              <span class="entry-label">AI 对话</span>
            </div>
          </div>
        </div>
      </div>

      <div class="home-sidebar">
        <!-- User Status -->
        <div class="card user-status-card" v-if="userStore.isLoggedIn">
          <div class="user-status-header">
            <el-avatar :size="44" class="user-avatar-lg">
              {{ userStore.username.charAt(0).toUpperCase() }}
            </el-avatar>
            <div class="user-status-info">
              <div class="user-status-name">{{ userStore.username }}</div>
              <div class="user-status-rating">
                <el-icon><TrendCharts /></el-icon>
                Rating: <strong>{{ userStore.userInfo?.rating || '--' }}</strong>
              </div>
            </div>
          </div>
          <div class="user-status-stats">
            <div class="mini-stat">
              <span class="mini-stat-val">{{ userStore.userInfo?.solvedCount || 0 }}</span>
              <span class="mini-stat-label">已解决</span>
            </div>
            <div class="mini-stat">
              <span class="mini-stat-val">{{ userStore.userInfo?.totalSubmissions || 0 }}</span>
              <span class="mini-stat-label">总提交</span>
            </div>
            <div class="mini-stat">
              <span class="mini-stat-val">{{ userStore.userInfo?.acceptRate || 0 }}%</span>
              <span class="mini-stat-label">通过率</span>
            </div>
          </div>
        </div>
        <div class="card welcome-card" v-else>
          <div class="welcome-icon">
            <svg width="40" height="40" viewBox="0 0 26 26" fill="none">
              <rect x="2" y="2" width="22" height="22" rx="6" fill="url(#wcGrad)" />
              <path d="M8 13l3 3 7-7" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
              <defs>
                <linearGradient id="wcGrad" x1="2" y1="2" x2="24" y2="24">
                  <stop offset="0%" stop-color="#52c41a"/>
                  <stop offset="100%" stop-color="#389e0d"/>
                </linearGradient>
              </defs>
            </svg>
          </div>
          <h3 class="welcome-title">欢迎来到 TerminalOJ</h3>
          <p class="welcome-desc">登录后可查看个人做题统计、提交记录和 AI 推荐</p>
          <el-button type="primary" style="width: 100%" round @click="$router.push('/login')">立即登录</el-button>
        </div>

        <!-- Daily Recommendations -->
        <div class="card">
          <div class="card-title">
            <el-icon><Sunny /></el-icon> 每日推荐
          </div>
          <div v-if="recommendations.length" class="recommend-list">
            <div
              v-for="(p, idx) in recommendations"
              :key="p.id"
              class="recommend-item"
              @click="$router.push(`/problem/${p.id}`)"
            >
              <span class="recommend-rank" :class="{ 'rank-top': idx < 3 }">{{ idx + 1 }}</span>
              <div class="recommend-info">
                <span class="recommend-title">{{ p.title }}</span>
              </div>
              <div class="recommend-meta">
                <el-tag :type="diffTagType(p.difficulty)" size="small" effect="plain">{{ p.difficulty }}</el-tag>
                <span v-if="p.rating" class="recommend-rating">{{ p.rating }}</span>
              </div>
            </div>
          </div>
          <div v-else-if="dailyChallenge" class="daily-card" @click="$router.push(`/problem/${dailyChallenge.problemId}`)">
            <div class="daily-date">{{ dailyChallenge.date }}</div>
            <div class="daily-title">#{{ dailyChallenge.problemId }} {{ dailyChallenge.title }}</div>
            <el-tag :type="diffTagType(dailyChallenge.difficulty)" size="small">{{ dailyChallenge.difficulty }}</el-tag>
          </div>
          <el-empty v-else description="暂无推荐" :image-size="60" />
        </div>

        <!-- Hot Problems -->
        <div class="card">
          <div class="card-title">
            <el-icon><TrendCharts /></el-icon> 热门题目
          </div>
          <div class="hot-problems">
            <div
              v-for="(p, idx) in hotProblems"
              :key="p.id"
              class="hot-item"
              @click="$router.push(`/problem/${p.id}`)"
            >
              <span class="hot-rank" :class="{ 'rank-top': idx < 3 }">{{ idx + 1 }}</span>
              <span class="hot-name">{{ p.title }}</span>
              <el-tag :type="diffTagType(p.difficulty)" size="small" effect="plain">{{ p.difficulty }}</el-tag>
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
import http from '@/api/index'
import AnnouncementBoard from '@/components/AnnouncementBoard.vue'

const userStore = useUserStore()
const hotProblems = ref([])
const dailyChallenge = ref(null)
const checkins = ref([])
const recommendations = ref([])
const problemCount = ref(5)

onMounted(async () => {
  const [problemRes, dailyRes, checkinRes, recRes] = await Promise.all([
    problemApi.getList({ page: 1, pageSize: 6 }),
    studyPlanApi.getDailyChallenge(),
    userStore.isLoggedIn ? studyPlanApi.getCheckins() : Promise.resolve({ data: { items: [] } }),
    http.get('/recommendations/daily').catch(() => ({ data: { items: [] } }))
  ])
  hotProblems.value = problemRes.data.list
  problemCount.value = problemRes.data.total || problemRes.data.list?.length || 5
  dailyChallenge.value = dailyRes.data
  checkins.value = checkinRes.data.items || []
  recommendations.value = recRes.data?.items || []
})

function diffTagType(d) {
  return d === '简单' ? 'success' : d === '中等' ? 'warning' : 'danger'
}
</script>

<style scoped>
.hero-section {
  background: linear-gradient(135deg, #2d6a1e 0%, #3d8c28 40%, #4a9d32 70%, #5cb840 100%);
  border-radius: 20px;
  padding: 48px 48px;
  margin-bottom: 28px;
  color: #fff;
  position: relative;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(45, 106, 30, 0.3), inset 0 1px 0 rgba(255,255,255,0.1);
}

/* Subtle pattern overlay */
.hero-section::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(circle at 80% 20%, rgba(255,255,255,0.12) 0%, transparent 50%),
    radial-gradient(circle at 20% 80%, rgba(255,255,255,0.06) 0%, transparent 40%),
    radial-gradient(circle at 60% 40%, rgba(232,168,56,0.06) 0%, transparent 30%);
  pointer-events: none;
}

/* Code-like decorative pattern */
.hero-section::after {
  content: '{ }';
  position: absolute;
  right: 60px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 120px;
  font-weight: 800;
  color: rgba(255,255,255,0.06);
  font-family: 'JetBrains Mono', monospace;
  pointer-events: none;
}

.hero-bg-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px),
    linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px);
  background-size: 20px 20px;
  pointer-events: none;
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 600px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: rgba(255,255,255,0.15);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 20px;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255,255,255,0.2);
}

.hero-title {
  font-family: 'Noto Sans SC', sans-serif;
  font-size: 40px;
  font-weight: 800;
  line-height: 1.2;
  margin-bottom: 16px;
  letter-spacing: -0.02em;
}

.hero-highlight {
  color: #ffd666;
  text-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.hero-desc {
  font-size: 15px;
  opacity: 0.9;
  margin-bottom: 28px;
  line-height: 1.7;
  max-width: 480px;
}

.hero-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 36px;
}

.hero-btn-primary {
  font-weight: 700;
  padding: 10px 28px;
  font-size: 15px;
  background: rgba(232, 168, 56, 0.9) !important;
  border: 1.5px solid rgba(255, 200, 80, 0.6) !important;
  color: #fff !important;
  text-shadow: 0 1px 2px rgba(0,0,0,0.15);
  box-shadow: 0 4px 16px rgba(232, 168, 56, 0.3);
}
.hero-btn-primary:hover {
  background: rgba(232, 168, 56, 1) !important;
  box-shadow: 0 6px 24px rgba(232, 168, 56, 0.4);
  transform: translateY(-1px);
}

.hero-btn-secondary {
  font-weight: 600;
  padding: 10px 24px;
  font-size: 15px;
  background: rgba(255,255,255,0.12) !important;
  border: 1.5px solid rgba(255,255,255,0.25) !important;
  color: #fff !important;
  backdrop-filter: blur(8px);
}
.hero-btn-secondary:hover {
  background: rgba(255,255,255,0.2) !important;
}

.hero-stats {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-pill {
  display: flex;
  flex-direction: column;
}

.stat-num {
  font-size: 28px;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.02em;
}

.stat-label {
  font-size: 12.5px;
  opacity: 0.75;
  margin-top: 2px;
}

.stat-divider {
  width: 1px;
  height: 36px;
  background: rgba(255,255,255,0.2);
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
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.entry-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 18px 12px;
  border-radius: 12px;
  border: 1px solid rgba(0,0,0,0.06);
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
}

.entry-card:hover {
  border-color: var(--accent-primary);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}

.entry-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.entry-label {
  font-size: 13px;
  font-weight: 500;
}

.entry-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.user-status-card {
  background: var(--gradient-card);
}

.user-status-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.user-avatar-lg {
  background: var(--gradient-amber) !important;
  color: #fff !important;
  font-size: 18px !important;
  font-weight: 700;
}

.user-status-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.user-status-rating {
  font-size: 13px;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 2px;
}
.user-status-rating strong {
  color: var(--accent-gold);
  font-weight: 700;
}

.user-status-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding-top: 14px;
  border-top: 1px solid var(--border-light);
}

.mini-stat {
  text-align: center;
}

.mini-stat-val {
  display: block;
  font-size: 20px;
  font-weight: 800;
  color: var(--accent-primary);
  line-height: 1.2;
}

.mini-stat-label {
  font-size: 11.5px;
  color: var(--text-muted);
  margin-top: 2px;
}

.welcome-card {
  text-align: center;
  background: var(--gradient-card);
}

.welcome-icon {
  margin-bottom: 14px;
}

.welcome-title {
  font-size: 17px;
  font-weight: 700;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.welcome-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 18px;
  line-height: 1.6;
}

.recommend-list, .hot-problems {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.recommend-item, .hot-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.recommend-item:hover, .hot-item:hover {
  background: var(--bg-hover);
}

.recommend-rank, .hot-rank {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  min-width: 20px;
  text-align: center;
}

.rank-top {
  color: var(--accent-gold);
}

.recommend-info {
  flex: 1;
  overflow: hidden;
}

.recommend-title, .hot-name {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.recommend-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.recommend-rating {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.hot-name {
  flex: 1;
}

.daily-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  padding: 8px 0;
}

.daily-date {
  font-size: 12px;
  color: var(--text-muted);
}

.daily-title {
  font-size: 15px;
  font-weight: 700;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

@media (max-width: 960px) {
  .home-grid {
    grid-template-columns: 1fr;
  }
  .entry-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .hero-title {
    font-size: 32px;
  }
  .hero-section {
    padding: 36px 28px;
  }
}
</style>
