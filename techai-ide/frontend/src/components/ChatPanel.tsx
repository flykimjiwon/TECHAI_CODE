import { useState, useRef, useEffect } from 'react'
import { Sparkles, Send, Trash2, BookOpen, Download } from 'lucide-react'
import { SendMessage, ClearChat, GetModel, GetKnowledgePacks, ToggleKnowledgePack, ExportChat, SaveSession, ListSessions, LoadSession } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface Message {
  role: 'user' | 'ai' | 'tool'
  content: string
  streaming?: boolean
}

export default function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [model, setModel] = useState('')
  const [showPacks, setShowPacks] = useState(false)
  const [packs, setPacks] = useState<{ id: string; name: string; category: string; enabled: boolean }[]>([])
  const chatRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    GetModel().then(setModel).catch(() => {})
    GetKnowledgePacks().then(setPacks).catch(() => {})

    // Listen to Wails events from Go backend.
    // EventsOn returns a cancel function — collect them for cleanup.
    const cancels: (() => void)[] = []

    cancels.push(EventsOn('chat:stream_start', () => {
      setStreaming(true)
      setMessages(prev => [...prev, { role: 'ai', content: '', streaming: true }])
    }))

    cancels.push(EventsOn('chat:chunk', (chunk: string) => {
      setMessages(prev => {
        const updated = [...prev]
        const last = updated[updated.length - 1]
        if (last && last.role === 'ai' && last.streaming) {
          updated[updated.length - 1] = { ...last, content: last.content + chunk }
        }
        return updated
      })
    }))

    cancels.push(EventsOn('chat:stream_done', () => {
      setStreaming(false)
      setMessages(prev => {
        const updated = [...prev]
        const last = updated[updated.length - 1]
        if (last && last.streaming) {
          updated[updated.length - 1] = { ...last, streaming: false }
        }
        return updated
      })
    }))

    cancels.push(EventsOn('chat:tool_start', (data: { name: string; args: string }) => {
      setMessages(prev => [...prev, { role: 'tool', content: `>> ${data.name} ${truncateArgs(data.args)}` }])
    }))

    cancels.push(EventsOn('chat:tool_done', (data: { name: string; result: string }) => {
      setMessages(prev => [...prev, { role: 'tool', content: `<< ${data.name}: ${data.result}` }])
    }))

    cancels.push(EventsOn('chat:error', (err: string) => {
      setStreaming(false)
      setMessages(prev => [...prev, { role: 'tool', content: `Error: ${err}` }])
    }))

    cancels.push(EventsOn('chat:cleared', () => {
      setMessages([])
    }))

    // Cleanup on unmount — prevents duplicate listeners
    return () => { cancels.forEach(fn => fn()) }
  }, [])

  useEffect(() => {
    if (chatRef.current) chatRef.current.scrollTop = chatRef.current.scrollHeight
  }, [messages])

  function send() {
    const text = input.trim()
    if (!text || streaming) return

    // Handle slash commands
    if (text.startsWith('/')) {
      handleSlashCommand(text)
      setInput('')
      return
    }

    setMessages(prev => [...prev, { role: 'user', content: text }])
    setInput('')
    SendMessage(text)
  }

  function handleSlashCommand(cmd: string) {
    const parts = cmd.split(' ')
    const command = parts[0].toLowerCase()
    switch (command) {
      case '/clear':
        ClearChat()
        break
      case '/export':
        ExportChat().then(p => {
          setMessages(prev => [...prev, { role: 'tool', content: `Exported to ${p}` }])
        }).catch(() => {})
        break
      case '/model':
        GetModel().then(m => {
          setMessages(prev => [...prev, { role: 'tool', content: `Current model: ${m}` }])
        })
        break
      case '/save':
        SaveSession(parts.slice(1).join(' ')).then(id => {
          setMessages(prev => [...prev, { role: 'tool', content: `Session saved: ${id}` }])
        }).catch(e => setMessages(prev => [...prev, { role: 'tool', content: `Error: ${e}` }]))
        break
      case '/sessions':
        ListSessions().then(sessions => {
          const list = sessions.length > 0
            ? sessions.map(s => `${s.id} — ${s.title} (${s.messages} msgs)`).join('\n')
            : 'No saved sessions'
          setMessages(prev => [...prev, { role: 'tool', content: list }])
        })
        break
      case '/load':
        if (parts[1]) {
          LoadSession(parts[1]).then(() => {
            setMessages(prev => [...prev, { role: 'tool', content: `Loaded session: ${parts[1]}` }])
          }).catch(e => setMessages(prev => [...prev, { role: 'tool', content: `Error: ${e}` }]))
        }
        break
      case '/help':
        setMessages(prev => [...prev, {
          role: 'tool',
          content: 'Commands: /clear /export /save [title] /sessions /load [id] /model /help'
        }])
        break
      default:
        setMessages(prev => [...prev, {
          role: 'tool',
          content: `Unknown command: ${command}. Try /help`
        }])
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      send()
    }
  }

  function handleClear() {
    ClearChat()
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
      {/* Header */}
      <div style={{
        padding: '12px 16px', borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'center', gap: 8
      }}>
        <Sparkles size={18} style={{ color: 'var(--accent)' }} />
        <span style={{ fontWeight: 600, fontSize: 14 }}>TECHAI</span>
        <span style={{
          background: 'var(--accent-glow)', color: 'var(--accent)',
          padding: '1px 7px', borderRadius: 4, fontSize: 9, fontWeight: 700,
          textTransform: 'uppercase', letterSpacing: '0.03em'
        }}>{model || 'Loading...'}</span>
        <button onClick={() => setShowPacks(p => !p)} style={{
          background: showPacks ? 'var(--accent-glow)' : 'none', border: 'none', cursor: 'pointer',
          color: showPacks ? 'var(--accent)' : 'var(--fg-dim)', padding: 4, borderRadius: 4,
          marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 3, fontSize: 10,
        }}>
          <BookOpen size={13} />
          <span>{packs.filter(p => p.enabled).length}</span>
        </button>
        <button onClick={() => ExportChat().then(p => alert('Exported: ' + p)).catch(() => {})} style={{
          background: 'none', border: 'none', cursor: 'pointer',
          color: 'var(--fg-dim)', padding: 4,
        }}>
          <Download size={13} />
        </button>
        <button onClick={handleClear} style={{
          background: 'none', border: 'none', cursor: 'pointer',
          color: 'var(--fg-dim)', padding: 4
        }}>
          <Trash2 size={14} />
        </button>
      </div>

      {/* Messages */}
      <div ref={chatRef} style={{
        flex: 1, overflowY: 'auto', padding: 14,
        display: 'flex', flexDirection: 'column', gap: 8
      }}>
        {messages.length === 0 && (
          <div style={{ color: 'var(--fg-dim)', fontSize: 13, textAlign: 'center', marginTop: 40 }}>
            Ask TECHAI anything about your code
          </div>
        )}
        {messages.map((msg, i) => {
          if (msg.role === 'tool') {
            const isCall = msg.content.startsWith('>>')
            const isResult = msg.content.startsWith('<<')
            return (
              <div key={i} style={{
                fontFamily: 'var(--font-code)', fontSize: 11, padding: '3px 10px',
                color: isCall ? 'var(--accent)' : isResult ? 'var(--success)' : 'var(--fg-muted)',
                opacity: 0.8, display: 'flex', alignItems: 'center', gap: 4,
                borderLeft: isCall ? '2px solid var(--accent)' : isResult ? '2px solid var(--success)' : '2px solid var(--fg-dim)',
                marginLeft: 8,
                background: 'var(--bg-hover)', borderRadius: '0 4px 4px 0',
              }}>
                {msg.content}
              </div>
            )
          }
          return (
            <div key={i} style={{
              maxWidth: '88%', padding: '10px 13px', borderRadius: 12,
              fontSize: 13, lineHeight: 1.55, whiteSpace: 'pre-wrap', wordBreak: 'break-word',
              alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
              background: msg.role === 'user' ? 'var(--bubble-user-bg)' : 'var(--bubble-ai-bg)',
              border: msg.role === 'ai' ? '1px solid var(--bubble-ai-border)' : 'none',
              borderBottomRightRadius: msg.role === 'user' ? 3 : 12,
              borderBottomLeftRadius: msg.role === 'ai' ? 3 : 12,
              color: msg.role === 'ai' ? 'var(--fg-secondary)' : 'var(--fg-primary)',
            }}>
              {msg.streaming ? msg.content : renderContent(msg.content)}
              {msg.streaming && (
                <span style={{
                  display: 'inline-block', width: 2, height: 14,
                  background: 'var(--accent)', animation: 'blink 1s step-end infinite',
                  verticalAlign: 'text-bottom', marginLeft: 1
                }} />
              )}
            </div>
          )
        })}
      </div>

      {/* Knowledge Packs */}
      {showPacks && (
        <div style={{
          borderTop: '1px solid var(--border)', maxHeight: 180, overflowY: 'auto',
          padding: '6px 10px', background: 'var(--bg-base)',
        }}>
          <div style={{ fontSize: 10, fontWeight: 700, color: 'var(--fg-muted)', textTransform: 'uppercase', marginBottom: 4 }}>
            Knowledge Packs
          </div>
          {(() => {
            const categories = [...new Set(packs.map(p => p.category))]
            return categories.map(cat => (
              <div key={cat} style={{ marginBottom: 4 }}>
                <div style={{ fontSize: 10, color: 'var(--fg-dim)', fontWeight: 600, marginBottom: 2 }}>{cat}</div>
                {packs.filter(p => p.category === cat).map(pack => (
                  <label key={pack.id} style={{
                    display: 'flex', alignItems: 'center', gap: 6, padding: '2px 4px',
                    fontSize: 11, color: 'var(--fg-secondary)', cursor: 'pointer',
                  }}>
                    <input type="checkbox" checked={pack.enabled}
                      onChange={e => {
                        const enabled = e.target.checked
                        ToggleKnowledgePack(pack.id, enabled)
                        setPacks(prev => prev.map(p => p.id === pack.id ? { ...p, enabled } : p))
                      }}
                      style={{ accentColor: 'var(--accent)' }}
                    />
                    {pack.name}
                  </label>
                ))}
              </div>
            ))
          })()}
        </div>
      )}

      {/* Input */}
      <div style={{ padding: '12px 14px', borderTop: '1px solid var(--border)' }}>
        <div style={{
          background: 'var(--bg-input)', border: '1px solid var(--border)',
          borderRadius: 8, padding: '8px 12px', transition: 'all 0.2s',
        }}>
          <textarea
            value={input} onChange={e => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={streaming ? 'AI responding...' : 'Ask TECHAI...'}
            disabled={streaming}
            rows={2}
            style={{
              width: '100%', background: 'transparent', border: 'none',
              color: 'var(--fg-primary)', resize: 'none',
              fontFamily: 'var(--font-ui)', fontSize: 12.5, outline: 'none',
              opacity: streaming ? 0.5 : 1,
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 6 }}>
            <span style={{ fontSize: 10, color: 'var(--fg-dim)' }}>Enter send / Shift+Enter newline</span>
            <Send size={16} style={{
              color: streaming ? 'var(--fg-dim)' : 'var(--accent)',
              cursor: streaming ? 'default' : 'pointer'
            }} onClick={send} />
          </div>
        </div>
      </div>
    </div>
  )
}

// Simple markdown renderer — handles code blocks and inline code.
function renderContent(text: string) {
  const parts: JSX.Element[] = []
  let key = 0

  // Split by code blocks ```...```
  const blocks = text.split(/(```[\s\S]*?```)/g)
  for (const block of blocks) {
    if (block.startsWith('```')) {
      const lines = block.slice(3, -3).split('\n')
      const lang = lines[0]?.trim() || ''
      const code = (lang ? lines.slice(1) : lines).join('\n')
      parts.push(
        <pre key={key++} style={{
          background: 'var(--bg-base)', padding: '8px 10px', borderRadius: 6,
          margin: '6px 0', overflow: 'auto', border: '1px solid var(--border)',
          fontFamily: 'var(--font-code)', fontSize: 11.5, lineHeight: 1.5,
          position: 'relative',
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            {lang && <span style={{ fontSize: 10, color: 'var(--fg-dim)' }}>{lang}</span>}
            <span style={{ display: 'flex', gap: 4, marginLeft: 'auto' }}>
              <button onClick={() => navigator.clipboard.writeText(code)} style={{
                background: 'var(--bg-active)', border: '1px solid var(--border)', borderRadius: 4,
                color: 'var(--fg-muted)', padding: '1px 6px', fontSize: 10, cursor: 'pointer',
                fontFamily: 'var(--font-ui)',
              }}>Copy</button>
            </span>
          </div>
          {code}
        </pre>
      )
    } else {
      // Handle inline code `...`
      const inlineParts = block.split(/(`[^`]+`)/g)
      const spans = inlineParts.map((part, i) => {
        if (part.startsWith('`') && part.endsWith('`')) {
          return <code key={i} style={{
            fontFamily: 'var(--font-code)', fontSize: 11.5,
            background: 'var(--bg-active)', padding: '1px 5px', borderRadius: 3,
            color: 'var(--accent)',
          }}>{part.slice(1, -1)}</code>
        }
        return <span key={i}>{part}</span>
      })
      parts.push(<span key={key++}>{spans}</span>)
    }
  }
  return <>{parts}</>
}

function truncateArgs(args: string): string {
  try {
    const parsed = JSON.parse(args)
    const keys = Object.keys(parsed)
    if (keys.length === 1) return String(parsed[keys[0]])
    return keys.map(k => `${k}=${parsed[k]}`).join(' ').slice(0, 60)
  } catch {
    return args.slice(0, 60)
  }
}
