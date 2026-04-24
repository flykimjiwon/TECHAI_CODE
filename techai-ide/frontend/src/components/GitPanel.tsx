import { useState, useEffect } from 'react'
import { GitBranch, RefreshCw, FileCode, Plus, Minus, Check } from 'lucide-react'
import { GetGitInfo, GitStage, GitUnstage, GitCommit, GitDiffFile } from '../../wailsjs/go/main/App'

interface GitChange {
  status: string
  file: string
}

interface Props {
  onFileSelect: (path: string) => void
}

export default function GitPanel({ onFileSelect }: Props) {
  const [branch, setBranch] = useState('')
  const [changes, setChanges] = useState<GitChange[]>([])
  const [loading, setLoading] = useState(false)
  const [commitMsg, setCommitMsg] = useState('')
  const [diffContent, setDiffContent] = useState('')
  const [diffFile, setDiffFile] = useState('')

  function refresh() {
    setLoading(true)
    GetGitInfo().then(info => {
      setBranch(info.branch || 'unknown')
      setChanges(info.changes || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  useEffect(() => { refresh() }, [])

  async function handleStage(file: string) {
    await GitStage(file)
    refresh()
  }

  async function handleUnstage(file: string) {
    await GitUnstage(file)
    refresh()
  }

  async function handleCommit() {
    if (!commitMsg.trim()) return
    try {
      await GitCommit(commitMsg.trim())
      setCommitMsg('')
      refresh()
    } catch (err) {
      import('./Toast').then(m => m.showToast('Commit failed: ' + err, 'error'))
    }
  }

  async function showDiff(file: string) {
    if (diffFile === file) { setDiffFile(''); setDiffContent(''); return }
    const diff = await GitDiffFile(file)
    setDiffFile(file)
    setDiffContent(diff || 'No changes')
  }

  return (
    <aside style={{
      width: '100%', height: '100%', background: 'var(--bg-sidebar)',
      display: 'flex', flexDirection: 'column', overflow: 'hidden'
    }}>
      <div style={{
        padding: '10px 14px', fontSize: 11, fontWeight: 700, textTransform: 'uppercase',
        letterSpacing: '0.05em', color: 'var(--fg-muted)',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between'
      }}>
        Source Control
        <RefreshCw size={12} style={{ cursor: 'pointer', opacity: loading ? 0.3 : 1 }} onClick={refresh} />
      </div>

      <div style={{ padding: '4px 14px 8px', display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, color: 'var(--fg-secondary)' }}>
        <GitBranch size={14} style={{ color: 'var(--accent)' }} />
        <span style={{ fontWeight: 500 }}>{branch}</span>
        {changes.length > 0 && <span style={{ color: 'var(--warning)', fontSize: 10 }}>*</span>}
      </div>

      {/* Commit input */}
      {changes.length > 0 && (
        <div style={{ padding: '0 10px 8px', display: 'flex', gap: 4 }}>
          <input value={commitMsg} onChange={e => setCommitMsg(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleCommit() }}
            placeholder="Commit message..."
            style={{
              flex: 1, padding: '4px 8px', borderRadius: 4, fontSize: 11,
              background: 'var(--bg-active)', border: '1px solid var(--border)',
              color: 'var(--fg-primary)', fontFamily: 'var(--font-ui)', outline: 'none',
            }}
          />
          <button onClick={handleCommit} style={{
            background: 'var(--accent)', border: 'none', borderRadius: 4, cursor: 'pointer',
            padding: '4px 8px', color: '#fff', display: 'flex', alignItems: 'center',
          }}>
            <Check size={12} />
          </button>
        </div>
      )}

      <div style={{
        padding: '4px 14px 4px', fontSize: 10, fontWeight: 600,
        color: 'var(--fg-muted)', textTransform: 'uppercase'
      }}>Changes ({changes.length})</div>

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {changes.length === 0 && (
          <div style={{ padding: '8px 14px', fontSize: 12, color: 'var(--fg-dim)' }}>No changes</div>
        )}
        {changes.map((c, i) => (
          <div key={i}>
            <div style={{
              padding: '3px 14px', display: 'flex', alignItems: 'center', gap: 6,
              fontSize: 12, cursor: 'pointer', color: 'var(--fg-secondary)',
              background: diffFile === c.file ? 'var(--bg-active)' : 'transparent',
              transition: 'background 0.1s',
            }}
            onMouseEnter={e => { if (diffFile !== c.file) e.currentTarget.style.background = 'var(--bg-hover)' }}
            onMouseLeave={e => { if (diffFile !== c.file) e.currentTarget.style.background = 'transparent' }}
            onClick={() => showDiff(c.file)}
            >
              <FileCode size={14} style={{ color: '#61afef', flexShrink: 0 }} />
              <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                onClick={e => { e.stopPropagation(); onFileSelect(c.file) }}
              >{c.file}</span>
              <span style={{
                fontSize: 10, fontWeight: 700,
                color: c.status === 'A' || c.status === '??' ? 'var(--success)' :
                       c.status === 'M' ? 'var(--warning)' :
                       c.status === 'D' ? 'var(--error)' : 'var(--fg-muted)'
              }}>{c.status}</span>
              <button onClick={e => { e.stopPropagation(); handleStage(c.file) }} style={{
                background: 'none', border: 'none', cursor: 'pointer', color: 'var(--success)', padding: 2,
              }}><Plus size={12} /></button>
            </div>
            {/* Inline diff */}
            {diffFile === c.file && diffContent && (
              <pre style={{
                padding: '4px 14px 4px 30px', fontSize: 10, lineHeight: 1.5,
                fontFamily: 'var(--font-code)', color: 'var(--fg-muted)',
                maxHeight: 120, overflowY: 'auto', background: 'var(--bg-base)',
                borderBottom: '1px solid var(--border)',
              }}>
                {diffContent.split('\n').map((line, j) => (
                  <div key={j} style={{
                    color: line.startsWith('+') ? 'var(--success)' :
                           line.startsWith('-') ? 'var(--error)' :
                           line.startsWith('@@') ? 'var(--accent)' : 'var(--fg-dim)',
                  }}>{line}</div>
                ))}
              </pre>
            )}
          </div>
        ))}
      </div>
    </aside>
  )
}
