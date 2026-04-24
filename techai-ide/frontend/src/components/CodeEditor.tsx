import { useEffect, useRef, useCallback } from 'react'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection, rectangularSelection, crosshairCursor, dropCursor, highlightSpecialChars, scrollPastEnd } from '@codemirror/view'
import { EditorState, Compartment } from '@codemirror/state'
import { defaultKeymap, indentWithTab, history, historyKeymap } from '@codemirror/commands'
import { searchKeymap, highlightSelectionMatches } from '@codemirror/search'
import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { LanguageSupport, indentOnInput, bracketMatching, foldGutter, foldKeymap, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { go } from '@codemirror/lang-go'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { yaml } from '@codemirror/lang-yaml'
import { markdown } from '@codemirror/lang-markdown'
import { json } from '@codemirror/lang-json'
import { css } from '@codemirror/lang-css'
import { html } from '@codemirror/lang-html'
import { sql } from '@codemirror/lang-sql'
import { rust } from '@codemirror/lang-rust'
import { java } from '@codemirror/lang-java'
import { php } from '@codemirror/lang-php'
import { cpp } from '@codemirror/lang-cpp'

// Dark theme matching designs-v3
const techaiTheme = EditorView.theme({
  '&': { backgroundColor: 'var(--bg-editor)', color: 'var(--fg-secondary)', fontSize: '12.5px' },
  '.cm-content': { fontFamily: 'var(--font-code)', lineHeight: '1.55', padding: '10px 0', caretColor: 'var(--accent)' },
  '.cm-gutters': { backgroundColor: 'var(--bg-editor)', color: 'var(--fg-dim)', border: 'none', borderRight: '1px solid var(--border)' },
  '.cm-gutter': { minWidth: '40px' },
  '.cm-activeLineGutter': { backgroundColor: 'transparent', color: 'var(--fg-primary)' },
  '.cm-activeLine': { backgroundColor: 'rgba(255,255,255,0.03)' },
  '.cm-cursor': { borderLeftColor: 'var(--accent)', borderLeftWidth: '2px' },
  '.cm-selectionBackground': { backgroundColor: 'rgba(59,130,246,0.2) !important' },
  '&.cm-focused .cm-selectionBackground': { backgroundColor: 'rgba(59,130,246,0.3) !important' },
  '.cm-matchingBracket': { backgroundColor: 'rgba(255,255,255,0.1)', outline: '1px solid rgba(255,255,255,0.2)' },
  '.cm-searchMatch': { backgroundColor: 'rgba(255,200,0,0.3)' },
  '.cm-searchMatch.cm-searchMatch-selected': { backgroundColor: 'rgba(255,200,0,0.5)' },
  '.cm-foldGutter .cm-gutterElement': { color: 'var(--fg-dim)', cursor: 'pointer', fontSize: '11px' },
  '.cm-scroller': { overflow: 'auto' },
  // Pastel muted — 트렌디한 저채도 톤
  '.cm-keyword': { color: '#a78bba' },       // 파스텔 퍼플
  '.cm-atom': { color: '#bf9070' },           // 부드러운 오렌지
  '.cm-number': { color: '#bf9070' },
  '.cm-string, .cm-string2': { color: '#8aad72' }, // 내추럴 초록
  '.cm-comment': { color: '#5a6475', fontStyle: 'italic' },
  '.cm-variableName': { color: '#c08088' },   // 로즈
  '.cm-variableName.cm-definition': { color: '#7ea8c9' }, // 스카이
  '.cm-typeName': { color: '#c4a86a' },       // 머스타드
  '.cm-className': { color: '#c4a86a' },
  '.cm-definition': { color: '#7ea8c9' },
  '.cm-function': { color: '#7ea8c9' },
  '.cm-propertyName': { color: '#c08088' },
  '.cm-operator': { color: '#7fa8b0' },       // 틸
  '.cm-punctuation': { color: '#7e8590' },
  '.cm-meta': { color: '#7e8590' },
  '.cm-tagName': { color: '#c08088' },
  '.cm-attributeName': { color: '#bf9070' },
  '.cm-bool': { color: '#bf9070' },
  '.cm-null': { color: '#7e8590' },
  '.cm-regexp': { color: '#8aad72' },
  '.cm-link': { color: '#7ea8c9', textDecoration: 'underline' },
  '.cm-heading': { color: '#e06c75', fontWeight: 'bold' },
  '.cm-emphasis': { fontStyle: 'italic' },
  '.cm-strong': { fontWeight: 'bold' },
}, { dark: true })

function getLang(ext: string): LanguageSupport | null {
  switch (ext) {
    case 'go': case 'mod': case 'sum': return go()
    case 'ts': case 'tsx': return javascript({ typescript: true, jsx: ext === 'tsx' })
    case 'js': case 'jsx': case 'mjs': case 'cjs': return javascript({ jsx: ext === 'jsx' })
    case 'py': case 'pyw': return python()
    case 'yaml': case 'yml': return yaml()
    case 'md': case 'mdx': return markdown()
    case 'json': case 'jsonc': return json()
    case 'css': case 'scss': case 'less': return css()
    case 'html': case 'htm': case 'ejs': case 'hbs': return html()
    case 'sql': return sql()
    case 'rs': return rust()
    case 'java': case 'kt': case 'kts': case 'scala': return java()
    case 'php': return php()
    case 'c': case 'cpp': case 'cc': case 'cxx': case 'h': case 'hpp': case 'cs': return cpp()
    case 'vue': case 'svelte': return html()
    default: return null
  }
}

interface Props {
  content: string
  filename: string
  onChange: (value: string) => void
  onCursorChange?: (line: number, col: number) => void
  onAskAI?: (selectedCode: string, filename: string) => void
}

export default function CodeEditor({ content, filename, onChange, onCursorChange, onAskAI }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const langCompartment = useRef(new Compartment())
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  useEffect(() => {
    if (!containerRef.current) return

    const ext = filename.split('.').pop()?.toLowerCase() || ''
    const lang = getLang(ext)

    const updateListener = EditorView.updateListener.of(update => {
      if (update.docChanged) {
        onChangeRef.current(update.state.doc.toString())
      }
      if (update.selectionSet && onCursorChange) {
        const pos = update.state.selection.main.head
        const line = update.state.doc.lineAt(pos)
        onCursorChange(line.number, pos - line.from + 1)
      }
    })

    const state = EditorState.create({
      doc: content,
      extensions: [
        lineNumbers(),
        highlightActiveLine(),
        highlightActiveLineGutter(),
        highlightSelectionMatches(),
        drawSelection(),
        dropCursor(),
        rectangularSelection(),
        crosshairCursor(),
        history(),
        bracketMatching(),
        closeBrackets(),
        indentOnInput(),
        foldGutter(),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        highlightSpecialChars(),
        scrollPastEnd(),
        keymap.of([
          ...defaultKeymap,
          ...searchKeymap,
          ...historyKeymap,
          ...closeBracketsKeymap,
          ...foldKeymap,
          indentWithTab,
        ]),
        langCompartment.current.of(lang ? [lang] : []),
        techaiTheme,
        updateListener,
        EditorState.tabSize.of(4),
        EditorView.clickAddsSelectionRange.of(e => e.altKey),
        keymap.of([
          // Cmd+G go to line
          { key: 'Mod-g', run: (view) => {
            const line = prompt('Go to line:')
            if (line) { const n = parseInt(line)
              if (n > 0 && n <= view.state.doc.lines) {
                view.dispatch({ selection: { anchor: view.state.doc.line(n).from }, scrollIntoView: true }); view.focus()
              }}; return true
          }},
          // Cmd+Shift+D duplicate line
          { key: 'Mod-Shift-d', run: (view) => {
            const { from } = view.state.selection.main
            const line = view.state.doc.lineAt(from)
            view.dispatch({ changes: { from: line.to, insert: '\n' + line.text } }); return true
          }},
          // Cmd+Shift+K delete line
          { key: 'Mod-Shift-k', run: (view) => {
            const { from } = view.state.selection.main
            const line = view.state.doc.lineAt(from)
            const delFrom = line.from === 0 ? line.from : line.from - 1
            view.dispatch({ changes: { from: delFrom, to: line.to } }); return true
          }},
          // Alt+Up/Down move line
          { key: 'Alt-ArrowUp', run: (view) => {
            const { from } = view.state.selection.main
            const line = view.state.doc.lineAt(from)
            if (line.number <= 1) return true
            const prev = view.state.doc.line(line.number - 1)
            view.dispatch({ changes: [
              { from: prev.from, to: line.to, insert: line.text + '\n' + prev.text }
            ], selection: { anchor: prev.from + line.text.length - (from - line.from) } }); return true
          }},
          { key: 'Alt-ArrowDown', run: (view) => {
            const { from } = view.state.selection.main
            const line = view.state.doc.lineAt(from)
            if (line.number >= view.state.doc.lines) return true
            const next = view.state.doc.line(line.number + 1)
            view.dispatch({ changes: [
              { from: line.from, to: next.to, insert: next.text + '\n' + line.text }
            ], selection: { anchor: line.from + next.text.length + 1 + (from - line.from) } }); return true
          }},
          // Cmd+/ toggle line comment
          { key: 'Mod-/', run: (view) => {
            const { from, to } = view.state.selection.main
            const fromLine = view.state.doc.lineAt(from).number
            const toLine = view.state.doc.lineAt(to).number
            const changes: {from: number; to: number; insert: string}[] = []
            let allCommented = true
            for (let i = fromLine; i <= toLine; i++) {
              const l = view.state.doc.line(i)
              if (!l.text.trimStart().startsWith('//')) { allCommented = false; break }
            }
            for (let i = fromLine; i <= toLine; i++) {
              const l = view.state.doc.line(i)
              if (allCommented) {
                const idx = l.text.indexOf('//')
                changes.push({ from: l.from + idx, to: l.from + idx + (l.text[idx+2] === ' ' ? 3 : 2), insert: '' })
              } else {
                changes.push({ from: l.from, to: l.from, insert: '// ' })
              }
            }
            view.dispatch({ changes }); return true
          }},
        ]),
      ],
    })

    const view = new EditorView({ state, parent: containerRef.current })
    viewRef.current = view

    // Right-click context menu for "Ask AI"
    containerRef.current.addEventListener('contextmenu', (e) => {
      const sel = view.state.sliceDoc(view.state.selection.main.from, view.state.selection.main.to)
      if (sel && sel.trim() && onAskAI) {
        e.preventDefault()
        showContextMenu(e.clientX, e.clientY, sel, filename, onAskAI)
      }
    })

    return () => { view.destroy(); viewRef.current = null }
  }, [filename])

  useEffect(() => {
    const view = viewRef.current
    if (!view) return
    const currentContent = view.state.doc.toString()
    if (currentContent !== content) {
      view.dispatch({
        changes: { from: 0, to: currentContent.length, insert: content }
      })
    }
  }, [content])

  const updateLang = useCallback((fname: string) => {
    const view = viewRef.current
    if (!view) return
    const ext = fname.split('.').pop()?.toLowerCase() || ''
    const lang = getLang(ext)
    view.dispatch({ effects: langCompartment.current.reconfigure(lang ? [lang] : []) })
  }, [])

  useEffect(() => { updateLang(filename) }, [filename, updateLang])

  return <div ref={containerRef} style={{ height: '100%', overflow: 'hidden' }} />
}

function showContextMenu(x: number, y: number, code: string, filename: string, onAskAI: (code: string, file: string) => void) {
  // Remove existing menu
  document.getElementById('ai-context-menu')?.remove()

  const menu = document.createElement('div')
  menu.id = 'ai-context-menu'
  menu.style.cssText = `position:fixed;left:${x}px;top:${y}px;z-index:9999;background:var(--bg-panel);border:1px solid var(--border);border-radius:8px;padding:4px 0;box-shadow:0 8px 24px rgba(0,0,0,0.4);min-width:180px;font-family:var(--font-ui);font-size:12px;`

  const items = [
    { label: '💡 Explain Selection', action: () => onAskAI(`Explain this code:\n\`\`\`\n${code}\n\`\`\``, filename) },
    { label: '🔧 Fix / Improve', action: () => onAskAI(`Fix or improve this code from ${filename}:\n\`\`\`\n${code}\n\`\`\``, filename) },
    { label: '📝 Add Comments', action: () => onAskAI(`Add comments to this code:\n\`\`\`\n${code}\n\`\`\``, filename) },
    { label: '🧪 Generate Tests', action: () => onAskAI(`Generate tests for this code from ${filename}:\n\`\`\`\n${code}\n\`\`\``, filename) },
    { label: '♻️ Refactor', action: () => onAskAI(`Refactor this code from ${filename}:\n\`\`\`\n${code}\n\`\`\``, filename) },
  ]

  items.forEach(item => {
    const btn = document.createElement('div')
    btn.textContent = item.label
    btn.style.cssText = `padding:6px 14px;cursor:pointer;color:var(--fg-secondary);transition:background 0.1s;`
    btn.onmouseenter = () => btn.style.background = 'var(--bg-hover)'
    btn.onmouseleave = () => btn.style.background = 'transparent'
    btn.onclick = () => { item.action(); menu.remove() }
    menu.appendChild(btn)
  })

  document.body.appendChild(menu)
  // Close on click outside
  setTimeout(() => {
    document.addEventListener('click', function close() {
      menu.remove()
      document.removeEventListener('click', close)
    })
  }, 100)
}
