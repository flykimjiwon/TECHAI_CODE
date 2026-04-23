import { useState, useRef, useEffect } from 'react'
import { Search, FileCode, Filter, X } from 'lucide-react'
import { SearchInFiles } from '../../wailsjs/go/main/App'

interface SearchResult {
  file: string
  line: number
  content: string
}

interface Props {
  onFileSelect: (path: string) => void
}

export default function SearchPanel({ onFileSelect }: Props) {
  const [query, setQuery] = useState('')
  const [include, setInclude] = useState('')
  const [exclude, setExclude] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [searched, setSearched] = useState(false)
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const debounceRef = useRef<number>(0)
  const inputRef = useRef<HTMLInputElement>(null)

  // Real-time search with debounce
  useEffect(() => {
    const q = query.trim()
    if (!q || q.length < 2) {
      if (searched) { setResults([]); setSearched(false) }
      return
    }
    clearTimeout(debounceRef.current)
    debounceRef.current = window.setTimeout(() => doSearch(q), 300)
    return () => clearTimeout(debounceRef.current)
  }, [query, include, exclude])

  function doSearch(q?: string) {
    const searchQuery = q || query.trim()
    if (!searchQuery) return
    setSearching(true)
    setSearched(true)
    SearchInFiles(searchQuery, '.').then(raw => {
      let res = raw || []
      // Client-side include filter
      if (include.trim()) {
        const patterns = include.split(',').map(p => p.trim().toLowerCase()).filter(Boolean)
        res = res.filter(r => patterns.some(p => {
          if (p.startsWith('*.')) return r.file.toLowerCase().endsWith(p.slice(1))
          return r.file.toLowerCase().includes(p)
        }))
      }
      // Client-side exclude filter
      if (exclude.trim()) {
        const patterns = exclude.split(',').map(p => p.trim().toLowerCase()).filter(Boolean)
        res = res.filter(r => !patterns.some(p => {
          if (p.startsWith('*.')) return r.file.toLowerCase().endsWith(p.slice(1))
          return r.file.toLowerCase().includes(p)
        }))
      }
      setResults(res)
      setSearching(false)
    }).catch(() => setSearching(false))
  }

  function toggleFile(file: string) {
    setCollapsed(prev => {
      const next = new Set(prev)
      if (next.has(file)) next.delete(file)
      else next.add(file)
      return next
    })
  }

  // Group results by file
  const grouped: Record<string, SearchResult[]> = {}
  for (const r of results) {
    if (!grouped[r.file]) grouped[r.file] = []
    grouped[r.file].push(r)
  }
  const fileCount = Object.keys(grouped).length

  return (
    <aside style={{
      width: '100%', height: '100%', background: 'var(--bg-sidebar)',
      display: 'flex', flexDirection: 'column', overflow: 'hidden'
    }}>
      <div style={{
        padding: '10px 14px', fontSize: 11, fontWeight: 700, textTransform: 'uppercase',
        letterSpacing: '0.05em', color: 'var(--fg-muted)',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      }}>
        Search
        <button onClick={() => setShowFilters(f => !f)} style={{
          background: showFilters ? 'var(--accent-glow)' : 'none', border: 'none',
          cursor: 'pointer', color: showFilters ? 'var(--accent)' : 'var(--fg-dim)',
          padding: 2, borderRadius: 4,
        }}>
          <Filter size={12} />
        </button>
      </div>

      <div style={{ padding: '0 10px 6px', display: 'flex', flexDirection: 'column', gap: 4 }}>
        {/* Main search */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 6,
          background: 'var(--bg-active)', border: '1px solid var(--border)',
          borderRadius: 6, padding: '5px 8px',
        }}>
          <Search size={13} style={{ color: 'var(--fg-muted)', flexShrink: 0 }} />
          <input ref={inputRef}
            value={query} onChange={e => setQuery(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') doSearch() }}
            placeholder="Search..."
            style={{
              flex: 1, background: 'none', border: 'none', outline: 'none',
              color: 'var(--fg-primary)', fontSize: 12, fontFamily: 'var(--font-ui)',
            }}
          />
          {query && (
            <button onClick={() => { setQuery(''); setResults([]); setSearched(false) }} style={{
              background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-dim)', padding: 0,
            }}><X size={12} /></button>
          )}
        </div>

        {/* Filters */}
        {showFilters && (
          <>
            <input value={include} onChange={e => setInclude(e.target.value)}
              placeholder="Include: *.go, *.ts, src/"
              style={filterInputStyle}
            />
            <input value={exclude} onChange={e => setExclude(e.target.value)}
              placeholder="Exclude: *.test.*, node_modules"
              style={filterInputStyle}
            />
          </>
        )}
      </div>

      {/* Results count */}
      {searched && !searching && (
        <div style={{ padding: '0 14px 4px', fontSize: 11, color: 'var(--fg-dim)' }}>
          {results.length} results in {fileCount} files
        </div>
      )}

      {/* Results */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {searching && <div style={{ padding: '8px 14px', fontSize: 12, color: 'var(--fg-muted)' }}>Searching...</div>}
        {!searching && searched && results.length === 0 && (
          <div style={{ padding: '8px 14px', fontSize: 12, color: 'var(--fg-dim)' }}>No results found</div>
        )}
        {Object.entries(grouped).map(([file, matches]) => (
          <div key={file}>
            <div onClick={() => toggleFile(file)} style={{
              padding: '4px 10px', fontSize: 11, fontWeight: 600,
              color: 'var(--accent)', cursor: 'pointer', display: 'flex',
              alignItems: 'center', gap: 4, userSelect: 'none',
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            >
              <span style={{ fontSize: 8, color: 'var(--fg-dim)' }}>
                {collapsed.has(file) ? '▸' : '▾'}
              </span>
              <FileCode size={12} />
              <span style={{ flex: 1 }}>{file}</span>
              <span style={{ fontSize: 10, color: 'var(--fg-dim)', background: 'var(--bg-active)', padding: '0 4px', borderRadius: 8 }}>
                {matches.length}
              </span>
            </div>
            {!collapsed.has(file) && matches.map((r, j) => (
              <div key={j} onClick={() => onFileSelect(r.file)} style={{
                padding: '2px 10px 2px 28px', cursor: 'pointer',
                fontFamily: 'var(--font-code)', fontSize: 11, color: 'var(--fg-muted)',
                whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
              }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-hover)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
              >
                <span style={{ color: 'var(--fg-dim)', marginRight: 6, fontSize: 10 }}>{r.line}</span>
                {highlightMatch(r.content, query)}
              </div>
            ))}
          </div>
        ))}
      </div>
    </aside>
  )
}

const filterInputStyle: React.CSSProperties = {
  padding: '4px 8px', borderRadius: 4, fontSize: 11,
  background: 'var(--bg-active)', border: '1px solid var(--border)',
  color: 'var(--fg-secondary)', fontFamily: 'var(--font-ui)', outline: 'none',
  width: '100%',
}

function highlightMatch(text: string, query: string) {
  if (!query) return <>{text}</>
  const lowerText = text.toLowerCase()
  const lowerQuery = query.toLowerCase()
  const parts: JSX.Element[] = []
  let lastIdx = 0
  let idx = lowerText.indexOf(lowerQuery)
  let key = 0
  while (idx !== -1) {
    if (idx > lastIdx) parts.push(<span key={key++}>{text.slice(lastIdx, idx)}</span>)
    parts.push(
      <span key={key++} style={{
        background: 'rgba(255,200,0,0.3)', color: 'var(--fg-primary)',
        borderRadius: 2, padding: '0 1px',
      }}>{text.slice(idx, idx + query.length)}</span>
    )
    lastIdx = idx + query.length
    idx = lowerText.indexOf(lowerQuery, lastIdx)
  }
  if (lastIdx < text.length) parts.push(<span key={key++}>{text.slice(lastIdx)}</span>)
  return <>{parts}</>
}
