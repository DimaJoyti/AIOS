// Performance monitoring and optimization utilities

interface PerformanceMetrics {
  loadTime: number
  domContentLoaded: number
  firstPaint: number
  firstContentfulPaint: number
  largestContentfulPaint: number
  firstInputDelay: number
  cumulativeLayoutShift: number
  memoryUsage?: number
  connectionType?: string
}

interface VitalMetrics {
  lcp: number // Largest Contentful Paint
  fid: number // First Input Delay
  cls: number // Cumulative Layout Shift
}

class PerformanceMonitor {
  private metrics: Partial<PerformanceMetrics> = {}
  private observers: PerformanceObserver[] = []

  constructor() {
    if (typeof window !== 'undefined') {
      this.initializeMonitoring()
    }
  }

  private initializeMonitoring() {
    // Monitor navigation timing
    this.monitorNavigationTiming()
    
    // Monitor paint timing
    this.monitorPaintTiming()
    
    // Monitor layout shifts
    this.monitorLayoutShifts()
    
    // Monitor first input delay
    this.monitorFirstInputDelay()
    
    // Monitor largest contentful paint
    this.monitorLargestContentfulPaint()
    
    // Monitor memory usage
    this.monitorMemoryUsage()
    
    // Monitor connection
    this.monitorConnection()
  }

  private monitorNavigationTiming() {
    window.addEventListener('load', () => {
      setTimeout(() => {
        const perfData = performance.timing
        this.metrics.loadTime = perfData.loadEventEnd - perfData.navigationStart
        this.metrics.domContentLoaded = perfData.domContentLoadedEventEnd - perfData.navigationStart
        
        this.reportMetrics()
      }, 0)
    })
  }

  private monitorPaintTiming() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-paint') {
            this.metrics.firstPaint = entry.startTime
          } else if (entry.name === 'first-contentful-paint') {
            this.metrics.firstContentfulPaint = entry.startTime
          }
        }
      })
      
      observer.observe({ entryTypes: ['paint'] })
      this.observers.push(observer)
    }
  }

  private monitorLayoutShifts() {
    if ('PerformanceObserver' in window) {
      let clsValue = 0
      
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!(entry as any).hadRecentInput) {
            clsValue += (entry as any).value
          }
        }
        this.metrics.cumulativeLayoutShift = clsValue
      })
      
      observer.observe({ entryTypes: ['layout-shift'] })
      this.observers.push(observer)
    }
  }

  private monitorFirstInputDelay() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          this.metrics.firstInputDelay = (entry as any).processingStart - entry.startTime
        }
      })
      
      observer.observe({ entryTypes: ['first-input'] })
      this.observers.push(observer)
    }
  }

  private monitorLargestContentfulPaint() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const lastEntry = entries[entries.length - 1]
        this.metrics.largestContentfulPaint = lastEntry.startTime
      })
      
      observer.observe({ entryTypes: ['largest-contentful-paint'] })
      this.observers.push(observer)
    }
  }

  private monitorMemoryUsage() {
    if ('memory' in performance) {
      const memory = (performance as any).memory
      this.metrics.memoryUsage = memory.usedJSHeapSize
    }
  }

  private monitorConnection() {
    if ('connection' in navigator) {
      const connection = (navigator as any).connection
      this.metrics.connectionType = connection.effectiveType
    }
  }

  private reportMetrics() {
    const isDevelopment = typeof window !== 'undefined' &&
      (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

    if (isDevelopment) {
      console.group('🚀 Performance Metrics')
      console.log('Load Time:', this.metrics.loadTime + 'ms')
      console.log('DOM Content Loaded:', this.metrics.domContentLoaded + 'ms')
      console.log('First Paint:', this.metrics.firstPaint + 'ms')
      console.log('First Contentful Paint:', this.metrics.firstContentfulPaint + 'ms')
      console.log('Largest Contentful Paint:', this.metrics.largestContentfulPaint + 'ms')
      console.log('First Input Delay:', this.metrics.firstInputDelay + 'ms')
      console.log('Cumulative Layout Shift:', this.metrics.cumulativeLayoutShift)
      console.log('Memory Usage:', this.formatBytes(this.metrics.memoryUsage || 0))
      console.log('Connection Type:', this.metrics.connectionType)
      console.groupEnd()
    }

    // Send to analytics service in production
    const isProduction = typeof window !== 'undefined' &&
      window.location.hostname !== 'localhost' &&
      window.location.hostname !== '127.0.0.1'

    if (isProduction) {
      this.sendToAnalytics(this.metrics)
    }
  }

  private formatBytes(bytes: number): string {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  private sendToAnalytics(metrics: Partial<PerformanceMetrics>) {
    // Implementation for sending metrics to analytics service
    // Example: Google Analytics, Mixpanel, custom endpoint
    console.log('Sending metrics to analytics:', metrics)
  }

  public getVitalMetrics(): Partial<VitalMetrics> {
    return {
      lcp: this.metrics.largestContentfulPaint || 0,
      fid: this.metrics.firstInputDelay || 0,
      cls: this.metrics.cumulativeLayoutShift || 0
    }
  }

  public getMetrics(): Partial<PerformanceMetrics> {
    return { ...this.metrics }
  }

  public cleanup() {
    this.observers.forEach(observer => observer.disconnect())
    this.observers = []
  }
}

// Resource loading optimization
export class ResourceOptimizer {
  private static preloadedResources = new Set<string>()

  static preloadResource(url: string, type: 'script' | 'style' | 'image' | 'font' = 'script') {
    if (this.preloadedResources.has(url)) return

    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = url
    
    switch (type) {
      case 'script':
        link.as = 'script'
        break
      case 'style':
        link.as = 'style'
        break
      case 'image':
        link.as = 'image'
        break
      case 'font':
        link.as = 'font'
        link.crossOrigin = 'anonymous'
        break
    }

    document.head.appendChild(link)
    this.preloadedResources.add(url)
  }

  static preloadCriticalResources() {
    // Preload critical fonts
    this.preloadResource('/fonts/inter-var.woff2', 'font')
    
    // Preload critical images
    this.preloadResource('/logo.svg', 'image')
    this.preloadResource('/og-image.png', 'image')
  }

  static optimizeImages() {
    // Lazy load images that are not in viewport
    const images = document.querySelectorAll('img[data-src]')
    
    if ('IntersectionObserver' in window) {
      const imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            const img = entry.target as HTMLImageElement
            img.src = img.dataset.src!
            img.removeAttribute('data-src')
            imageObserver.unobserve(img)
          }
        })
      })

      images.forEach(img => imageObserver.observe(img))
    }
  }
}

// Bundle size analyzer
export class BundleAnalyzer {
  static analyzeChunks() {
    const isDevelopment = typeof window !== 'undefined' &&
      (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

    if (isDevelopment) {
      // Analyze webpack chunks
      const chunks = (window as any).__webpack_require__?.cache
      if (chunks) {
        const chunkSizes = Object.keys(chunks).map(key => ({
          id: key,
          size: JSON.stringify(chunks[key]).length
        }))

        console.table(chunkSizes.sort((a, b) => b.size - a.size))
      }
    }
  }

  static trackComponentRenders() {
    const isDevelopment = typeof window !== 'undefined' &&
      (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

    if (isDevelopment) {
      // Track React component renders
      const originalConsoleLog = console.log
      let renderCount = 0

      console.log = (...args) => {
        if (args[0]?.includes?.('render')) {
          renderCount++
        }
        originalConsoleLog.apply(console, args)
      }

      setInterval(() => {
        if (renderCount > 0) {
          console.warn(`🔄 ${renderCount} renders in the last second`)
          renderCount = 0
        }
      }, 1000)
    }
  }
}

// Initialize performance monitoring
export const performanceMonitor = new PerformanceMonitor()

// Export utilities
export { PerformanceMonitor, ResourceOptimizer, BundleAnalyzer }

// Auto-initialize optimizations
if (typeof window !== 'undefined') {
  // Preload critical resources
  ResourceOptimizer.preloadCriticalResources()
  
  // Optimize images when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      ResourceOptimizer.optimizeImages()
    })
  } else {
    ResourceOptimizer.optimizeImages()
  }
  
  // Analyze bundle in development
  const isDevelopment = typeof window !== 'undefined' &&
    (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

  if (isDevelopment) {
    BundleAnalyzer.analyzeChunks()
    BundleAnalyzer.trackComponentRenders()
  }
}

export default performanceMonitor
