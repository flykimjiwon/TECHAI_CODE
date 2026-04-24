import { useState, useEffect, useRef } from 'react'
import { Search, FileCode, FileText, File } from 'lucide-react'
import { WalkProject } from '../../wailsjs/go/main/App'

interface FileEntry {
  name: string
  path: string
  isDir: boolean
  kids?: FileEntry[]
}

interface Props {
  open: boolean
  onClose: () => void
  onSelect: (path: string) => void
}

export default function QuickOpen({ open, onClose, onSelect }: Props) {
  const [query, setQuery] = useState('')
  const [allFiles, setAllFiles] = useState<string[]>([])
  const [selected, setSelected] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (open) {
      setQuery('')
      setSelected(0)
      inputRef.current?.focus()
      // Flatten file tree
      WalkProject().then(tree => {
        const files: string[] = []
        function walk(items: FileEntry[], prefix: string) {
          for (const item of items) {
            const p = prefix ? prefix + '/' + item.name : item.name
            if (!item.isDir) files.push(item.path)
            if (item.kids) walk(item.kids, p)
          }
        }
        walk(tree, '')
        setAllFiles(files)
      }).catch(() => {})
    }
  }, [open])

  const filtered = query
    ? allFiles.filter(f => {
        const name = f.split('/').pop()?.toLowerCase() || ''
        const q = query.toLowerCase()
        return name.includes(q) || f.toLowerCase().includes(q)
      }).slice(0, 15)
    : allFiles.slice(0, 15)

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') { onClose(); return }
    if (e.key === 'ArrowDown') { e.preventDefault(); setSelected(s => Math.min(s + 1, filtered.length - 1)) }
    if (e.key === 'ArrowUp') { e.preventDefault(); setSelected(s => Math.max(s - 1, 0)) }
    if (e.key === 'Enter' && filtered[selected]) {
      onSelect(filtered[selected])
      onClose()
    }
  }

  function getIcon(name: string) {
    const ext = name.split('.').pop()?.toLowerCase()
    if (['go', 'ts', 'tsx', 'js', 'jsx', 'py', 'rs'].includes(ext || ''))
      return <FileCode size={14} style={{ color: '#61afef', flexShrink: 0 }} />
    if (['md', 'txt'].includes(ext || ''))
      return <FileText size={14} style={{ color: 'var(--fg-muted)', flexShrink: 0 }} />
    return <File size={14} style={{ color: 'var(--fg-muted)', flexShrink: 0 }} />
  }

  if (!open) return null

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(0,0,0,0.4)', zIndex: 1000,
      display: 'flex', justifyContent: 'center', paddingTop: 80,
    }} onClick={onClose}>
      <div onClick={e => e.stopPropagation()} style={{
        width: 500, background: 'var(--bg-panel)', border: '1px solid var(--border)',
        borderRadius: 10, boxShadow: '0 16px 48px rgba(0,0,0,0.4)',
        overflow: 'hidden', maxHeight: 400,
      }}>
        {/* Search input */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 8, padding: '10px 14px',
          borderBottom: '1px solid var(--border)',
        }}>
          <Search size={16} style={{ color: 'var(--fg-muted)' }} />
          <input
            ref={inputRef}
            value={query} onChange={e => { setQuery(e.target.value); setSelected(0) }}
            onKeyDown={handleKeyDown}
            placeholder="Type to search files..."
            style={{
              flex: 1, background: 'none', border: 'none', outline: 'none',
              color: 'var(--fg-primary)', fontSize: 14, fontFamily: 'var(--font-ui)',
            }}
          />
        </div>

        {/* Results */}
        <div style={{ maxHeight: 320, overflowY: 'auto' }}>
          {filtered.map((file, i) => {
            const name = file.split('/').pop() || file
            const dir = file.substring(0, file.length - name.length - 1)
            return (
              <div key={file} ref={i === selected ? el => el?.scrollIntoView({ block: 'nearest' }) : undefined}
                onClick={() => { onSelect(file); onClose() }} style={{
                padding: '6px 14px', display: 'flex', alignItems: 'center', gap: 8,
                cursor: 'pointer', fontSize: 13,
                background: i === selected ? 'var(--bg-active)' : 'transparent',
                color: i === selected ? 'var(--fg-primary)' : 'var(--fg-secondary)',
              }}
              onMouseEnter={() => setSelected(i)}
              >
                {getIcon(name)}
                <span style={{ fontWeight: 500 }}>{name}</span>
                <span style={{ fontSize: 11, color: 'var(--fg-dim)', marginLeft: 'auto' }}>{dir}</span>
              </div>
            )
          })}
          {filtered.length === 0 && (
            <div style={{ padding: '12px 14px', color: 'var(--fg-dim)', fontSize: 13 }}>No files found</div>
          )}
        </div>
      </div>
    </div>
  )
}
