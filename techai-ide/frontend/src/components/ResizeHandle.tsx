import { useRef, useCallback } from 'react'

interface Props {
  direction: 'horizontal' | 'vertical'
  onResize: (delta: number) => void
}

export default function ResizeHandle({ direction, onResize }: Props) {
  const dragging = useRef(false)
  const lastPos = useRef(0)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    dragging.current = true
    lastPos.current = direction === 'horizontal' ? e.clientX : e.clientY

    const handleMouseMove = (e: MouseEvent) => {
      if (!dragging.current) return
      const pos = direction === 'horizontal' ? e.clientX : e.clientY
      const delta = pos - lastPos.current
      lastPos.current = pos
      onResize(delta)
    }

    const handleMouseUp = () => {
      dragging.current = false
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
    document.body.style.cursor = direction === 'horizontal' ? 'col-resize' : 'row-resize'
    document.body.style.userSelect = 'none'
  }, [direction, onResize])

  return (
    <div
      onMouseDown={handleMouseDown}
      style={{
        ...(direction === 'horizontal'
          ? { width: 6, cursor: 'col-resize', borderLeft: '1px solid var(--border)' }
          : { height: 6, cursor: 'row-resize', borderTop: '1px solid var(--border)' }),
        flexShrink: 0,
        background: 'transparent',
        transition: 'background 0.15s',
      }}
      onMouseEnter={e => (e.currentTarget.style.background = 'var(--accent-glow)')}
      onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
    />
  )
}
