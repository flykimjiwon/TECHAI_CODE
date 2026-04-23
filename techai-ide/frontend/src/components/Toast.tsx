import { useState, useEffect, useCallback } from 'react'
import { CheckCircle, AlertCircle, Info, X } from 'lucide-react'

interface ToastItem {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
}

let toastId = 0
let addToastFn: ((msg: string, type: 'success' | 'error' | 'info') => void) | null = null

// Global function to show toast from anywhere
export function showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
  if (addToastFn) addToastFn(message, type)
}

export default function ToastContainer() {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const addToast = useCallback((message: string, type: 'success' | 'error' | 'info') => {
    const id = ++toastId
    setToasts(prev => [...prev, { id, message, type }])
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 3000)
  }, [])

  useEffect(() => {
    addToastFn = addToast
    return () => { addToastFn = null }
  }, [addToast])

  if (toasts.length === 0) return null

  return (
    <div style={{
      position: 'fixed', bottom: 40, right: 16, zIndex: 2000,
      display: 'flex', flexDirection: 'column', gap: 6,
    }}>
      {toasts.map(t => {
        const Icon = t.type === 'success' ? CheckCircle : t.type === 'error' ? AlertCircle : Info
        const color = t.type === 'success' ? 'var(--success)' : t.type === 'error' ? 'var(--error)' : 'var(--accent)'
        return (
          <div key={t.id} style={{
            background: 'var(--bg-panel)', border: '1px solid var(--border)',
            borderLeft: `3px solid ${color}`, borderRadius: 8,
            padding: '8px 12px', display: 'flex', alignItems: 'center', gap: 8,
            boxShadow: '0 8px 24px rgba(0,0,0,0.3)', fontSize: 12,
            color: 'var(--fg-primary)', minWidth: 240, maxWidth: 360,
            animation: 'slideIn 0.2s ease-out',
          }}>
            <Icon size={14} style={{ color, flexShrink: 0 }} />
            <span style={{ flex: 1 }}>{t.message}</span>
            <button onClick={() => setToasts(prev => prev.filter(x => x.id !== t.id))} style={{
              background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-dim)', padding: 2
            }}><X size={12} /></button>
          </div>
        )
      })}
    </div>
  )
}
