/**
 * Environment detection utilities for client-side code
 * 
 * Since process.env is not available in the browser, we use alternative methods
 * to detect the environment based on the hostname and other browser APIs.
 */

// Cache the environment detection to avoid repeated calculations
let cachedEnvironment: 'development' | 'production' | 'test' | null = null

/**
 * Detects if the application is running in development mode
 * Based on hostname (localhost, 127.0.0.1) and other indicators
 */
export function isDevelopment(): boolean {
  if (typeof window === 'undefined') {
    // Server-side: use NODE_ENV if available
    return process.env.NODE_ENV === 'development'
  }

  // Client-side: check hostname
  const hostname = window.location.hostname
  return hostname === 'localhost' || 
         hostname === '127.0.0.1' || 
         hostname.startsWith('192.168.') ||
         hostname.endsWith('.local')
}

/**
 * Detects if the application is running in production mode
 */
export function isProduction(): boolean {
  if (typeof window === 'undefined') {
    // Server-side: use NODE_ENV if available
    return process.env.NODE_ENV === 'production'
  }

  // Client-side: opposite of development
  return !isDevelopment()
}

/**
 * Detects if the application is running in test mode
 */
export function isTest(): boolean {
  if (typeof window === 'undefined') {
    // Server-side: use NODE_ENV if available
    return process.env.NODE_ENV === 'test'
  }

  // Client-side: check for test indicators
  return !!(window as any).__TEST__ || 
         !!(window as any).jest ||
         !!(window as any).jasmine ||
         !!(window as any).mocha
}

/**
 * Gets the current environment
 */
export function getEnvironment(): 'development' | 'production' | 'test' {
  if (cachedEnvironment) {
    return cachedEnvironment
  }

  if (isTest()) {
    cachedEnvironment = 'test'
  } else if (isDevelopment()) {
    cachedEnvironment = 'development'
  } else {
    cachedEnvironment = 'production'
  }

  return cachedEnvironment
}

/**
 * Checks if debugging should be enabled
 */
export function isDebugMode(): boolean {
  if (typeof window === 'undefined') {
    return false
  }

  // Check for debug flags
  const urlParams = new URLSearchParams(window.location.search)
  const hasDebugParam = urlParams.has('debug') || urlParams.has('dev')
  const hasDebugStorage = localStorage.getItem('debug') === 'true'
  
  return isDevelopment() || hasDebugParam || hasDebugStorage
}

/**
 * Checks if analytics should be enabled
 */
export function shouldEnableAnalytics(): boolean {
  if (typeof window === 'undefined') {
    return false
  }

  // Don't track in development or if user has opted out
  const hasOptedOut = localStorage.getItem('analytics-opt-out') === 'true'
  const isDoNotTrack = navigator.doNotTrack === '1'
  
  return isProduction() && !hasOptedOut && !isDoNotTrack
}

/**
 * Checks if error reporting should be enabled
 */
export function shouldEnableErrorReporting(): boolean {
  if (typeof window === 'undefined') {
    return false
  }

  // Enable error reporting in production, but respect user privacy
  const hasOptedOut = localStorage.getItem('error-reporting-opt-out') === 'true'
  
  return isProduction() && !hasOptedOut
}

/**
 * Gets the application version from package.json or meta tag
 */
export function getAppVersion(): string {
  if (typeof window === 'undefined') {
    return 'unknown'
  }

  // Try to get version from meta tag
  const versionMeta = document.querySelector('meta[name="version"]')
  if (versionMeta) {
    return versionMeta.getAttribute('content') || 'unknown'
  }

  // Fallback to a default version
  return '1.0.0'
}

/**
 * Gets build information
 */
export function getBuildInfo(): {
  version: string
  environment: string
  buildTime?: string
  gitCommit?: string
} {
  const buildTime = typeof window !== 'undefined' 
    ? document.querySelector('meta[name="build-time"]')?.getAttribute('content')
    : undefined

  const gitCommit = typeof window !== 'undefined'
    ? document.querySelector('meta[name="git-commit"]')?.getAttribute('content')
    : undefined

  return {
    version: getAppVersion(),
    environment: getEnvironment(),
    buildTime,
    gitCommit
  }
}

/**
 * Console logging utility that respects environment
 */
export const logger = {
  debug: (...args: any[]) => {
    if (isDebugMode()) {
      console.debug('[DEBUG]', ...args)
    }
  },
  
  info: (...args: any[]) => {
    if (isDevelopment() || isDebugMode()) {
      console.info('[INFO]', ...args)
    }
  },
  
  warn: (...args: any[]) => {
    console.warn('[WARN]', ...args)
  },
  
  error: (...args: any[]) => {
    console.error('[ERROR]', ...args)
  },
  
  performance: (name: string, data: any) => {
    if (isDevelopment() || isDebugMode()) {
      console.group(`🚀 Performance [${name}]`)
      console.table(data)
      console.groupEnd()
    }
  }
}

/**
 * Feature flags based on environment
 */
export const features = {
  enablePerformanceMonitoring: isDevelopment() || isDebugMode(),
  enableDetailedLogging: isDevelopment() || isDebugMode(),
  enableAnalytics: shouldEnableAnalytics(),
  enableErrorReporting: shouldEnableErrorReporting(),
  enableDevTools: isDevelopment(),
  enableBundleAnalysis: isDevelopment(),
  enableA11yChecks: isDevelopment() || isDebugMode()
}

export default {
  isDevelopment,
  isProduction,
  isTest,
  getEnvironment,
  isDebugMode,
  shouldEnableAnalytics,
  shouldEnableErrorReporting,
  getAppVersion,
  getBuildInfo,
  logger,
  features
}
