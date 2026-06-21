# TerminalOJ 刷题编辑器体验改进进度报告

> 最后更新：2026-05-06  
> 阶段目标：在不改动后端接口契约的前提下，增强题目详情页代码编辑体验，为 Monaco Editor 增加按题目与语言隔离的本地草稿保护能力，并完成构建与基础验证。

---

## 一、改进背景与目标

在线判题系统的核心使用场景是“阅读题面 -> 编写代码 -> 调试思路 -> 提交评测”。当前项目已经具备题目详情三栏布局、Monaco 编辑器、多语言模板、AI 助手和提交评测能力，但编辑器存在一个典型体验空洞：用户刷新页面、切换路由、误关闭标签页或临时离开时，未提交代码容易丢失。

本次改进选择从这个高频痛点切入，目标如下：

- [x] 为 `CodeEditor.vue` 增加自动草稿保存能力
- [x] 草稿按题目维度隔离，避免不同题目的代码互相覆盖
- [x] 草稿按语言维度隔离，避免 C++ / Java / Python3 / Go 互相覆盖
- [x] 进入题目详情页时自动恢复最近草稿
- [x] 在编辑器工具栏显示草稿状态和最近暂存时间
- [x] 提供手动“恢复草稿”入口
- [x] 保持现有提交、重置、语言切换、双向绑定行为可用
- [x] 完成前端生产构建和后端基础测试

---

## 二、变更清单

| 文件 | 类型 | 本次变更 |
|------|------|----------|
| `frontend/src/components/CodeEditor.vue` | 前端组件 | 新增 `draftKey` prop、本地草稿读写、自动保存、防抖暂存、恢复按钮、草稿状态展示 |
| `frontend/src/views/ProblemDetail.vue` | 页面视图 | 为编辑器传入 `problem-${route.params.id}`，实现题目级隔离 |
| `WORK_SUMMARY.md` | 文档 | 按 Progress 文档格式补充详细工作记录、验证结果、风险与后续计划 |

本次未改动后端接口、数据库结构、路由配置和 API 请求层，因此不会影响现有前后端契约。

---

## 三、功能完成情况

| # | 功能点 | 状态 | 落地位置 |
|---|--------|------|----------|
| 1 | 编辑器支持外部传入草稿命名空间 | 已完成 | `CodeEditor.vue` 的 `draftKey` prop |
| 2 | 输入代码后自动暂存 | 已完成 | `editor.onDidChangeModelContent` -> `scheduleDraftSave` |
| 3 | 暂存写入浏览器本地存储 | 已完成 | `saveDraft` 使用 `window.localStorage` |
| 4 | 进入页面自动读取草稿 | 已完成 | `onMounted` 内调用 `readDraft` |
| 5 | 按题目隔离草稿 | 已完成 | `ProblemDetail.vue` 传入 `problem-${route.params.id}` |
| 6 | 按语言隔离草稿 | 已完成 | `storageKey` 拼接 `draftKey` 与 `lang` |
| 7 | 工具栏展示暂存状态 | 已完成 | `draftStatus` + `.draft-status` |
| 8 | 手动恢复草稿 | 已完成 | “恢复草稿”按钮 + `restoreDraft` |
| 9 | 页面卸载前兜底保存 | 已完成 | `onBeforeUnmount` 调用 `saveDraft` |
| 10 | 构建验证 | 已完成 | `npm run build` |
| 11 | 后端基础测试 | 已完成 | `go test ./...` |

---

## 四、用户流程

### 1. 正常刷题流程

```
进入题目详情页
        |
        v
ProblemDetail.vue 根据 route.params.id 生成草稿 key
        |
        v
CodeEditor.vue 初始化 Monaco Editor
        |
        v
读取 localStorage 中对应题目 + 语言的草稿
        |
        v
有草稿则恢复，没有草稿则使用语言模板
        |
        v
用户输入代码
        |
        v
500ms 防抖后自动暂存
```

### 2. 草稿隔离规则

草稿存储 key 的格式为：

```text
terminal-oj:code-draft:${draftKey}:${lang}
```

题目详情页传入的 `draftKey` 为：

```js
`problem-${route.params.id}`
```

因此实际保存形态类似：

```text
terminal-oj:code-draft:problem-1001:cpp
terminal-oj:code-draft:problem-1001:python
terminal-oj:code-draft:problem-1002:cpp
```

这样可以保证：

- 题目 1001 的 C++ 草稿不会覆盖题目 1002 的 C++ 草稿
- 题目 1001 的 C++ 草稿不会覆盖题目 1001 的 Python 草稿
- 用户在切换语言时，如果当前内容仍是模板代码，会优先尝试恢复目标语言草稿

---

## 五、重点难点详解

### 难点 1：Monaco Editor 与 Vue 双向绑定之间的草稿同步

**问题**：Monaco Editor 不是普通 `<textarea>`，其内容变化来自编辑器内部 model。项目原本通过 `editor.onDidChangeModelContent` 向外 `emit('update:modelValue')`，本次新增草稿保存时必须复用同一个变化入口，避免出现 Vue 状态和 Monaco 状态不一致。

**实现方案**：

```js
editor.onDidChangeModelContent(() => {
  const nextCode = editor.getValue()
  emit('update:modelValue', nextCode)
  scheduleDraftSave(nextCode)
})
```

这样一次编辑会同时完成两件事：

- 更新父组件中的 `code`
- 触发本地草稿防抖保存

**学习要点**：

- Monaco 的数据源是 editor model，不应绕过 `editor.getValue()` 读取内容
- Vue 的 `v-model` 仍作为父子组件之间的数据同步通道
- 草稿保存放在同一个内容变化监听里，行为更集中，减少状态分叉

---

### 难点 2：自动保存不能过于频繁

**问题**：如果用户每输入一个字符都立刻写入 `localStorage`，虽然浏览器能承受，但会造成不必要的同步 IO，也会让状态提示频繁闪动。

**实现方案**：使用 500ms 防抖保存。

```js
function scheduleDraftSave(value) {
  if (!props.draftKey) return
  clearTimeout(saveTimer)
  draftStatus.value = '正在暂存...'
  saveTimer = setTimeout(() => saveDraft(value), 500)
}
```

组件卸载时再做一次兜底保存：

```js
onBeforeUnmount(() => {
  clearTimeout(saveTimer)
  if (editor) {
    saveDraft(editor.getValue())
  }
  editor?.dispose()
})
```

**学习要点**：

- 防抖保存降低写入频率
- 卸载前立即保存避免最后一次输入还在防抖窗口内就离开页面
- `clearTimeout(saveTimer)` 可以避免组件销毁后仍执行延迟任务

---

### 难点 3：语言切换时如何处理当前代码

**问题**：编辑器支持 C++ / Java / Python3 / Go。切换语言时有两种情况：

1. 当前内容还是模板代码，用户还没有真正开始写
2. 当前内容已经是用户代码，不应该被语言模板直接覆盖

**实现方案**：通过 `isTemplateCode` 判断当前内容是否与任一语言模板一致。

```js
function isTemplateCode(value) {
  return Object.values(TEMPLATES).some(t => value.trim() === t.trim())
}
```

语言切换逻辑：

```js
function onLangChange(lang) {
  if (editor) {
    const model = editor.getModel()
    const currentCode = editor.getValue()
    monaco.editor.setModelLanguage(model, lang)
    if (!currentCode.trim() || isTemplateCode(currentCode)) {
      const draft = readDraft(lang)
      editor.setValue(draft?.code || TEMPLATES[lang] || '')
      draftStatus.value = draft?.updatedAt ? `已恢复 ${formatTime(draft.updatedAt)}` : '草稿保护已开启'
    } else {
      scheduleDraftSave(currentCode)
    }
  }
  refreshDraftState(lang)
  emit('change-language', lang)
}
```

**学习要点**：

- 切换 Monaco 语言时只需要 `setModelLanguage`，不需要重建编辑器
- 用户已写代码时优先保护当前内容
- 当前内容为空或仍是模板时，恢复目标语言草稿更符合用户预期

---

### 难点 4：localStorage 异常处理

**问题**：`localStorage` 并非在所有环境中都 100% 可用。例如隐私模式、存储空间不足、浏览器策略限制都可能导致读写失败。

**实现方案**：读写都包裹 `try/catch`，失败时只更新状态文案，不阻断编辑器使用。

```js
function readDraft(lang = currentLang.value) {
  const key = storageKey(lang)
  if (!key) return null
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (typeof parsed?.code === 'string') return parsed
  } catch {
    draftStatus.value = '本地暂存不可用'
  }
  return null
}
```

**学习要点**：

- 本地草稿是体验增强，不应成为核心编辑功能的硬依赖
- 失败时保留编辑器、提交按钮和模板功能可用
- 状态提示让用户知道当前暂存能力不可用

---

## 六、核心代码说明

### 1. 新增组件入参

`CodeEditor.vue` 新增：

```js
draftKey: { type: String, default: '' }
```

当不传 `draftKey` 时，编辑器保持原有行为，不启用草稿保存。这让组件仍可复用于不需要本地暂存的场景。

### 2. 草稿状态

```js
const draftStatus = ref('草稿保护已开启')
const hasSavedDraft = ref(false)
```

- `draftStatus` 用于显示当前草稿状态
- `hasSavedDraft` 用于控制“恢复草稿”按钮是否可点击

### 3. 存储结构

```js
window.localStorage.setItem(key, JSON.stringify({
  code: value,
  language: lang,
  updatedAt
}))
```

保留 `language` 和 `updatedAt` 的原因：

- 后续可以做草稿列表、最近编辑时间、提交前提醒等功能
- `updatedAt` 可直接用于工具栏展示最近暂存时间

### 4. 题目页接入

`ProblemDetail.vue` 中：

```vue
<CodeEditor
  ref="codeEditorRef"
  v-model="code"
  :language="language"
  :draft-key="`problem-${route.params.id}`"
  @change-language="lang => language = lang"
/>
```

这一处改动让编辑器拥有题目上下文，但仍不需要知道题目对象本身，组件边界比较清晰。

---

## 七、验证与运行

### 1. 前端构建验证

已执行：

```bash
npm run build
```

结果：构建通过。

构建输出中仍存在部分大 chunk 提示：

| 模块 | 原因 |
|------|------|
| `monaco-editor` | 编辑器自身较重，包含语言、worker、编辑器内核 |
| `element-plus` | 组件库整体依赖体积较大 |
| `AIChat` | 组合了 AI 组件、Markdown、LaTeX、代码高亮等依赖 |

当前项目已经在 `frontend/vite.config.js` 中配置 `manualChunks`，后续可继续做更细粒度拆分。

### 2. 后端基础测试

虽然本次未修改后端，仍执行了基础测试确认项目状态：

```bash
go test ./...
```

结果：通过。当前后端各包显示 `[no test files]`，表示编译与包加载通过，但尚缺少具体单元测试。

### 3. 本地开发服务器

已启动 Vite 开发服务器：

```text
http://127.0.0.1:5173/
```

可在浏览器中进入题目详情页，验证以下流程：

1. 登录或使用 Mock 登录进入题目详情
2. 在编辑器中输入代码
3. 刷新页面或离开后重新进入
4. 确认代码草稿自动恢复
5. 切换语言，确认不同语言的草稿互不覆盖

---

## 八、风险与边界

| 风险点 | 当前处理 | 后续可优化 |
|--------|----------|------------|
| 浏览器禁用 localStorage | 捕获异常，提示“本地暂存不可用” | 增加 IndexedDB 兜底或服务端草稿 |
| 提交成功后草稿仍保留 | 当前保留，避免误清空用户代码 | 增加提交成功后的清理偏好 |
| 多标签页同时编辑同一题 | 后保存者覆盖先保存者 | 使用 `storage` 事件提示草稿冲突 |
| localStorage 空间有限 | 当前按题目和语言持续保留 | 增加草稿过期清理策略 |
| 当前没有端到端测试 | 已做构建和基础测试 | 使用 Playwright 覆盖刷新恢复流程 |

---

## 九、后续待办

- [ ] 提交成功后提供“保留草稿 / 清理草稿”的用户偏好
- [ ] 在提交记录详情页增加“一键带回编辑器”能力
- [ ] 增加草稿管理入口，展示最近编辑题目、语言和更新时间
- [ ] 支持多标签页草稿冲突提示
- [ ] 对 Monaco 语言包做更细的懒加载优化
- [ ] 为题目详情页增加 Playwright 测试，覆盖草稿保存、刷新恢复、语言切换隔离
- [ ] 后端补充 handler 和 judger 的表驱动单测，提高回归验证可信度

---

## 十、与现有 Progress 文档的对齐

| 对齐项 | 本次文档处理 |
|--------|--------------|
| 标题 | 使用“TerminalOJ + 模块 + 进度报告”命名 |
| 更新时间 | 顶部声明最后更新日期 |
| 阶段目标 | 顶部说明本次改进范围和交付目标 |
| 完成清单 | 使用表格列出功能点、状态和落地位置 |
| 架构说明 | 使用流程图说明题目页、编辑器和 localStorage 的关系 |
| 难点详解 | 按问题、实现方案、学习要点拆分 |
| 验证记录 | 明确记录 `npm run build`、`go test ./...` 和开发服务器地址 |
| 后续计划 | 使用待办清单沉淀下一步可执行事项 |

---

## 十一、总结

本次改进没有扩大系统边界，也没有引入新的依赖，而是在现有 Monaco 编辑器封装之上补齐了一个真实刷题场景中非常关键的体验能力。自动草稿保护让题目详情页更像一个可靠的工作台：用户可以更放心地探索思路、切换语言、打开 AI 助手讨论实现，而不用担心一次刷新就丢掉正在形成的代码。
