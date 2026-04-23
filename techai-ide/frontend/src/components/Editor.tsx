import { useState, useEffect, useRef, useCallback, CSSProperties } from 'react'
import { FileCode, X, Save } from 'lucide-react'
import { ReadFile, WriteFile } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import CodeEditor from './CodeEditor'
import { showToast } from './Toast'

interface Tab {
  path: string
  name: string
  content: string
  modified: boolean
}

interface Props {
  filePath: string | null
  onCursorChange?: (line: number, col: number, lang: string) => void
}

const kbd: CSSProperties = {
  background: 'var(--bg-active)', padding: '2px 6px', borderRadius: 4,
  fontSize: 10, fontFamily: 'var(--font-code)', border: '1px solid var(--border)',
  display: 'inline-block',
}

function isImage(name: string): boolean {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  return ['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg', 'ico'].includes(ext)
}

function detectLang(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const map: Record<string, string> = {
    go: 'Go', ts: 'TypeScript', tsx: 'TypeScript React', js: 'JavaScript', jsx: 'JavaScript React',
    py: 'Python', rs: 'Rust', java: 'Java', md: 'Markdown', json: 'JSON', yaml: 'YAML', yml: 'YAML',
    css: 'CSS', scss: 'SCSS', html: 'HTML', sql: 'SQL', sh: 'Shell', bash: 'Shell',
    toml: 'TOML', xml: 'XML', txt: 'Plain Text', mod: 'Go Module', sum: 'Go Sum',
  }
  return map[ext] || 'Plain Text'
}

export default function Editor({ filePath, onCursorChange }: Props) {
  const [tabs, setTabs] = useState<Tab[]>([])
  const [activeTab, setActiveTab] = useState<string | null>(null)
  const [saveFlash, setSaveFlash] = useState(false)
  const [findOpen, setFindOpen] = useState(false)
  const [findQuery, setFindQuery] = useState('')
  const findRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!filePath) return
    const existing = tabs.find(t => t.path === filePath)
    if (existing) {
      setActiveTab(filePath)
      return
    }
    ReadFile(filePath).then(content => {
      const name = filePath.split('/').pop() || filePath
      setTabs(prev => {
        if (prev.find(t => t.path === filePath)) return prev
        return [...prev, { path: filePath, name, content, modified: false }]
      })
      setActiveTab(filePath)
    }).catch(console.error)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filePath])

  useEffect(() => {
    const cancel = EventsOn('file:changed', (path: string) => {
      ReadFile(path).then(content => {
        setTabs(prev => prev.map(t => t.path === path ? { ...t, content, modified: false } : t))
      }).catch(() => {})
    })
    return cancel
  }, [])

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        saveCurrentFile()
      }
      if ((e.metaKey || e.ctrlKey) && e.key === 'f') {
        e.preventDefault()
        setFindOpen(prev => !prev)
        setTimeout(() => findRef.current?.focus(), 50)
      }
      if (e.key === 'Escape' && findOpen) {
        setFindOpen(false)
      }
      if ((e.metaKey || e.ctrlKey) && e.key === 'w') {
        e.preventDefault()
        if (activeTab) {
          const tab = tabs.find(t => t.path === activeTab)
          if (tab?.modified && !confirm('Unsaved changes. Close?')) return
          setTabs(prev => {
            const next = prev.filter(t => t.path !== activeTab)
            setActiveTab(next.length > 0 ? next[next.length - 1].path : null)
            return next
          })
        }
      }
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [activeTab, tabs, findOpen])

  const saveCurrentFile = useCallback(() => {
    const tab = tabs.find(t => t.path === activeTab)
    if (!tab || !tab.modified) return
    WriteFile(tab.path, tab.content).then(() => {
      setTabs(prev => prev.map(t => t.path === activeTab ? { ...t, modified: false } : t))
      setSaveFlash(true)
      setTimeout(() => setSaveFlash(false), 800)
      showToast(`Saved ${tab.name}`, 'success')
    }).catch(err => showToast(`Save failed: ${err}`, 'error'))
  }, [activeTab, tabs])

  function handleContentChange(value: string) {
    setTabs(prev => prev.map(t =>
      t.path === activeTab ? { ...t, content: value, modified: true } : t
    ))
  }

  function closeTab(path: string, e: React.MouseEvent) {
    e.stopPropagation()
    const tab = tabs.find(t => t.path === path)
    if (tab?.modified && !confirm('Unsaved changes. Close anyway?')) return
    setTabs(prev => prev.filter(t => t.path !== path))
    if (activeTab === path) {
      const remaining = tabs.filter(t => t.path !== path)
      setActiveTab(remaining.length > 0 ? remaining[remaining.length - 1].path : null)
    }
  }

  const current = tabs.find(t => t.path === activeTab)

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      {/* Tab Bar */}
      <div style={{
        height: 36, background: 'var(--bg-activity)', display: 'flex',
        borderBottom: '1px solid var(--border)', overflowX: 'auto', alignItems: 'stretch',
      }}>
        {tabs.map(tab => (
          <div key={tab.path} onClick={() => setActiveTab(tab.path)} style={{
            padding: '0 14px', display: 'flex', alignItems: 'center', gap: 7,
            fontSize: 12.5, color: activeTab === tab.path ? 'var(--fg-primary)' : 'var(--fg-muted)',
            background: activeTab === tab.path ? 'var(--bg-editor)' : 'var(--bg-activity)',
            borderRight: '1px solid var(--border)', cursor: 'pointer',
            minWidth: 110, position: 'relative',
          }}>
            {activeTab === tab.path && <span style={{
              position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: 'var(--accent)'
            }} />}
            <FileCode size={14} style={{ color: '#61afef', flexShrink: 0 }} />
            <span>{tab.name}</span>
            {tab.modified && <span style={{ color: 'var(--warning)', fontSize: 18, lineHeight: 1 }}>&#9679;</span>}
            <span onClick={(e) => closeTab(tab.path, e)} style={{
              marginLeft: 'auto', opacity: 0.4, cursor: 'pointer', lineHeight: 1
            }}>
              <X size={12} />
            </span>
          </div>
        ))}
        {saveFlash && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, padding: '0 12px', fontSize: 11, color: 'var(--success)' }}>
            <Save size={12} /> Saved
          </div>
        )}
      </div>

      {/* Breadcrumb */}
      {current && (
        <div style={{
          padding: '3px 14px', fontSize: 11, color: 'var(--fg-dim)',
          borderBottom: '1px solid var(--border)', background: 'var(--bg-editor)',
          display: 'flex', alignItems: 'center', gap: 4,
          fontFamily: 'var(--font-code)',
        }}>
          {current.path.split('/').map((part, i, arr) => (
            <span key={i}>
              {i > 0 && <span style={{ margin: '0 2px', color: 'var(--fg-dim)' }}>/</span>}
              <span style={{ color: i === arr.length - 1 ? 'var(--fg-primary)' : 'var(--fg-dim)' }}>{part}</span>
            </span>
          ))}
        </div>
      )}

      {/* Find Bar */}
      {findOpen && (
        <div style={{
          display: 'flex', alignItems: 'center', gap: 6, padding: '4px 10px',
          background: 'var(--bg-panel)', borderBottom: '1px solid var(--border)',
        }}>
          <input ref={findRef} value={findQuery} onChange={e => setFindQuery(e.target.value)}
            onKeyDown={e => { if (e.key === 'Escape') setFindOpen(false) }}
            placeholder="Find... (CodeMirror Ctrl+F also works)"
            style={{
              width: 200, padding: '4px 8px', borderRadius: 4, fontSize: 12,
              background: 'var(--bg-base)', border: '1px solid var(--border)',
              color: 'var(--fg-primary)', fontFamily: 'var(--font-ui)', outline: 'none',
            }}
          />
          <button onClick={() => setFindOpen(false)} style={{
            background: 'none', border: 'none', color: 'var(--fg-dim)', cursor: 'pointer', fontSize: 14,
          }}>&times;</button>
        </div>
      )}

      {/* Code Area — CodeMirror / Image Preview / Welcome */}
      <div style={{ flex: 1, overflow: 'hidden', background: 'var(--bg-editor)', minHeight: 0 }}>
        {current && isImage(current.name) ? (
          <div style={{
            height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center',
            flexDirection: 'column', gap: 8,
          }}>
            <img src={`file://${current.path}`} alt={current.name}
              style={{ maxWidth: '90%', maxHeight: '80%', objectFit: 'contain', borderRadius: 8, border: '1px solid var(--border)' }}
              onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
            />
            <span style={{ fontSize: 11, color: 'var(--fg-dim)' }}>{current.name}</span>
          </div>
        ) : current ? (
          <CodeEditor
            content={current.content}
            filename={current.name}
            onChange={handleContentChange}
            onCursorChange={(line, col) => {
              if (onCursorChange && current) onCursorChange(line, col, detectLang(current.name))
            }}
          />
        ) : (
          <div style={{
            height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: 'var(--fg-dim)', fontSize: 13, flexDirection: 'column', gap: 16,
            fontFamily: 'var(--font-ui)',
          }}>
            <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--accent)', opacity: 0.2, letterSpacing: -1 }}>
              TECHAI IDE
            </div>
            <div style={{ textAlign: 'center', lineHeight: 1.8 }}>
              Open a file from the explorer<br />
              <span style={{ fontSize: 11 }}>or ask TECHAI to create one</span>
            </div>
            <div style={{
              marginTop: 8, fontSize: 11, color: 'var(--fg-dim)',
              display: 'grid', gridTemplateColumns: 'auto auto', gap: '4px 24px', alignItems: 'center',
            }}>
              <kbd style={kbd}>Cmd+P</kbd><span>Quick Open</span>
              <kbd style={kbd}>Cmd+F</kbd><span>Find</span>
              <kbd style={kbd}>Cmd+B</kbd><span>Toggle Sidebar</span>
              <kbd style={kbd}>Cmd+J</kbd><span>Toggle Terminal</span>
              <kbd style={kbd}>Cmd+S</kbd><span>Save</span>
              <kbd style={kbd}>Cmd+,</kbd><span>Theme</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
