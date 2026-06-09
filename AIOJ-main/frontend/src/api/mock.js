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
      accepted: i % 4 === 0,
      attempted: i % 3 === 0,
      favorite: i % 5 === 0
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

const STATUSES = ['Pending', 'Queueing', 'Compiling', 'Running', 'Accepted', 'Wrong Answer', 'Compile Error', 'Runtime Error', 'Time Limit Exceeded', 'Memory Limit Exceeded', 'Output Limit Exceeded', 'System Error']
const LANGUAGES = ['cpp', 'python', 'go']

function generateSubmissions(count = 80) {
  const list = []
  const now = Date.now()
  for (let i = 0; i < count; i++) {
    const status = STATUSES[Math.floor(Math.random() * STATUSES.length)]
      list.push({
        id: 100000 + i,
        problemId: 1000 + Math.floor(Math.random() * 50) + 1,
        problemTitle: PROBLEMS[Math.floor(Math.random() * 50)].title,
        traceId: `mock-trace-${100000 + i}`,
        status,
        language: LANGUAGES[Math.floor(Math.random() * LANGUAGES.length)],
        runtime: status === 'Accepted' ? Math.floor(Math.random() * 500) + 10 : 0,
        runtimeMs: status === 'Accepted' ? Math.floor(Math.random() * 500) + 10 : 0,
        memory: status === 'Accepted' ? (Math.random() * 64 + 1).toFixed(1) : '0.0',
        memoryKb: status === 'Accepted' ? Math.floor(Math.random() * 65536) + 1024 : 0,
        compileOutput: status === 'Compile Error' ? 'mock compile output' : '',
        errorMessage: status === 'System Error' ? 'mock system error' : '',
        createdAt: new Date(now - i * 3600000 * Math.random() * 48).toISOString(),
        updatedAt: new Date(now - i * 3600000 * Math.random() * 24).toISOString(),
        caseResults: [],
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
  id: 2,
  username: 'admin',
  role: 'admin',
  email: 'admin@terminaloj.com',
  avatar: '',
  bio: '题库与系统管理员',
  rating: 1800,
  rank: 1,
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

const ADMIN_USERS = [
  { id: 1, username: 'coder_test', email: 'test@terminaloj.com', role: 'user', rating: 1520, registeredAt: '2026-03-15' },
  { id: 2, username: 'admin', email: 'admin@terminaloj.com', role: 'admin', rating: 1800, registeredAt: '2026-03-10' }
]

const AUDIT_LOGS = [
  { id: 1, userId: 2, username: 'admin', userRole: 'admin', resourceType: 'problem', resourceId: '1001', action: 'publish', detail: 'published version 1', createdAt: new Date().toISOString() },
  { id: 2, userId: 2, username: 'admin', userRole: 'admin', resourceType: 'rejudge_job', resourceId: '1', action: 'create', detail: 'created rejudge job for problem 1001', createdAt: new Date().toISOString() }
]

const STUDY_PLANS = [
  {
    id: 1,
    title: '哈希与字符串入门',
    description: '适合刚开始刷题的用户，覆盖哈希表、字符串和基础动态规划。',
    difficulty: '简单',
    tags: ['哈希表', '字符串', '动态规划'],
    problemCount: 3,
    completedCount: 1,
    items: [
      { id: 1, problemId: 1001, orderNo: 1, title: '两数之和', difficulty: '简单' },
      { id: 2, problemId: 1002, orderNo: 2, title: '最长回文子串', difficulty: '中等' },
      { id: 3, problemId: 1004, orderNo: 3, title: '零钱兑换', difficulty: '中等' }
    ]
  },
  {
    id: 2,
    title: '图搜索与进阶结构',
    description: '围绕搜索、图论和堆结构的练习题单。',
    difficulty: '中等',
    tags: ['搜索', '图论', '堆'],
    problemCount: 2,
    completedCount: 0,
    items: [
      { id: 4, problemId: 1005, orderNo: 1, title: '岛屿数量', difficulty: '中等' },
      { id: 5, problemId: 1003, orderNo: 2, title: '合并 K 个升序链表', difficulty: '困难' }
    ]
  }
]

const DAILY_CHALLENGE = {
  id: 1,
  problemId: 1002,
  title: '最长回文子串',
  difficulty: '中等',
  date: '2026-06-09'
}

const STUDY_CHECKINS = [
  { id: 1, userId: 2, date: '2026-06-10', count: 2, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: 2, userId: 2, date: '2026-06-09', count: 1, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }
]

const SOLUTIONS = []

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

  async getAdminUsers() {
    await delay(200)
    return wrap({ items: ADMIN_USERS })
  },

  async updateAdminUserRole(id, data) {
    await delay(200)
    const user = ADMIN_USERS.find(item => item.id === Number(id))
    if (!user) throw new Error('用户不存在')
    user.role = data.role
    return wrap({ id: user.id, username: user.username, role: user.role })
  },

  async getAuditLogs({ action = '', resourceType = '', username = '' } = {}) {
    await delay(200)
    let items = [...AUDIT_LOGS]
    if (action) items = items.filter(item => item.action === action)
    if (resourceType) items = items.filter(item => item.resourceType === resourceType)
    if (username) items = items.filter(item => item.username === username)
    return wrap({ items })
  },

  async getStudyPlans() {
    await delay(160)
    return wrap({ items: STUDY_PLANS.map(({ items, ...rest }) => rest) })
  },

  async getStudyPlanDetail(id) {
    await delay(160)
    const plan = STUDY_PLANS.find(item => item.id === Number(id))
    if (!plan) throw new Error('题单不存在')
    return wrap(plan)
  },

  async getDailyChallenge() {
    await delay(120)
    return wrap(DAILY_CHALLENGE)
  },

  async getStudyCheckins() {
    await delay(120)
    return wrap({ items: STUDY_CHECKINS })
  },

  async getProblems({ page = 1, pageSize = 20, keyword = '', difficulty = '', tag = '', status = '' } = {}) {
    await delay(400)
    let filtered = [...PROBLEMS]
    if (keyword) filtered = filtered.filter(p => p.title.includes(keyword) || String(p.id).includes(keyword))
    if (difficulty) filtered = filtered.filter(p => p.difficulty === difficulty)
    if (tag) filtered = filtered.filter(p => p.tags.includes(tag))
    if (status === 'accepted') filtered = filtered.filter(p => p.accepted)
    if (status === 'attempted') filtered = filtered.filter(p => p.attempted)
    if (status === 'favorite') filtered = filtered.filter(p => p.favorite)
    if (status === 'unattempted') filtered = filtered.filter(p => !p.attempted)
    const start = (page - 1) * pageSize
    return wrap({ list: filtered.slice(start, start + pageSize), total: filtered.length })
  },

  async getProblemDetail(id) {
    await delay(300)
    const base = PROBLEMS.find(p => p.id === Number(id))
    if (!base) throw new Error('题目不存在')
    const solutions = SOLUTIONS.filter(item => item.problemId === Number(id) && item.isPublished)
    const mySolution = SOLUTIONS.find(item => item.problemId === Number(id) && item.userId === USER_PROFILE.id) || {}
    return wrap({
      ...base,
      ...PROBLEM_DETAIL_TEMPLATE,
      status: base.status || 'published',
      constraints: base.constraints || 'mock constraints',
      editorial: base.editorial || 'mock editorial',
      samples: base.samples || [{ caseNo: 1, input: '1 2', expected: '3', explanation: '' }],
      testCases: base.testCases || [{ caseNo: 1, input: '1 2', expected: '3', isHidden: false }],
      favorite: Boolean(base.favorite),
      solutions,
      mySolution,
      templates: base.templates || [
        { language: 'cpp', code: '#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    return 0;\n}\n' },
        { language: 'python', code: 'print()' },
        { language: 'go', code: 'package main\n\nfunc main() {\n}\n' }
      ],
      versions: base.versions || [{ id: 1, versionNo: 1, title: base.title, difficulty: base.difficulty, createdAt: new Date().toISOString(), publishedAt: new Date().toISOString() }]
    })
  },

  async saveProblemSolution(id, data) {
    await delay(200)
    const existing = SOLUTIONS.find(item => item.problemId === Number(id) && item.userId === USER_PROFILE.id)
    if (existing) {
      Object.assign(existing, data, { updatedAt: new Date().toISOString() })
      return wrap(existing)
    }
    const created = {
      id: Date.now(),
      problemId: Number(id),
      userId: USER_PROFILE.id,
      username: USER_PROFILE.username,
      title: data.title,
      content: data.content,
      language: data.language,
      isPublished: data.isPublished,
      updatedAt: new Date().toISOString()
    }
    SOLUTIONS.push(created)
    return wrap(created)
  },

  async getMySolutions() {
    await delay(160)
    return wrap({ items: SOLUTIONS.filter(item => item.userId === USER_PROFILE.id) })
  },

  async getMySolutionDetail(id) {
    await delay(120)
    const item = SOLUTIONS.find(row => row.id === Number(id) && row.userId === USER_PROFILE.id)
    if (!item) throw new Error('题解不存在')
    return wrap(item)
  },

  async getSolutionDetail(id) {
    await delay(120)
    const item = SOLUTIONS.find(row => row.id === Number(id))
    if (!item) throw new Error('题解不存在')
    return wrap(item)
  },

  async favoriteProblem(id) {
    await delay(120)
    const problem = PROBLEMS.find(p => p.id === Number(id))
    if (!problem) throw new Error('题目不存在')
    problem.favorite = true
    return wrap({ problemId: Number(id), favorite: true })
  },

  async unfavoriteProblem(id) {
    await delay(120)
    const problem = PROBLEMS.find(p => p.id === Number(id))
    if (!problem) throw new Error('题目不存在')
    problem.favorite = false
    return wrap({ problemId: Number(id), favorite: false })
  },

  async runProblemCode(id, data) {
    await delay(800)
    const problem = PROBLEMS.find(p => p.id === Number(id))
    if (!problem) throw new Error('题目不存在')
    const status = data.code?.includes('compile_error') ? 'Compile Error' : 'Accepted'
    return wrap({
      traceId: `mock-run-${Date.now()}`,
      problemId: Number(id),
      problemTitle: problem.title,
      source: 'run',
      status,
      language: data.language,
      runtime: 12,
      runtimeMs: 12,
      memory: '1.5',
      memoryKb: 1536,
      compileOutput: status === 'Compile Error' ? 'mock compile output' : '',
      errorMessage: '',
      customInput: data.customInput || '',
      stdout: status === 'Accepted' ? (data.customInput || 'mock run output') : '',
      stderr: '',
      caseResults: [
        {
          caseNo: 1,
          status,
          runtimeMs: 12,
          memoryKb: 1536,
          stdoutBytes: status === 'Accepted' ? (data.customInput || 'mock run output').length : 0,
          stderrBytes: 0,
          signal: '',
          stdoutPreview: status === 'Accepted' ? (data.customInput || 'mock run output') : '',
          stderrPreview: ''
        }
      ]
    })
  },

  async createProblem(data) {
    await delay(300)
    const problem = {
      id: Number(data.id),
      title: data.title,
      difficulty: data.difficulty,
      difficultyScore: data.difficultyScore,
      tags: data.tags || [],
      acceptRate: '0.0',
      submitCount: 0,
      accepted: false,
      status: data.status || 'draft',
      reviewComment: data.reviewComment || '',
      content: data.content,
      constraints: data.constraints || '',
      editorial: data.editorial || '',
      timeLimit: data.timeLimit,
      memoryLimit: data.memoryLimit,
      outputLimitKb: data.outputLimitKb,
      source: data.source,
      samples: data.samples || [],
      testCases: data.testCases || [],
      templates: data.templates || [],
      versions: [{ id: Date.now(), versionNo: 1, title: data.title, difficulty: data.difficulty, createdAt: new Date().toISOString(), publishedAt: data.status === 'published' ? new Date().toISOString() : '' }]
    }
    PROBLEMS.unshift(problem)
    return wrap(problem)
  },

  async updateProblem(id, data) {
    await delay(300)
    const index = PROBLEMS.findIndex(p => p.id === Number(id))
    if (index < 0) throw new Error('题目不存在')
    const current = PROBLEMS[index]
    const nextVersionNo = (current.versions?.[0]?.versionNo || 0) + 1
    PROBLEMS[index] = {
      ...current,
      ...data,
      versions: [{ id: Date.now(), versionNo: nextVersionNo, title: data.title || current.title, difficulty: data.difficulty || current.difficulty, createdAt: new Date().toISOString(), publishedAt: current.status === 'published' ? new Date().toISOString() : '' }, ...(current.versions || [])]
    }
    return wrap(PROBLEMS[index])
  },

  async getProblemVersions(id) {
    await delay(180)
    const problem = PROBLEMS.find(p => p.id === Number(id))
    if (!problem) throw new Error('题目不存在')
    return wrap({ problemId: Number(id), items: problem.versions || [] })
  },

  async publishProblem(id, data) {
    await delay(220)
    const index = PROBLEMS.findIndex(p => p.id === Number(id))
    if (index < 0) throw new Error('题目不存在')
    PROBLEMS[index] = {
      ...PROBLEMS[index],
      status: 'published',
      reviewComment: data.reviewComment || '',
      publishedAt: new Date().toISOString()
    }
    return wrap(PROBLEMS[index])
  },

  async rollbackProblem(id, data) {
    await delay(220)
    const index = PROBLEMS.findIndex(p => p.id === Number(id))
    if (index < 0) throw new Error('题目不存在')
    const version = (PROBLEMS[index].versions || []).find(v => v.id === data.versionId)
    if (!version) throw new Error('目标版本不存在')
    PROBLEMS[index] = {
      ...PROBLEMS[index],
      title: version.title,
      difficulty: version.difficulty
    }
    return wrap(PROBLEMS[index])
  },

  async rejudgeProblem(id, data) {
    await delay(220)
    return wrap({
      id: Date.now(),
      problemId: Number(id),
      targetVersionId: data.targetVersionId || null,
      status: 'pending',
      reason: data.reason || '',
      totalSubmissions: 0,
      processedCount: 0,
      succeededCount: 0,
      failedCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    })
  },

  async getRejudgeJobs(id) {
    await delay(120)
    return wrap({ problemId: Number(id), items: [] })
  },

  async deleteProblem(id) {
    await delay(300)
    const index = PROBLEMS.findIndex(p => p.id === Number(id))
    if (index < 0) throw new Error('题目不存在')
    PROBLEMS.splice(index, 1)
    return wrap({ deleted: true, id: Number(id) })
  },

  async submitCode({ problemId, language, code }) {
    await delay(1500)
    const statuses = ['Accepted', 'Wrong Answer', 'Time Limit Exceeded', 'Compile Error', 'Memory Limit Exceeded', 'Output Limit Exceeded']
    const status = statuses[Math.floor(Math.random() * statuses.length)]
    const runtimeMs = status === 'Accepted' ? Math.floor(Math.random() * 200) + 20 : 0
    const memoryKb = status === 'Accepted' ? Math.floor(Math.random() * 16384) + 1024 : 0
    const id = Date.now()
    return wrap({
      id,
      problemId,
      traceId: `mock-trace-${id}`,
      status,
      language,
      runtime: runtimeMs,
      runtimeMs,
      memory: (memoryKb / 1024).toFixed(1),
      memoryKb,
      compileOutput: status === 'Compile Error' ? 'mock compile output' : '',
      errorMessage: '',
      caseResults: [
        {
          submissionId: id,
          caseNo: 1,
          status,
          runtimeMs,
          memoryKb,
          stdoutBytes: status === 'Output Limit Exceeded' ? 4096 : 0,
          stderrBytes: 0,
          signal: '',
          stdoutPreview: '',
          stderrPreview: ''
        }
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
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

  async getSubmissionCases(id) {
    await delay(120)
    const sub = SUBMISSIONS.find(s => s.id === Number(id))
    return wrap({ submissionId: Number(id), items: sub?.caseResults || [] })
  },

  async getSubmissionOutput(id) {
    await delay(120)
    const sub = SUBMISSIONS.find(s => s.id === Number(id))
    const first = sub?.caseResults?.[0]
    return wrap({ submissionId: Number(id), stdout: first?.stdoutPreview || '', stderr: first?.stderrPreview || '' })
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
