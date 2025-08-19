'use client'

import { 
  memo, 
  useMemo, 
  useCallback, 
  useState, 
  useEffect, 
  useRef,
  Suspense,
  lazy,
  forwardRef
} from 'react'
import { motion } from 'framer-motion'
import { LoadingSpinner } from './LoadingStates'

// Intersection Observer hook for lazy loading
export function useIntersectionObserver(
  options: IntersectionObserverInit = {}
) {
  const [isIntersecting, setIsIntersecting] = useState(false)
  const [hasIntersected, setHasIntersected] = useState(false)
  const elementRef = useRef<HTMLElement>(null)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsIntersecting(entry.isIntersecting)
        if (entry.isIntersecting && !hasIntersected) {
          setHasIntersected(true)
        }
      },
      {
        threshold: 0.1,
        rootMargin: '50px',
        ...options
      }
    )

    observer.observe(element)

    return () => {
      observer.unobserve(element)
    }
  }, [hasIntersected, options])

  return { elementRef, isIntersecting, hasIntersected }
}

// Lazy loading wrapper component
export function LazyLoad({ 
  children, 
  fallback = <LoadingSpinner />,
  className = '',
  once = true
}: {
  children: React.ReactNode
  fallback?: React.ReactNode
  className?: string
  once?: boolean
}) {
  const { elementRef, hasIntersected, isIntersecting } = useIntersectionObserver()
  
  const shouldRender = once ? hasIntersected : isIntersecting

  return (
    <div ref={elementRef} className={className}>
      {shouldRender ? children : fallback}
    </div>
  )
}

// Image with lazy loading and optimization
export const OptimizedImage = memo(forwardRef<HTMLImageElement, {
  src: string
  alt: string
  width?: number
  height?: number
  className?: string
  placeholder?: string
  onLoad?: () => void
  onError?: () => void
}>(({ 
  src, 
  alt, 
  width, 
  height, 
  className = '', 
  placeholder,
  onLoad,
  onError 
}, ref) => {
  const [isLoaded, setIsLoaded] = useState(false)
  const [hasError, setHasError] = useState(false)
  const { elementRef, hasIntersected } = useIntersectionObserver()

  const handleLoad = useCallback(() => {
    setIsLoaded(true)
    onLoad?.()
  }, [onLoad])

  const handleError = useCallback(() => {
    setHasError(true)
    onError?.()
  }, [onError])

  return (
    <div 
      ref={elementRef} 
      className={`relative overflow-hidden ${className}`}
      style={{ width, height }}
    >
      {/* Placeholder */}
      {!isLoaded && !hasError && (
        <div className="absolute inset-0 bg-gray-200 dark:bg-gray-700 animate-pulse flex items-center justify-center">
          {placeholder || <LoadingSpinner size="sm" />}
        </div>
      )}

      {/* Error state */}
      {hasError && (
        <div className="absolute inset-0 bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-gray-400">
          <span className="text-sm">Failed to load</span>
        </div>
      )}

      {/* Actual image */}
      {hasIntersected && (
        <img
          ref={ref}
          src={src}
          alt={alt}
          width={width}
          height={height}
          className={`transition-opacity duration-300 ${
            isLoaded ? 'opacity-100' : 'opacity-0'
          }`}
          onLoad={handleLoad}
          onError={handleError}
          loading="lazy"
        />
      )}
    </div>
  )
}))

OptimizedImage.displayName = 'OptimizedImage'

// Virtual scrolling for large lists
export function VirtualList<T>({
  items,
  itemHeight,
  containerHeight,
  renderItem,
  className = '',
  overscan = 5
}: {
  items: T[]
  itemHeight: number
  containerHeight: number
  renderItem: (item: T, index: number) => React.ReactNode
  className?: string
  overscan?: number
}) {
  const [scrollTop, setScrollTop] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)

  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan)
  const endIndex = Math.min(
    items.length - 1,
    Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
  )

  const visibleItems = useMemo(() => {
    return items.slice(startIndex, endIndex + 1).map((item, index) => ({
      item,
      index: startIndex + index
    }))
  }, [items, startIndex, endIndex])

  const totalHeight = items.length * itemHeight

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop)
  }, [])

  return (
    <div
      ref={containerRef}
      className={`overflow-auto ${className}`}
      style={{ height: containerHeight }}
      onScroll={handleScroll}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        {visibleItems.map(({ item, index }) => (
          <div
            key={index}
            style={{
              position: 'absolute',
              top: index * itemHeight,
              height: itemHeight,
              width: '100%'
            }}
          >
            {renderItem(item, index)}
          </div>
        ))}
      </div>
    </div>
  )
}

// Debounced input hook
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

// Memoized search hook
export function useSearch<T>(
  items: T[],
  searchTerm: string,
  searchFields: (keyof T)[],
  options: {
    caseSensitive?: boolean
    debounceMs?: number
  } = {}
) {
  const { caseSensitive = false, debounceMs = 300 } = options
  const debouncedSearchTerm = useDebounce(searchTerm, debounceMs)

  return useMemo(() => {
    if (!debouncedSearchTerm.trim()) return items

    const term = caseSensitive 
      ? debouncedSearchTerm 
      : debouncedSearchTerm.toLowerCase()

    return items.filter(item =>
      searchFields.some(field => {
        const value = String(item[field])
        const searchValue = caseSensitive ? value : value.toLowerCase()
        return searchValue.includes(term)
      })
    )
  }, [items, debouncedSearchTerm, searchFields, caseSensitive])
}

// Performance monitoring hook
export function usePerformanceMonitor(name: string) {
  const startTime = useRef<number>()
  const [metrics, setMetrics] = useState<{
    duration?: number
    memory?: number
  }>({})

  useEffect(() => {
    startTime.current = performance.now()

    return () => {
      if (startTime.current) {
        const duration = performance.now() - startTime.current
        
        // Get memory usage if available
        const memory = (performance as any).memory?.usedJSHeapSize

        setMetrics({ duration, memory })

        // Log in development
        if (process.env.NODE_ENV === 'development') {
          console.log(`Performance [${name}]:`, {
            duration: `${duration.toFixed(2)}ms`,
            memory: memory ? `${(memory / 1024 / 1024).toFixed(2)}MB` : 'N/A'
          })
        }
      }
    }
  }, [name])

  return metrics
}

// Memoized component wrapper
export function withMemo<P extends object>(
  Component: React.ComponentType<P>,
  areEqual?: (prevProps: P, nextProps: P) => boolean
) {
  const MemoizedComponent = memo(Component, areEqual)
  MemoizedComponent.displayName = `withMemo(${Component.displayName || Component.name})`
  return MemoizedComponent
}

// Bundle splitting helper
export function createLazyComponent<T extends React.ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
  fallback?: React.ReactNode
) {
  const LazyComponent = lazy(importFn)
  
  return function WrappedLazyComponent(props: React.ComponentProps<T>) {
    return (
      <Suspense fallback={fallback || <LoadingSpinner />}>
        <LazyComponent {...props} />
      </Suspense>
    )
  }
}

// Resource preloader
export function usePreloader(resources: string[]) {
  const [loadedResources, setLoadedResources] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(false)

  const preload = useCallback(async () => {
    setIsLoading(true)
    
    const promises = resources.map(async (resource) => {
      if (loadedResources.has(resource)) return resource

      try {
        if (resource.endsWith('.js')) {
          await import(resource)
        } else if (resource.match(/\.(jpg|jpeg|png|gif|webp)$/i)) {
          return new Promise((resolve, reject) => {
            const img = new Image()
            img.onload = () => resolve(resource)
            img.onerror = reject
            img.src = resource
          })
        } else {
          await fetch(resource)
        }
        
        setLoadedResources(prev => new Set(prev).add(resource))
        return resource
      } catch (error) {
        console.warn(`Failed to preload resource: ${resource}`, error)
        throw error
      }
    })

    try {
      await Promise.allSettled(promises)
    } finally {
      setIsLoading(false)
    }
  }, [resources, loadedResources])

  return { preload, loadedResources, isLoading }
}

// Optimized animation wrapper
export function OptimizedMotion({
  children,
  reduceMotion = false,
  ...motionProps
}: {
  children: React.ReactNode
  reduceMotion?: boolean
  [key: string]: any
}) {
  // Disable animations if user prefers reduced motion
  const prefersReducedMotion = typeof window !== 'undefined' 
    ? window.matchMedia('(prefers-reduced-motion: reduce)').matches 
    : false

  if (reduceMotion || prefersReducedMotion) {
    return <div>{children}</div>
  }

  return <motion.div {...motionProps}>{children}</motion.div>
}
