import { useState, useEffect } from 'react'
import { Palette, X } from 'lucide-react'

const themes = [
  { id: '', name: 'Slate', dot: 'linear-gradient(135deg,#3b82f6,#10b981)', dark: true },
  { id: 't-cursor', name: 'Cursor', dot: '#f54e00', dark: true },
  { id: 't-linear', name: 'Linear', dot: '#7c3aed', dark: true },
  { id: 't-github', name: 'GitHub Dark', dot: '#58a6ff', dark: true },
  { id: 't-dracula', name: 'Dracula', dot: '#bd93f9', dark: true },
  { id: 't-nord', name: 'Nord', dot: '#88c0d0', dark: true },
  { id: 't-onedark', name: 'One Dark', dot: '#61afef', dark: true },
  { id: 't-monokai', name: 'Monokai', dot: '#ffd866', dark: true },
  { id: 't-solarized', name: 'Solarized', dot: '#268bd2', dark: true },
  { id: 't-claude', name: 'Claude', dot: '#c96442', dark: false },
  { id: 't-vercel', name: 'Vercel', dot: 'linear-gradient(135deg,#000,#fff)', dark: false },
]

interface Props {
  open: boolean
  onClose: () => void
}

export default function ThemePicker({ open, onClose }: Props) {
  const [current, setCurrent] = useState('')

  useEffect(() => {
    const saved = localStorage.getItem('techai-theme') || ''
    setCurrent(saved)
    document.body.className = saved
  }, [])

  function select(id: string) {
    setCurrent(id)
    document.body.className = id
    localStorage.setItem('techai-theme', id)
  }

  if (!open) return null

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(0,0,0,0.5)', zIndex: 1000,
      display: 'flex', alignItems: 'center', justifyContent: 'center',
    }} onClick={onClose}>
      <div onClick={e => e.stopPropagation()} style={{
        background: 'var(--bg-panel)', border: '1px solid var(--border)',
        borderRadius: 12, padding: 20, width: 360,
        boxShadow: '0 16px 48px rgba(0,0,0,0.3)',
      }}>
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Palette size={16} style={{ color: 'var(--accent)' }} />
            <span style={{ fontWeight: 600, fontSize: 14 }}>Theme</span>
          </div>
          <button onClick={onClose} style={{
            background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-muted)', padding: 4
          }}>
            <X size={16} />
          </button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {themes.map(t => (
            <button key={t.id} onClick={() => select(t.id)} style={{
              display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px',
              background: current === t.id ? 'var(--bg-active)' : 'transparent',
              border: 'none', borderRadius: 6, cursor: 'pointer',
              color: 'var(--fg-primary)', fontFamily: 'var(--font-ui)', fontSize: 13,
              transition: 'background 0.1s', textAlign: 'left', width: '100%',
            }}
            onMouseEnter={e => { if (current !== t.id) e.currentTarget.style.background = 'var(--bg-hover)' }}
            onMouseLeave={e => { if (current !== t.id) e.currentTarget.style.background = 'transparent' }}
            >
              <span style={{
                width: 14, height: 14, borderRadius: '50%',
                background: t.dot, flexShrink: 0,
                border: '1px solid rgba(255,255,255,0.1)'
              }} />
              <span>{t.name}</span>
              <span style={{ marginLeft: 'auto', fontSize: 10, color: 'var(--fg-dim)' }}>
                {t.dark ? 'Dark' : 'Light'}
              </span>
              {current === t.id && <span style={{ color: 'var(--accent)', fontSize: 12 }}>&#10003;</span>}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
