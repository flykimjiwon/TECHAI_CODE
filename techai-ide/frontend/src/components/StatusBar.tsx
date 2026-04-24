import { useState, useEffect } from 'react'
import { GitBranch, Cloud, Check, AlertCircle } from 'lucide-react'
import { GetGitInfo, GetModel, GetCwd } from '../../wailsjs/go/main/App'

interface Props {
  line?: number
  col?: number
  lang?: string
  onGitClick?: () => void
  onSettingsClick?: () => void
}

export default function StatusBar({ line, col, lang, onGitClick, onSettingsClick }: Props) {
  const [branch, setBranch] = useState('main')
  const [dirty, setDirty] = useState(false)
  const [changes, setChanges] = useState(0)
  const [model, setModel] = useState('')
  const [cwd, setCwd] = useState('')

  useEffect(() => {
    const refresh = () => {
      GetGitInfo().then(info => {
        setBranch(info.branch || 'unknown')
        setDirty(info.isDirty)
        setChanges(info.changes?.length || 0)
      }).catch(() => {})
    }
    GetModel().then(setModel).catch(() => {})
    GetCwd().then(path => {
      const parts = path.split('/')
      setCwd(parts.length > 2 ? '~/' + parts.slice(-2).join('/') : path)
    }).catch(() => {})
    refresh()
    const interval = setInterval(refresh, 5000)
    return () => clearInterval(interval)
  }, [])

  const si: React.CSSProperties = {
    display: 'flex', alignItems: 'center', gap: 4, cursor: 'pointer',
    padding: '0 4px', borderRadius: 2, transition: 'background 0.1s',
  }

  return (
    <footer style={{
      height: 24, background: 'var(--status-bg)', color: 'var(--status-fg, #fff)',
      display: 'flex', alignItems: 'center', padding: '0 12px',
      fontSize: 11, fontWeight: 500, justifyContent: 'space-between',
      fontFamily: 'var(--font-ui)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={si} onClick={onGitClick}
          onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.15)')}
          onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
          <GitBranch size={12} /> {branch}{dirty ? '*' : ''}
          {changes > 0 && <span style={{ fontSize: 10 }}>({changes})</span>}
        </span>
        <span style={si}>
          {dirty ? <AlertCircle size={11} /> : <Check size={11} />}
          {dirty ? `${changes} changes` : 'Clean'}
        </span>
        <span style={si} onClick={onSettingsClick}
          onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.15)')}
          onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
          {model}
        </span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        {line != null && <span>Ln {line}, Col {col}</span>}
        <span>{lang || 'Plain Text'}</span>
        <span>{cwd}</span>
        <span>UTF-8</span>
        <span>v0.1.0</span>
      </div>
    </footer>
  )
}
