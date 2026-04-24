// Platform detection
export const isMac = navigator.platform.toUpperCase().includes('MAC')
export const isWindows = navigator.platform.toUpperCase().includes('WIN')
export const modKey = isMac ? 'Cmd' : 'Ctrl'
export const modSymbol = isMac ? '⌘' : 'Ctrl'

// Path utilities — handles both / and \ separators
export function pathBasename(p: string): string {
  const sep = p.includes('\\') ? '\\' : '/'
  return p.split(sep).pop() || p
}

export function pathDirname(p: string): string {
  const sep = p.includes('\\') ? '\\' : '/'
  const parts = p.split(sep)
  parts.pop()
  return parts.join(sep)
}

export function pathSplit(p: string): string[] {
  return p.split(/[/\\]/)
}

export function pathShorten(p: string): string {
  const parts = pathSplit(p)
  return parts.length > 2 ? '~/' + parts.slice(-2).join('/') : p
}
