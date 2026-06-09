const delay = (ms = 300) => new Promise(r => setTimeout(r, ms + Math.random() * 200))

const wrap = data => ({ code: 0, message: 'ok', data })

const ALGORITHMS = ['动态规划', '贪心', '搜索', '图论', '数学', '字符串', '数据结构', '模拟', '排序', '二分']
const DIFFICULTIES = ['简单', '中等', '困难']

function generateProblems(count = 50) {
  const list = []
  for (let i = 1; i <= count; i++) {
    const diff = DIFFICULTIES[i % 3]
    list.push({
      id: 1000 + i,
      title: `${['两数之和', '最长回文子串', '合并区间', '接雨水', '全排列', '最短路径', '背包问题', '编辑距离', '岛屿数量', '二叉树遍历'][i % 10]} ${i > 10 ? 'II' : ''}`.trim(),
      difficulty: diff,
      difficultyScore: diff === '简单' ? 800 + (i % 5) * 100 : diff === '中等' ? 1300 + (i % 5) * 100 : 1800 + (i % 5) * 100,
      tags: [ALGORITHMS[i % 10], ALGORITHMS[(i + 3) % 10]],
      acceptRate: (40 + Math.random() * 50).toFixed(1),
      submitCount: Math.floor(100 + Math.random() * 5000),
      accepted: i % 4 === 0
    })
  }
  return list
}

const PROBLEMS = generateProblems()

const PROBLEM_DETAIL_TEMPLATE = {
  content: `## 题目描述

给定一个整数数组 \`nums\` 和一个整数目标值 \`target\`，请你在该数组中找出和为目标值的两个整数，并返回它们的数组下标。

你可以假设每种输入只会对应一个答案，并且你不能使用两次相同的元素。

## 输入格式

第一行包含两个整数 $n$ 和 $target$，其中 $1 \\leq n \\leq 10^5$，$-10^9 \\leq target \\leq 10^9$。

第二行包含 $n$ 个整数 $a_1, a_2, \\ldots, a_n$，其中 $-10^9 \\leq a_i \\leq 10^9$。

## 输出格式

输出两个整数，表示和为 $target$ 的两个数的下标（从 0 开始），用空格分隔。

## 样例

### 输入
\`\`\`
4 9
2 7 11 15
\`\`\`

### 输出
\`\`\`
0 1
\`\`\`

## 提示

- 时间复杂度要求：$O(n)$
- 空间复杂度要求：$O(n)$

可以考虑使用哈希表来优化查找过程。`,
  timeLimit: 1000,
  memoryLimit: 256,
  source: 'TerminalOJ 原创题目'
}

const STATUSES = ['Accepted', 'Wrong Answer', 'Time Limit Exceeded', 'Runtime Error', 'Compilation Error', 'Pending']
const LANGUAGES = ['C++', 'Java', 'Python3', 'Go']

function generateSubmissions(count = 80) {
  const list = []
  const now = Date.now()
  for (let i = 0; i < count; i++) {
    const status = STATUSES[Math.floor(Math.random() * STATUSES.length)]
    list.push({
      id: 100000 + i,
      problemId: 1000 + Math.floor(Math.random() * 50) + 1,
      problemTitle: PROBLEMS[Math.floor(Math.random() * 50)].title,
      status,
      language: LANGUAGES[Math.floor(Math.random() * LANGUAGES.length)],
      runtime: status === 'Accepted' ? Math.floor(Math.random() * 500) + 10 : null,
      memory: status === 'Accepted' ? (Math.random() * 64 + 1).toFixed(1) : null,
      createdAt: new Date(now - i * 3600000 * Math.random() * 48).toISOString(),
      codeLength: Math.floor(Math.random() * 2000) + 200
    })
  }
  return list.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
}

const SUBMISSIONS = generateSubmissions()

const ANNOUNCEMENTS = [
  { id: 1, title: '🎉 TerminalOJ 正式上线！', content: '欢迎使用 TerminalOJ 在线评测系统，祝大家刷题愉快！', date: '2026-04-01', type: 'success' },
  { id: 2, title: '📢 新增 AI 辅助训练功能', content: '现在你可以在做题时使用 AI 助手获取思路提示，同时支持独立的 AI 训练模式。', date: '2026-04-03', type: 'info' },
  { id: 3, title: '🔧 系统维护通知', content: '4月10日 02:00-04:00 将进行系统维护，届时评测服务暂停。', date: '2026-04-05', type: 'warning' },
  { id: 4, title: '🏆 每周竞赛开放报名', content: '第一期每周竞赛将于4月12日 19:00 开始，欢迎报名参加！', date: '2026-04-06', type: 'primary' }
]

const USER_PROFILE = {
  id: 1,
  username: 'coder_test',
  email: 'test@terminaloj.com',
  avatar: '',
  bio: '热爱算法的开发者',
  rating: 1520,
  rank: 42,
  solvedCount: 28,
  totalSubmissions: 65,
  acceptRate: '43.1',
  registeredAt: '2026-03-15',
  solvedByDifficulty: { '简单': 15, '中等': 10, '困难': 3 },
  solvedByAlgorithm: {
    '动态规划': 8, '贪心': 5, '搜索': 4, '图论': 3,
    '数学': 3, '字符串': 2, '数据结构': 2, '模拟': 1
  },
  recentActivity: [
    { date: '2026-04-06', count: 3 }, { date: '2026-04-05', count: 5 },
    { date: '2026-04-04', count: 2 }, { date: '2026-04-03', count: 0 },
    { date: '2026-04-02', count: 4 }, { date: '2026-04-01', count: 1 }
  ]
}

export const mockApi = {
  async login({ username, password }) {
    await delay(500)
    if (!username || !password) throw new Error('请输入用户名和密码')
    return wrap({
      token: 'mock_jwt_' + btoa(username) + '_' + Date.now(),
      user: { ...USER_PROFILE, username }
    })
  },

  async register({ username, email, password }) {
    await delay(500)
    if (!username || !email || !password) throw new Error('请填写完整信息')
    return wrap({ message: '注册成功' })
  },

  async getProfile() {
    await delay(300)
    return wrap(USER_PROFILE)
  },

  async updateProfile(data) {
    await delay(300)
    return wrap({ ...USER_PROFILE, ...data })
  },

  async getProblems({ page = 1, pageSize = 20, keyword = '', difficulty = '', tag = '' } = {}) {
    await delay(400)
    let filtered = [...PROBLEMS]
    if (keyword) filtered = filtered.filter(p => p.title.includes(keyword) || String(p.id).includes(keyword))
    if (difficulty) filtered = filtered.filter(p => p.difficulty === difficulty)
    if (tag) filtered = filtered.filter(p => p.tags.includes(tag))
    const start = (page - 1) * pageSize
    return wrap({ list: filtered.slice(start, start + pageSize), total: filtered.length })
  },

  async getProblemDetail(id) {
    await delay(300)
    const base = PROBLEMS.find(p => p.id === Number(id))
    if (!base) throw new Error('题目不存在')
    return wrap({ ...base, ...PROBLEM_DETAIL_TEMPLATE })
  },

  async submitCode({ problemId, language, code }) {
    await delay(1500)
    const statuses = ['Accepted', 'Wrong Answer', 'Time Limit Exceeded', 'Accepted', 'Accepted']
    const status = statuses[Math.floor(Math.random() * statuses.length)]
    return wrap({
      id: Date.now(),
      problemId,
      status,
      language,
      runtime: status === 'Accepted' ? Math.floor(Math.random() * 200) + 20 : null,
      memory: status === 'Accepted' ? (Math.random() * 32 + 2).toFixed(1) : null,
      createdAt: new Date().toISOString()
    })
  },

  async getSubmissions({ page = 1, pageSize = 20, problemId = '', status = '', sortBy = 'time' } = {}) {
    await delay(400)
    let filtered = [...SUBMISSIONS]
    if (problemId) filtered = filtered.filter(s => s.problemId === Number(problemId))
    if (status) filtered = filtered.filter(s => s.status === status)
    if (sortBy === 'problemId') filtered.sort((a, b) => a.problemId - b.problemId)
    const start = (page - 1) * pageSize
    return wrap({ list: filtered.slice(start, start + pageSize), total: filtered.length })
  },

  async getSubmissionDetail(id) {
    await delay(200)
    const sub = SUBMISSIONS.find(s => s.id === Number(id))
    return wrap(sub || null)
  },

  async aiChat({ message, history, problem_id, conversation_id }) {
    await delay(800 + Math.random() * 1200)
    const contextNote = problem_id ? `\n\n> 当前关联题目 ID: ${problem_id}` : ''
    const replies = [
      `这是一个很好的问题！让我来分析一下：\n\n首先，我们需要理解问题的核心：\n\n1. **分析输入输出**：仔细观察给定的样例\n2. **选择合适的算法**：根据时间复杂度要求选择\n3. **实现与优化**：编写代码并进行优化\n\n你可以尝试使用 **哈希表** 来优化查找过程，时间复杂度为 $O(n)$。\n\n\`\`\`cpp\nunordered_map<int, int> mp;\nfor (int i = 0; i < n; i++) {\n    if (mp.count(target - nums[i])) {\n        return {mp[target - nums[i]], i};\n    }\n    mp[nums[i]] = i;\n}\n\`\`\`${contextNote}`,
      `让我帮你理清思路：\n\n这道题可以用 **动态规划** 来解决。\n\n### 状态定义\n设 $dp[i]$ 表示以第 $i$ 个元素结尾的最优解。\n\n### 状态转移\n$$dp[i] = \\max_{j < i}(dp[j] + w(j, i))$$\n\n### 边界条件\n- $dp[0] = 0$\n\n### 复杂度分析\n- 时间：$O(n^2)$，可以用数据结构优化到 $O(n \\log n)$\n- 空间：$O(n)$${contextNote}`,
      `好的，我来给你一些提示：\n\n**关键观察**：这道题本质上是一个 **图论问题**。\n\n1. 将每个元素看作图中的节点\n2. 根据条件建边\n3. 然后在图上进行 BFS/DFS\n\n> 💡 提示：注意边界条件的处理，特别是当 $n = 1$ 的情况。\n\n如果你需要更详细的解释，请告诉我具体哪个部分不理解。${contextNote}`
    ]
    return wrap({
      reply: replies[Math.floor(Math.random() * replies.length)],
      conversationId: conversation_id || 'mock_conv_' + Date.now(),
      provider: 'mock'
    })
  },

  async getAIHistory() {
    await delay(300)
    return wrap({ conversations: [] })
  },

  async getAIMessages() {
    await delay(200)
    return wrap({ conversation: null, messages: [] })
  },

  async aiCodeDiagnosis({ problemId, language, code }) {
    await delay(700)
    const issues = []
    if (!code?.trim()) {
      issues.push({ severity: 'error', message: '代码为空', hint: '请先输入待诊断代码。' })
    }
    if (code?.includes('TODO')) {
      issues.push({ severity: 'warning', message: '代码中包含 TODO 占位', hint: '提交前补齐逻辑。' })
    }
    if ((code?.match(/{/g) || []).length !== (code?.match(/}/g) || []).length) {
      issues.push({ severity: 'error', message: '花括号数量不匹配', hint: '检查代码块闭合。' })
    }
    if (issues.length === 0) {
      issues.push({ severity: 'info', message: 'Mock 检查未发现明显语法级问题', hint: '继续用边界用例验证。' })
    }
    const rawMarkdown = `### 代码诊断\n\n题目：#${problemId}，语言：${language}\n\n#### 发现的问题\n\n${issues.map(i => `- **${i.severity}**：${i.message}。${i.hint}`).join('\n')}\n\n#### 建议\n\n- 先跑样例，再补充极值和重复数据。\n- 根据题目约束重新核对时间复杂度。`
    return wrap({
      summary: 'Mock 代码诊断完成。',
      issues,
      suggestions: ['补充边界用例', '检查复杂度', '确认输入输出格式'],
      rawMarkdown,
      provider: 'mock'
    })
  },

  async aiKnowledgeGraph({ problemId, scope = 'recent' } = {}) {
    await delay(700)
    const nodes = [
      { id: 'user', label: '当前用户', type: 'user', weight: 1 },
      { id: 'tag:动态规划', label: '动态规划', type: 'algorithm', weight: 8 },
      { id: 'tag:图论', label: '图论', type: 'algorithm', weight: 3 },
      { id: 'status:Accepted', label: 'Accepted', type: 'status', weight: 12 }
    ]
    if (problemId) nodes.push({ id: `problem:${problemId}`, label: `题目 ${problemId}`, type: 'problem', weight: 1 })
    const edges = [
      { source: 'user', target: 'tag:动态规划', type: 'strong_at', weight: 8 },
      { source: 'user', target: 'tag:图论', type: 'need_practice', weight: 3 },
      { source: 'user', target: 'status:Accepted', type: 'has_result', weight: 12 }
    ]
    const rawMarkdown = `### 学习知识图谱\n\n已按 \`${scope}\` 范围生成 Mock 图谱。\n\n- 节点数：${nodes.length}\n- 关系数：${edges.length}\n\n#### 建议\n\n- 继续巩固动态规划的状态设计。\n- 增加图论最短路和连通性题目练习。`
    return wrap({ summary: 'Mock 知识图谱生成完成。', nodes, edges, rawMarkdown, provider: 'mock' })
  },

  async aiSolve({ problemId, question = '', level = 'hint' }) {
    await delay(800)
    return wrap({
      answer: `### #${problemId} 解题辅助\n\n当前级别：\`${level}\`。\n\n先从暴力思路出发，确认状态或数据结构设计，再根据约束优化。${question ? `\n\n你的问题：\n\n> ${question}` : ''}`,
      hints: ['手算样例观察规律', '确认边界条件', '写出复杂度再提交'],
      complexity: 'Mock 模式下建议目标复杂度控制在题目约束可接受范围内。',
      provider: 'mock'
    })
  },

  async getAnnouncements() {
    await delay(200)
    return wrap(ANNOUNCEMENTS)
  }
}
