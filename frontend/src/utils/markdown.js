import { marked } from 'marked'
import hljs from 'highlight.js'
import katex from 'katex'

marked.setOptions({
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  },
  breaks: true,
  gfm: true
})

function renderLatex(text) {
  // block LaTeX: $$...$$
  text = text.replace(/\$\$([\s\S]+?)\$\$/g, (match, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: true, throwOnError: false })
    } catch {
      return match
    }
  })
  // inline LaTeX: $...$  (but not $$)
  text = text.replace(/(?<!\$)\$(?!\$)(.+?)(?<!\$)\$(?!\$)/g, (match, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: false, throwOnError: false })
    } catch {
      return match
    }
  })
  return text
}

export function renderMarkdown(src) {
  if (!src) return ''
  const withLatex = renderLatex(src)
  return marked.parse(withLatex)
}
