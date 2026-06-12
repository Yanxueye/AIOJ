<template>
  <div class="code-editor" :class="{ 'is-fullscreen': isFullscreen }">
    <div class="editor-toolbar">
      <el-select v-model="currentLang" size="small" style="width: 140px" @change="onLangChange">
        <el-option v-for="lang in languages" :key="lang.value" :label="lang.label" :value="lang.value" />
      </el-select>
      <el-select v-model="fontSize" size="small" style="width: 100px" @change="onFontSizeChange">
        <el-option v-for="s in [12, 13, 14, 15, 16, 18]" :key="s" :label="`${s}px`" :value="s" />
      </el-select>
      <el-tooltip content="自动换行" placement="top">
        <el-button size="small" :type="wordWrap === 'on' ? 'primary' : ''" text @click="toggleWordWrap">
          <el-icon><Operation /></el-icon>
        </el-button>
      </el-tooltip>
      <span v-if="draftKey" class="draft-status">
        <el-icon><CircleCheck /></el-icon>
        {{ draftStatus }}
      </span>
      <div class="toolbar-spacer" />
      <el-button v-if="draftKey" size="small" :disabled="!hasSavedDraft" @click="restoreDraft">
        <el-icon><FolderOpened /></el-icon>恢复草稿
      </el-button>
      <el-button size="small" @click="resetCode">
        <el-icon><RefreshRight /></el-icon>重置
      </el-button>
      <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏编辑'" placement="top">
        <el-button size="small" @click="toggleFullscreen">
          <el-icon><FullScreen v-if="!isFullscreen" /><Aim v-else /></el-icon>
        </el-button>
      </el-tooltip>
    </div>
    <div ref="editorContainer" class="editor-container" />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { FullScreen, Aim, Operation } from '@element-plus/icons-vue'
import * as monaco from 'monaco-editor'

const props = defineProps({
  modelValue: { type: String, default: '' },
  language: { type: String, default: 'cpp' },
  draftKey: { type: String, default: '' },
  legacyDraftKey: { type: String, default: '' },
  templates: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['update:modelValue', 'change-language'])

const languages = [
  { label: 'C++', value: 'cpp' },
  { label: 'Python3', value: 'python' },
  { label: 'Go', value: 'go' }
]

const TEMPLATES = {
  cpp: '#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    \n    return 0;\n}\n',
  python: 'import sys\ninput = sys.stdin.readline\n\ndef solve():\n    pass\n\nsolve()\n',
  go: 'package main\n\nimport (\n    "fmt"\n)\n\nfunc main() {\n    fmt.Println()\n}\n'
}

function mergedTemplates() {
  return { ...TEMPLATES, ...props.templates }
}

const currentLang = ref(props.language)
const fontSize = ref(14)
const wordWrap = ref('on')
const isFullscreen = ref(false)
const draftStatus = ref('草稿保护已开启')
const hasSavedDraft = ref(false)
const editorContainer = ref(null)
let editor = null
let saveTimer = null

onMounted(() => {
  // Migrate legacy draft if exists
  if (props.legacyDraftKey) {
    migrateDrafts()
  }
  window.addEventListener('keydown', handleKeydown)
  nextTick(() => {
    const initialDraft = readDraft(currentLang.value)
    const initialValue = props.modelValue || initialDraft?.code || mergedTemplates()[currentLang.value] || ''
    if (initialDraft?.updatedAt) {
      draftStatus.value = `已恢复 ${formatTime(initialDraft.updatedAt)}`
    }

    editor = monaco.editor.create(editorContainer.value, {
      value: initialValue,
      language: currentLang.value,
      theme: 'vs-dark',
      fontSize: fontSize.value,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      automaticLayout: true,
      tabSize: 4,
      lineNumbers: 'on',
      roundedSelection: true,
      cursorBlinking: 'smooth',
      smoothScrolling: true,
      padding: { top: 12 },
      wordWrap: wordWrap.value
    })

    editor.onDidChangeModelContent(() => {
      const nextCode = editor.getValue()
      emit('update:modelValue', nextCode)
      scheduleDraftSave(nextCode)
    })

    emit('update:modelValue', initialValue)
    refreshDraftState()
  })
})

watch(fontSize, val => {
  editor?.updateOptions({ fontSize: val })
  nextTick(() => editor?.layout())
})

function onFontSizeChange() {
  editor?.updateOptions({ fontSize: fontSize.value })
  nextTick(() => editor?.layout())
}

function toggleWordWrap() {
  wordWrap.value = wordWrap.value === 'on' ? 'off' : 'on'
  editor?.updateOptions({ wordWrap: wordWrap.value })
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => editor?.layout())
}

// Handle ESC key to exit fullscreen
function handleKeydown(e) {
  if (e.key === 'Escape' && isFullscreen.value) {
    isFullscreen.value = false
    nextTick(() => editor?.layout())
  }
}

watch(() => props.modelValue, val => {
  if (editor && val !== editor.getValue()) {
    editor.setValue(val)
  }
})

function onLangChange(lang) {
  if (editor) {
    const model = editor.getModel()
    monaco.editor.setModelLanguage(model, lang)
    const draft = readDraft(lang)
    editor.setValue(draft?.code || mergedTemplates()[lang] || '')
    draftStatus.value = draft?.updatedAt ? `已恢复 ${formatTime(draft.updatedAt)}` : '草稿保护已开启'
  }
  refreshDraftState(lang)
  emit('change-language', lang)
}

function resetCode() {
  if (editor) {
    editor.setValue(mergedTemplates()[currentLang.value] || '')
  }
}

onBeforeUnmount(() => {
  clearTimeout(saveTimer)
  window.removeEventListener('keydown', handleKeydown)
  if (editor) {
    saveDraft(editor.getValue())
  }
  editor?.dispose()
})

defineExpose({ getCode: () => editor?.getValue() || '', setCode: val => editor?.setValue(val) })

function isTemplateCode(value) {
  return Object.values(mergedTemplates()).some(t => value.trim() === t.trim())
}

function storageKey(lang = currentLang.value) {
  return props.draftKey ? `terminal-oj:code-draft:${props.draftKey}:${lang}` : ''
}

function legacyStorageKey(lang = currentLang.value) {
  return props.legacyDraftKey ? `terminal-oj:code-draft:${props.legacyDraftKey}:${lang}` : ''
}

function readDraft(lang = currentLang.value) {
  const key = storageKey(lang)
  if (!key) return null
  try {
    const raw = window.localStorage.getItem(key)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (typeof parsed?.code === 'string') return parsed
      if (typeof raw === 'string') return { code: raw, updatedAt: null }
    }
  } catch {
    draftStatus.value = '本地暂存不可用'
  }
  return null
}

function saveDraft(value, lang = currentLang.value) {
  const key = storageKey(lang)
  if (!key) return
  try {
    const updatedAt = Date.now()
    window.localStorage.setItem(key, JSON.stringify({ code: value, language: lang, updatedAt }))
    hasSavedDraft.value = Boolean(value.trim())
    draftStatus.value = `已暂存 ${formatTime(updatedAt)}`
  } catch {
    draftStatus.value = '本地暂存不可用'
  }
}

function scheduleDraftSave(value) {
  if (!props.draftKey) return
  clearTimeout(saveTimer)
  draftStatus.value = '正在暂存...'
  saveTimer = setTimeout(() => saveDraft(value), 500)
}

function restoreDraft() {
  const draft = readDraft()
  if (!editor || !draft?.code) return
  editor.setValue(draft.code)
  draftStatus.value = draft.updatedAt ? `已恢复 ${formatTime(draft.updatedAt)}` : '已恢复草稿'
}

function refreshDraftState(lang = currentLang.value) {
  hasSavedDraft.value = Boolean(readDraft(lang)?.code?.trim())
}

function formatTime(value) {
  return new Date(value).toLocaleTimeString('zh-CN', { hour12: false })
}

function migrateDrafts() {
  if (!props.legacyDraftKey || !props.draftKey) return
  for (const lang of ['cpp', 'python', 'go']) {
    const legacyKey = `terminal-oj:code-draft:${props.legacyDraftKey}:${lang}`
    const newKey = `terminal-oj:code-draft:${props.draftKey}:${lang}`
    try {
      const legacy = window.localStorage.getItem(legacyKey)
      const existing = window.localStorage.getItem(newKey)
      if (legacy && !existing) {
        window.localStorage.setItem(newKey, legacy)
        window.localStorage.removeItem(legacyKey)
      } else if (legacy) {
        window.localStorage.removeItem(legacyKey)
      }
    } catch { /* ignore */ }
  }
}
</script>

<style scoped>
.code-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  border-radius: var(--radius-sm);
  overflow: hidden;
  border: 1px solid var(--border-color);
}
.code-editor.is-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 9999;
  border-radius: 0;
  background: #1e1e1e;
}
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--editor-bg);
  border-bottom: 1px solid var(--editor-border);
}
.toolbar-spacer {
  flex: 1;
}
.draft-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--accent-green);
  font-size: 12px;
  white-space: nowrap;
}
.editor-container {
  flex: 1;
  min-height: 300px;
}
</style>
