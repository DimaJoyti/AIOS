'use client'

import { useState, useEffect } from 'react'
import { 
  ChartBarIcon, 
  CpuChipIcon, 
  DocumentTextIcon, 
  ChatBubbleLeftRightIcon,
  Cog6ToothIcon,
  SparklesIcon,
  ClockIcon,
  UserGroupIcon,
  ServerIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import Link from 'next/link'

interface SystemStats {
  totalRequests: number
  activeModels: number
  cacheHitRate: number
  avgLatency: number
  totalCost: number
  documentsProcessed: number
  activeSessions: number
  systemHealth: 'healthy' | 'degraded' | 'unhealthy'
}

interface RecentActivity {
  id: string
  type: 'ai_request' | 'document_upload' | 'template_execution' | 'system_event'
  description: string
  timestamp: Date
  status: 'success' | 'error' | 'pending'
}

export default function Dashboard() {
  const [stats, setStats] = useState<SystemStats>({
    totalRequests: 0,
    activeModels: 0,
    cacheHitRate: 0,
    avgLatency: 0,
    totalCost: 0,
    documentsProcessed: 0,
    activeSessions: 0,
    systemHealth: 'healthy'
  })

  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Simulate loading data from APIs
    const loadDashboardData = async () => {
      try {
        // In a real implementation, these would be actual API calls
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        setStats({
          totalRequests: 15847,
          activeModels: 8,
          cacheHitRate: 87.3,
          avgLatency: 245,
          totalCost: 127.45,
          documentsProcessed: 2341,
          activeSessions: 23,
          systemHealth: 'healthy'
        })

        setRecentActivity([
          {
            id: '1',
            type: 'ai_request',
            description: 'GPT-4 text generation completed',
            timestamp: new Date(Date.now() - 5 * 60 * 1000),
            status: 'success'
          },
          {
            id: '2',
            type: 'document_upload',
            description: 'Research paper uploaded and processed',
            timestamp: new Date(Date.now() - 12 * 60 * 1000),
            status: 'success'
          },
          {
            id: '3',
            type: 'template_execution',
            description: 'Summarization template executed',
            timestamp: new Date(Date.now() - 18 * 60 * 1000),
            status: 'success'
          },
          {
            id: '4',
            type: 'system_event',
            description: 'Cache cleanup completed',
            timestamp: new Date(Date.now() - 25 * 60 * 1000),
            status: 'success'
          }
        ])
      } catch (error) {
        console.error('Failed to load dashboard data:', error)
      } finally {
        setLoading(false)
      }
    }

    loadDashboardData()
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center space-x-4">
              <div className="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <SparklesIcon className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">AIOS Dashboard</h1>
                <p className="text-sm text-gray-500">AI-Powered Operating System Control Center</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className={`flex items-center space-x-2 px-3 py-1 rounded-full text-sm font-medium ${
                stats.systemHealth === 'healthy' 
                  ? 'bg-green-100 text-green-800' 
                  : stats.systemHealth === 'degraded'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-red-100 text-red-800'
              }`}>
                <div className={`w-2 h-2 rounded-full ${
                  stats.systemHealth === 'healthy' 
                    ? 'bg-green-500' 
                    : stats.systemHealth === 'degraded'
                    ? 'bg-yellow-500'
                    : 'bg-red-500'
                }`}></div>
                System {stats.systemHealth}
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <StatCard
            title="Total Requests"
            value={stats.totalRequests.toLocaleString()}
            icon={<ChartBarIcon className="w-6 h-6" />}
            color="blue"
            change="+12.5%"
          />
          <StatCard
            title="Active Models"
            value={stats.activeModels.toString()}
            icon={<CpuChipIcon className="w-6 h-6" />}
            color="purple"
            change="+2"
          />
          <StatCard
            title="Cache Hit Rate"
            value={`${stats.cacheHitRate}%`}
            icon={<ServerIcon className="w-6 h-6" />}
            color="green"
            change="+3.2%"
          />
          <StatCard
            title="Avg Latency"
            value={`${stats.avgLatency}ms`}
            icon={<ClockIcon className="w-6 h-6" />}
            color="orange"
            change="-15ms"
          />
        </div>

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <QuickActionCard
            title="AI Chat"
            description="Start a conversation with AI models"
            icon={<ChatBubbleLeftRightIcon className="w-8 h-8" />}
            href="/dashboard/chat"
            color="blue"
          />
          <QuickActionCard
            title="Documents"
            description="Upload and manage documents"
            icon={<DocumentTextIcon className="w-8 h-8" />}
            href="/dashboard/documents"
            color="green"
          />
          <QuickActionCard
            title="Templates"
            description="Manage prompt templates"
            icon={<SparklesIcon className="w-8 h-8" />}
            href="/dashboard/templates"
            color="purple"
          />
          <QuickActionCard
            title="Settings"
            description="Configure system settings"
            icon={<Cog6ToothIcon className="w-8 h-8" />}
            href="/dashboard/settings"
            color="gray"
          />
        </div>

        {/* Recent Activity */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Recent Activity</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {recentActivity.map((activity) => (
              <div key={activity.id} className="px-6 py-4 flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className={`w-2 h-2 rounded-full ${
                    activity.status === 'success' 
                      ? 'bg-green-500' 
                      : activity.status === 'error'
                      ? 'bg-red-500'
                      : 'bg-yellow-500'
                  }`}></div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">{activity.description}</p>
                    <p className="text-xs text-gray-500">
                      {activity.timestamp.toLocaleTimeString()} • {activity.type.replace('_', ' ')}
                    </p>
                  </div>
                </div>
                <div className={`px-2 py-1 rounded-full text-xs font-medium ${
                  activity.status === 'success' 
                    ? 'bg-green-100 text-green-800' 
                    : activity.status === 'error'
                    ? 'bg-red-100 text-red-800'
                    : 'bg-yellow-100 text-yellow-800'
                }`}>
                  {activity.status}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function StatCard({ title, value, icon, color, change }: {
  title: string
  value: string
  icon: React.ReactNode
  color: string
  change: string
}) {
  const colorClasses = {
    blue: 'bg-blue-500',
    purple: 'bg-purple-500',
    green: 'bg-green-500',
    orange: 'bg-orange-500',
    red: 'bg-red-500',
    gray: 'bg-gray-500'
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          <p className="text-sm text-gray-500">{change} from last hour</p>
        </div>
        <div className={`p-3 rounded-lg ${colorClasses[color as keyof typeof colorClasses]} text-white`}>
          {icon}
        </div>
      </div>
    </motion.div>
  )
}

function QuickActionCard({ title, description, icon, href, color }: {
  title: string
  description: string
  icon: React.ReactNode
  href: string
  color: string
}) {
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700',
    green: 'from-green-500 to-green-600 hover:from-green-600 hover:to-green-700',
    purple: 'from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700',
    gray: 'from-gray-500 to-gray-600 hover:from-gray-600 hover:to-gray-700'
  }

  return (
    <Link href={href}>
      <motion.div
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
        className={`bg-gradient-to-r ${colorClasses[color as keyof typeof colorClasses]} text-white rounded-lg p-6 cursor-pointer transition-all duration-200`}
      >
        <div className="flex items-center space-x-3 mb-3">
          {icon}
          <h3 className="text-lg font-semibold">{title}</h3>
        </div>
        <p className="text-sm opacity-90">{description}</p>
      </motion.div>
    </Link>
  )
}
