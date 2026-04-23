import { useState, useEffect } from 'react'
import { GitBranch, GitCommit, RefreshCw, Tag } from 'lucide-react'
import { GetGitGraph, GetGitBranches, GetGitInfo } from '../../wailsjs/go/main/App'

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

      {/* Branch list */}
      <div style={{
        padding: '8px 16px', borderBottom: '1px solid var(--border)',
        display: 'flex', flexWrap: 'wrap', gap: 6,
      }}>
        {branches.map(br => (
          <span key={br} style={{
            padding: '2px 8px', borderRadius: 10, fontSize: 11, fontFamily: 'var(--font-code)',
            background: br === currentBranch ? 'var(--accent-glow)' : 'var(--bg-active)',
            color: br === currentBranch ? 'var(--accent)' : 'var(--fg-muted)',
            border: br === currentBranch ? '1px solid var(--accent)' : '1px solid var(--border)',
          }}>
            {br}
          </span>
        ))}
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
