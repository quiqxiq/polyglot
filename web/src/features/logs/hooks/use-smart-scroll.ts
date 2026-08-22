import { useEffect, useRef, useState, useCallback } from 'react'

interface UseSmartScrollOptions {
  dependencies: unknown[]
  isAutoScrollEnabled?: boolean
  threshold?: number // Distance in px from bottom to consider "at bottom" (default: 80)
}

export function useSmartScroll({
  dependencies,
  isAutoScrollEnabled = true,
  threshold = 80,
}: UseSmartScrollOptions) {
  const containerRef = useRef<HTMLDivElement>(null)
  const userScrolledUpRef = useRef<boolean>(false)
  const isProgrammaticScrollRef = useRef<boolean>(false)
  const scrollLockTimeoutRef = useRef<number | null>(null)
  const hasMountedRef = useRef(false)
  const [showScrollBottomBtn, setShowScrollBottomBtn] = useState(false)

  // Recompute "am I at the bottom?" and sync state to match
  const evaluatePosition = useCallback(() => {
    const el = containerRef.current
    if (!el) return

    const { scrollTop, scrollHeight, clientHeight } = el
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight

    if (distanceFromBottom > threshold) {
      userScrolledUpRef.current = true
      setShowScrollBottomBtn(true)
    } else {
      userScrolledUpRef.current = false
      setShowScrollBottomBtn(false)
    }
  }, [threshold])

  // Listen to manual user scroll
  const handleScroll = useCallback(() => {
    if (!containerRef.current) return

    // A *smooth* scrollTo fires MANY 'scroll' events over its whole
    // animation (one per frame), not just one. So while we're auto
    // scrolling programmatically, ignore every event until it's done -
    // otherwise frame #2 onward (mid-animation, not yet at the bottom)
    // gets misread as "the user scrolled up".
    if (isProgrammaticScrollRef.current) return

    evaluatePosition()
  }, [evaluatePosition])

  // Scroll to bottom (used by "Jump to latest" button and by auto-scroll)
  const scrollToBottom = useCallback((smooth = false) => {
    const el = containerRef.current
    if (!el) return

    if (scrollLockTimeoutRef.current !== null) {
      window.clearTimeout(scrollLockTimeoutRef.current)
    }

    isProgrammaticScrollRef.current = true
    userScrolledUpRef.current = false
    setShowScrollBottomBtn(false)

    requestAnimationFrame(() => {
      el.scrollTo({
        top: el.scrollHeight,
        behavior: smooth ? 'smooth' : 'instant',
      })
    })

    // Only release the lock once the animation has realistically
    // finished, then re-check where we actually landed - this covers the
    // case where the user grabs the scrollbar mid-animation.
    scrollLockTimeoutRef.current = window.setTimeout(() => {
      isProgrammaticScrollRef.current = false
      evaluatePosition()
    }, smooth ? 600 : 50)
  }, [evaluatePosition])

  // Explicitly toggling auto-scroll ON -> snap to bottom immediately
  useEffect(() => {
    if (!containerRef.current) return
    if (isAutoScrollEnabled) {
      scrollToBottom(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAutoScrollEnabled])

  // Auto-scroll when the tracked dependencies (e.g. the logs array) update
  useEffect(() => {
    if (!containerRef.current) return

    // 1. Auto-scroll OFF -> do absolutely nothing, no matter what changed.
    if (!isAutoScrollEnabled) return

    // 2. User is reading older logs -> don't yank them back down.
    if (userScrolledUpRef.current) return

    // 3. Auto-scroll ON and user is at the bottom -> follow the new logs.
    //    Snap on the very first run (mount / initial batch load) so a big
    //    initial dump of logs doesn't slowly animate past; smooth after
    //    that for each incremental update.
    scrollToBottom(hasMountedRef.current)
    hasMountedRef.current = true
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...dependencies, isAutoScrollEnabled])

  // Clean up any pending timer on unmount
  useEffect(() => {
    return () => {
      if (scrollLockTimeoutRef.current !== null) {
        window.clearTimeout(scrollLockTimeoutRef.current)
      }
    }
  }, [])

  return {
    containerRef,
    handleScroll,
    scrollToBottom,
    showScrollBottomBtn,
  }
}