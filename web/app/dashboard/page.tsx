'use client'

import { useState, useEffect, useMemo } from 'react'
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
  ExclamationTriangleIcon,
  ArrowUpIcon,
  ArrowDownIcon,
  EyeIcon,
  BoltIcon,
  CloudIcon,
  CircleStackIcon,
  CurrencyDollarIcon,
  BeakerIcon,
  RocketLaunchIcon,
  ShieldCheckIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import Link from 'next/link'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts'

interface SystemStats {
  totalRequests: number
  activeModels: number
  cacheHitRate: number
  avgLatency: number
  totalCost: number
  documentsProcessed: number
  activeSessions: number
  systemHealth: 'healthy' | 'degraded' | 'unhealthy'
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  networkIn: number
  networkOut: number
  errorRate: number
  uptime: number
  requestsPerSecond: number
}

interface RecentActivity {
  id: string
  type: 'ai_request' | 'document_upload' | 'template_execution' | 'system_event' | 'user_action' | 'alert'
  description: string
  timestamp: Date
  status: 'success' | 'error' | 'pending' | 'warning'
  user?: string
  duration?: number
  metadata?: Record<string, any>
}

interface ChartDataPoint {
  time: string
  requests: number
  latency: number
  errors: number
  cost: number
  cpu: number
  memory: number
}

interface ModelUsage {
  name: string
  requests: number
  cost: number
  avgLatency: number
  errorRate: number
  color: string
}

interface AlertItem {
  id: string
  type: 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: Date
  resolved: boolean
}

export default function Dashboard() {
  const [mounted, setMounted] = useState(false)
  const [stats, setStats] = useState<SystemStats>({
    totalRequests: 0,
    activeModels: 0,
    cacheHitRate: 0,
    avgLatency: 0,
    totalCost: 0,
    documentsProcessed: 0,
    activeSessions: 0,
    systemHealth: 'healthy',
    cpuUsage: 0,
    memoryUsage: 0,
    diskUsage: 0,
    networkIn: 0,
    networkOut: 0,
    errorRate: 0,
    uptime: 0,
    requestsPerSecond: 0
  })

  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([])
  const [chartData, setChartData] = useState<ChartDataPoint[]>([])
  const [modelUsage, setModelUsage] = useState<ModelUsage[]>([])
  const [alerts, setAlerts] = useState<AlertItem[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedTimeRange, setSelectedTimeRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h')

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
          systemHealth: 'healthy',
          cpuUsage: 68.5,
          memoryUsage: 74.2,
          diskUsage: 45.8,
          networkIn: 1.2,
          networkOut: 0.8,
          errorRate: 0.02,
          uptime: 99.97,
          requestsPerSecond: 42.3
        })

        // Generate chart data for the last 24 hours
        const chartData = Array.from({ length: 24 }, (_, i) => {
          const time = new Date(Date.now() - (23 - i) * 60 * 60 * 1000)
          return {
            time: time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
            requests: Math.floor(Math.random() * 100) + 50,
            latency: Math.floor(Math.random() * 200) + 150,
            errors: Math.floor(Math.random() * 5),
            cost: Math.random() * 10 + 5,
            cpu: Math.random() * 30 + 40,
            memory: Math.random() * 20 + 60
          }
        })
        setChartData(chartData)

        // Model usage data
        setModelUsage([
          { name: 'GPT-4', requests: 8432, cost: 89.23, avgLatency: 1200, errorRate: 0.01, color: '#3b82f6' },
          { name: 'GPT-3.5', requests: 5621, cost: 23.45, avgLatency: 800, errorRate: 0.02, color: '#8b5cf6' },
          { name: 'Claude-3', requests: 1794, cost: 14.77, avgLatency: 950, errorRate: 0.015, color: '#10b981' },
          { name: 'Gemini', requests: 892, cost: 8.91, avgLatency: 750, errorRate: 0.025, color: '#f59e0b' }
        ])

        // Alerts data
        setAlerts([
          {
            id: '1',
            type: 'warning',
            title: 'High Memory Usage',
            message: 'Memory usage is above 70% threshold',
            timestamp: new Date(Date.now() - 10 * 60 * 1000),
            resolved: false
          },
          {
            id: '2',
            type: 'info',
            title: 'Model Update Available',
            message: 'GPT-4 Turbo update is available',
            timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000),
            resolved: false
          }
        ])

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

  useEffect(() => {
    setMounted(true)
  }, [])

  // Generate real-time data updates
  useEffect(() => {
    const interval = setInterval(() => {
      setStats(prev => ({
        ...prev,
        requestsPerSecond: Math.random() * 20 + 30,
        cpuUsage: Math.random() * 30 + 50,
        memoryUsage: Math.random() * 20 + 60,
        activeSessions: Math.floor(Math.random() * 10) + 20
      }))
    }, 3000)

    return () => clearInterval(interval)
  }, [])

  const timeRangeOptions = [
    { value: '1h', label: '1 Hour' },
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' },
    { value: '30d', label: '30 Days' }
  ]

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center">
        <div className="text-center">
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
            className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"
          />
          <p className="text-gray-600 dark:text-gray-400 text-lg">Loading Epic Dashboard...</p>
        </div>
      </div>
    )
  }

  if (!mounted) {
    return null
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20">
      {/* Enhanced Header */}
      <header className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <motion.div
                className="w-12 h-12 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg"
                whileHover={{ scale: 1.05, rotate: 5 }}
                transition={{ type: "spring", stiffness: 400, damping: 10 }}
              >
                <SparklesIcon className="w-7 h-7 text-white" />
              </motion.div>
              <div>
                <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                  AIOS Epic Dashboard
                </h1>
                <p className="text-sm text-gray-600 dark:text-gray-400">Real-time AI System Monitoring & Control</p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {/* Time Range Selector */}
              <select
                value={selectedTimeRange}
                onChange={(e) => setSelectedTimeRange(e.target.value as any)}
                className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                {timeRangeOptions.map(option => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>

              {/* System Health Badge */}
              <motion.div
                className={`flex items-center space-x-2 px-4 py-2 rounded-full text-sm font-medium shadow-lg ${
                  stats.systemHealth === 'healthy'
                    ? 'bg-gradient-to-r from-green-500 to-emerald-500 text-white'
                    : stats.systemHealth === 'degraded'
                    ? 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white'
                    : 'bg-gradient-to-r from-red-500 to-pink-500 text-white'
                }`}
                animate={{ scale: [1, 1.02, 1] }}
                transition={{ duration: 2, repeat: Infinity }}
              >
                <motion.div
                  className="w-2 h-2 bg-white rounded-full"
                  animate={{ opacity: [1, 0.5, 1] }}
                  transition={{ duration: 1, repeat: Infinity }}
                />
                System {stats.systemHealth}
              </motion.div>

              {/* Alerts Badge */}
              {alerts.filter(a => !a.resolved).length > 0 && (
                <motion.div
                  className="relative"
                  whileHover={{ scale: 1.05 }}
                >
                  <ExclamationTriangleIcon className="w-6 h-6 text-amber-500" />
                  <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                    {alerts.filter(a => !a.resolved).length}
                  </span>
                </motion.div>
              )}
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Enhanced Stats Grid */}
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, staggerChildren: 0.1 }}
        >
          <EnhancedStatCard
            title="Total Requests"
            value={stats.totalRequests.toLocaleString()}
            icon={<ChartBarIcon className="w-6 h-6" />}
            color="blue"
            change="+12.5%"
            trend="up"
            realTimeValue={stats.requestsPerSecond}
            realTimeLabel="req/sec"
          />
          <EnhancedStatCard
            title="Active Models"
            value={stats.activeModels.toString()}
            icon={<CpuChipIcon className="w-6 h-6" />}
            color="purple"
            change="+2"
            trend="up"
            realTimeValue={stats.activeSessions}
            realTimeLabel="sessions"
          />
          <EnhancedStatCard
            title="Cache Hit Rate"
            value={`${stats.cacheHitRate}%`}
            icon={<ServerIcon className="w-6 h-6" />}
            color="green"
            change="+3.2%"
            trend="up"
            realTimeValue={stats.errorRate}
            realTimeLabel="% errors"
          />
          <EnhancedStatCard
            title="Avg Latency"
            value={`${stats.avgLatency}ms`}
            icon={<ClockIcon className="w-6 h-6" />}
            color="orange"
            change="-15ms"
            trend="down"
            realTimeValue={stats.uptime}
            realTimeLabel="% uptime"
          />
        </motion.div>

        {/* System Resources Grid */}
        <motion.div
          className="grid grid-cols-1 md:grid-cols-3 gap-6"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <ResourceCard
            title="CPU Usage"
            value={stats.cpuUsage}
            icon={<CpuChipIcon className="w-5 h-5" />}
            color="blue"
            max={100}
          />
          <ResourceCard
            title="Memory Usage"
            value={stats.memoryUsage}
            icon={<CircleStackIcon className="w-5 h-5" />}
            color="purple"
            max={100}
          />
          <ResourceCard
            title="Disk Usage"
            value={stats.diskUsage}
            icon={<ServerIcon className="w-5 h-5" />}
            color="green"
            max={100}
          />
        </motion.div>

        {/* Charts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Performance Chart */}
          <motion.div
            className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
          >
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Performance Metrics</h3>
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                <span className="text-sm text-gray-600 dark:text-gray-400">Requests</span>
                <div className="w-3 h-3 bg-purple-500 rounded-full ml-4"></div>
                <span className="text-sm text-gray-600 dark:text-gray-400">Latency</span>
              </div>
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="time" stroke="#6b7280" fontSize={12} />
                <YAxis stroke="#6b7280" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(255, 255, 255, 0.95)',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="requests"
                  stroke="#3b82f6"
                  strokeWidth={3}
                  dot={{ fill: '#3b82f6', strokeWidth: 2, r: 4 }}
                  activeDot={{ r: 6, stroke: '#3b82f6', strokeWidth: 2 }}
                />
                <Line
                  type="monotone"
                  dataKey="latency"
                  stroke="#8b5cf6"
                  strokeWidth={3}
                  dot={{ fill: '#8b5cf6', strokeWidth: 2, r: 4 }}
                  activeDot={{ r: 6, stroke: '#8b5cf6', strokeWidth: 2 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </motion.div>

          {/* Model Usage Chart */}
          <motion.div
            className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
          >
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">Model Usage Distribution</h3>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={modelUsage}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={120}
                  paddingAngle={5}
                  dataKey="requests"
                >
                  {modelUsage.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(255, 255, 255, 0.95)',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                  }}
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </motion.div>
        </div>

        {/* Enhanced Quick Actions */}
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.5 }}
        >
          <QuickActionCard
            title="AI Chat"
            description="Start a conversation with AI models"
            icon={<ChatBubbleLeftRightIcon className="w-8 h-8" />}
            href="/dashboard/chat"
            color="blue"
          />
          <QuickActionCard
            title="Knowledge Base"
            description="Upload and manage documents"
            icon={<DocumentTextIcon className="w-8 h-8" />}
            href="/dashboard/documents"
            color="green"
          />
          <QuickActionCard
            title="Projects"
            description="Manage your AI projects"
            icon={<RocketLaunchIcon className="w-8 h-8" />}
            href="/dashboard/projects"
            color="purple"
          />
          <QuickActionCard
            title="Settings"
            description="Configure system settings"
            icon={<Cog6ToothIcon className="w-8 h-8" />}
            href="/dashboard/settings"
            color="gray"
          />
        </motion.div>

        {/* Activity and Alerts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Enhanced Recent Activity */}
          <motion.div
            className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.6 }}
          >
            <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Recent Activity</h2>
                <motion.div
                  className="w-2 h-2 bg-green-500 rounded-full"
                  animate={{ scale: [1, 1.2, 1], opacity: [1, 0.7, 1] }}
                  transition={{ duration: 2, repeat: Infinity }}
                />
              </div>
            </div>
            <div className="max-h-96 overflow-y-auto">
              <AnimatePresence>
                {recentActivity.map((activity, index) => (
                  <motion.div
                    key={activity.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 20 }}
                    transition={{ delay: index * 0.1 }}
                    className="px-6 py-4 border-b border-gray-100 dark:border-gray-700/50 last:border-b-0 hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <div className={`w-3 h-3 rounded-full ${
                          activity.status === 'success'
                            ? 'bg-green-500'
                            : activity.status === 'error'
                            ? 'bg-red-500'
                            : activity.status === 'warning'
                            ? 'bg-yellow-500'
                            : 'bg-blue-500'
                        }`}></div>
                        <div>
                          <p className="text-sm font-medium text-gray-900 dark:text-white">{activity.description}</p>
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            {activity.timestamp.toLocaleTimeString()} • {activity.type.replace('_', ' ')}
                            {activity.user && ` • ${activity.user}`}
                          </p>
                        </div>
                      </div>
                      <div className={`px-2 py-1 rounded-full text-xs font-medium ${
                        activity.status === 'success'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : activity.status === 'error'
                          ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                          : activity.status === 'warning'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                          : 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                      }`}>
                        {activity.status}
                      </div>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          </motion.div>

          {/* System Alerts */}
          <motion.div
            className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.7 }}
          >
            <div className="px-6 py-4 border-b border-gray-200/50 dark:border-gray-700/50">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">System Alerts</h2>
            </div>
            <div className="p-6 space-y-4">
              {alerts.length === 0 ? (
                <div className="text-center py-8">
                  <ShieldCheckIcon className="w-12 h-12 text-green-500 mx-auto mb-3" />
                  <p className="text-gray-500 dark:text-gray-400">All systems operational</p>
                </div>
              ) : (
                alerts.map((alert) => (
                  <motion.div
                    key={alert.id}
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className={`p-4 rounded-xl border-l-4 ${
                      alert.type === 'error'
                        ? 'bg-red-50 border-red-500 dark:bg-red-900/20'
                        : alert.type === 'warning'
                        ? 'bg-yellow-50 border-yellow-500 dark:bg-yellow-900/20'
                        : 'bg-blue-50 border-blue-500 dark:bg-blue-900/20'
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div>
                        <h4 className={`font-medium ${
                          alert.type === 'error'
                            ? 'text-red-800 dark:text-red-400'
                            : alert.type === 'warning'
                            ? 'text-yellow-800 dark:text-yellow-400'
                            : 'text-blue-800 dark:text-blue-400'
                        }`}>
                          {alert.title}
                        </h4>
                        <p className={`text-sm mt-1 ${
                          alert.type === 'error'
                            ? 'text-red-600 dark:text-red-300'
                            : alert.type === 'warning'
                            ? 'text-yellow-600 dark:text-yellow-300'
                            : 'text-blue-600 dark:text-blue-300'
                        }`}>
                          {alert.message}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                          {alert.timestamp.toLocaleString()}
                        </p>
                      </div>
                      {!alert.resolved && (
                        <motion.div
                          className={`w-2 h-2 rounded-full ${
                            alert.type === 'error' ? 'bg-red-500' : alert.type === 'warning' ? 'bg-yellow-500' : 'bg-blue-500'
                          }`}
                          animate={{ scale: [1, 1.2, 1], opacity: [1, 0.7, 1] }}
                          transition={{ duration: 2, repeat: Infinity }}
                        />
                      )}
                    </div>
                  </motion.div>
                ))
              )}
            </div>
          </motion.div>
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

// Enhanced Stat Card Component
function EnhancedStatCard({ title, value, icon, color, change, trend, realTimeValue, realTimeLabel }: {
  title: string
  value: string
  icon: React.ReactNode
  color: string
  change: string
  trend: 'up' | 'down'
  realTimeValue: number
  realTimeLabel: string
}) {
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600',
    purple: 'from-purple-500 to-purple-600',
    green: 'from-green-500 to-green-600',
    orange: 'from-orange-500 to-orange-600',
    red: 'from-red-500 to-red-600'
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ y: -4, shadow: "0 20px 25px -5px rgb(0 0 0 / 0.1)" }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6 relative overflow-hidden"
    >
      {/* Background Gradient */}
      <div className={`absolute top-0 right-0 w-32 h-32 bg-gradient-to-br ${colorClasses[color as keyof typeof colorClasses]} opacity-10 rounded-full -mr-16 -mt-16`} />

      <div className="relative">
        <div className="flex items-center justify-between mb-4">
          <div className={`p-3 rounded-xl bg-gradient-to-r ${colorClasses[color as keyof typeof colorClasses]} text-white shadow-lg`}>
            {icon}
          </div>
          <div className={`flex items-center space-x-1 text-sm font-medium ${
            trend === 'up' ? 'text-green-600' : 'text-red-600'
          }`}>
            {trend === 'up' ? <ArrowUpIcon className="w-4 h-4" /> : <ArrowDownIcon className="w-4 h-4" />}
            {change}
          </div>
        </div>

        <div>
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">{title}</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mb-2">{value}</p>
          <div className="flex items-center justify-between">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {realTimeValue.toFixed(1)} {realTimeLabel}
            </p>
            <motion.div
              className="w-2 h-2 bg-green-500 rounded-full"
              animate={{ scale: [1, 1.2, 1], opacity: [1, 0.7, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
            />
          </div>
        </div>
      </div>
    </motion.div>
  )
}

// Resource Card Component
function ResourceCard({ title, value, icon, color, max }: {
  title: string
  value: number
  icon: React.ReactNode
  color: string
  max: number
}) {
  const percentage = (value / max) * 100
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600',
    purple: 'from-purple-500 to-purple-600',
    green: 'from-green-500 to-green-600',
    orange: 'from-orange-500 to-orange-600'
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6"
    >
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-3">
          <div className={`p-2 rounded-lg bg-gradient-to-r ${colorClasses[color as keyof typeof colorClasses]} text-white`}>
            {icon}
          </div>
          <div>
            <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">{value.toFixed(1)}%</p>
          </div>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mb-2">
        <motion.div
          className={`h-2 rounded-full bg-gradient-to-r ${colorClasses[color as keyof typeof colorClasses]}`}
          initial={{ width: 0 }}
          animate={{ width: `${percentage}%` }}
          transition={{ duration: 1, ease: "easeOut" }}
        />
      </div>

      <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400">
        <span>0%</span>
        <span>100%</span>
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
