import { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { CodeIcon, Cross2Icon, DotFilledIcon } from '@radix-ui/react-icons'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useDevicesContext } from './devices-provider'
import { getErrorMessage } from '../lib/formatters'

export function DeviceTerminalDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useDevicesContext()
  const isTerminalOpen = open === 'terminal' && Boolean(currentRow)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const terminalRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState<boolean>(false)

  useEffect(() => {
    if (!isTerminalOpen || !currentRow) return

    let isSubscribed = true
    let term: Terminal | null = null
    let fitAddon: FitAddon | null = null
    let ws: WebSocket | null = null
    let resizeObserver: ResizeObserver | null = null

    // Give DOM a microtask to attach containerRef element inside Dialog
    const timer = setTimeout(() => {
      if (!isSubscribed || !containerRef.current) return

      // Clean up previous
      if (terminalRef.current) {
        terminalRef.current.dispose()
        terminalRef.current = null
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }

      // Initialize Xterm
      term = new Terminal({
        cursorBlink: true,
        fontFamily: 'ui-monospace, "JetBrains Mono", "SFMono-Regular", Consolas, monospace',
        fontSize: 13,
        rows: 24,
        cols: 80,
        theme: {
          background: '#090d16',
          foreground: '#dce6e4',
          cursor: '#2fb8ac',
          selectionBackground: 'rgba(47, 184, 172, 0.3)',
        },
      })

      fitAddon = new FitAddon()
      term.loadAddon(fitAddon)

      containerRef.current.innerHTML = ''
      term.open(containerRef.current)

      try {
        fitAddon.fit()
      } catch {
        // ignore initial fit error before mount
      }

      terminalRef.current = term
      fitAddonRef.current = fitAddon

      // Intro banner
      term.writeln(`\x1b[36m=== RouterOS / SSH Terminal ===\x1b[0m`)
      term.writeln(
        `Connecting SSH to \x1b[33m${currentRow.name}\x1b[0m (${currentRow.host}:${
          currentRow.sshPort || 22
        })...\r\n`
      )

      // Connect WebSocket
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const host =
        window.location.port === '5173'
          ? `${window.location.hostname}:8080`
          : window.location.host
      const wsUrl = `${proto}://${host}/ws/devices/${currentRow.id}/terminal`

      try {
        ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          if (!isSubscribed) return
          setIsConnected(true)
          term?.writeln(`\x1b[32m[Connected to ${currentRow.name}]\x1b[0m\r\n`)
          term?.focus()
          if (term && fitAddon) {
            try {
              fitAddon.fit()
            } catch {
              // ignore
            }
          }
        }

        ws.onmessage = async (ev) => {
          if (!term) return
          if (typeof ev.data === 'string') {
            term.write(ev.data)
          } else if (ev.data instanceof Blob) {
            const text = await ev.data.text()
            term.write(text)
          } else if (ev.data instanceof ArrayBuffer) {
            term.write(new Uint8Array(ev.data))
          }
        }

        ws.onerror = () => {
          if (!isSubscribed) return
          setIsConnected(false)
          term?.writeln(
            `\r\n\x1b[31m[WebSocket Connection Error - Check if backend is running on port 8080]\x1b[0m`
          )
        }

        ws.onclose = () => {
          if (!isSubscribed) return
          setIsConnected(false)
          term?.writeln(`\r\n\x1b[31m[SSH / RouterOS Terminal Session Closed]\x1b[0m`)
        }
      } catch (err: unknown) {
        term.writeln(
          `\x1b[31m[Failed to connect: ${getErrorMessage(err, 'Connection error')}]\x1b[0m\r\n`
        )
        setIsConnected(false)
      }

      // Input handling
      const dataDisposable = term.onData((data) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(data)
        }
      })

      // Resize observer
      resizeObserver = new ResizeObserver(() => {
        if (fitAddon && term) {
          try {
            fitAddon.fit()
            if (ws && ws.readyState === WebSocket.OPEN) {
              ws.send(
                JSON.stringify({
                  type: 'resize',
                  cols: term.cols,
                  rows: term.rows,
                })
              )
            }
          } catch {
            // ignore
          }
        }
      })

      if (containerRef.current) {
        resizeObserver.observe(containerRef.current)
      }

      return () => {
        dataDisposable.dispose()
      }
    }, 50)

    return () => {
      isSubscribed = false
      clearTimeout(timer)
      if (resizeObserver) resizeObserver.disconnect()
      if (ws) ws.close()
      if (term) term.dispose()
      terminalRef.current = null
      wsRef.current = null
    }
  }, [isTerminalOpen, currentRow])

  return (
    <Dialog
      open={isTerminalOpen}
      onOpenChange={() => {
        setOpen(null)
        setCurrentRow(null)
      }}
    >
      <DialogContent className='sm:max-w-3xl bg-[#090d16] text-white border-slate-800 p-0 overflow-hidden shadow-2xl'>
        <DialogHeader className='flex flex-row items-center justify-between px-4 py-3 border-b border-slate-800 space-y-0 bg-[#0c121e]'>
          <DialogTitle className='flex items-center gap-2 text-sm font-mono font-medium text-slate-200'>
            <span
              className={`inline-flex items-center ${
                isConnected ? 'text-emerald-400' : 'text-slate-500'
              }`}
            >
              <DotFilledIcon className='h-4 w-4' />
            </span>
            <CodeIcon className='h-4 w-4 text-emerald-400' />
            <span>
              {currentRow?.name} ({currentRow?.host}:{currentRow?.sshPort || 22}) &gt;_
            </span>
          </DialogTitle>

          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7 text-slate-400 hover:text-white hover:bg-slate-800'
            onClick={() => {
              setOpen(null)
              setCurrentRow(null)
            }}
          >
            <Cross2Icon className='h-4 w-4' />
          </Button>
        </DialogHeader>

        <div className='p-3 min-h-[380px] bg-[#090d16] font-mono text-xs'>
          <div ref={containerRef} className='h-[360px] w-full block' />
        </div>
      </DialogContent>
    </Dialog>
  )
}
