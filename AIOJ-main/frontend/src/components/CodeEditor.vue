<template>
  <div class="code-editor">
    <div class="editor-toolbar">
      <el-select v-model="currentLang" size="small" style="width: 140px" @change="onLangChange">
        <el-option v-for="lang in languages" :key="lang.value" :label="lang.label" :value="lang.value" />
      </el-select>
      <el-select v-model="fontSize" size="small" style="width: 100px">
        <el-option v-for="s in [12, 13, 14, 15, 16, 18]" :key="s" :label="`${s}px`" :value="s" />
      </el-select>
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
    </div>
    <div ref="editorContainer" class="editor-container" />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as monaco from 'monaco-editor'

const props = defineProps({
  modelValue: { type: String, default: '' },
  language: { type: String, default: 'cpp' },
  draftKey: { type: String, default: '' }
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

const currentLang = ref(props.language)
const fontSize = ref(14)
const draftStatus = ref('草稿保护已开启')
const hasSavedDraft = ref(false)
const editorContainer = ref(null)
let editor = null
let saveTimer = null

onMounted(() => {
  nextTick(() => {
    const initialDraft = readDraft(currentLang.value)
    const initialValue = props.modelValue || initialDraft?.code || TEMPLATES[currentLang.value] || ''
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
      padding: { top: 12 }
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
})

watch(() => props.modelValue, val => {
  if (editor && val !== editor.getValue()) {
    editor.setValue(val)
  }
})

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

function resetCode() {
  if (editor) {
    editor.setValue(TEMPLATES[currentLang.value] || '')
  }
}

onBeforeUnmount(() => {
  clearTimeout(saveTimer)
  if (editor) {
    saveDraft(editor.getValue())
  }
  editor?.dispose()
})

defineExpose({ getCode: () => editor?.getValue() || '', setCode: val => editor?.setValue(val) })

function isTemplateCode(value) {
  return Object.values(TEMPLATES).some(t => value.trim() === t.trim())
}

function storageKey(lang = currentLang.value) {
  return props.draftKey ? `terminal-oj:code-draft:${props.draftKey}:${lang}` : ''
}

function readDraft(lang = currentLang.value) {
  const key = storageKey(lang)
  if (!key) return null
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (typeof parsed?.code === 'string') return parsed
    if (typeof raw === 'string') return { code: raw, updatedAt: null }
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
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #1e1e2e;
  border-bottom: 1px solid #2d2d3f;
}
.toolbar-spacer {
  flex: 1;
}
.draft-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #a6e3a1;
  font-size: 12px;
  white-space: nowrap;
}
.editor-container {
  flex: 1;
  min-height: 300px;
}
</style>
