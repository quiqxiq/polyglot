import { useEffect, useRef } from 'react'
import type { TrafficDataPoint } from '../../types'
import { formatBps } from '../../lib/formatters'

interface DeviceTrafficChartProps {
  trafficHistory: TrafficDataPoint[]
  maxSamples?: number
  className?: string
}

export function DeviceTrafficChart({
  trafficHistory,
  maxSamples = 60,
  className = 'w-full h-full block',
}: DeviceTrafficChartProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.max(120, Math.round(rect.width || 280))
    const h = Math.max(50, Math.round(rect.height || 96))

    if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
      canvas.width = w * dpr
      canvas.height = h * dpr
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)

    if (trafficHistory.length < 2) {
      ctx.fillStyle = '#94a3b8'
      ctx.font = '11px ui-sans-serif, system-ui, sans-serif'
      ctx.textAlign = 'center'
      ctx.fillText('Menghubungkan ke traffic stream...', w / 2, h / 2 + 4)
      return
    }

    const padTop = 14
    const padBottom = 6
    const padX = 4
    const chartH = h - padTop - padBottom

    // Find max peak
    const actualMax = Math.max(0, ...trafficHistory.map((s) => Math.max(s.rx, s.tx)))
    const scale = Math.max(1000, actualMax * 1.1)
    const latestTime = trafficHistory[trafficHistory.length - 1]?.time ?? Date.now()
    const firstTime = trafficHistory[0]?.time ?? latestTime
    const sampleInterval = Math.max(
      1000,
      trafficHistory.length > 1
        ? (latestTime - firstTime) / (trafficHistory.length - 1)
        : 1000
    )
    const windowMs = sampleInterval * Math.max(maxSamples - 1, 1)
    const xFor = (time: number) =>
      padX + Math.max(0, Math.min(1, (time - (latestTime - windowMs)) / windowMs)) * (w - padX * 2)

    // Draw subtle grid lines
    ctx.strokeStyle = 'rgba(148, 163, 184, 0.12)'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(padX, padTop)
    ctx.lineTo(w - padX, padTop)
    ctx.moveTo(padX, padTop + chartH / 2)
    ctx.lineTo(w - padX, padTop + chartH / 2)
    ctx.moveTo(padX, h - padBottom)
    ctx.lineTo(w - padX, h - padBottom)
    ctx.stroke()

    // Draw Line & Gradient Area helper
    const drawSeries = (
      key: 'rx' | 'tx',
      strokeColor: string,
      fillColorTop: string,
      fillColorBottom: string
    ) => {
      if (trafficHistory.length === 0) return

      // Gradient Fill
      const grad = ctx.createLinearGradient(0, padTop, 0, h - padBottom)
      grad.addColorStop(0, fillColorTop)
      grad.addColorStop(1, fillColorBottom)

      ctx.beginPath()
       const firstX = xFor(trafficHistory[0].time)
      ctx.moveTo(firstX, h - padBottom)

       trafficHistory.forEach((s) => {
         const x = xFor(s.time)
        const y = h - padBottom - (s[key] / scale) * chartH
        ctx.lineTo(x, y)
      })

       const lastX = xFor(trafficHistory[trafficHistory.length - 1].time)
      ctx.lineTo(lastX, h - padBottom)
      ctx.closePath()
      ctx.fillStyle = grad
      ctx.fill()

      // Stroke line
      ctx.beginPath()
      trafficHistory.forEach((s, i) => {
         const x = xFor(s.time)
        const y = h - padBottom - (s[key] / scale) * chartH
        if (i === 0) ctx.moveTo(x, y)
        else ctx.lineTo(x, y)
      })
      ctx.strokeStyle = strokeColor
      ctx.lineWidth = 1.75
      ctx.lineJoin = 'round'
      ctx.stroke()
    }

    // Render RX (Cyan / Emerald) & TX (Amber / Gold)
    drawSeries(
      'rx',
      '#2fb8ac',
      'rgba(47, 184, 172, 0.25)',
      'rgba(47, 184, 172, 0.0)'
    )
    drawSeries(
      'tx',
      '#e8a33d',
      'rgba(232, 163, 61, 0.2)',
      'rgba(232, 163, 61, 0.0)'
    )

    // Top Peak bandwidth text
    ctx.textAlign = 'left'
    ctx.fillStyle = '#64748b'
    ctx.font = '10px ui-monospace, monospace'
    ctx.fillText(`${formatBps(actualMax)} peak`, padX + 2, 10)
  }, [trafficHistory, maxSamples])

  return <canvas ref={canvasRef} className={className} />
}
