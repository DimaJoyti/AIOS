'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { ExclamationTriangleIcon, ArrowPathIcon, HomeIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface State {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ error, errorInfo })

    // Log error to console in development
    const isDevelopment = typeof window !== 'undefined' &&
      (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

    if (isDevelopment) {
      console.error('ErrorBoundary caught an error:', error, errorInfo)
    }

    // Call custom error handler if provided
    this.props.onError?.(error, errorInfo)

    // In production, you might want to send this to an error reporting service
    // Example: Sentry.captureException(error, { contexts: { react: errorInfo } })
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  handleGoHome = () => {
    window.location.href = '/dashboard'
  }

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Default error UI
      return (
        <div className="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex items-center justify-center p-6">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-md w-full bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-xl border border-gray-200/50 dark:border-gray-700/50 p-8 text-center"
          >
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
              className="w-16 h-16 bg-gradient-to-r from-red-500 to-pink-600 rounded-full flex items-center justify-center mx-auto mb-6"
            >
              <ExclamationTriangleIcon className="w-8 h-8 text-white" />
            </motion.div>

            <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Oops! Something went wrong
            </h1>
            
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              We encountered an unexpected error. Don't worry, our team has been notified and we're working on a fix.
            </p>

            {typeof window !== 'undefined' &&
             (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') &&
             this.state.error && (
              <details className="mb-6 text-left">
                <summary className="cursor-pointer text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Error Details (Development)
                </summary>
                <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 text-xs font-mono text-gray-800 dark:text-gray-200 overflow-auto max-h-32">
                  <div className="text-red-600 dark:text-red-400 font-semibold mb-2">
                    {this.state.error.name}: {this.state.error.message}
                  </div>
                  <div className="whitespace-pre-wrap">
                    {this.state.error.stack}
                  </div>
                  {this.state.errorInfo && (
                    <div className="mt-4 pt-4 border-t border-gray-300 dark:border-gray-600">
                      <div className="text-blue-600 dark:text-blue-400 font-semibold mb-2">
                        Component Stack:
                      </div>
                      <div className="whitespace-pre-wrap">
                        {this.state.errorInfo.componentStack}
                      </div>
                    </div>
                  )}
                </div>
              </details>
            )}

            <div className="flex flex-col sm:flex-row gap-3">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={this.handleRetry}
                className="flex-1 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 shadow-lg flex items-center justify-center space-x-2"
              >
                <ArrowPathIcon className="w-4 h-4" />
                <span>Try Again</span>
              </motion.button>
              
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={this.handleGoHome}
                className="flex-1 px-6 py-3 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-600 transition-all duration-200 shadow-lg flex items-center justify-center space-x-2"
              >
                <HomeIcon className="w-4 h-4" />
                <span>Go Home</span>
              </motion.button>
            </div>

            <p className="text-xs text-gray-500 dark:text-gray-400 mt-6">
              Error ID: {Date.now().toString(36).toUpperCase()}
            </p>
          </motion.div>
        </div>
      )
    }

    return this.props.children
  }
}

// Hook version for functional components
export function useErrorHandler() {
  return (error: Error, errorInfo?: ErrorInfo) => {
    console.error('Error caught by useErrorHandler:', error, errorInfo)

    // In production, send to error reporting service
    const isProduction = typeof window !== 'undefined' &&
      window.location.hostname !== 'localhost' &&
      window.location.hostname !== '127.0.0.1'

    if (isProduction) {
      // Example: Sentry.captureException(error, { contexts: { react: errorInfo } })
    }
  }
}

// Higher-order component for wrapping components with error boundary
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  fallback?: ReactNode,
  onError?: (error: Error, errorInfo: ErrorInfo) => void
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary fallback={fallback} onError={onError}>
      <Component {...props} />
    </ErrorBoundary>
  )
  
  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`
  
  return WrappedComponent
}

// Lightweight error boundary for specific sections
export function SectionErrorBoundary({ 
  children, 
  title = "Section Error",
  description = "This section encountered an error."
}: {
  children: ReactNode
  title?: string
  description?: string
}) {
  return (
    <ErrorBoundary
      fallback={
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-6 text-center">
          <ExclamationTriangleIcon className="w-8 h-8 text-red-500 mx-auto mb-3" />
          <h3 className="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">{title}</h3>
          <p className="text-red-700 dark:text-red-300 text-sm">{description}</p>
          <button
            onClick={() => window.location.reload()}
            className="mt-4 px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors text-sm"
          >
            Reload Page
          </button>
        </div>
      }
    >
      {children}
    </ErrorBoundary>
  )
}

export default ErrorBoundary
