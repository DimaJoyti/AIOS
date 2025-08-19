'use client'

import { motion } from 'framer-motion'
import { SparklesIcon, CpuChipIcon, ChartBarIcon } from '@heroicons/react/24/outline'

// Generic loading spinner
export function LoadingSpinner({ size = 'md', color = 'blue' }: {
  size?: 'sm' | 'md' | 'lg'
  color?: 'blue' | 'purple' | 'green' | 'orange'
}) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-6 h-6',
    lg: 'w-8 h-8'
  }

  const colorClasses = {
    blue: 'text-blue-500',
    purple: 'text-purple-500',
    green: 'text-green-500',
    orange: 'text-orange-500'
  }

  return (
    <motion.div
      animate={{ rotate: 360 }}
      transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
      className={`${sizeClasses[size]} ${colorClasses[color]}`}
    >
      <svg className="w-full h-full" fill="none" viewBox="0 0 24 24">
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    </motion.div>
  )
}

// Skeleton loading components
export function SkeletonBox({ className = "" }: { className?: string }) {
  return (
    <motion.div
      animate={{ opacity: [0.5, 1, 0.5] }}
      transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
      className={`bg-gray-200 dark:bg-gray-700 rounded ${className}`}
    />
  )
}

export function SkeletonText({ lines = 1, className = "" }: { 
  lines?: number
  className?: string 
}) {
  return (
    <div className={`space-y-2 ${className}`}>
      {Array.from({ length: lines }).map((_, i) => (
        <SkeletonBox 
          key={i} 
          className={`h-4 ${i === lines - 1 ? 'w-3/4' : 'w-full'}`} 
        />
      ))}
    </div>
  )
}

// Dashboard loading state
export function DashboardSkeleton() {
  return (
    <div className="h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex flex-col">
      {/* Header skeleton */}
      <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <SkeletonBox className="w-12 h-12 rounded-xl" />
            <div>
              <SkeletonBox className="w-48 h-8 mb-2" />
              <SkeletonBox className="w-64 h-4" />
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <SkeletonBox className="w-32 h-10 rounded-xl" />
            <SkeletonBox className="w-24 h-10 rounded-xl" />
          </div>
        </div>
      </div>

      {/* Stats skeleton */}
      <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4">
              <div className="flex items-center justify-between mb-3">
                <SkeletonBox className="w-8 h-8 rounded-lg" />
                <SkeletonBox className="w-6 h-4" />
              </div>
              <SkeletonBox className="w-16 h-8 mb-1" />
              <SkeletonBox className="w-20 h-3" />
            </div>
          ))}
        </div>
      </div>

      {/* Main content skeleton */}
      <div className="flex-1 p-6 space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-6">
              <SkeletonBox className="w-48 h-6 mb-4" />
              <SkeletonBox className="w-full h-64" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

// Chat loading state
export function ChatSkeleton() {
  return (
    <div className="h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex">
      {/* Sidebar skeleton */}
      <div className="w-80 bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-r border-gray-200/50 dark:border-gray-700/50 p-6">
        <SkeletonBox className="w-full h-12 rounded-xl mb-6" />
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-white/80 dark:bg-gray-800/80 rounded-xl p-4">
              <SkeletonBox className="w-3/4 h-4 mb-2" />
              <SkeletonBox className="w-1/2 h-3" />
            </div>
          ))}
        </div>
      </div>

      {/* Chat area skeleton */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <SkeletonBox className="w-8 h-8 rounded-lg" />
              <SkeletonBox className="w-32 h-6" />
            </div>
            <SkeletonBox className="w-24 h-8 rounded-lg" />
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 p-6 space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className={`flex ${i % 2 === 0 ? 'justify-start' : 'justify-end'}`}>
              <div className="max-w-2xl bg-white/80 dark:bg-gray-800/80 rounded-2xl p-4">
                <SkeletonText lines={2} />
              </div>
            </div>
          ))}
        </div>

        {/* Input area */}
        <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-t border-gray-200/50 dark:border-gray-700/50 p-6">
          <SkeletonBox className="w-full h-12 rounded-xl" />
        </div>
      </div>
    </div>
  )
}

// Project card skeleton
export function ProjectCardSkeleton() {
  return (
    <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-3">
          <SkeletonBox className="w-3 h-3 rounded-full" />
          <SkeletonBox className="w-16 h-5 rounded-full" />
        </div>
        <SkeletonBox className="w-6 h-6 rounded" />
      </div>

      <SkeletonBox className="w-3/4 h-6 mb-2" />
      <SkeletonText lines={2} className="mb-4" />

      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <SkeletonBox className="w-16 h-4" />
          <SkeletonBox className="w-8 h-4" />
        </div>
        <SkeletonBox className="w-full h-2 rounded-full" />
      </div>

      <div className="flex items-center justify-between mb-4">
        <SkeletonBox className="w-20 h-4" />
        <SkeletonBox className="w-12 h-4" />
      </div>

      <div className="flex items-center justify-between">
        <SkeletonBox className="w-16 h-4" />
        <SkeletonBox className="w-20 h-8 rounded-lg" />
      </div>
    </div>
  )
}

// Full page loading with branded animation
export function FullPageLoading({ message = "Loading..." }: { message?: string }) {
  return (
    <div className="fixed inset-0 bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex items-center justify-center z-50">
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center"
      >
        <motion.div
          animate={{ 
            rotate: [0, 360],
            scale: [1, 1.1, 1]
          }}
          transition={{ 
            rotate: { duration: 2, repeat: Infinity, ease: "linear" },
            scale: { duration: 1, repeat: Infinity, ease: "easeInOut" }
          }}
          className="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-lg"
        >
          <SparklesIcon className="w-8 h-8 text-white" />
        </motion.div>
        
        <motion.h2
          animate={{ opacity: [0.5, 1, 0.5] }}
          transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
          className="text-xl font-semibold text-gray-900 dark:text-white mb-2"
        >
          {message}
        </motion.h2>
        
        <div className="flex items-center justify-center space-x-1">
          {Array.from({ length: 3 }).map((_, i) => (
            <motion.div
              key={i}
              animate={{ opacity: [0.3, 1, 0.3] }}
              transition={{ 
                duration: 1.5, 
                repeat: Infinity, 
                delay: i * 0.2,
                ease: "easeInOut"
              }}
              className="w-2 h-2 bg-blue-500 rounded-full"
            />
          ))}
        </div>
      </motion.div>
    </div>
  )
}

// Button loading state
export function ButtonLoading({ children, loading, ...props }: {
  children: React.ReactNode
  loading: boolean
  [key: string]: any
}) {
  return (
    <button {...props} disabled={loading || props.disabled}>
      <div className="flex items-center justify-center space-x-2">
        {loading && <LoadingSpinner size="sm" />}
        <span className={loading ? 'opacity-70' : ''}>{children}</span>
      </div>
    </button>
  )
}
