import { useEffect } from 'react'
import { X } from 'lucide-react'

interface Props {
  open: boolean
  onClose: () => void
}

export default function AboutDialog({ open, onClose }: Props) {
  useEffect(() => {
    if (!open) return
    const fn = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', fn)
    return () => window.removeEventListener('keydown', fn)
  }, [open, onClose])

  if (!open) return null
  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(0,0,0,0.5)', zIndex: 1000,
      display: 'flex', alignItems: 'center', justifyContent: 'center',
    }} onClick={onClose}>
      <div onClick={e => e.stopPropagation()} role="dialog" aria-modal="true" style={{
        background: 'var(--bg-panel)', border: '1px solid var(--border)',
        borderRadius: 12, padding: 28, width: 340, textAlign: 'center',
        boxShadow: '0 16px 48px rgba(0,0,0,0.3)', position: 'relative',
      }}>
        <button onClick={onClose} title="Close" style={{
          position: 'absolute', right: 12, top: 12,
          background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-dim)', padding: 6,
        }}><X size={16} /></button>

        <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--accent)', marginBottom: 8 }}>
          TECHAI IDE
        </div>
        <div style={{ fontSize: 12, color: 'var(--fg-muted)', marginBottom: 16 }}>
          v0.2.0 — AI-Powered Code Editor
        </div>
        <div style={{ fontSize: 11, color: 'var(--fg-dim)', lineHeight: 1.8, marginBottom: 16 }}>
          Built with Go + Wails + React<br />
          Powered by Qwen3-Coder<br />
          12MB single binary<br />
          macOS / Windows
        </div>
        <div style={{ fontSize: 10, color: 'var(--fg-dim)' }}>
          TECHAI CODE — Shinhan Bank Tech Innovation
        </div>
      </div>
    </div>
  )
}
