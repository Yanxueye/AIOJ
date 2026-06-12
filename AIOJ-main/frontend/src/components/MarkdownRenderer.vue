<template>
  <div ref="container" class="markdown-body" v-html="rendered"></div>
</template>

<script setup>
import { computed, ref, watch, nextTick } from 'vue'
import { renderMarkdown } from '@/utils/markdown'

const props = defineProps({
  content: { type: String, default: '' }
})

const container = ref(null)
const rendered = computed(() => renderMarkdown(props.content))

function attachCopyButtons() {
  if (!container.value) return
  const blocks = container.value.querySelectorAll('pre')
  blocks.forEach(pre => {
    if (pre.querySelector('.code-copy-btn')) return // already added
    const btn = document.createElement('button')
    btn.className = 'code-copy-btn'
    btn.textContent = '复制'
    btn.title = '复制代码'
    btn.onclick = () => {
      const code = pre.querySelector('code')
      const text = code ? code.textContent : pre.textContent
      navigator.clipboard.writeText(text).then(() => {
        btn.textContent = '已复制'
        setTimeout(() => { btn.textContent = '复制' }, 2000)
      }).catch(() => {
        btn.textContent = '失败'
        setTimeout(() => { btn.textContent = '复制' }, 2000)
      })
    }
    pre.style.position = 'relative'
    pre.appendChild(btn)
  })
}

watch(rendered, () => nextTick(attachCopyButtons))
</script>

<style>
@import 'highlight.js/styles/github.css';
@import 'katex/dist/katex.min.css';

.markdown-body {
  font-size: 15px;
  line-height: 1.8;
  color: var(--text-primary);
  word-wrap: break-word;
}
.markdown-body h1, .markdown-body h2, .markdown-body h3 {
  margin: 20px 0 12px;
  font-weight: 600;
  line-height: 1.4;
}
.markdown-body h2 {
  font-size: 20px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}
.markdown-body h3 {
  font-size: 16px;
}
.markdown-body p {
  margin: 8px 0;
}
.markdown-body code {
  background: var(--bg-hover);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
}
.markdown-body pre {
  background: var(--editor-bg);
  color: var(--editor-text);
  border-radius: var(--radius-sm);
  padding: 16px;
  overflow-x: auto;
  margin: 12px 0;
}
.markdown-body pre code {
  background: transparent;
  padding: 0;
  color: inherit;
  font-size: 13px;
  line-height: 1.6;
}
.markdown-body blockquote {
  border-left: 4px solid var(--accent-blue);
  padding: 8px 16px;
  margin: 12px 0;
  background: var(--accent-primary-bg);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  color: var(--text-secondary);
}
.markdown-body ul, .markdown-body ol {
  padding-left: 24px;
  margin: 8px 0;
}
.markdown-body li {
  margin: 4px 0;
}
.markdown-body strong {
  font-weight: 600;
}
.markdown-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 12px 0;
}
.markdown-body th, .markdown-body td {
  border: 1px solid var(--border-color);
  padding: 8px 12px;
  text-align: left;
}
.markdown-body th {
  background: var(--bg-hover);
  font-weight: 600;
}
.markdown-body .katex-display {
  margin: 16px 0;
  overflow-x: auto;
}
.markdown-body pre {
  position: relative;
}
.markdown-body .code-copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px 10px;
  font-size: 12px;
  background: rgba(255,255,255,0.1);
  color: var(--text-muted, #666);
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 4px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.markdown-body pre:hover .code-copy-btn {
  opacity: 1;
}
.markdown-body .code-copy-btn:hover {
  background: rgba(255,255,255,0.2);
  color: var(--text-primary, #fff);
}
</style>
