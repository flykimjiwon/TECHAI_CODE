import { useState, useEffect } from 'react'
import { GitBranch, X } from 'lucide-react'
import { GitDiffFile } from '../../wailsjs/go/main/App'

interface Props {
  filePath: string
  onClose: () => void
}

export default function DiffView({ filePath, onClose }: Props) {
  const [diff, setDiff] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    GitDiffFile(filePath).then(d => {
      setDiff(d || 'No changes')
      setLoading(false)
    }).catch(() => { setDiff('Error loading diff'); setLoading(false) })
  }, [filePath])

  const lines = diff.split('\n')
  const fileName = filePath.split('/').pop()

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden', background: 'var(--bg-editor)' }}>
      {/* Header */}
      <div style={{
        padding: '8px 14px', borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'center', gap: 8, background: 'var(--bg-activity)',
        fontSize: 12,
      }}>
        <GitBranch size={14} style={{ color: 'var(--accent)' }} />
        <span style={{ fontWeight: 600 }}>Diff: {fileName}</span>
        <span style={{ color: 'var(--fg-dim)', fontSize: 11 }}>
          +{lines.filter(l => l.startsWith('+')).length - 1} / -{lines.filter(l => l.startsWith('-')).length - 1}
        </span>
        <button onClick={onClose} style={{
          marginLeft: 'auto', background: 'none', border: 'none', cursor: 'pointer',
          color: 'var(--fg-dim)', padding: 4,
        }}>
          <X size={14} />
        </button>
      </div>

      {/* Diff content */}
      <div style={{ flex: 1, overflow: 'auto', fontFamily: 'var(--font-code)', fontSize: 12, lineHeight: 1.6 }}>
        {loading ? (
          <div style={{ padding: 20, color: 'var(--fg-dim)' }}>Loading diff...</div>
        ) : (
          lines.map((line, i) => {
            let bg = 'transparent'
            let color = 'var(--fg-secondary)'
            let prefix = ' '

            if (line.startsWith('+++') || line.startsWith('---')) {
              color = 'var(--fg-muted)'
              bg = 'rgba(255,255,255,0.02)'
            } else if (line.startsWith('@@')) {
              color = 'var(--accent)'
              bg = 'rgba(59,130,246,0.06)'
            } else if (line.startsWith('+')) {
              color = 'var(--success)'
              bg = 'rgba(16,185,129,0.08)'
              prefix = '+'
            } else if (line.startsWith('-')) {
              color = 'var(--error)'
              bg = 'rgba(239,68,68,0.08)'
              prefix = '-'
            }

            return (
              <div key={i} style={{
                padding: '0 14px', background: bg, color,
                display: 'flex', minHeight: 20,
              }}>
                <span style={{ width: 40, textAlign: 'right', paddingRight: 10, color: 'var(--fg-dim)', userSelect: 'none', flexShrink: 0 }}>
                  {i + 1}
                </span>
                <span style={{ width: 14, textAlign: 'center', flexShrink: 0, fontWeight: 600 }}>
                  {prefix !== ' ' ? prefix : ''}
                </span>
                <span style={{ whiteSpace: 'pre', flex: 1 }}>{line.slice(1) || line}</span>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
