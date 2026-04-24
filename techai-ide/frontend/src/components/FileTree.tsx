import { useState, useEffect, useCallback, useRef } from 'react'
import { ChevronDown, ChevronRight, Folder, FolderOpen, FileCode, FileText, File, RefreshCw, FolderPlus, FilePlus, Trash2, FileJson, FileType, Cog, Image, Database, Terminal as TermIcon, Globe, Lock, FileVideo, FileAudio, FileArchive, FileSpreadsheet, BookOpen, Palette, Package, Shield, TestTube, Wrench, GitBranch } from 'lucide-react'
import { ListFiles, OpenFolder, WriteFile, DeleteFile, RenameFile, GetCwd, OpenInBrowser, StartLiveServer } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface FileEntry {
  name: string
  path: string
  isDir: boolean
  size: number
  kids?: FileEntry[]
}

interface Props {
  onFileSelect: (path: string) => void
  selectedFile: string
}

export default function FileTree({ onFileSelect, selectedFile }: Props) {
  const [tree, setTree] = useState<FileEntry[]>([])
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [projectName, setProjectName] = useState('')
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; path: string; isDir: boolean } | null>(null)
  const [filter, setFilter] = useState('')
  const [renamingPath, setRenamingPath] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  // Track modifier keys globally (Wails may not pass them in click events)
  const modKeyDown = useRef(false)
  useEffect(() => {
    const down = (e: KeyboardEvent) => { if (e.metaKey || e.ctrlKey) modKeyDown.current = true }
    const up = (e: KeyboardEvent) => { if (!e.metaKey && !e.ctrlKey) modKeyDown.current = false }
    window.addEventListener('keydown', down)
    window.addEventListener('keyup', up)
    return () => { window.removeEventListener('keydown', down); window.removeEventListener('keyup', up) }
  }, [])
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null) // path pending delete
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const refresh = useCallback(() => {
    ListFiles('.', 3).then(setTree).catch(console.error)
    GetCwd().then(cwd => {
      const parts = cwd.split('/')
      setProjectName(parts[parts.length - 1] || cwd)
    }).catch(() => {})
  }, [])

  useEffect(() => {
    refresh()
    const cancels = [
      EventsOn('file:changed', () => setTimeout(refresh, 300)),
      EventsOn('tree:refresh', () => setTimeout(refresh, 300)),
    ]
    return () => cancels.forEach(fn => fn())
  }, [refresh])

  function toggleDir(path: string) {
    setExpanded(prev => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  async function handleOpenFolder() {
    try {
      const dir = await OpenFolder()
      if (dir) refresh()
    } catch {}
  }

  function handleContextMenu(e: React.MouseEvent, path: string, isDir: boolean) {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, path, isDir })
  }

  async function handleNewFile(atPath?: string) {
    let base = atPath || (contextMenu?.isDir ? contextMenu.path : contextMenu?.path.substring(0, contextMenu.path.lastIndexOf('/')))
    if (!base) { const { GetCwd } = await import('../../wailsjs/go/main/App'); base = await GetCwd() }
    if (!base) return
    const newPath = base + '/newfile'
    try {
      await WriteFile(newPath, '')
      refresh()
      onFileSelect(newPath)
      // Start rename immediately
      setTimeout(() => { setRenamingPath(newPath); setRenameValue('newfile') }, 100)
    } catch (e) {
      import('./Toast').then(m => m.showToast(`Error: ${e}`, 'error'))
    }
    setContextMenu(null)
  }

  async function handleDelete(path?: string) {
    const target = path || contextMenu?.path
    if (!target) return
    try {
      await DeleteFile(target)
      import('./Toast').then(m => m.showToast('Deleted', 'success'))
      refresh()
    } catch (e) {
      import('./Toast').then(m => m.showToast(`Error: ${e}`, 'error'))
    }
    setContextMenu(null)
    setDeleteConfirm(null)
  }

  async function submitRename() {
    if (!renamingPath || !renameValue.trim()) { setRenamingPath(null); return }
    const dir = renamingPath.substring(0, renamingPath.lastIndexOf('/'))
    const newPath = dir + '/' + renameValue.trim()
    if (newPath !== renamingPath) {
      try {
        await RenameFile(renamingPath, newPath)
        refresh()
      } catch (e) {
        import('./Toast').then(m => m.showToast(`Error: ${e}`, 'error'))
      }
    }
    setRenamingPath(null)
  }

  function getIcon(entry: FileEntry) {
    if (entry.isDir) {
      const dname = entry.name.toLowerCase()
      const open = expanded.has(entry.path)
      const Icon = open ? FolderOpen : Folder
      // Color-code special folders
      let color = 'var(--accent)'
      if (dname === 'src' || dname === 'lib' || dname === 'internal') color = '#61afef'
      if (dname === 'test' || dname === 'tests' || dname === '__tests__') color = '#98c379'
      if (dname === 'node_modules' || dname === 'vendor') color = 'var(--fg-dim)'
      if (dname === 'public' || dname === 'static' || dname === 'assets') color = '#c678dd'
      if (dname === 'components') color = '#61afef'
      if (dname === 'pages' || dname === 'app' || dname === 'views') color = '#e5c07b'
      if (dname === 'api' || dname === 'routes') color = '#d19a66'
      if (dname === 'config' || dname === 'configs') color = '#e06c75'
      if (dname === 'docs' || dname === 'doc') color = '#56b6c2'
      if (dname === 'build' || dname === 'dist' || dname === 'out') color = '#f59e0b'
      if (dname === '.git' || dname === '.github') color = '#F05032'
      if (dname === 'cmd') color = '#00ADD8'
      if (dname === 'pkg') color = '#00ADD8'
      if (dname === 'frontend' || dname === 'client') color = '#3178C6'
      if (dname === 'backend' || dname === 'server') color = '#68217A'
      return <Icon size={15} style={{ color }} />
    }
    const ext = entry.name.split('.').pop()?.toLowerCase() || ''
    const name = entry.name.toLowerCase()
    const s = 15
    // Languages
    if (ext === 'go' || ext === 'mod' || ext === 'sum') return <FileCode size={s} style={{ color: '#00ADD8' }} />
    if (ext === 'ts' || ext === 'tsx') return <FileType size={s} style={{ color: '#3178C6' }} />
    if (ext === 'js' || ext === 'jsx' || ext === 'mjs' || ext === 'cjs') return <FileCode size={s} style={{ color: '#F7DF1E' }} />
    if (ext === 'py' || ext === 'pyw' || ext === 'pyi') return <FileCode size={s} style={{ color: '#3776AB' }} />
    if (ext === 'rs') return <FileCode size={s} style={{ color: '#CE422B' }} />
    if (ext === 'java' || ext === 'kt' || ext === 'kts') return <FileCode size={s} style={{ color: '#ED8B00' }} />
    if (ext === 'c' || ext === 'h') return <FileCode size={s} style={{ color: '#A8B9CC' }} />
    if (ext === 'cpp' || ext === 'cc' || ext === 'cxx' || ext === 'hpp') return <FileCode size={s} style={{ color: '#00599C' }} />
    if (ext === 'cs') return <FileCode size={s} style={{ color: '#68217A' }} />
    if (ext === 'rb') return <FileCode size={s} style={{ color: '#CC342D' }} />
    if (ext === 'php') return <FileCode size={s} style={{ color: '#777BB4' }} />
    if (ext === 'swift') return <FileCode size={s} style={{ color: '#F05138' }} />
    if (ext === 'dart') return <FileCode size={s} style={{ color: '#0175C2' }} />
    if (ext === 'lua') return <FileCode size={s} style={{ color: '#000080' }} />
    if (ext === 'r' || ext === 'rmd') return <FileCode size={s} style={{ color: '#276DC3' }} />
    if (ext === 'scala') return <FileCode size={s} style={{ color: '#DC322F' }} />
    if (ext === 'ex' || ext === 'exs') return <FileCode size={s} style={{ color: '#6E4A7E' }} />
    if (ext === 'zig') return <FileCode size={s} style={{ color: '#F7A41D' }} />
    if (ext === 'v' || ext === 'sv') return <FileCode size={s} style={{ color: '#5D87BF' }} />
    // Web
    if (ext === 'html' || ext === 'htm' || ext === 'ejs' || ext === 'hbs') return <Globe size={s} style={{ color: '#E34F26' }} />
    if (ext === 'css') return <Palette size={s} style={{ color: '#1572B6' }} />
    if (ext === 'scss' || ext === 'sass' || ext === 'less') return <Palette size={s} style={{ color: '#CC6699' }} />
    if (ext === 'vue') return <FileCode size={s} style={{ color: '#4FC08D' }} />
    if (ext === 'svelte') return <FileCode size={s} style={{ color: '#FF3E00' }} />
    if (ext === 'astro') return <FileCode size={s} style={{ color: '#FF5D01' }} />
    // Data
    if (ext === 'json' || ext === 'jsonc') return <FileJson size={s} style={{ color: '#d19a66' }} />
    if (ext === 'yaml' || ext === 'yml') return <FileJson size={s} style={{ color: '#CB171E' }} />
    if (ext === 'toml') return <FileJson size={s} style={{ color: '#9C4121' }} />
    if (ext === 'xml' || ext === 'plist') return <FileJson size={s} style={{ color: '#e37933' }} />
    if (ext === 'csv' || ext === 'tsv' || ext === 'xls' || ext === 'xlsx') return <FileSpreadsheet size={s} style={{ color: '#217346' }} />
    if (ext === 'sql') return <Database size={s} style={{ color: '#336791' }} />
    if (ext === 'graphql' || ext === 'gql') return <Database size={s} style={{ color: '#E10098' }} />
    if (ext === 'proto') return <Database size={s} style={{ color: '#4285F4' }} />
    // Config
    if (['ini', 'cfg', 'conf', 'properties'].includes(ext) || name === 'dockerfile' || name === '.dockerignore')
      return <Cog size={s} style={{ color: '#2496ED' }} />
    if (name === 'makefile' || name === 'cmakelists.txt' || ext === 'mk') return <Wrench size={s} style={{ color: 'var(--fg-muted)' }} />
    if (name.startsWith('.env')) return <Lock size={s} style={{ color: '#ECD53F' }} />
    if (name === '.gitignore' || name === '.gitattributes') return <GitBranch size={s} style={{ color: '#F05032' }} />
    if (ext === 'nginx' || name === 'nginx.conf') return <Cog size={s} style={{ color: '#009639' }} />
    // Package managers
    if (name === 'package.json' || name === 'package-lock.json') return <Package size={s} style={{ color: '#CB3837' }} />
    if (name === 'cargo.toml' || name === 'cargo.lock') return <Package size={s} style={{ color: '#CE422B' }} />
    if (name === 'go.mod' || name === 'go.sum') return <Package size={s} style={{ color: '#00ADD8' }} />
    if (name === 'requirements.txt' || name === 'pyproject.toml' || name === 'pipfile') return <Package size={s} style={{ color: '#3776AB' }} />
    if (name === 'gemfile' || name === 'gemfile.lock') return <Package size={s} style={{ color: '#CC342D' }} />
    // Shell
    if (['sh', 'bash', 'zsh', 'fish', 'ps1', 'bat', 'cmd'].includes(ext)) return <TermIcon size={s} style={{ color: 'var(--success)' }} />
    // Docs
    if (ext === 'md' || ext === 'mdx') return <BookOpen size={s} style={{ color: '#083fa1' }} />
    if (ext === 'txt' || ext === 'rtf') return <FileText size={s} style={{ color: 'var(--fg-muted)' }} />
    if (ext === 'pdf') return <FileText size={s} style={{ color: '#FF0000' }} />
    if (ext === 'doc' || ext === 'docx') return <FileText size={s} style={{ color: '#2B579A' }} />
    if (name === 'license' || name === 'licence' || name.startsWith('license')) return <Shield size={s} style={{ color: '#D4AA00' }} />
    if (name === 'readme' || name === 'readme.md') return <BookOpen size={s} style={{ color: '#083fa1' }} />
    // Media
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'tiff', 'ico', 'svg'].includes(ext))
      return <Image size={s} style={{ color: '#c678dd' }} />
    if (['mp4', 'webm', 'avi', 'mov', 'mkv'].includes(ext))
      return <FileVideo size={s} style={{ color: '#FF6F61' }} />
    if (['mp3', 'wav', 'ogg', 'flac', 'aac'].includes(ext))
      return <FileAudio size={s} style={{ color: '#1DB954' }} />
    // Archives
    if (['zip', 'tar', 'gz', 'bz2', 'xz', '7z', 'rar', 'war', 'jar'].includes(ext))
      return <FileArchive size={s} style={{ color: '#FFC107' }} />
    // Test files
    if (name.includes('test') || name.includes('spec') || name.includes('_test.'))
      return <TestTube size={s} style={{ color: '#98c379' }} />
    // Lock files
    if (ext === 'lock' || name.endsWith('.lock')) return <Lock size={s} style={{ color: 'var(--fg-dim)' }} />
    // Certificates
    if (['pem', 'crt', 'key', 'cert'].includes(ext)) return <Shield size={s} style={{ color: '#E06C75' }} />
    // Wasm
    if (ext === 'wasm') return <FileCode size={s} style={{ color: '#654FF0' }} />
    return <File size={s} style={{ color: 'var(--fg-dim)' }} />
  }

  function renderItems(items: FileEntry[], depth: number) {
    const filtered = filter ? items.filter(e => e.isDir || e.name.toLowerCase().includes(filter.toLowerCase())) : items
    return filtered.map(entry => (
      <div key={entry.path}>
        <div
          onClick={(e) => {
            if (entry.isDir) { toggleDir(entry.path); return }
            // Shift+Click: range selection
            if (e.shiftKey && selectedFile) {
              e.preventDefault()
              const allFiles: string[] = []
              function collect(items: FileEntry[]) { for (const f of items) { if (!f.isDir) allFiles.push(f.path); if (f.kids) collect(f.kids) } }
              collect(tree)
              const startIdx = allFiles.indexOf(selectedFile)
              const endIdx = allFiles.indexOf(entry.path)
              if (startIdx >= 0 && endIdx >= 0) {
                setSelected(new Set(allFiles.slice(Math.min(startIdx, endIdx), Math.max(startIdx, endIdx) + 1)))
              }
              return
            }
            // Normal click: always clear selection + open file
            setSelected(new Set())
            onFileSelect(entry.path)
          }}
          onContextMenu={e => handleContextMenu(e, entry.path, entry.isDir)}
          draggable={!entry.isDir}
          onDragStart={e => { if (!entry.isDir) e.dataTransfer.setData('text/plain', entry.path) }}
          onDoubleClick={() => { if (!entry.isDir) onFileSelect(entry.path) }}
          className="tree-item-row"
          style={{
            padding: '4px 12px', paddingLeft: 12 + depth * 14, minHeight: 24,
            fontSize: 12.5, display: 'flex', alignItems: 'center', gap: 6,
            cursor: 'pointer',
            color: selectedFile === entry.path || selected.has(entry.path) ? 'var(--fg-primary)' : 'var(--fg-secondary)',
            background: selected.has(entry.path) ? 'rgba(59,130,246,0.15)' : selectedFile === entry.path ? 'var(--bg-active)' : 'transparent',
            transition: 'background 0.1s',
          }}
          onMouseEnter={e => { if (selectedFile !== entry.path) e.currentTarget.style.background = 'var(--bg-hover)' }}
          onMouseLeave={e => { if (selectedFile !== entry.path) e.currentTarget.style.background = 'transparent' }}
        >
          {entry.isDir ? (expanded.has(entry.path) ? <ChevronDown size={14} style={{ color: 'var(--fg-dim)', flexShrink: 0 }} /> : <ChevronRight size={14} style={{ color: 'var(--fg-dim)', flexShrink: 0 }} />) : <span style={{ width: 14, flexShrink: 0 }} />}
          {getIcon(entry)}
          {renamingPath === entry.path ? (
            <input autoFocus value={renameValue} onChange={e => setRenameValue(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') submitRename(); if (e.key === 'Escape') setRenamingPath(null) }}
              onBlur={submitRename}
              onClick={e => e.stopPropagation()}
              style={{
                flex: 1, padding: '1px 4px', fontSize: 12, borderRadius: 3,
                background: 'var(--bg-base)', border: '1px solid var(--accent)', outline: 'none',
                color: 'var(--fg-primary)', fontFamily: 'var(--font-code)',
              }}
            />
          ) : (
            <>
              <span style={{ fontWeight: entry.isDir ? 500 : 400, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', flex: 1 }}>{entry.name}</span>
              {!entry.isDir && entry.size > 0 && (
                <span style={{ fontSize: 9, color: 'var(--fg-dim)', flexShrink: 0 }}>{formatSize(entry.size)}</span>
              )}
            </>
          )}
        </div>
        {entry.isDir && expanded.has(entry.path) && entry.kids && renderItems(entry.kids, depth + 1)}
      </div>
    ))
  }

  return (
    <aside style={{
      width: '100%', height: '100%', background: 'var(--bg-sidebar)',
      display: 'flex', flexDirection: 'column', overflow: 'hidden'
    }} tabIndex={0} onClick={() => { setContextMenu(null); setDeleteConfirm(null) }}
      onKeyDown={e => {
        // Cmd+A toggle select all files
        if ((e.metaKey || e.ctrlKey) && e.key === 'a') {
          e.preventDefault(); e.stopPropagation()
          if (selected.size > 0) {
            setSelected(new Set()) // deselect all
          } else {
            const allPaths = new Set<string>()
            function collect(items: FileEntry[]) {
              for (const f of items) {
                if (!f.isDir) allPaths.add(f.path)
                if (f.kids) collect(f.kids)
              }
            }
            collect(tree)
            setSelected(allPaths)
          }
        }
        // Delete selected
        if (e.key === 'Backspace' || e.key === 'Delete') {
          if (selected.size > 0) {
            selected.forEach(p => DeleteFile(p).catch(() => {}))
            import('./Toast').then(m => m.showToast(`Deleted ${selected.size} file(s)`, 'success'))
            setSelected(new Set())
            refresh()
          }
        }
      }}
      onContextMenu={e => {
        // Empty space right-click
        if ((e.target as HTMLElement).closest('.tree-item-row')) return
        e.preventDefault()
        setContextMenu({ x: e.clientX, y: e.clientY, path: '', isDir: true })
      }}
      onDragOver={e => { e.preventDefault(); e.stopPropagation() }}
      onDrop={async e => {
        e.preventDefault(); e.stopPropagation()
        const files = e.dataTransfer.files
        if (!files.length) return
        for (let i = 0; i < files.length; i++) {
          const file = files[i]
          const buf = await file.arrayBuffer()
          await WriteFile(file.name, new TextDecoder().decode(buf))
        }
        import('./Toast').then(m => m.showToast(`${files.length} file(s) added`, 'success'))
        refresh()
      }}
    >
      <div style={{
        padding: '10px 14px', fontSize: 11, fontWeight: 700, textTransform: 'uppercase',
        letterSpacing: '0.05em', color: 'var(--fg-muted)',
        display: 'flex', alignItems: 'center', gap: 6
      }}>
        <span style={{ flex: 1 }}>{projectName || 'Explorer'}</span>
        <FilePlus size={13} style={{ cursor: 'pointer', opacity: 0.6 }} onClick={() => handleNewFile()} />
        <FolderPlus size={13} style={{ cursor: 'pointer', opacity: 0.6 }} onClick={handleOpenFolder} />
        <RefreshCw size={12} style={{ cursor: 'pointer', opacity: 0.6 }} onClick={refresh} />
      </div>
      {/* Filter */}
      <div style={{ padding: '0 10px 4px' }}>
        <input value={filter} onChange={e => setFilter(e.target.value)}
          placeholder="Filter files..."
          style={{
            width: '100%', padding: '3px 8px', borderRadius: 4, fontSize: 11,
            background: 'var(--bg-active)', border: '1px solid var(--border)',
            color: 'var(--fg-primary)', fontFamily: 'var(--font-ui)', outline: 'none',
            display: filter || tree.length > 20 ? 'block' : 'none',
          }}
        />
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '2px 0' }}>
        {renderItems(tree, 0)}
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <div style={{
          position: 'fixed', left: contextMenu.x, top: contextMenu.y,
          background: 'var(--bg-panel)', border: '1px solid var(--border)',
          borderRadius: 6, padding: '4px 0', boxShadow: '0 8px 24px rgba(0,0,0,0.3)',
          zIndex: 999, minWidth: 160,
        }}>
          <div onClick={() => handleNewFile()} style={ctxStyle}>
            <FilePlus size={13} /> New File
          </div>
          <div onClick={() => {
            if (!contextMenu) return
            setRenamingPath(contextMenu.path)
            setRenameValue(contextMenu.path.split('/').pop() || '')
            setContextMenu(null)
          }} style={ctxStyle}>
            <FilePlus size={13} /> Rename
          </div>
          {contextMenu && !contextMenu.isDir && /\.(html?|htm)$/i.test(contextMenu.path) && (
            <>
              <div onClick={() => { OpenInBrowser(contextMenu!.path); setContextMenu(null) }} style={ctxStyle}>
                <Globe size={13} /> Open in Browser
              </div>
              <div onClick={() => {
                const dir = contextMenu!.path.substring(0, contextMenu!.path.lastIndexOf('/'))
                StartLiveServer(dir || '.'); setContextMenu(null)
              }} style={ctxStyle}>
                <Globe size={13} style={{ color: 'var(--success)' }} /> Live Server
              </div>
            </>
          )}
          {deleteConfirm === contextMenu?.path ? (
            <div style={{ padding: '4px 10px', display: 'flex', gap: 4 }}>
              <button onClick={() => handleDelete(contextMenu?.path)} style={{
                flex: 1, padding: '4px', borderRadius: 4, border: 'none', cursor: 'pointer',
                background: 'var(--error)', color: '#fff', fontSize: 11, fontFamily: 'var(--font-ui)',
              }}>Delete</button>
              <button onClick={() => { setDeleteConfirm(null); setContextMenu(null) }} style={{
                flex: 1, padding: '4px', borderRadius: 4, border: '1px solid var(--border)', cursor: 'pointer',
                background: 'var(--bg-active)', color: 'var(--fg-secondary)', fontSize: 11, fontFamily: 'var(--font-ui)',
              }}>Cancel</button>
            </div>
          ) : (
            <div onClick={() => setDeleteConfirm(contextMenu?.path || null)} style={{...ctxStyle, color: 'var(--error)'}}>
              <Trash2 size={13} /> Delete
            </div>
          )}
        </div>
      )}
    </aside>
  )
}

const ctxStyle: React.CSSProperties = {
  padding: '6px 14px', fontSize: 12, display: 'flex', alignItems: 'center', gap: 8,
  cursor: 'pointer', color: 'var(--fg-secondary)', transition: 'background 0.1s',
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + 'K'
  return (bytes / 1024 / 1024).toFixed(1) + 'M'
}
