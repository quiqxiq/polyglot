import { useEffect, useRef } from 'react'
import type { PingDataPoint } from '../../types'

interface DevicePingSparklineProps {
  pingHistory: PingDataPoint[]
  maxSamples?: number
  className?: string
}

export function DevicePingSparkline({
  pingHistory,
  maxSamples = 40,
  className = 'w-full h-full block',
}: DevicePingSparklineProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.max(80, Math.round(rect.width || 128))
    const h = Math.max(16, Math.round(rect.height || 28))

    if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
      canvas.width = w * dpr
      canvas.height = h * dpr
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)

    if (pingHistory.length === 0) {
      ctx.fillStyle = '#94a3b8'
      ctx.font = '10px ui-sans-serif, system-ui, sans-serif'
      ctx.fillText('Live ping...', 4, h / 2 + 3)
      return
    }

    const barW = w / maxSamples
    const maxMs = Math.max(40, ...pingHistory.map((p) => p.ms))

    pingHistory.forEach((p, i) => {
      const x = w - (pingHistory.length - i) * barW
      const barH = Math.max(2, Math.min(h, (p.ms / maxMs) * h))

      let color = '#2fb8ac' // Emerald/Teal normal
      if (p.ms > 100) {
        color = '#f43f5e' // Rose high latency
      } else if (p.ms > 50) {
        color = '#e8a33d' // Amber moderate latency
      }

      ctx.fillStyle = color
      ctx.globalAlpha = 0.35 + 0.65 * ((i + 1) / pingHistory.length)
      ctx.fillRect(x + 0.5, h - barH, Math.max(1, barW - 1.5), barH)
    })
    ctx.globalAlpha = 1
  }, [pingHistory, maxSamples])

  return <canvas ref={canvasRef} className={className} />
}

