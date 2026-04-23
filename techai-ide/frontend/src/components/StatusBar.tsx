import { useState, useEffect } from 'react'
import { GitBranch, Cloud } from 'lucide-react'
import { GetGitInfo, GetModel, GetCwd } from '../../wailsjs/go/main/App'

interface Props {
  line?: number
  col?: number
  lang?: string
}

export default function StatusBar({ line, col, lang }: Props) {
  const [branch, setBranch] = useState('main')
  const [dirty, setDirty] = useState(false)
  const [model, setModel] = useState('')
  const [cwd, setCwd] = useState('')

  useEffect(() => {
    GetGitInfo().then(info => {
      setBranch(info.branch || 'unknown')
      setDirty(info.isDirty)
    }).catch(() => {})
    GetModel().then(setModel).catch(() => {})
    GetCwd().then(path => {
      const parts = path.split('/')
      setCwd(parts.length > 2 ? '~/' + parts.slice(-2).join('/') : path)
    }).catch(() => {})
  }, [])

  return (
    <footer style={{
      height: 24, background: 'var(--status-bg)', color: 'var(--status-fg, #fff)',
      display: 'flex', alignItems: 'center', padding: '0 12px',
      fontSize: 11, fontWeight: 500, justifyContent: 'space-between',
      fontFamily: 'var(--font-ui)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <GitBranch size={12} /> {branch}{dirty ? '*' : ''}
        </span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Cloud size={12} /> Ready
        </span>
        <span>{model}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        {line != null && <span>Ln {line}, Col {col}</span>}
        <span>{lang || 'Plain Text'}</span>
        <span>{cwd}</span>
        <span>UTF-8</span>
        <span>TECHAI IDE v0.1.0</span>
      </div>
    </footer>
  )
}
