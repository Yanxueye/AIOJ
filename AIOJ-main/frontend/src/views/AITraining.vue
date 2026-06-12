<template>
  <div class="ai-training-page">
    <div class="training-sidebar">
      <div class="sidebar-header">
        <h3>AI 训练</h3>
        <el-button type="primary" size="small" @click="handleNewChat">
          <el-icon><Plus /></el-icon>新对话
        </el-button>
      </div>
      <div class="sidebar-desc">
        独立 AI 对话模式，不关联任何特定题目。你可以自由讨论算法、数据结构、编程技巧等任何话题。
      </div>
      <el-button
        class="graph-button"
        type="success"
        plain
        :loading="aiStore.loading"
        @click="handleBuildGraph"
      >
        整理我的知识图谱
      </el-button>
      <el-divider />
      <div class="sidebar-section">
        <div class="section-title">快速提问</div>
        <div class="quick-prompts">
          <el-button
            v-for="prompt in quickPrompts"
            :key="prompt"
            size="small"
            plain
            @click="sendQuick(prompt)"
          >
            {{ prompt }}
          </el-button>
        </div>
      </div>
      <el-divider />
      <div class="sidebar-section">
        <div class="section-title">提示</div>
        <ul class="tips-list">
          <li>支持 Markdown 和 LaTeX 渲染</li>
          <li>AI 可以帮你分析算法复杂度</li>
          <li>可以请求代码示例和思路提示</li>
          <li>在题目页面使用 AI 将自动关联题目上下文</li>
        </ul>
      </div>
    </div>

    <div class="training-main">
      <AIChat :problem-context="null" />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAIStore } from '@/stores/ai'
import AIChat from '@/components/AIChat.vue'

const aiStore = useAIStore()

const quickPrompts = [
  '什么是动态规划？',
  '如何分析时间复杂度？',
  '常见的排序算法有哪些？',
  '图论入门推荐？',
  '如何准备算法竞赛？',
  '递归和迭代的区别？'
]

function handleNewChat() {
  aiStore.startNewConversation()
}

function sendQuick(prompt) {
  aiStore.sendMessage(prompt)
}

async function handleBuildGraph() {
  try {
    await aiStore.buildKnowledgeGraph({ scope: 'recent' })
  } catch {
    return
  }
}

onMounted(() => {
  if (aiStore.currentMessages.length === 0) {
    aiStore.startNewConversation()
  }
})
</script>

<style scoped>
.ai-training-page {
  display: flex;
  height: calc(100vh - 60px);
}
.training-sidebar {
  width: 300px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  padding: 20px;
  overflow-y: auto;
  flex-shrink: 0;
}
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.sidebar-header h3 {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.sidebar-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}
.graph-button {
  width: 100%;
  margin-top: 14px;
}
.sidebar-section {
  margin-bottom: 4px;
}
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-bottom: 12px;
}
.quick-prompts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.quick-prompts .el-button {
  font-size: 12px;
}
.tips-list {
  padding-left: 18px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 2;
}

.training-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.training-main :deep(.ai-chat) {
  border: none;
  border-radius: 0;
  height: 100%;
}
</style>
