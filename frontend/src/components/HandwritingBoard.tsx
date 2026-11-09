// PoC: handwriting-style ink rendered client-side with perfect-freehand.
// Kept snapshot-based (no realtime) to match the v1 design.
import { useCallback, useEffect, useRef, useState } from 'react'
import type { PointerEvent as ReactPointerEvent } from 'react'
import { getStroke } from 'perfect-freehand'
import './HandwritingBoard.css'

type Point = [number, number, number]

type Stroke = {
  points: Point[]
  color: string
  size: number
  eraser: boolean
  simulatePressure: boolean
}

type Tool = 'pen' | 'eraser'

const getOutlinePath = (points: number[][]): Path2D => {
  const path = new Path2D()
  if (points.length === 0) return path
  path.moveTo(points[0][0], points[0][1])
  for (let i = 0; i < points.length; i++) {
    const [x0, y0] = points[i]
    const [x1, y1] = points[(i + 1) % points.length]
    path.quadraticCurveTo(x0, y0, (x0 + x1) / 2, (y0 + y1) / 2)
  }
  path.closePath()
  return path
}

export default function HandwritingBoard() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const strokesRef = useRef<Stroke[]>([])
  const currentStrokeRef = useRef<Stroke | null>(null)

  const [tool, setTool] = useState<Tool>('pen')
  const [color, setColor] = useState('#111827')
  const [size, setSize] = useState(8)
  const [hasContent, setHasContent] = useState(false)
  const [snapshot, setSnapshot] = useState<string | null>(null)

  const render = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = window.devicePixelRatio || 1
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)

    const strokes = currentStrokeRef.current
      ? [...strokesRef.current, currentStrokeRef.current]
      : strokesRef.current

    for (const stroke of strokes) {
      const outline = getStroke(stroke.points, {
        size: stroke.size,
        thinning: 0.65,
        smoothing: 0.5,
        streamline: 0.5,
        simulatePressure: stroke.simulatePressure,
        start: { taper: 0, cap: true },
        end: { taper: 0, cap: true },
      })
      ctx.globalCompositeOperation = stroke.eraser
        ? 'destination-out'
        : 'source-over'
      ctx.fillStyle = stroke.eraser ? '#000' : stroke.color
      ctx.fill(getOutlinePath(outline))
    }
    ctx.globalCompositeOperation = 'source-over'
  }, [])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const resize = () => {
      const dpr = window.devicePixelRatio || 1
      const rect = canvas.getBoundingClientRect()
      canvas.width = Math.max(1, Math.round(rect.width * dpr))
      canvas.height = Math.max(1, Math.round(rect.height * dpr))
      render()
    }

    resize()
    const observer = new ResizeObserver(resize)
    observer.observe(canvas)
    return () => observer.disconnect()
  }, [render])

  const pointFromEvent = (
    event: PointerEvent,
    rect: DOMRect,
    pointerType: string,
  ): Point => [
    event.clientX - rect.left,
    event.clientY - rect.top,
    pointerType === 'pen' ? event.pressure : 0,
  ]

  const handlePointerDown = (event: ReactPointerEvent<HTMLCanvasElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    const eraser = tool === 'eraser'
    currentStrokeRef.current = {
      points: [
        pointFromEvent(
          event.nativeEvent,
          event.currentTarget.getBoundingClientRect(),
          event.pointerType,
        ),
      ],
      color,
      size: eraser ? size * 2.5 : size,
      eraser,
      simulatePressure: event.pointerType !== 'pen',
    }
    setHasContent(true)
    setSnapshot(null)
    render()
  }

  const handlePointerMove = (event: ReactPointerEvent<HTMLCanvasElement>) => {
    const stroke = currentStrokeRef.current
    if (!stroke) return

    const native = event.nativeEvent
    const events =
      typeof native.getCoalescedEvents === 'function'
        ? native.getCoalescedEvents()
        : [native]
    const rect = event.currentTarget.getBoundingClientRect()

    for (const coalesced of events) {
      stroke.points.push(
        pointFromEvent(coalesced, rect, event.pointerType),
      )
    }
    render()
  }

  const handlePointerUp = (event: ReactPointerEvent<HTMLCanvasElement>) => {
    const stroke = currentStrokeRef.current
    if (!stroke) return

    strokesRef.current = [...strokesRef.current, stroke]
    currentStrokeRef.current = null
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }
    render()
  }

  const handleUndo = () => {
    strokesRef.current = strokesRef.current.slice(0, -1)
    setHasContent(strokesRef.current.length > 0)
    setSnapshot(null)
    render()
  }

  const handleClear = () => {
    strokesRef.current = []
    currentStrokeRef.current = null
    setHasContent(false)
    setSnapshot(null)
    render()
  }

  const handleSubmit = () => {
    const canvas = canvasRef.current
    if (!canvas) return

    const output = document.createElement('canvas')
    output.width = canvas.width
    output.height = canvas.height
    const ctx = output.getContext('2d')
    if (!ctx) return

    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, output.width, output.height)
    ctx.drawImage(canvas, 0, 0)

    const image = output.toDataURL('image/png')
    setSnapshot(image)
    console.info('[PoC] snapshot ready for backend:', image.length, 'bytes')
  }

  return (
    <section className="poc-board">
      <header className="poc-board__header">
        <div>
          <h1>
            Handwriting board <span className="poc-board__badge">PoC</span>
          </h1>
          <p>
            Pressure-aware ink via <code>perfect-freehand</code>. Everything is
            local; only the PNG snapshot is sent.
          </p>
        </div>
        <div className="poc-board__actions">
          <button
            type="button"
            className="poc-board__button"
            aria-pressed={tool === 'pen'}
            onClick={() => setTool('pen')}
          >
            Pen
          </button>
          <button
            type="button"
            className="poc-board__button"
            aria-pressed={tool === 'eraser'}
            onClick={() => setTool('eraser')}
          >
            Eraser
          </button>
          <label className="poc-board__color">
            <span>Color</span>
            <input
              type="color"
              value={color}
              onChange={(event) => setColor(event.target.value)}
            />
          </label>
          <label className="poc-board__size">
            <span>Size</span>
            <input
              type="range"
              min={2}
              max={30}
              value={size}
              onChange={(event) => setSize(Number(event.target.value))}
            />
          </label>
          <button
            type="button"
            className="poc-board__button"
            onClick={handleUndo}
            disabled={!hasContent}
          >
            Undo
          </button>
          <button
            type="button"
            className="poc-board__button"
            onClick={handleClear}
            disabled={!hasContent}
          >
            Clear
          </button>
          <button
            type="button"
            className="poc-board__button poc-board__button--primary"
            onClick={handleSubmit}
            disabled={!hasContent}
          >
            Submit
          </button>
        </div>
      </header>

      <div className="poc-board__surface">
        <canvas
          ref={canvasRef}
          className="poc-board__canvas"
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onPointerCancel={handlePointerUp}
        />
      </div>

      {snapshot && (
        <div className="poc-board__snapshot">
          <img src={snapshot} alt="Captured handwriting snapshot" />
          <p>PNG snapshot ready. This is what a backend submission would receive.</p>
        </div>
      )}
    </section>
  )
}
