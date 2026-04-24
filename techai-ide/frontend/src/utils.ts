// Platform detection
export const isMac = navigator.platform.toUpperCase().includes('MAC')
export const modKey = isMac ? 'Cmd' : 'Ctrl'
export const modSymbol = isMac ? '⌘' : 'Ctrl'
