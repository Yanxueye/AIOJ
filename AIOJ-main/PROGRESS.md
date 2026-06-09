# TerminalOJ 前端开发进度文档

> 最后更新：2026-04-06

---

## 一、技术栈选型与理由

| 层级 | 选型 | 选型理由 |
|------|------|----------|
| 框架 | Vue 3 + Composition API | 组合式 API 使逻辑复用和状态管理更直观，`<script setup>` 语法减少模板代码 |
| 构建 | Vite 5 | 原生 ESM 开发服务器，冷启动 < 2 秒，HMR 毫秒级响应 |
| 路由 | Vue Router 4 | 支持路由懒加载 `() => import()` + 全局前置守卫实现认证拦截 |
| 状态 | Pinia | Vue 官方推荐，相比 Vuex 去掉了 mutations，API 更简洁 |
| UI | Element Plus | Vue 3 生态中最成熟的组件库，表格/表单/分页/弹窗等 OJ 场景高频组件完备 |
| 图表 | ECharts 5 + vue-echarts | 饼图/柱状图配置灵活，vue-echarts 封装了响应式尺寸和自动销毁 |
| 编辑器 | Monaco Editor | VSCode 同款内核，语法高亮/智能提示/多语言支持，OJ 场景的标准选择 |
| 渲染 | marked + KaTeX + highlight.js | 三者组合实现 Markdown + LaTeX 数学公式 + 代码块高亮的完整渲染链路 |
| 请求 | Axios | 拦截器机制天然适合统一处理 JWT 注入和错误响应 |

---

## 二、业务模块实现清单

### 1. 导航与首页

- [x] 顶部固定导航栏（`position: fixed`），包含首页 / 题库 / 评测 / AI 训练入口
- [x] 用户头像下拉菜单（个人中心 / 退出登录），未登录时展示登录/注册按钮
- [x] 首页 Hero 区域（CSS 渐变背景 `linear-gradient`）+ 数据统计概览
- [x] 公告栏组件，支持点击弹窗查看公告详情
- [x] 四宫格快捷入口卡片（hover 上浮动效 `transform: translateY(-2px)`）
- [x] 侧边栏展示个人做题简报与热门题目

### 2. 题目查询

- [x] 题目列表页，支持分页（20 / 50 / 100 条切换）
- [x] 关键字搜索（题号 + 题目名称模糊匹配，`setTimeout` 防抖 300ms）
- [x] 难度与算法标签下拉筛选
- [x] 表格列排序（难度自定义排序函数 / 分数 / 通过率）
- [x] 已通过题目状态标记（绿色 `CircleCheckFilled` 图标）
- [x] 点击整行跳转至题目详情

### 3. 题目详情（三栏分割页面）

- [x] **左栏**：题目描述（Markdown + LaTeX 渲染），显示时间/内存限制
- [x] **中栏**：Monaco 代码编辑器，支持 C++ / Java / Python3 / Go 切换
- [x] **右栏**：AI 助手面板，自动关联当前题目上下文
- [x] 右栏可通过工具栏按钮选择性开启/关闭
- [x] 各栏之间支持鼠标拖拽调整宽度
- [x] 编辑器内置各语言模板代码，支持字号切换与代码重置
- [x] 提交后结果以动画滑入面板展示（状态 / 运行时间 / 内存占用）

### 4. 提交评测

- [x] 题目详情页内提交按钮，调用 API 并返回评测结果
- [x] 评测状态列表页，展示全部历史提交记录
- [x] 三维度筛选：按时间/题号排序、按评测状态筛选、按题号搜索
- [x] 分页控件 + 手动刷新按钮
- [x] 状态字段彩色标注（绿 AC / 红 WA / 橙 TLE / 紫 CE / 灰 Pending）

### 5. 个人学习

- [x] 用户资料卡片（字母头像 / 用户名 / 简介 / 邮箱 / 注册日期）
- [x] 四项核心指标卡片（Rating / 已解决 / 总提交 / 通过率），图标+数据布局
- [x] 难度分布环形饼图（简单绿 / 中等橙 / 困难红）
- [x] 算法分类水平柱状图（渐变色填充，按数量降序排列）
- [x] 个人资料编辑弹窗（邮箱 / 个人简介，200 字限制）
- [x] AI 训练入口按钮

### 6. 注册与登录

- [x] 登录 / 注册页面（居中卡片 + 渐变全屏背景）
- [x] Element Plus 表单校验（用户名必填 / 密码 ≥ 6 位 / 邮箱格式 / 密码确认一致性）
- [x] JWT Token 持久化（localStorage）
- [x] 路由守卫：未认证重定向登录页，已认证跳过登录页
- [x] 登录后跳转回原目标页面（`redirect` query 参数）

### 7. AI 训练

- [x] 独立 AI 对话页面（不绑定题目上下文）+ 侧边栏快捷提问
- [x] 题目详情页内 AI 助手（自动携带 `problem_id` 上下文）
- [x] 消息历史双向排列（用户右对齐蓝色气泡 / AI 左对齐白色气泡）
- [x] AI 回复支持 Markdown / LaTeX / 代码高亮渲染
- [x] 加载中三点跳动动画指示器
- [x] 新对话 / 清空对话功能

---

## 三、重点难点详解

### 难点 1：三栏可拖拽分割布局

**问题**：题目详情页需要同时展示题目描述、代码编辑器、AI 助手三个面板，且用户可以拖拽调整各面板宽度，还能选择性关闭 AI 面板。

**实现方案**：

```
┌──────────────┬─┬──────────────┬─┬──────────────┐
│  题目描述     │↔│  代码编辑器   │↔│  AI 助手      │
│  flex: 1     │ │  flex: 1     │ │  flex: 0.8   │
└──────────────┴─┴──────────────┴─┴──────────────┘
                ↑ divider (5px, cursor: col-resize)
```

- 三个面板使用 CSS Flexbox 布局，每个面板通过 `flex` 值控制宽度比例
- 面板之间插入 5px 宽的 `divider` 元素，设置 `cursor: col-resize`
- 拖拽逻辑：`mousedown` 记录起始 X 坐标和初始 flex 值 → `mousemove` 计算位移比例并更新两侧面板的 flex → `mouseup` 清除监听
- 设置 `min-width: 200px` + `flex >= 0.3` 防止面板被拖至不可见
- AI 面板通过 `v-if` 条件渲染，关闭后编辑器和题目描述自动均分空间

**核心代码逻辑**（`ProblemDetail.vue`）：

```js
// 响应式 flex 比例
const panelFlex = reactive({ problem: 1, editor: 1, ai: 0.8 })

function startResize(e, side) {
  resizeState = { side, startX: e.clientX, startFlex: { ...panelFlex } }
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

function onResize(e) {
  const dx = e.clientX - resizeState.startX
  const scale = dx / window.innerWidth * 3  // 归一化到 flex 比例空间
  if (resizeState.side === 'left') {
    panelFlex.problem = Math.max(0.3, resizeState.startFlex.problem + scale)
    panelFlex.editor = Math.max(0.3, resizeState.startFlex.editor - scale)
  } else {
    panelFlex.editor = Math.max(0.3, resizeState.startFlex.editor + scale)
    panelFlex.ai = Math.max(0.3, resizeState.startFlex.ai - scale)
  }
}
```

**学习要点**：
- Flexbox 的 `flex` 属性本质是比例分配，修改任意一个面板的 flex 值会自动影响其他面板的实际宽度
- 全局 `mousemove` / `mouseup` 监听必须在 `onBeforeUnmount` 中清除，防止内存泄漏
- 可进一步参考 [Split.js](https://split.js.org/) 了解更完善的分割面板库实现

---

### 难点 2：Monaco Editor 集成

**问题**：Monaco Editor 是一个重量级编辑器（打包后约 3MB），需要正确初始化、语言切换、双向数据绑定，以及组件卸载时的资源释放。

**实现方案**：

```js
// 创建编辑器实例
editor = monaco.editor.create(container, {
  value: initialCode,
  language: 'cpp',
  theme: 'vs-dark',
  automaticLayout: true,  // 关键：自动适应容器大小变化
  minimap: { enabled: false },
  scrollBeyondLastLine: false
})

// 监听内容变化，向外 emit
editor.onDidChangeModelContent(() => {
  emit('update:modelValue', editor.getValue())
})

// 语言切换（不重建编辑器，而是切换 model 语言）
function onLangChange(lang) {
  const model = editor.getModel()
  monaco.editor.setModelLanguage(model, lang)
}

// 组件卸载时必须 dispose
onBeforeUnmount(() => { editor?.dispose() })
```

**踩坑记录**：
1. `automaticLayout: true` 是必须的，否则在 flex 面板拖拽改变宽度后编辑器不会重新计算布局
2. 语言切换不要销毁重建编辑器，用 `setModelLanguage` 只切换语法高亮规则，保留用户已输入的代码
3. 切换语言时检测当前内容是否为模板代码，如果是则自动替换为新语言的模板

**学习要点**：
- Monaco Editor 官方文档：https://microsoft.github.io/monaco-editor/
- `automaticLayout` 内部使用 `ResizeObserver` 监听容器尺寸，如果浏览器不支持可用 `editor.layout()` 手动触发
- 生产环境建议配置 Vite 的 `manualChunks` 将 Monaco 单独分包，避免首屏加载过大

---

### 难点 3：Markdown + LaTeX 混合渲染

**问题**：题目描述和 AI 回复中同时包含 Markdown 语法和 LaTeX 数学公式（行内 `$...$` 和块级 `$$...$$`），需要两者正确共存而不互相干扰。

**实现方案**：先渲染 LaTeX → 再渲染 Markdown，因为 KaTeX 输出的 HTML 不会被 marked 的 Markdown 语法误解析。

```js
import { marked } from 'marked'
import katex from 'katex'

function renderLatex(text) {
  // 第一步：处理块级公式 $$...$$
  text = text.replace(/\$\$([\s\S]+?)\$\$/g, (match, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: true, throwOnError: false })
    } catch { return match }
  })
  // 第二步：处理行内公式 $...$（用负向前瞻排除 $$）
  text = text.replace(/(?<!\$)\$(?!\$)(.+?)(?<!\$)\$(?!\$)/g, (match, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: false, throwOnError: false })
    } catch { return match }
  })
  return text
}

export function renderMarkdown(src) {
  const withLatex = renderLatex(src)   // LaTeX → HTML
  return marked.parse(withLatex)       // Markdown → HTML
}
```

**正则细节**：
- `$$...$$` 使用 `[\s\S]+?` 匹配（非贪婪），允许公式内换行
- `$...$` 使用 `(?<!\$)` 负向后瞻和 `(?!\$)` 负向前瞻，确保不会匹配到 `$$` 中的单个 `$`
- `throwOnError: false` 让 KaTeX 在遇到非法公式时降级显示原文而不是抛出异常

**学习要点**：
- KaTeX 文档：https://katex.org/docs/api.html
- 需要在全局 CSS 中引入 `katex/dist/katex.min.css`，否则公式字体和布局不生效
- 渲染顺序至关重要：如果先 Markdown 后 LaTeX，Markdown 会把 `$` 当作普通文本处理，导致 `_下标_` 等语法与 LaTeX 冲突

---

### 难点 4：JWT 认证与路由守卫

**问题**：需要实现完整的前端认证流程——登录获取 Token、持久化存储、请求自动携带、过期处理、以及页面级别的权限控制。

**实现方案**（三层拦截体系）：

```
[路由守卫]           → 页面级：未登录用户不能进入受保护页面
[Axios 请求拦截器]   → 请求级：每个请求自动注入 Authorization 头
[Axios 响应拦截器]   → 响应级：401 时自动清除 Token 并跳转登录页
```

**路由守卫**（`router/index.js`）：

```js
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.auth && !userStore.isLoggedIn) {
    // 记录目标路径，登录后跳回
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if (to.meta.guest && userStore.isLoggedIn) {
    // 已登录用户访问登录/注册页，直接跳首页
    next({ name: 'home' })
  } else {
    next()
  }
})
```

**Axios 拦截器**（`api/index.js`）：

```js
// 请求拦截：注入 Token
http.interceptors.request.use(config => {
  const token = localStorage.getItem('toj_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截：401 自动登出
http.interceptors.response.use(
  response => response.data,  // 解包：直接返回 data 层
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('toj_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

**学习要点**：
- JWT 的 Token 存储有 `localStorage`（持久） vs `sessionStorage`（关标签页即失效） vs 内存变量（刷新即失效） 三种方案，各有安全性和便利性取舍
- 生产环境应配合 `Refresh Token` 机制实现无感续签，避免 Access Token 过期时用户被强制登出
- Vue Router 的 `beforeEach` 是异步的，可以在其中 `await` 后端验证接口

---

### 难点 5：Mock 数据层架构

**问题**：后端尚未开发，前端需要独立运行并完整验证所有业务功能，Mock 层还需要在后端就绪后能无痛切换。

**实现方案**（策略模式）：

```
src/api/
├── index.js      → Axios 实例 + USE_MOCK 开关
├── mock.js       → 所有 Mock 数据和模拟逻辑（50题/80提交/用户画像/公告）
├── user.js       → 每个方法根据 USE_MOCK 决定走 mockApi 还是 http
├── problem.js
├── submission.js
└── ai.js
```

每个 API 模块的写法统一为：

```js
import http, { USE_MOCK } from './index'
import { mockApi } from './mock'

export const problemApi = {
  getList:   params => USE_MOCK ? mockApi.getProblems(params) : http.get('/problems', { params }),
  getDetail: id     => USE_MOCK ? mockApi.getProblemDetail(id) : http.get(`/problems/${id}`)
}
```

Mock 函数内部用 `await delay(300~500)` 模拟网络延迟，使 loading 状态和骨架屏等 UX 效果可在开发阶段验证。

**切换方式**：修改 `src/api/index.js` 中 `const USE_MOCK = false` 即可，所有 API 调用自动走真实后端。无需修改任何业务代码。

**学习要点**：
- 这种「同接口双实现」模式本质上是策略模式的简化应用
- 更完善的方案可以用 [MSW (Mock Service Worker)](https://mswjs.io/) 在网络层拦截，对业务代码完全透明
- Mock 数据应尽量覆盖边界情况（空列表、超长文本、特殊字符等），提前暴露前端渲染问题

---

### 难点 6：ECharts 响应式图表

**问题**：个人中心需要展示难度分布饼图和算法分类柱状图，且图表需要在窗口缩放和容器尺寸变化时自动重绘。

**实现方案**：

使用 `vue-echarts` 组件，它在内部通过 `ResizeObserver` 实现了自动 resize，只需传入 `autoresize` 属性：

```html
<v-chart :option="chartOption" autoresize style="height: 280px" />
```

ECharts 采用按需引入以减小包体积：

```js
import { use } from 'echarts/core'
import { BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
```

**学习要点**：
- ECharts 5 支持 Tree-shaking，只引入使用的图表类型和组件可以大幅减小打包体积
- `option` 对象使用 Vue 的 `computed` 包装，当数据变化时图表会自动更新
- 柱状图的渐变色通过 `itemStyle.color` 传入 `{ type: 'linear', colorStops: [...] }` 实现

---

### 难点 7：AI 对话的双模式复用

**问题**：AI 聊天组件在两个场景中使用——题目详情页（关联题目上下文）和独立训练页（无上下文），需要同一个组件支持两种模式。

**实现方案**：

`AIChat.vue` 组件接收 `problemContext` prop：

```html
<!-- 题目详情页：传入题目对象 -->
<AIChat :problem-context="problem" />

<!-- 独立训练页：传 null -->
<AIChat :problem-context="null" />
```

Pinia AI Store 中，`sendMessage` 方法将 `problemContext` 传给后端：

```js
async function sendMessage(content, problemContext = null) {
  addMessage('user', content)
  const res = await aiApi.chat({
    message: content,
    history: currentMessages.value.map(m => ({ role: m.role, content: m.content })),
    problem_id: problemContext?.id || null  // null 时后端不关联题目
  })
  addMessage('assistant', res.data.reply)
}
```

通过 `startNewConversation(problemContext)` 方法，在题目页进入时自动注入系统消息标记上下文：

```js
function startNewConversation(problemContext = null) {
  clearMessages()
  if (problemContext) {
    addMessage('system', `当前题目上下文：[${problemContext.id}] ${problemContext.title}`)
  }
}
```

**学习要点**：
- 组件通过 props 控制行为差异，而非维护两份代码，是 Vue 组件设计的核心原则
- AI 对话历史通过 `history` 数组传给后端，后端可用此实现多轮对话的上下文理解
- 后续可扩展为 SSE (Server-Sent Events) 实现流式输出，逐字显示 AI 回复

---

## 四、工程化实践

| 项目 | 状态 | 说明 |
|------|------|------|
| Mock 数据层 | ✅ | 50 道题目 + 80 条提交 + 用户画像 + 4 条公告 |
| API 文档 | ✅ | `frontend/API.md`，覆盖 11 个接口，含完整请求/响应示例 |
| Mock / 真实后端切换 | ✅ | `USE_MOCK` 单一标志位控制 |
| 代码分割 | ✅ | 路由懒加载 + `manualChunks` 将 Monaco / ECharts / Element Plus / Vue 分包 |
| 构建验证 | ✅ | `npm run build` 零错误通过 |
| CSS 设计系统 | ✅ | 全局 CSS 变量定义颜色、圆角、阴影、滚动条等 |
| 响应式 | ✅ | 关键页面 `@media (max-width: 768/960px)` 适配 |

---

## 五、后端（已完成，详见 `backend/PROGRESS.md`）

- [x] Go 后端服务搭建（Gin + Gorm）
- [x] MySQL 数据库表设计（users / problems / submissions / announcements / conversations / messages）
- [x] JWT 签发与验证中间件 + 每用户令牌桶限流
- [x] RabbitMQ 接入（带内存降级），评测任务异步写库
- [x] gRPC 判题服务（proto + 自定义 JSON Codec + Docker 镜像）
- [x] AI 对话接口（会话持久化 + Mock 回复，可切换至真实 LLM）
- [ ] 真实沙箱（isolate / nsjail）替换 `cmd/judger` 的 MockSandbox
- [ ] AI 接入真实 LLM（`config.ai.enabled = true` 时启用）

---

## 六、目录结构

```
AIOJ/
├── PROGRESS.md                  # 本文档
├── frontend/
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── API.md                   # 前端请求 API 接口文档
│   └── src/
│       ├── main.js              # 应用入口：挂载 Vue / Pinia / Router / Element Plus
│       ├── App.vue              # 根组件：条件渲染导航栏
│       ├── router/index.js      # 路由定义 + JWT 前置守卫
│       ├── stores/              # Pinia 状态模块
│       │   ├── user.js          # 登录/登出/Token持久化/个人信息
│       │   ├── problem.js       # 题目列表与详情
│       │   ├── submission.js    # 提交与评测结果
│       │   └── ai.js            # AI 对话历史与消息管理
│       ├── api/                 # 请求层
│       │   ├── index.js         # Axios 实例 + 拦截器 + USE_MOCK 开关
│       │   ├── mock.js          # 完整 Mock 数据与模拟逻辑
│       │   ├── user.js          # 用户认证 & 信息 API
│       │   ├── problem.js       # 题目 & 公告 API
│       │   ├── submission.js    # 提交评测 API
│       │   └── ai.js            # AI 对话 API
│       ├── utils/markdown.js    # Markdown + LaTeX 渲染管线
│       ├── components/          # 可复用组件
│       │   ├── NavBar.vue       # 顶部导航栏
│       │   ├── AnnouncementBoard.vue  # 公告栏
│       │   ├── CodeEditor.vue   # Monaco 代码编辑器封装
│       │   ├── AIChat.vue       # AI 聊天组件（支持题目上下文/独立模式）
│       │   ├── MarkdownRenderer.vue   # Markdown/LaTeX 渲染组件
│       │   └── StatsCharts.vue  # 难度饼图 + 算法柱状图
│       ├── views/               # 页面视图
│       │   ├── Home.vue         # 首页（Hero + 公告 + 快捷入口 + 侧边栏）
│       │   ├── Login.vue        # 登录页
│       │   ├── Register.vue     # 注册页
│       │   ├── ProblemList.vue  # 题目列表（搜索/筛选/分页/排序）
│       │   ├── ProblemDetail.vue # 题目详情（三栏分割 + 拖拽 + 提交）
│       │   ├── SubmissionStatus.vue # 评测状态（多维筛选 + 分页）
│       │   ├── Profile.vue      # 个人中心（统计卡片 + 图表 + 资料编辑）
│       │   └── AITraining.vue   # 独立 AI 训练（侧边栏 + 聊天主体）
│       └── assets/styles/
│           └── global.css       # CSS 变量 + 全局样式 + 滚动条 + 过渡动画
└── worker.skill                 # 开发需求文档
```
