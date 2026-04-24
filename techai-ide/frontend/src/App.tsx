import { useState, useEffect, useCallback } from 'react'
import { EventsOn } from '../wailsjs/runtime/runtime'
import ActivityBar from './components/ActivityBar'
import FileTree from './components/FileTree'
import SearchPanel from './components/SearchPanel'
import GitPanel from './components/GitPanel'
import Editor from './components/Editor'
import Terminal from './components/Terminal'
import ChatPanel from './components/ChatPanel'
import StatusBar from './components/StatusBar'
import ThemePicker from './components/ThemePicker'
import SettingsPanel from './components/SettingsPanel'
import QuickOpen from './components/QuickOpen'
import ResizeHandle from './components/ResizeHandle'
import GitGraph from './components/GitGraph'
import DiffView from './components/DiffView'
import ToastContainer from './components/Toast'
import AboutDialog from './components/AboutDialog'
import CommandPalette from './components/CommandPalette'

function App() {
  const [activePanel, setActivePanel] = useState('files')
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [showTheme, setShowTheme] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [cursor, setCursor] = useState({ line: 1, col: 1, lang: 'Plain Text' })
  const [showQuickOpen, setShowQuickOpen] = useState(false)
  const [showAbout, setShowAbout] = useState(false)
  const [showCommandPalette, setShowCommandPalette] = useState(false)
  const [splitFile, setSplitFile] = useState<string | null>(null)
  const [diffFile, setDiffFile] = useState<string | null>(null)

  // Resizable panel sizes
  const [sidebarWidth, setSidebarWidth] = useState(240)
  const [terminalHeight, setTerminalHeight] = useState(200)
  const [chatWidth, setChatWidth] = useState(360)
  const [showTerminal, setShowTerminal] = useState(true)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)

  function handlePanelSelect(panel: string) {
    if (panel === 'settings') { setShowTheme(true); return }
    if (panel === 'account') { setShowSettings(true); return }
    setActivePanel(prev => prev === panel ? '' : panel)
  }

  const sidebarOpen = activePanel === 'files' || activePanel === 'search' || activePanel === 'git'

  const resizeSidebar = useCallback((d: number) => setSidebarWidth(w => Math.max(160, Math.min(400, w + d))), [])
  const resizeTerminal = useCallback((d: number) => setTerminalHeight(h => Math.max(80, Math.min(500, h - d))), [])
  const resizeChat = useCallback((d: number) => setChatWidth(w => Math.max(260, Math.min(600, w - d))), [])

  // Native menu events
  useEffect(() => {
    const cancels = [
      EventsOn('menu:settings', () => setShowSettings(true)),
      EventsOn('menu:theme', () => setShowTheme(true)),
      EventsOn('menu:quickopen', () => setShowQuickOpen(true)),
      EventsOn('menu:terminal', () => setShowTerminal(t => !t)),
      EventsOn('menu:panel', (panel: string) => handlePanelSelect(panel)),
      EventsOn('menu:findinfiles', () => setActivePanel('search')),
      EventsOn('menu:about', () => setShowAbout(true)),
      EventsOn('preview:open', (url: string) => setPreviewUrl(url)),
      EventsOn('menu:openfolder', () => {
        import('../wailsjs/go/main/App').then(({ OpenFolder }) => {
          OpenFolder().then(dir => { if (dir) window.location.reload() })
        })
      }),
    ]
    return () => cancels.forEach(fn => fn())
  }, [])

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey
      if (mod && e.key === '1') { e.preventDefault(); handlePanelSelect('files') }
      if (mod && e.key === '2') { e.preventDefault(); handlePanelSelect('search') }
      if (mod && e.key === '3') { e.preventDefault(); handlePanelSelect('git') }
      if (mod && e.key === ',') { e.preventDefault(); setShowTheme(true) }
      if (mod && e.shiftKey && e.key === 'f') { e.preventDefault(); setActivePanel('search') }
      if (mod && e.key === 'b') { e.preventDefault(); handlePanelSelect('files') }
      if (mod && e.key === 'p') { e.preventDefault(); setShowQuickOpen(true) }
      if (mod && e.key === '`') { e.preventDefault(); setShowTerminal(t => !t) }
      if (mod && e.key === 'j') { e.preventDefault(); setShowTerminal(t => !t) }
      if (mod && e.shiftKey && e.key === 'p') { e.preventDefault(); setShowCommandPalette(true) }
      if (mod && e.key === '\\') { e.preventDefault(); setSplitFile(prev => prev ? null : selectedFile) }
      if (mod && (e.key === '=' || e.key === '+')) { e.preventDefault(); document.documentElement.style.fontSize = (parseFloat(getComputedStyle(document.documentElement).fontSize) + 1) + 'px' }
      if (mod && e.key === '-') { e.preventDefault(); document.documentElement.style.fontSize = Math.max(10, parseFloat(getComputedStyle(document.documentElement).fontSize) - 1) + 'px' }
      if (mod && e.key === '0') { e.preventDefault(); document.documentElement.style.fontSize = '13px' }
      if (e.key === 'F11') { e.preventDefault(); document.fullscreenElement ? document.exitFullscreen() : document.documentElement.requestFullscreen() }
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [])

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* macOS traffic light spacing — hide on Windows */}
      {navigator.platform.includes('Mac') && (
        <div style={{ height: 28, background: 'var(--bg-activity)', borderBottom: '1px solid var(--border)' }} />
      )}

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <ActivityBar active={activePanel} onSelect={handlePanelSelect} />

        {/* Sidebar */}
        {sidebarOpen && (
          <>
            <div style={{ width: sidebarWidth, flexShrink: 0, overflow: 'hidden' }}>
              {activePanel === 'files' && <FileTree onFileSelect={setSelectedFile} selectedFile={selectedFile || ''} />}
              {activePanel === 'search' && <SearchPanel onFileSelect={setSelectedFile} />}
              {activePanel === 'git' && <GitPanel onFileSelect={(path) => setDiffFile(path)} />}
            </div>
            <ResizeHandle direction="horizontal" onResize={resizeSidebar} />
          </>
        )}

        {/* Center: Editor + Terminal */}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, overflow: 'hidden' }}
          onDragOver={e => e.preventDefault()}
          onDrop={e => {
            const path = e.dataTransfer.getData('text/plain')
            if (path) { e.preventDefault(); setSelectedFile(path) }
          }}
        >
          <div style={{ flex: 1, minHeight: 0, overflow: 'hidden', display: 'flex' }}>
            {diffFile ? (
              <DiffView filePath={diffFile} onClose={() => setDiffFile(null)} />
            ) : activePanel === 'git' ? (
              <GitGraph />
            ) : (
              <>
                <div style={{ flex: 1, overflow: 'hidden' }}>
                  <Editor filePath={selectedFile} onCursorChange={(l, c, lang) => setCursor({ line: l, col: c, lang })}
                    onAskAI={(prompt) => { import('../wailsjs/go/main/App').then(m => m.SendMessage(prompt)) }} />
                </div>
                {splitFile && !previewUrl && (
                  <>
                    <div style={{ width: 3, background: 'transparent', borderLeft: '1px solid var(--border)', cursor: 'col-resize' }} />
                    <div style={{ flex: 1, overflow: 'hidden' }}>
                      <Editor filePath={splitFile} onCursorChange={(l, c, lang) => setCursor({ line: l, col: c, lang })}
                        onAskAI={(prompt) => { import('../wailsjs/go/main/App').then(m => m.SendMessage(prompt)) }} />
                    </div>
                  </>
                )}
                {previewUrl && (
                  <>
                    <div style={{ width: 3, background: 'transparent', borderLeft: '1px solid var(--border)' }} />
                    <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
                      <div style={{
                        height: 32, background: 'var(--bg-activity)', display: 'flex', alignItems: 'center',
                        padding: '0 12px', gap: 8, borderBottom: '1px solid var(--border)', fontSize: 12,
                      }}>
                        <span style={{ color: 'var(--success)', fontSize: 10 }}>●</span>
                        <input value={previewUrl} onChange={e => setPreviewUrl(e.target.value)}
                          onKeyDown={e => {
                            if (e.key === 'Enter') {
                              const f = document.getElementById('preview-frame') as HTMLIFrameElement
                              if (f) f.src = previewUrl || ''
                            }
                          }}
                          style={{
                            flex: 1, background: 'var(--bg-base)', border: '1px solid var(--border)',
                            borderRadius: 4, padding: '2px 8px', color: 'var(--fg-primary)',
                            fontFamily: 'var(--font-code)', fontSize: 11, outline: 'none',
                          }}
                        />
                        <button title="Reload" onClick={() => {
                          const f = document.getElementById('preview-frame') as HTMLIFrameElement
                          if (f) f.src = previewUrl || ''
                        }} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-dim)', fontSize: 13 }}>↻</button>
                        <button title="Close preview" onClick={() => setPreviewUrl(null)} style={{
                          background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-dim)', fontSize: 14
                        }}>×</button>
                      </div>
                      <iframe id="preview-frame" src={previewUrl} style={{
                        flex: 1, border: 'none', background: '#fff', width: '100%',
                      }} />
                    </div>
                  </>
                )}
              </>
            )}
          </div>
          {/* Terminal toggle bar */}
          <div onClick={() => setShowTerminal(t => !t)} style={{
            height: 22, background: 'var(--bg-activity)', borderTop: '1px solid var(--border)',
            display: 'flex', alignItems: 'center', padding: '0 12px', cursor: 'pointer',
            fontSize: 10, color: 'var(--fg-muted)', gap: 6, flexShrink: 0,
          }}>
            <span style={{ fontSize: 8 }}>{showTerminal ? '▼' : '▶'}</span>
            Terminal
            <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-code)', fontSize: 9, color: 'var(--fg-dim)' }}>⌘J</span>
          </div>
          {showTerminal && (
            <>
              <ResizeHandle direction="vertical" onResize={resizeTerminal} />
              <div style={{ height: terminalHeight, flexShrink: 0, overflow: 'hidden' }}>
                <Terminal />
              </div>
            </>
          )}
        </div>

        {/* Chat resize handle + panel */}
        <ResizeHandle direction="horizontal" onResize={resizeChat} />
        <div style={{
          width: chatWidth, flexShrink: 0,
          background: 'var(--bg-panel)',
          display: 'flex', flexDirection: 'column',
          boxShadow: '-4px 0 16px rgba(0,0,0,0.1)',
        }}>
          <ChatPanel />
        </div>
      </div>

      <StatusBar line={cursor.line} col={cursor.col} lang={cursor.lang}
        onGitClick={() => setActivePanel('git')}
        onSettingsClick={() => setShowSettings(true)} />

      <ThemePicker open={showTheme} onClose={() => setShowTheme(false)} />
      <SettingsPanel open={showSettings} onClose={() => setShowSettings(false)} />
      <QuickOpen open={showQuickOpen} onClose={() => setShowQuickOpen(false)} onSelect={setSelectedFile} />
      <ToastContainer />
      <AboutDialog open={showAbout} onClose={() => setShowAbout(false)} />
      <CommandPalette
        open={showCommandPalette}
        onClose={() => setShowCommandPalette(false)}
        actions={{
          openFiles: () => setActivePanel('files'),
          openSearch: () => setActivePanel('search'),
          openGit: () => setActivePanel('git'),
          openSettings: () => setShowSettings(true),
          openTheme: () => setShowTheme(true),
          toggleTerminal: () => setShowTerminal(t => !t),
          openQuickOpen: () => setShowQuickOpen(true),
          clearChat: () => { import('../wailsjs/go/main/App').then(m => m.ClearChat()) },
          exportChat: () => { import('../wailsjs/go/main/App').then(m => m.ExportChat()) },
          openFolder: () => { import('../wailsjs/go/main/App').then(m => m.OpenFolder().then(d => { if(d) window.location.reload() })) },
          toggleSplit: () => setSplitFile(prev => prev ? null : selectedFile),
          openPreview: (url: string) => setPreviewUrl(url),
        }}
      />
    </div>
  )
}

export default App
