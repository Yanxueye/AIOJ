<template>
  <div class="markdown-body" v-html="rendered"></div>
</template>

<script setup>
import { computed } from 'vue'
import { renderMarkdown } from '@/utils/markdown'

const props = defineProps({
  content: { type: String, default: '' }
})

const rendered = computed(() => renderMarkdown(props.content))
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
  background: #f6f8fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
}
.markdown-body pre {
  background: #1e1e2e;
  color: #cdd6f4;
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
  background: #f0f5ff;
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
  background: #f5f7fa;
  font-weight: 600;
}
.markdown-body .katex-display {
  margin: 16px 0;
  overflow-x: auto;
}
</style>
