import { useState, useEffect } from 'react'
import { Settings, X, Save, Check } from 'lucide-react'
import { GetSettings, SaveSettings } from '../../wailsjs/go/main/App'

interface Props {
  open: boolean
  onClose: () => void
}

export default function SettingsPanel({ open, onClose }: Props) {
  const [baseURL, setBaseURL] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [model, setModel] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (open) {
      GetSettings().then(s => {
        setBaseURL(s.baseURL || '')
        setApiKey(s.apiKey || '')
        setModel(s.model || '')
        setSaved(false)
      }).catch(() => {})
    }
  }, [open])

  async function handleSave() {
    try {
      await SaveSettings(baseURL, apiKey, model)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (err) {
      alert('Save failed: ' + err)
    }
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
        borderRadius: 12, padding: 24, width: 440,
        boxShadow: '0 16px 48px rgba(0,0,0,0.3)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Settings size={16} style={{ color: 'var(--accent)' }} />
            <span style={{ fontWeight: 600, fontSize: 15 }}>Settings</span>
          </div>
          <button onClick={onClose} style={{
            background: 'none', border: 'none', cursor: 'pointer', color: 'var(--fg-muted)', padding: 4
          }}><X size={16} /></button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Field label="API Base URL" value={baseURL} onChange={setBaseURL} placeholder="https://api.novita.ai/openai" />
          <Field label="API Key" value={apiKey} onChange={setApiKey} placeholder="sk-..." type="password" />
          <Field label="Model" value={model} onChange={setModel} placeholder="qwen/qwen3-coder-30b-a3b-instruct" />

          <button onClick={handleSave} style={{
            padding: '10px 16px', borderRadius: 8, border: 'none', cursor: 'pointer',
            background: saved ? 'var(--success)' : 'var(--accent)',
            color: '#fff', fontWeight: 600, fontSize: 13, fontFamily: 'var(--font-ui)',
            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
            transition: 'background 0.2s',
          }}>
            {saved ? <><Check size={14} /> Saved</> : <><Save size={14} /> Save & Reconnect</>}
          </button>
        </div>

        <div style={{ marginTop: 16, fontSize: 11, color: 'var(--fg-dim)', lineHeight: 1.6 }}>
          Settings are saved to ~/.tgc/config.yaml (shared with TUI).
        </div>
      </div>
    </div>
  )
}

function Field({ label, value, onChange, placeholder, type }: {
  label: string; value: string; onChange: (v: string) => void; placeholder?: string; type?: string
}) {
  return (
    <div>
      <label style={{ fontSize: 11, fontWeight: 600, color: 'var(--fg-muted)', marginBottom: 4, display: 'block' }}>{label}</label>
      <input
        value={value} onChange={e => onChange(e.target.value)}
        placeholder={placeholder} type={type}
        style={{
          width: '100%', padding: '8px 12px', borderRadius: 6,
          background: 'var(--bg-base)', border: '1px solid var(--border)',
          color: 'var(--fg-primary)', fontSize: 13, fontFamily: 'var(--font-code)',
          outline: 'none', transition: 'border-color 0.2s',
        }}
        onFocus={e => e.target.style.borderColor = 'var(--accent)'}
        onBlur={e => e.target.style.borderColor = 'var(--border)'}
      />
    </div>
  )
}
