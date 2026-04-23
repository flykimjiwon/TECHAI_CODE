import { useEffect, useRef, useCallback } from 'react'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection, rectangularSelection, crosshairCursor, dropCursor } from '@codemirror/view'
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
  // One Dark syntax colors
  '.cm-keyword': { color: '#c678dd' },
  '.cm-atom': { color: '#d19a66' },
  '.cm-number': { color: '#d19a66' },
  '.cm-string, .cm-string2': { color: '#98c379' },
  '.cm-comment': { color: '#5c6370', fontStyle: 'italic' },
  '.cm-variableName': { color: '#e06c75' },
  '.cm-variableName.cm-definition': { color: '#61afef' },
  '.cm-typeName': { color: '#e5c07b' },
  '.cm-className': { color: '#e5c07b' },
  '.cm-definition': { color: '#61afef' },
  '.cm-function': { color: '#61afef' },
  '.cm-propertyName': { color: '#e06c75' },
  '.cm-operator': { color: '#56b6c2' },
  '.cm-punctuation': { color: '#abb2bf' },
  '.cm-meta': { color: '#abb2bf' },
  '.cm-tagName': { color: '#e06c75' },
  '.cm-attributeName': { color: '#d19a66' },
  '.cm-bool': { color: '#d19a66' },
  '.cm-null': { color: '#d19a66' },
  '.cm-regexp': { color: '#98c379' },
  '.cm-link': { color: '#61afef', textDecoration: 'underline' },
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
}

export default function CodeEditor({ content, filename, onChange, onCursorChange }: Props) {
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
      ],
    })

    const view = new EditorView({ state, parent: containerRef.current })
    viewRef.current = view

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
