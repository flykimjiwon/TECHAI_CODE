import { useEffect, useRef, useState } from 'react'
import { Plus, X, ChevronDown, Trash2 } from 'lucide-react'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import 'xterm/css/xterm.css'
import { WriteTerminal, StartTerminal, GetAvailableShells, GetCurrentShell, SetShell, ResizeTerminal } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface TermTab { id: number; label: string }
let tabCounter = 0

export default function Terminal() {
  const [tabs, setTabs] = useState<TermTab[]>([{ id: 0, label: 'bash' }])
  const [activeTab, setActiveTab] = useState(0)
  const [shells, setShells] = useState<string[]>([])
  const [currentShell, setCurrentShell] = useState('')
  const [showShellPicker, setShowShellPicker] = useState(false)
  const termRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const fitRef = useRef<FitAddon | null>(null)

  useEffect(() => {
    if (!termRef.current) return
    // Read terminal colors from CSS variables (theme-aware)
    const cs = getComputedStyle(document.documentElement)
    const cv = (v: string, fb: string) => cs.getPropertyValue(v).trim() || fb
    const termBg = cv('--term-bg', cv('--bg-terminal', '#08080a'))
    const termFg = cv('--term-fg', cv('--fg-secondary', '#a1a1aa'))
    const termCursor = cv('--term-cursor', cv('--accent', '#3b82f6'))

    const term = new XTerm({
      fontSize: 12,
      fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
      theme: {
        background: termBg, foreground: termFg, cursor: termCursor,
        selectionBackground: 'rgba(59,130,246,0.3)',
        black: termBg,
        red: cv('--code-variable', '#c08088'),
        green: cv('--code-string', '#8aad72'),
        yellow: cv('--code-type', '#c4a86a'),
        blue: cv('--code-function', '#7ea8c9'),
        magenta: cv('--code-keyword', '#a78bba'),
        cyan: cv('--code-operator', '#7fa8b0'),
        white: termFg,
        brightBlack: cv('--code-comment', '#5a6475'),
        brightRed: cv('--error', '#ef4444'),
        brightGreen: cv('--success', '#10b981'),
        brightYellow: cv('--warning', '#f59e0b'),
        brightBlue: cv('--accent', '#3b82f6'),
        brightMagenta: cv('--code-keyword', '#a78bba'),
        brightCyan: cv('--code-operator', '#7fa8b0'),
        brightWhite: cv('--fg-primary', '#e4e4e7'),
      },
      cursorBlink: true, scrollback: 5000, allowTransparency: true,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.loadAddon(new WebLinksAddon((_event, url) => {
      // Open localhost URLs in IDE preview, others in external browser
      if (url.includes('localhost') || url.includes('127.0.0.1')) {
        import('../../wailsjs/runtime/runtime').then(r => r.EventsEmit('preview:open', url))
      } else {
        window.open(url, '_blank')
      }
    }))
    term.open(termRef.current)
    setTimeout(() => { fit.fit(); ResizeTerminal(term.rows, term.cols).catch(() => {}) }, 100)
    term.onData(data => WriteTerminal(data))
    const obs = new ResizeObserver(() => { fit.fit(); ResizeTerminal(term.rows, term.cols).catch(() => {}) })
    obs.observe(termRef.current)
    xtermRef.current = term; fitRef.current = fit
    return () => { obs.disconnect(); term.dispose(); xtermRef.current = null }
  }, [])

  useEffect(() => {
    const cancel = EventsOn('term:output', (data: string) => { xtermRef.current?.write(data) })
    return cancel
  }, [])

  // Watch for theme changes and update xterm colors
  useEffect(() => {
    const observer = new MutationObserver(() => {
      if (!xtermRef.current) return
      const cs = getComputedStyle(document.documentElement)
      const cv = (v: string, fb: string) => cs.getPropertyValue(v).trim() || fb
      xtermRef.current.options.theme = {
        background: cv('--term-bg', cv('--bg-terminal', '#08080a')),
        foreground: cv('--term-fg', cv('--fg-secondary', '#a1a1aa')),
        cursor: cv('--term-cursor', cv('--accent', '#3b82f6')),
        selectionBackground: 'rgba(59,130,246,0.3)',
        black: cv('--term-bg', '#08080a'),
        red: cv('--code-variable', '#c08088'),
        green: cv('--code-string', '#8aad72'),
        yellow: cv('--code-type', '#c4a86a'),
        blue: cv('--code-function', '#7ea8c9'),
        magenta: cv('--code-keyword', '#a78bba'),
        cyan: cv('--code-operator', '#7fa8b0'),
        white: cv('--term-fg', '#a1a1aa'),
        brightBlack: cv('--code-comment', '#5a6475'),
        brightRed: cv('--error', '#ef4444'),
        brightGreen: cv('--success', '#10b981'),
        brightYellow: cv('--warning', '#f59e0b'),
        brightBlue: cv('--accent', '#3b82f6'),
        brightMagenta: cv('--code-keyword', '#a78bba'),
        brightCyan: cv('--code-operator', '#7fa8b0'),
        brightWhite: cv('--fg-primary', '#e4e4e7'),
      }
      // Force redraw
      xtermRef.current.refresh(0, xtermRef.current.rows - 1)
    })
    observer.observe(document.body, { attributes: true, attributeFilter: ['class'] })
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    // Start terminal on mount (on-demand, not at app startup)
    StartTerminal().catch(() => {})
    GetAvailableShells().then(setShells).catch(() => {})
    GetCurrentShell().then(s => {
      setCurrentShell(s)
      const name = s.includes('\\') ? s.split('\\').pop() : s.split('/').pop()
      setTabs([{ id: 0, label: name || 'shell' }])
    }).catch(() => {})
  }, [])

  function addTab() {
    tabCounter++
    const name = currentShell.split('/').pop() || 'bash'
    setTabs(prev => [...prev, { id: tabCounter, label: `${name} (${tabCounter})` }])
    setActiveTab(tabCounter)
    StartTerminal().catch(() => {})
  }
  function closeTab(id: number, e: React.MouseEvent) {
    e.stopPropagation()
    if (tabs.length <= 1) return
    setTabs(prev => prev.filter(t => t.id !== id))
    if (activeTab === id) { const r = tabs.filter(t => t.id !== id); setActiveTab(r[r.length-1].id) }
  }
  async function switchShell(shell: string) {
    setShowShellPicker(false)
    try { await SetShell(shell); setCurrentShell(shell); xtermRef.current?.clear()
      setTabs([{id:0,label:shell.split('/').pop()||'bash'}]); setActiveTab(0) } catch {}
  }

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: 'var(--bg-terminal, #08080a)' }}>
      <div style={{ display:'flex', alignItems:'stretch', borderBottom:'1px solid var(--border)', minHeight:28, background:'var(--bg-activity)' }}>
        {tabs.map(tab => (
          <div key={tab.id} onClick={() => setActiveTab(tab.id)} style={{
            padding:'0 10px', display:'flex', alignItems:'center', gap:6, fontSize:11, cursor:'pointer',
            borderRight:'1px solid var(--border)', color:activeTab===tab.id?'var(--fg-primary)':'var(--fg-muted)',
            background:activeTab===tab.id?'var(--bg-terminal, #08080a)':'transparent', position:'relative',
          }}>
            {activeTab===tab.id && <span style={{position:'absolute',top:0,left:0,right:0,height:2,background:'var(--success)'}}/>}
            {tab.label}
            {tabs.length>1 && <span onClick={e=>closeTab(tab.id,e)} style={{opacity:.4,cursor:'pointer'}}><X size={10}/></span>}
          </div>
        ))}
        <button onClick={addTab} style={{background:'none',border:'none',cursor:'pointer',padding:'0 8px',color:'var(--fg-dim)',display:'flex',alignItems:'center'}}>
          <Plus size={13}/>
        </button>
        <button title="Clear terminal" onClick={() => { xtermRef.current?.clear(); xtermRef.current?.write('\x1b[2J\x1b[H') }} style={{background:'none',border:'none',cursor:'pointer',padding:'0 6px',color:'var(--fg-dim)',display:'flex',alignItems:'center'}}>
          <Trash2 size={11}/>
        </button>
        <div style={{marginLeft:'auto',position:'relative'}}>
          <button onClick={()=>setShowShellPicker(p=>!p)} style={{
            background:'none',border:'none',cursor:'pointer',padding:'0 10px',color:'var(--fg-dim)',
            display:'flex',alignItems:'center',gap:3,fontSize:10,height:'100%',
          }}>
            {currentShell.split('/').pop()||'bash'} <ChevronDown size={10}/>
          </button>
          {showShellPicker && (
            <div style={{position:'absolute',right:0,top:'100%',zIndex:100,background:'var(--bg-panel)',border:'1px solid var(--border)',
              borderRadius:6,boxShadow:'0 8px 24px rgba(0,0,0,0.3)',minWidth:180,padding:'4px 0'}}>
              {shells.map(s=>(
                <div key={s} onClick={()=>switchShell(s)} style={{
                  padding:'6px 12px',fontSize:12,cursor:'pointer',fontFamily:'var(--font-code)',
                  color:s===currentShell?'var(--accent)':'var(--fg-secondary)',
                  background:s===currentShell?'var(--bg-active)':'transparent',
                }}
                onMouseEnter={e=>{if(s!==currentShell)e.currentTarget.style.background='var(--bg-hover)'}}
                onMouseLeave={e=>{if(s!==currentShell)e.currentTarget.style.background='transparent'}}
                >{s}</div>
              ))}
            </div>
          )}
        </div>
      </div>
      <div ref={termRef} style={{ flex:1, padding:'4px 0 0 4px' }}/>
    </div>
  )
}
