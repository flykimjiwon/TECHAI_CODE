import { useState, useEffect } from 'react'
import { GitBranch, GitCommit, RefreshCw, Tag, Plus, ArrowDown, ArrowUp } from 'lucide-react'
import { GetGitGraph, GetGitBranches, GetGitInfo, GitCheckout, GitCreateBranch, GitPull, GitPush } from '../../wailsjs/go/main/App'
import { showToast } from './Toast'

interface GitLogEntry {
  hash: string
  short: string
  author: string
  date: string
  message: string
  refs: string
}

export default function GitGraph() {
  const [commits, setCommits] = useState<GitLogEntry[]>([])
  const [branches, setBranches] = useState<string[]>([])
  const [currentBranch, setCurrentBranch] = useState('')
  const [loading, setLoading] = useState(true)
  const [newBranch, setNewBranch] = useState('')
  const [showNewBranch, setShowNewBranch] = useState(false)

  function refresh() {
    setLoading(true)
    Promise.all([
      GetGitGraph(50),
      GetGitBranches(),
      GetGitInfo(),
    ]).then(([graph, br, info]) => {
      setCommits(graph || [])
      setBranches(br || [])
      setCurrentBranch(info.branch || '')
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  useEffect(() => { refresh() }, [])

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden', background: 'var(--bg-editor)' }}>
      {/* Header */}
      <div style={{
        padding: '10px 16px', borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'center', gap: 8, background: 'var(--bg-activity)',
      }}>
        <GitBranch size={16} style={{ color: 'var(--accent)' }} />
        <span style={{ fontWeight: 600, fontSize: 13 }}>Git Graph</span>
        <span style={{ fontSize: 11, color: 'var(--fg-muted)' }}>— {currentBranch}</span>
        <button onClick={refresh} style={{
          marginLeft: 'auto', background: 'none', border: 'none', cursor: 'pointer',
          color: 'var(--fg-dim)', padding: 4,
        }}>
          <RefreshCw size={13} style={{ opacity: loading ? 0.3 : 1 }} />
        </button>
      </div>

      {/* Branch list + actions */}
      <div style={{
        padding: '8px 16px', borderBottom: '1px solid var(--border)',
        display: 'flex', flexWrap: 'wrap', gap: 6, alignItems: 'center',
      }}>
        {branches.map(br => (
          <span key={br} onClick={async () => {
            if (br === currentBranch) return
            try {
              await GitCheckout(br)
              showToast(`Switched to ${br}`, 'success')
              refresh()
            } catch (e) { showToast(`Checkout failed: ${e}`, 'error') }
          }} style={{
            padding: '2px 8px', borderRadius: 10, fontSize: 11, fontFamily: 'var(--font-code)',
            background: br === currentBranch ? 'var(--accent-glow)' : 'var(--bg-active)',
            color: br === currentBranch ? 'var(--accent)' : 'var(--fg-muted)',
            border: br === currentBranch ? '1px solid var(--accent)' : '1px solid var(--border)',
            cursor: br === currentBranch ? 'default' : 'pointer',
          }}>
            {br}
          </span>
        ))}
        {/* New branch */}
        {showNewBranch ? (
          <input autoFocus value={newBranch} onChange={e => setNewBranch(e.target.value)}
            placeholder="branch name"
            onKeyDown={async e => {
              if (e.key === 'Enter' && newBranch.trim()) {
                try { await GitCreateBranch(newBranch.trim()); showToast(`Created: ${newBranch}`, 'success'); refresh() }
                catch (err) { showToast(`Error: ${err}`, 'error') }
                setShowNewBranch(false); setNewBranch('')
              }
              if (e.key === 'Escape') { setShowNewBranch(false); setNewBranch('') }
            }}
            style={{
              padding: '2px 8px', borderRadius: 10, fontSize: 11, fontFamily: 'var(--font-code)',
              background: 'var(--bg-base)', border: '1px solid var(--accent)', outline: 'none',
              color: 'var(--fg-primary)', width: 120,
            }}
          />
        ) : (
          <button onClick={() => setShowNewBranch(true)} style={{
            background: 'var(--bg-active)', border: '1px solid var(--border)', borderRadius: 10,
            color: 'var(--fg-muted)', padding: '2px 8px', fontSize: 11, cursor: 'pointer',
            display: 'flex', alignItems: 'center', gap: 3,
          }}>
            <Plus size={10} /> branch
          </button>
        )}
        {/* Pull/Push */}
        <div style={{ marginLeft: 'auto', display: 'flex', gap: 4 }}>
          <button onClick={async () => {
            try { const r = await GitPull(); showToast('Pulled: ' + (r || 'up to date'), 'success') }
            catch (e) { showToast(`Pull failed: ${e}`, 'error') }
            refresh()
          }} style={gitActionBtn}>
            <ArrowDown size={11} /> Pull
          </button>
          <button onClick={async () => {
            try { const r = await GitPush(); showToast('Pushed: ' + (r || 'done'), 'success') }
            catch (e) { showToast(`Push failed: ${e}`, 'error') }
          }} style={gitActionBtn}>
            <ArrowUp size={11} /> Push
          </button>
        </div>
      </div>

      {/* Commit list */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '4px 0' }}>
        {commits.map((c, i) => (
          <div key={c.hash || i} style={{
            padding: '6px 16px', display: 'flex', alignItems: 'flex-start', gap: 10,
            borderBottom: '1px solid var(--border)',
            fontSize: 12,
          }}
          onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-hover)')}
          onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
          >
            {/* Graph dot */}
            <div style={{
              width: 20, display: 'flex', flexDirection: 'column', alignItems: 'center',
              paddingTop: 2, flexShrink: 0,
            }}>
              <GitCommit size={14} style={{ color: c.refs ? 'var(--accent)' : 'var(--fg-dim)' }} />
              {i < commits.length - 1 && (
                <div style={{ width: 1, flex: 1, background: 'var(--border)', minHeight: 10 }} />
              )}
            </div>

            {/* Content */}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                {/* Refs (branches, tags) */}
                {c.refs && c.refs.split(',').map(ref => {
                  const r = ref.trim()
                  if (!r) return null
                  const isHead = r.includes('HEAD')
                  const isTag = r.includes('tag:')
                  return (
                    <span key={r} style={{
                      padding: '1px 6px', borderRadius: 8, fontSize: 10, fontWeight: 600,
                      fontFamily: 'var(--font-code)',
                      background: isHead ? 'var(--accent-glow)' : isTag ? 'rgba(245,158,11,0.15)' : 'var(--bg-active)',
                      color: isHead ? 'var(--accent)' : isTag ? 'var(--warning)' : 'var(--fg-muted)',
                      display: 'inline-flex', alignItems: 'center', gap: 3,
                    }}>
                      {isTag ? <Tag size={9} /> : <GitBranch size={9} />}
                      {r.replace('HEAD -> ', '').replace('tag: ', '')}
                    </span>
                  )
                })}
                <span style={{ color: 'var(--fg-primary)', fontWeight: 500 }}>{c.message}</span>
              </div>
              <div style={{ display: 'flex', gap: 12, marginTop: 3, color: 'var(--fg-dim)', fontSize: 11 }}>
                <span style={{ fontFamily: 'var(--font-code)', color: 'var(--fg-muted)' }}>{c.short}</span>
                <span>{c.author}</span>
                <span>{c.date}</span>
              </div>
            </div>
          </div>
        ))}
        {commits.length === 0 && !loading && (
          <div style={{ padding: 20, textAlign: 'center', color: 'var(--fg-dim)' }}>No commits found</div>
        )}
      </div>
    </div>
  )
}

const gitActionBtn: React.CSSProperties = {
  background: 'var(--bg-active)', border: '1px solid var(--border)', borderRadius: 6,
  color: 'var(--fg-muted)', padding: '3px 8px', fontSize: 11, cursor: 'pointer',
  display: 'flex', alignItems: 'center', gap: 3, fontFamily: 'var(--font-ui)',
}
