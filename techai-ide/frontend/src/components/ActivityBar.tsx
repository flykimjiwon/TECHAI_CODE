import { useState, useEffect } from 'react'
import { Files, Search, GitBranch, User, Settings } from 'lucide-react'
import { GetGitInfo } from '../../wailsjs/go/main/App'

interface Props {
  active: string
  onSelect: (panel: string) => void
}

export default function ActivityBar({ active, onSelect }: Props) {
  const [gitCount, setGitCount] = useState(0)

  useEffect(() => {
    GetGitInfo().then(info => setGitCount(info.changes?.length || 0)).catch(() => {})
    const interval = setInterval(() => {
      GetGitInfo().then(info => setGitCount(info.changes?.length || 0)).catch(() => {})
    }, 5000) // refresh every 5s
    return () => clearInterval(interval)
  }, [])

  const top = [
    { id: 'files', icon: Files, label: 'Explorer', badge: 0 },
    { id: 'search', icon: Search, label: 'Search', badge: 0 },
    { id: 'git', icon: GitBranch, label: 'Git', badge: gitCount },
  ]
  const bottom = [
    { id: 'account', icon: User, label: 'Account', badge: 0 },
    { id: 'settings', icon: Settings, label: 'Settings', badge: 0 },
  ]

  const renderIcon = (item: typeof top[0], isBottom = false) => (
    <button
      key={item.id}
      title={item.label}
      onClick={() => onSelect(item.id)}
      style={{
        background: 'none', border: 'none', cursor: 'pointer', padding: 10,
        borderRadius: 8,
        color: active === item.id && !isBottom ? 'var(--accent)' : 'var(--fg-muted)',
        position: 'relative', transition: 'all 0.15s',
      }}
      onMouseEnter={e => (e.currentTarget.style.color = 'var(--fg-primary)', e.currentTarget.style.background = 'var(--bg-hover)')}
      onMouseLeave={e => (
        e.currentTarget.style.color = active === item.id && !isBottom ? 'var(--accent)' : 'var(--fg-muted)',
        e.currentTarget.style.background = 'none'
      )}
    >
      {active === item.id && !isBottom && <span style={{
        position: 'absolute', left: 0, top: 10, bottom: 10, width: 2,
        background: 'var(--accent)', borderRadius: '0 2px 2px 0'
      }} />}
      <item.icon size={20} />
      {item.badge > 0 && (
        <span style={{
          position: 'absolute', top: 4, right: 4,
          background: 'var(--accent)', color: '#fff',
          fontSize: 9, fontWeight: 700, minWidth: 16, height: 16,
          borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontFamily: 'var(--font-ui)',
        }}>
          {item.badge > 9 ? '9+' : item.badge}
        </span>
      )}
    </button>
  )

  return (
    <aside style={{
      width: 48, background: 'var(--bg-activity)', borderRight: '1px solid var(--border)',
      display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '10px 0', gap: 2
    }}>
      {top.map(item => renderIcon(item))}
      <div style={{ flex: 1 }} />
      {bottom.map(item => renderIcon(item, true))}
    </aside>
  )
}
