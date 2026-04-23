import { useEffect, useRef, useState } from 'react'
import { Plus, X, ChevronDown } from 'lucide-react'
import { WriteTerminal, StartTerminal, GetAvailableShells, GetCurrentShell, SetShell } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface TermTab {
  id: number
  output: string
}

let tabCounter = 0

export default function Terminal() {
  const [tabs, setTabs] = useState<TermTab[]>([{ id: 0, output: '' }])
  const [activeTab, setActiveTab] = useState(0)
  const [shells, setShells] = useState<string[]>([])
  const [currentShell, setCurrentShell] = useState('')
  const [showShellPicker, setShowShellPicker] = useState(false)
  const outputRef = useRef<HTMLPreElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const [cmdHistory, setCmdHistory] = useState<string[]>([])
  const [historyIdx, setHistoryIdx] = useState(-1)

  useEffect(() => {
    GetAvailableShells().then(setShells).catch(() => {})
    GetCurrentShell().then(setCurrentShell).catch(() => {})
  }, [])

  useEffect(() => {
    const cancel = EventsOn('term:output', (data: string) => {
      const clean = data
        .replace(/\x1b\[[0-9;?]*[a-zA-Z]/g, '')
        .replace(/\x1b\][^\x07]*\x07/g, '')
        .replace(/\x1b\[[\d;]*m/g, '')
        .replace(/\x1b[()][AB012]/g, '')
        .replace(/\r\n/g, '\n')
        .replace(/\r/g, '')
      setTabs(prev => prev.map(t =>
        t.id === activeTab ? { ...t, output: t.output + clean } : t
      ))
    })
    return cancel
  }, [activeTab])

  useEffect(() => {
    if (outputRef.current) outputRef.current.scrollTop = outputRef.current.scrollHeight
  }, [tabs, activeTab])

  function addTab() {
    tabCounter++
    setTabs(prev => [...prev, { id: tabCounter, output: '' }])
    setActiveTab(tabCounter)
    StartTerminal().catch(() => {})
  }

  function closeTab(id: number, e: React.MouseEvent) {
    e.stopPropagation()
    if (tabs.length <= 1) return
    setTabs(prev => prev.filter(t => t.id !== id))
    if (activeTab === id) {
      const remaining = tabs.filter(t => t.id !== id)
      setActiveTab(remaining[remaining.length - 1].id)
    }
  }

  async function switchShell(shell: string) {
    setShowShellPicker(false)
    try {
      await SetShell(shell)
      setCurrentShell(shell)
      setTabs([{ id: 0, output: '' }])
      setActiveTab(0)
    } catch {}
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      const val = inputRef.current?.value || ''
      if (val.trim()) setCmdHistory(prev => [val, ...prev].slice(0, 50))
      setHistoryIdx(-1)
      WriteTerminal(val + '\n')
      if (inputRef.current) inputRef.current.value = ''
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (cmdHistory.length > 0) {
        const next = Math.min(historyIdx + 1, cmdHistory.length - 1)
        setHistoryIdx(next)
        if (inputRef.current) inputRef.current.value = cmdHistory[next]
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIdx > 0) {
        const next = historyIdx - 1
        setHistoryIdx(next)
        if (inputRef.current) inputRef.current.value = cmdHistory[next]
      } else {
        setHistoryIdx(-1)
        if (inputRef.current) inputRef.current.value = ''
      }
    } else if (e.key === 'c' && e.ctrlKey) {
      WriteTerminal('\x03')
    } else if (e.key === 'd' && e.ctrlKey) {
      WriteTerminal('\x04')
    }
  }

  const currentTab = tabs.find(t => t.id === activeTab)
  const shellName = currentShell.split('/').pop() || 'bash'

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: 'var(--bg-terminal)' }}>
      {/* Tab bar */}
      <div style={{
        display: 'flex', alignItems: 'stretch', borderBottom: '1px solid var(--border)',
        minHeight: 28, background: 'var(--bg-activity)',
      }}>
        {tabs.map(tab => (
          <div key={tab.id} onClick={() => setActiveTab(tab.id)} style={{
            padding: '0 10px', display: 'flex', alignItems: 'center', gap: 6,
            fontSize: 11, cursor: 'pointer', borderRight: '1px solid var(--border)',
            color: activeTab === tab.id ? 'var(--fg-primary)' : 'var(--fg-muted)',
            background: activeTab === tab.id ? 'var(--bg-terminal)' : 'transparent',
            position: 'relative',
          }}>
            {activeTab === tab.id && <span style={{
              position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: 'var(--success)'
            }} />}
            {shellName} {tab.id > 0 ? `(${tab.id})` : ''}
            {tabs.length > 1 && (
              <span onClick={e => closeTab(tab.id, e)} style={{ opacity: 0.4, cursor: 'pointer' }}>
                <X size={10} />
              </span>
            )}
          </div>
        ))}
        <button onClick={addTab} style={{
          background: 'none', border: 'none', cursor: 'pointer', padding: '0 8px',
          color: 'var(--fg-dim)', display: 'flex', alignItems: 'center',
        }}>
          <Plus size={13} />
        </button>

        {/* Shell selector */}
        <div style={{ marginLeft: 'auto', position: 'relative' }}>
          <button onClick={() => setShowShellPicker(p => !p)} style={{
            background: 'none', border: 'none', cursor: 'pointer', padding: '0 10px',
            color: 'var(--fg-dim)', display: 'flex', alignItems: 'center', gap: 3,
            fontSize: 10, height: '100%',
          }}>
            {shellName} <ChevronDown size={10} />
          </button>
          {showShellPicker && (
            <div style={{
              position: 'absolute', right: 0, top: '100%', zIndex: 100,
              background: 'var(--bg-panel)', border: '1px solid var(--border)',
              borderRadius: 6, boxShadow: '0 8px 24px rgba(0,0,0,0.3)',
              minWidth: 180, padding: '4px 0',
            }}>
              {shells.map(s => (
                <div key={s} onClick={() => switchShell(s)} style={{
                  padding: '6px 12px', fontSize: 12, cursor: 'pointer',
                  color: s === currentShell ? 'var(--accent)' : 'var(--fg-secondary)',
                  background: s === currentShell ? 'var(--bg-active)' : 'transparent',
                  fontFamily: 'var(--font-code)',
                }}
                onMouseEnter={e => { if (s !== currentShell) e.currentTarget.style.background = 'var(--bg-hover)' }}
                onMouseLeave={e => { if (s !== currentShell) e.currentTarget.style.background = 'transparent' }}
                >
                  {s}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <pre ref={outputRef} style={{
        flex: 1, padding: '6px 14px', fontFamily: 'var(--font-code)',
        fontSize: 12, lineHeight: 1.5, overflowY: 'auto', margin: 0,
        color: 'var(--fg-secondary)', whiteSpace: 'pre-wrap', wordBreak: 'break-all',
      }}>
        {currentTab?.output || ''}
      </pre>

      <div style={{
        padding: '4px 14px 6px', display: 'flex', alignItems: 'center', gap: 6,
        borderTop: '1px solid var(--border)',
      }}>
        <span style={{ color: 'var(--success)', fontFamily: 'var(--font-code)', fontSize: 13, fontWeight: 700 }}>$</span>
        <input
          ref={inputRef}
          onKeyDown={handleKeyDown}
          placeholder="Type command..."
          style={{
            flex: 1, background: 'none', border: 'none', outline: 'none',
            color: 'var(--fg-primary)', fontFamily: 'var(--font-code)', fontSize: 12,
          }}
        />
      </div>
    </div>
  )
}
