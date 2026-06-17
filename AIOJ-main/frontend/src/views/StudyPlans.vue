<template>
  <div class="page-container study-plans-page">
    <div class="page-header">
      <h2>学习计划</h2>
      <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon> 新建题单</el-button>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="showDialog" :title="editing ? '编辑题单' : '新建题单'" width="600px" @closed="resetForm">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题" required><el-input v-model="form.title" placeholder="题单标题" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.desc" type="textarea" :rows="2" placeholder="题单描述" /></el-form-item>
        <el-form-item label="难度">
          <el-radio-group v-model="form.diff">
            <el-radio-button v-for="d in diffs" :key="d" :value="d">{{ d }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="添加题目">
          <el-select
            v-model="selectedProblem"
            filterable
            remote
            reserve-keyword
            placeholder="搜索题目ID或标题"
            :remote-method="searchProblems"
            :loading="searchLoading"
            clearable
            value-key="id"
            style="width:100%"
            @change="addProblem"
          >
            <el-option v-for="p in searchResults" :key="p.id" :label="`#${p.id} ${p.title}`" :value="p">
              <span style="float:left">{{ p.title }}</span>
              <span style="float:right;color:var(--text-muted);font-size:13px">#{{ p.id }} {{ p.difficulty }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="按标签添加">
          <el-input v-model="tagInput" placeholder="输入标签名，如 哈希表 动态规划，回车添加" @keyup.enter="addByTags" />
          <div style="font-size:11px;color:var(--text-muted);margin-top:4px">按标签查找未做的题目并加入题单</div>
        </el-form-item>
        <!-- 已选题目列表 -->
        <el-form-item v-if="form.problems.length" label="已选题目">
          <div style="display:flex;flex-wrap:wrap;gap:6px">
            <el-tag
              v-for="(p,i) in form.problems"
              :key="p.id"
              closable
              size="small"
              @close="form.problems.splice(i,1)"
            >
              #{{ p.id }} {{ p.title }}
            </el-tag>
          </div>
          <div style="font-size:11px;color:var(--text-muted);margin-top:4px">共 {{ form.problems.length }} 道，可拖拽排序</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ editing ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input v-model="searchQuery" placeholder="搜索题单..." clearable @input="fetchPlans" style="max-width:400px">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
    </div>

    <!-- 我的题单 -->
    <div class="section-title" v-if="myPlans.length">我的题单</div>
    <div class="plans-grid">
      <div v-for="plan in myPlans" :key="plan.id" class="plan-card card" @click="viewPlan(plan.id)">
        <div class="plan-card-header">
          <h3>{{ plan.title }}</h3>
          <el-tag v-if="plan.isFavorited && !plan.isOwner" size="small" type="warning" effect="plain">收藏</el-tag>
          <el-tag v-if="plan.difficulty" size="small" effect="plain">{{ plan.difficulty }}</el-tag>
        </div>
        <p v-if="plan.description" class="plan-desc">{{ plan.description }}</p>
        <div class="plan-stats">
          <span>{{ plan.completedCount || 0 }}/{{ plan.problemCount || 0 }} 完成</span>
        </div>
        <div class="plan-actions" @click.stop>
          <el-button size="small" text @click="viewPlan(plan.id)">查看</el-button>
          <el-button size="small" text @click="toggleFav(plan)" :type="plan.isFavorited ? 'warning' : ''">
            {{ plan.isFavorited ? '★ 已收藏' : '☆ 收藏' }}
          </el-button>
          <template v-if="plan.isOwner">
            <el-button size="small" text type="primary" @click="editPlan(plan)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(plan.id)">
              <template #reference><el-button size="small" text type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </div>
      </div>
    </div>

    <!-- 发现题单 -->
    <div class="section-title" v-if="otherPlans.length">发现题单</div>
    <div class="plans-grid">
      <div v-for="plan in otherPlans" :key="plan.id" class="plan-card card" @click="viewPlan(plan.id)">
        <div class="plan-card-header">
          <h3>{{ plan.title }}</h3>
          <el-tag v-if="plan.difficulty" size="small" effect="plain">{{ plan.difficulty }}</el-tag>
        </div>
        <p v-if="plan.description" class="plan-desc">{{ plan.description }}</p>
        <div class="plan-stats">
          <span>{{ plan.completedCount || 0 }}/{{ plan.problemCount || 0 }} 完成</span>
          <span v-if="plan.username" class="plan-owner">by {{ plan.username }}</span>
        </div>
        <div class="plan-actions" @click.stop>
          <el-button size="small" text @click="viewPlan(plan.id)">查看</el-button>
          <el-button size="small" text @click="toggleFav(plan)" :type="plan.isFavorited ? 'warning' : ''">
            {{ plan.isFavorited ? '★ 已收藏' : '☆ 收藏' }}
          </el-button>
        </div>
      </div>
    </div>

    <el-empty v-if="!loading && !plans.length" description="暂无题单，点击上方按钮创建" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/api/index'
import { ElMessage } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

const router = useRouter()
const plans = ref([])
const loading = ref(false)
const myPlans = computed(() => plans.value.filter(p => p.isOwner || p.isFavorited))
const otherPlans = computed(() => plans.value.filter(p => !p.isOwner && !p.isFavorited))
const showDialog = ref(false)
const saving = ref(false)
const editing = ref(null)
const searchQuery = ref('')
const searchResults = ref([])
const searchLoading = ref(false)
const selectedProblem = ref(null)
const tagInput = ref('')
const diffs = ['简单', '中等', '困难']
const form = reactive({ title: '', desc: '', diff: '', problems: [] })

function resetForm() {
  form.title = ''; form.desc = ''; form.diff = ''; form.problems = []
  editing.value = null; selectedProblem.value = null; tagInput.value = ''
}

function openCreate() { resetForm(); showDialog.value = true }

function editPlan(plan) {
  editing.value = plan
  form.title = plan.title; form.desc = plan.description || ''; form.diff = plan.difficulty || ''
  form.problems = plan.items?.map(item => ({ id: item.problemId, title: item.title, difficulty: item.difficulty })) || []
  showDialog.value = true
}

async function searchProblems(query) {
  if (!query) { searchResults.value = []; return }
  searchLoading.value = true
  try {
    const r = await http.get('/problems', { params: { search: query, pageSize: 8 } })
    searchResults.value = r.data?.items || []
  } catch { searchResults.value = [] }
  finally { searchLoading.value = false }
}

function addProblem(p) {
  if (p && !form.problems.find(x => x.id === p.id)) form.problems.push(p)
  selectedProblem.value = null
}

async function addByTags() {
  const tags = tagInput.value.split(/[,，\s]+/).map(s => s.trim()).filter(Boolean)
  if (!tags.length) return
  try {
    const r = await http.post('/knowledge/problems-by-tags', { tags, onlyUntried: true })
    const added = (r.data?.untried || []).filter(p => !form.problems.find(x => x.id === p.id))
    added.forEach(p => form.problems.push(p))
    ElMessage.success(`从标签添加了 ${added.length} 道题`)
  } catch { ElMessage.error('搜索失败') }
  tagInput.value = ''
}

async function handleSave() {
  if (!form.title.trim()) { ElMessage.warning('请输入标题'); return }
  if (!form.problems.length) { ElMessage.warning('请添加题目'); return }
  const problemIDs = form.problems.map(p => p.id)
  saving.value = true
  try {
    if (editing.value) {
      await http.put(`/study-plans/${editing.value.id}`, { title: form.title, description: form.desc, difficulty: form.diff, problemIDs })
      ElMessage.success('已保存')
    } else {
      await http.post('/study-plans', { title: form.title, description: form.desc, difficulty: form.diff, problemIDs })
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    fetchPlans()
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

async function handleDelete(id) {
  try { await http.delete(`/study-plans/${id}`); ElMessage.success('已删除'); fetchPlans() } catch { ElMessage.error('删除失败') }
}

async function toggleFav(plan) {
  try { const r = await http.post(`/study-plans/${plan.id}/favorite`); plan.isFavorited = r.data?.favorited } catch {}
}

async function fetchPlans() {
  loading.value = true
  try {
    const params = {}
    if (searchQuery.value) params.q = searchQuery.value
    const r = await http.get('/study-plans', { params })
    plans.value = r.data?.items || []
  } catch { plans.value = [] }
  finally { loading.value = false }
}

function viewPlan(id) { router.push(`/study-plans/${id}`) }

onMounted(fetchPlans)
</script>

<style scoped>
.study-plans-page { max-width: 1040px; margin: 0 auto; padding: 24px 0 }

.page-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 28px; padding-bottom: 20px;
  border-bottom: 1px solid var(--border-light);
}
.page-header h2 {
  font-family: 'JetBrains Mono', 'Cascadia Code', 'Fira Code', monospace;
  font-size: 26px; font-weight: 700; letter-spacing: -0.03em;
  background: linear-gradient(135deg, var(--text-primary) 30%, var(--accent-gold, #e6a23c) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text; margin: 0;
}

.search-bar { margin-bottom: 24px }
.search-bar :deep(.el-input__wrapper) {
  border-radius: 10px; transition: box-shadow 0.25s;
}
.search-bar :deep(.el-input__wrapper:hover) { box-shadow: 0 0 0 1px var(--accent-gold, #e6a23c) inset }

.section-title {
  font-family: 'JetBrains Mono', monospace;
  font-size: 15px; font-weight: 700; letter-spacing: 0.04em;
  margin: 28px 0 14px; color: var(--text-primary);
  display: flex; align-items: center; gap: 10px;
}
.section-title::after {
  content: ''; flex: 1; height: 1px;
  background: linear-gradient(90deg, var(--border-light) 0%, transparent 100%);
}

.plans-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(310px, 1fr));
  gap: 14px; margin-bottom: 12px;
}

.plan-card {
  padding: 22px 24px 16px; cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.25s, border-color 0.25s;
  position: relative; overflow: hidden;
  border-radius: 14px;
  border: 1px solid var(--border-light);
  background: var(--bg-card);
}
.plan-card::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px;
  background: linear-gradient(90deg, var(--accent-gold, #e6a23c), var(--accent-primary));
  opacity: 0; transition: opacity 0.25s;
}
.plan-card:hover::before { opacity: 1 }
.plan-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0,0,0,0.15), 0 0 0 1px var(--accent-gold, #e6a23c40);
  border-color: var(--accent-gold, #e6a23c30);
}

.plan-card-header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px }
.plan-card-header h3 {
  font-size: 15px; font-weight: 700; margin: 0;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  font-family: 'JetBrains Mono', monospace; letter-spacing: -0.01em;
}

.plan-desc {
  font-size: 13px; color: var(--text-secondary); line-height: 1.65;
  margin: 0 0 10px; display: -webkit-box; -webkit-line-clamp: 2;
  -webkit-box-orient: vertical; overflow: hidden;
}

.plan-stats {
  display: flex; align-items: center; gap: 14px;
  font-size: 11px; color: var(--text-muted); flex-wrap: wrap;
  font-family: 'JetBrains Mono', monospace;
}
.plan-owner { color: var(--accent-gold, #e6a23c) }

.plan-actions {
  margin-top: 14px; display: flex; gap: 6px;
  border-top: 1px solid var(--border-light); padding-top: 12px;
}

/* Card entrance animation */
.plan-card {
  animation: cardIn 0.4s ease-out both;
}
.plan-card:nth-child(1) { animation-delay: 0.02s }
.plan-card:nth-child(2) { animation-delay: 0.06s }
.plan-card:nth-child(3) { animation-delay: 0.10s }
.plan-card:nth-child(4) { animation-delay: 0.14s }
.plan-card:nth-child(5) { animation-delay: 0.18s }
.plan-card:nth-child(6) { animation-delay: 0.22s }
.plan-card:nth-child(7) { animation-delay: 0.26s }
.plan-card:nth-child(8) { animation-delay: 0.30s }

@keyframes cardIn {
  from { opacity: 0; transform: translateY(16px) scale(0.97) }
  to   { opacity: 1; transform: translateY(0) scale(1) }
}
</style>
