import { Files, Search, GitBranch, User, Settings } from 'lucide-react'

interface Props {
  active: string
  onSelect: (panel: string) => void
}

export default function ActivityBar({ active, onSelect }: Props) {
  const top = [
    { id: 'files', icon: Files, label: 'Explorer' },
    { id: 'search', icon: Search, label: 'Search' },
    { id: 'git', icon: GitBranch, label: 'Git' },
  ]
  const bottom = [
    { id: 'account', icon: User, label: 'Account' },
    { id: 'settings', icon: Settings, label: 'Settings' },
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
