'use client'

import { useState, useEffect, useMemo } from 'react'
import { 
  CpuChipIcon,
  CircleStackIcon,
  ServerIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  BoltIcon,
  ChartBarIcon,
  EyeIcon,
  ArrowPathIcon,
  PlayIcon,
  PauseIcon,
  StopIcon,
  AdjustmentsHorizontalIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  DocumentTextIcon,
  ShieldCheckIcon,
  WifiIcon,
  CloudIcon,
  BeakerIcon,
  CogIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { 
  LineChart, 
  Line, 
  AreaChart, 
  Area, 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell
} from 'recharts'

interface SystemMetric {
  timestamp: Date
  cpu: number
  memory: number
  disk: number
  network: number
  requests: number
  errors: number
  latency: number
}

interface LogEntry {
  id: string
  timestamp: Date
  level: 'info' | 'warn' | 'error' | 'debug'
  service: string
  message: string
  metadata?: Record<string, any>
}

interface ServiceHealth {
  name: string
  status: 'healthy' | 'degraded' | 'unhealthy'
  uptime: number
  lastCheck: Date
  responseTime: number
  errorRate: number
  version: string
}

export default function MonitoringPage() {
  const [mounted, setMounted] = useState(false)
  const [metrics, setMetrics] = useState<SystemMetric[]>([])
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [services, setServices] = useState<ServiceHealth[]>([
    {
      name: 'AI Chat Service',
      status: 'healthy',
      uptime: 99.97,
      lastCheck: new Date(),
      responseTime: 245,
      errorRate: 0.02,
      version: '1.2.3'
    },
    {
      name: 'Document Processing',
      status: 'healthy',
      uptime: 99.85,
      lastCheck: new Date(),
      responseTime: 1200,
      errorRate: 0.05,
      version: '2.1.0'
    },
    {
      name: 'Knowledge Base',
      status: 'degraded',
      uptime: 98.5,
      lastCheck: new Date(),
      responseTime: 3500,
      errorRate: 0.15,
      version: '1.8.2'
    },
    {
      name: 'Analytics Engine',
      status: 'healthy',
      uptime: 99.92,
      lastCheck: new Date(),
      responseTime: 180,
      errorRate: 0.01,
      version: '3.0.1'
    },
    {
      name: 'Authentication',
      status: 'healthy',
      uptime: 99.99,
      lastCheck: new Date(),
      responseTime: 95,
      errorRate: 0.001,
      version: '2.5.4'
    },
    {
      name: 'File Storage',
      status: 'unhealthy',
      uptime: 95.2,
      lastCheck: new Date(),
      responseTime: 8500,
      errorRate: 2.5,
      version: '1.4.7'
    }
  ])

  const [selectedTimeRange, setSelectedTimeRange] = useState<'1h' | '6h' | '24h' | '7d'>('1h')
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [refreshInterval, setRefreshInterval] = useState(5000)
  const [logFilter, setLogFilter] = useState<string>('all')
  const [logSearch, setLogSearch] = useState('')
  const [showFilters, setShowFilters] = useState(false)

  // Generate mock metrics data
  useEffect(() => {
    const generateMetrics = () => {
      const now = new Date()
      const points = selectedTimeRange === '1h' ? 60 : 
                   selectedTimeRange === '6h' ? 72 : 
                   selectedTimeRange === '24h' ? 96 : 168

      const newMetrics: SystemMetric[] = Array.from({ length: points }, (_, i) => {
        const timestamp = new Date(now.getTime() - (points - i) * (selectedTimeRange === '1h' ? 60000 : 
                                                                   selectedTimeRange === '6h' ? 300000 : 
                                                                   selectedTimeRange === '24h' ? 900000 : 3600000))
        return {
          timestamp,
          cpu: Math.random() * 30 + 40 + Math.sin(i / 10) * 15,
          memory: Math.random() * 20 + 60 + Math.sin(i / 8) * 10,
          disk: Math.random() * 10 + 45,
          network: Math.random() * 50 + 25,
          requests: Math.random() * 100 + 50 + Math.sin(i / 5) * 30,
          errors: Math.random() * 5,
          latency: Math.random() * 200 + 150 + Math.sin(i / 12) * 50
        }
      })
      setMetrics(newMetrics)
    }

    generateMetrics()
  }, [selectedTimeRange])

  // Generate mock logs
  useEffect(() => {
    const generateLogs = () => {
      const levels: LogEntry['level'][] = ['info', 'warn', 'error', 'debug']
      const services = ['ai-chat', 'doc-processor', 'knowledge-base', 'analytics', 'auth', 'storage']
      const messages = [
        'Request processed successfully',
        'Database connection established',
        'Cache miss for key: user_session_123',
        'AI model inference completed',
        'Document uploaded and queued for processing',
        'User authentication successful',
        'Memory usage threshold exceeded',
        'Network timeout occurred',
        'Service health check passed',
        'Background job completed'
      ]

      const newLogs: LogEntry[] = Array.from({ length: 100 }, (_, i) => ({
        id: `log-${i}`,
        timestamp: new Date(Date.now() - Math.random() * 3600000),
        level: levels[Math.floor(Math.random() * levels.length)],
        service: services[Math.floor(Math.random() * services.length)],
        message: messages[Math.floor(Math.random() * messages.length)],
        metadata: Math.random() > 0.7 ? {
          userId: `user_${Math.floor(Math.random() * 1000)}`,
          requestId: `req_${Math.random().toString(36).substr(2, 9)}`,
          duration: Math.floor(Math.random() * 1000)
        } : undefined
      }))

      setLogs(newLogs.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime()))
    }

    generateLogs()
  }, [])

  // Auto-refresh functionality
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      // Update metrics with new data point
      setMetrics(prev => {
        const newPoint: SystemMetric = {
          timestamp: new Date(),
          cpu: Math.random() * 30 + 40,
          memory: Math.random() * 20 + 60,
          disk: Math.random() * 10 + 45,
          network: Math.random() * 50 + 25,
          requests: Math.random() * 100 + 50,
          errors: Math.random() * 5,
          latency: Math.random() * 200 + 150
        }
        return [...prev.slice(1), newPoint]
      })

      // Update service health
      setServices(prev => prev.map(service => ({
        ...service,
        lastCheck: new Date(),
        responseTime: service.responseTime + (Math.random() - 0.5) * 100,
        errorRate: Math.max(0, service.errorRate + (Math.random() - 0.5) * 0.1)
      })))
    }, refreshInterval)

    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval])

  const filteredLogs = useMemo(() => {
    return logs.filter(log => {
      const matchesLevel = logFilter === 'all' || log.level === logFilter
      const matchesSearch = logSearch === '' || 
        log.message.toLowerCase().includes(logSearch.toLowerCase()) ||
        log.service.toLowerCase().includes(logSearch.toLowerCase())
      return matchesLevel && matchesSearch
    })
  }, [logs, logFilter, logSearch])

  const currentMetrics = useMemo(() => {
    if (metrics.length === 0) return null
    const latest = metrics[metrics.length - 1]
    return {
      cpu: latest.cpu,
      memory: latest.memory,
      disk: latest.disk,
      network: latest.network,
      requests: latest.requests,
      errors: latest.errors,
      latency: latest.latency
    }
  }, [metrics])

  const systemHealth = useMemo(() => {
    const healthyServices = services.filter(s => s.status === 'healthy').length
    const totalServices = services.length
    const overallHealth = healthyServices / totalServices

    return {
      status: overallHealth >= 0.8 ? 'healthy' : overallHealth >= 0.6 ? 'degraded' : 'unhealthy',
      healthyServices,
      totalServices,
      percentage: Math.round(overallHealth * 100)
    }
  }, [services])

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return null
  }

  return (
    <div className="h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex flex-col">
      {/* Enhanced Header */}
      <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <motion.div
              className="w-12 h-12 bg-gradient-to-r from-green-500 to-teal-600 rounded-xl flex items-center justify-center shadow-lg"
              whileHover={{ scale: 1.05, rotate: 5 }}
              transition={{ type: "spring", stiffness: 400, damping: 10 }}
            >
              <ChartBarIcon className="w-7 h-7 text-white" />
            </motion.div>
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                System Monitoring
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Real-time system health and performance monitoring
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            {/* System Health Badge */}
            <motion.div
              className={`flex items-center space-x-2 px-4 py-2 rounded-full text-sm font-medium shadow-lg ${
                systemHealth.status === 'healthy' 
                  ? 'bg-gradient-to-r from-green-500 to-emerald-500 text-white' 
                  : systemHealth.status === 'degraded'
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
              <span>System {systemHealth.status} ({systemHealth.percentage}%)</span>
            </motion.div>

            {/* Time Range Selector */}
            <select 
              value={selectedTimeRange}
              onChange={(e) => setSelectedTimeRange(e.target.value as any)}
              className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:border-transparent"
            >
              <option value="1h">Last Hour</option>
              <option value="6h">Last 6 Hours</option>
              <option value="24h">Last 24 Hours</option>
              <option value="7d">Last 7 Days</option>
            </select>

            {/* Auto Refresh Toggle */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`p-3 rounded-xl transition-all duration-200 shadow-lg ${
                autoRefresh 
                  ? 'bg-green-500 text-white' 
                  : 'bg-white dark:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 border border-gray-300 dark:border-gray-600'
              }`}
            >
              {autoRefresh ? <PauseIcon className="w-5 h-5" /> : <PlayIcon className="w-5 h-5" />}
            </motion.button>

            {/* Filters Toggle */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setShowFilters(!showFilters)}
              className={`p-3 rounded-xl transition-all duration-200 shadow-lg ${
                showFilters 
                  ? 'bg-green-500 text-white' 
                  : 'bg-white dark:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 border border-gray-300 dark:border-gray-600'
              }`}
            >
              <FunnelIcon className="w-5 h-5" />
            </motion.button>
          </div>
        </div>
      </div>

      {/* Enhanced Filters */}
      <AnimatePresence>
        {showFilters && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4"
          >
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Log Level:</label>
                <select
                  value={logFilter}
                  onChange={(e) => setLogFilter(e.target.value)}
                  className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:border-transparent"
                >
                  <option value="all">All Levels</option>
                  <option value="info">Info</option>
                  <option value="warn">Warning</option>
                  <option value="error">Error</option>
                  <option value="debug">Debug</option>
                </select>
              </div>

              <div className="flex items-center space-x-2">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Refresh Rate:</label>
                <select
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(Number(e.target.value))}
                  className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:border-transparent"
                >
                  <option value={1000}>1 second</option>
                  <option value={5000}>5 seconds</option>
                  <option value={10000}>10 seconds</option>
                  <option value={30000}>30 seconds</option>
                </select>
              </div>

              <div className="relative">
                <MagnifyingGlassIcon className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search logs..."
                  value={logSearch}
                  onChange={(e) => setLogSearch(e.target.value)}
                  className="pl-9 pr-4 py-2 w-64 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:border-transparent"
                />
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        <div className="grid grid-cols-1 lg:grid-cols-3 h-full">
          {/* Metrics Dashboard */}
          <div className="lg:col-span-2 flex flex-col">
            {/* Current Metrics */}
            <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 p-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Current Metrics</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <MetricCard
                  title="CPU Usage"
                  value={currentMetrics?.cpu.toFixed(1) || '0'}
                  unit="%"
                  icon={<CpuChipIcon className="w-5 h-5" />}
                  color="blue"
                  trend={currentMetrics?.cpu > 70 ? 'up' : 'stable'}
                />
                <MetricCard
                  title="Memory"
                  value={currentMetrics?.memory.toFixed(1) || '0'}
                  unit="%"
                  icon={<CircleStackIcon className="w-5 h-5" />}
                  color="purple"
                  trend={currentMetrics?.memory > 80 ? 'up' : 'stable'}
                />
                <MetricCard
                  title="Disk Usage"
                  value={currentMetrics?.disk.toFixed(1) || '0'}
                  unit="%"
                  icon={<ServerIcon className="w-5 h-5" />}
                  color="green"
                  trend="stable"
                />
                <MetricCard
                  title="Avg Latency"
                  value={currentMetrics?.latency.toFixed(0) || '0'}
                  unit="ms"
                  icon={<ClockIcon className="w-5 h-5" />}
                  color="orange"
                  trend={currentMetrics?.latency > 300 ? 'up' : 'down'}
                />
              </div>
            </div>

            {/* Charts */}
            <div className="flex-1 p-6 space-y-6 overflow-y-auto">
              {/* Performance Chart */}
              <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">System Performance</h3>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={metrics}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="timestamp"
                      stroke="#6b7280"
                      fontSize={12}
                      tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                    />
                    <YAxis stroke="#6b7280" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'rgba(255, 255, 255, 0.95)',
                        border: 'none',
                        borderRadius: '12px',
                        boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                      }}
                      labelFormatter={(value) => new Date(value).toLocaleString()}
                    />
                    <Line
                      type="monotone"
                      dataKey="cpu"
                      stroke="#3b82f6"
                      strokeWidth={2}
                      name="CPU %"
                    />
                    <Line
                      type="monotone"
                      dataKey="memory"
                      stroke="#8b5cf6"
                      strokeWidth={2}
                      name="Memory %"
                    />
                    <Line
                      type="monotone"
                      dataKey="latency"
                      stroke="#f59e0b"
                      strokeWidth={2}
                      name="Latency (ms)"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>

              {/* Request & Error Chart */}
              <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Requests & Errors</h3>
                <ResponsiveContainer width="100%" height={250}>
                  <AreaChart data={metrics}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="timestamp"
                      stroke="#6b7280"
                      fontSize={12}
                      tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                    />
                    <YAxis stroke="#6b7280" fontSize={12} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'rgba(255, 255, 255, 0.95)',
                        border: 'none',
                        borderRadius: '12px',
                        boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                      }}
                      labelFormatter={(value) => new Date(value).toLocaleString()}
                    />
                    <Area
                      type="monotone"
                      dataKey="requests"
                      stackId="1"
                      stroke="#10b981"
                      fill="#10b981"
                      fillOpacity={0.6}
                      name="Requests"
                    />
                    <Area
                      type="monotone"
                      dataKey="errors"
                      stackId="2"
                      stroke="#ef4444"
                      fill="#ef4444"
                      fillOpacity={0.8}
                      name="Errors"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Service Health & Logs */}
          <div className="border-l border-gray-200/50 dark:border-gray-700/50 flex flex-col">
            {/* Service Health */}
            <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 p-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Service Health</h2>
              <div className="space-y-3">
                {services.map((service) => (
                  <ServiceHealthCard key={service.name} service={service} />
                ))}
              </div>
            </div>

            {/* Live Logs */}
            <div className="flex-1 flex flex-col">
              <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Live Logs</h2>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {filteredLogs.length} entries
                  </span>
                </div>
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-2">
                <AnimatePresence>
                  {filteredLogs.slice(0, 50).map((log, index) => (
                    <LogEntry key={log.id} log={log} index={index} />
                  ))}
                </AnimatePresence>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Metric Card Component
function MetricCard({ title, value, unit, icon, color, trend }: {
  title: string
  value: string
  unit: string
  icon: React.ReactNode
  color: 'blue' | 'purple' | 'green' | 'orange'
  trend: 'up' | 'down' | 'stable'
}) {
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600',
    purple: 'from-purple-500 to-purple-600',
    green: 'from-green-500 to-green-600',
    orange: 'from-orange-500 to-orange-600'
  }

  const trendColors = {
    up: 'text-red-500',
    down: 'text-green-500',
    stable: 'text-gray-500'
  }

  return (
    <motion.div
      whileHover={{ y: -2 }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-4"
    >
      <div className="flex items-center justify-between mb-2">
        <div className={`p-2 rounded-lg bg-gradient-to-r ${colorClasses[color]} text-white`}>
          {icon}
        </div>
        <div className={`text-xs font-medium ${trendColors[trend]}`}>
          {trend === 'up' ? '↗' : trend === 'down' ? '↘' : '→'}
        </div>
      </div>
      <div className="text-2xl font-bold text-gray-900 dark:text-white">
        {value}<span className="text-sm text-gray-500 dark:text-gray-400 ml-1">{unit}</span>
      </div>
      <div className="text-xs text-gray-500 dark:text-gray-400">{title}</div>
    </motion.div>
  )
}

// Service Health Card Component
function ServiceHealthCard({ service }: { service: ServiceHealth }) {
  const getStatusColor = (status: ServiceHealth['status']) => {
    switch (status) {
      case 'healthy': return 'text-green-600 bg-green-100 dark:bg-green-900/30 dark:text-green-400'
      case 'degraded': return 'text-yellow-600 bg-yellow-100 dark:bg-yellow-900/30 dark:text-yellow-400'
      case 'unhealthy': return 'text-red-600 bg-red-100 dark:bg-red-900/30 dark:text-red-400'
    }
  }

  const getStatusIcon = (status: ServiceHealth['status']) => {
    switch (status) {
      case 'healthy': return <CheckCircleIcon className="w-4 h-4" />
      case 'degraded': return <ExclamationTriangleIcon className="w-4 h-4" />
      case 'unhealthy': return <XCircleIcon className="w-4 h-4" />
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-lg border border-gray-200/50 dark:border-gray-700/50 p-3"
    >
      <div className="flex items-center justify-between mb-2">
        <h4 className="font-medium text-gray-900 dark:text-white text-sm">{service.name}</h4>
        <span className={`px-2 py-1 text-xs rounded-full font-medium flex items-center space-x-1 ${getStatusColor(service.status)}`}>
          {getStatusIcon(service.status)}
          <span>{service.status}</span>
        </span>
      </div>

      <div className="grid grid-cols-2 gap-2 text-xs text-gray-500 dark:text-gray-400">
        <div>
          <span className="font-medium">Uptime:</span> {service.uptime.toFixed(2)}%
        </div>
        <div>
          <span className="font-medium">Response:</span> {service.responseTime.toFixed(0)}ms
        </div>
        <div>
          <span className="font-medium">Errors:</span> {service.errorRate.toFixed(2)}%
        </div>
        <div>
          <span className="font-medium">Version:</span> {service.version}
        </div>
      </div>

      <div className="mt-2 text-xs text-gray-400 dark:text-gray-500">
        Last check: {service.lastCheck.toLocaleTimeString()}
      </div>
    </motion.div>
  )
}

// Log Entry Component
function LogEntry({ log, index }: { log: LogEntry; index: number }) {
  const getLevelColor = (level: LogEntry['level']) => {
    switch (level) {
      case 'info': return 'text-blue-600 bg-blue-100 dark:bg-blue-900/30 dark:text-blue-400'
      case 'warn': return 'text-yellow-600 bg-yellow-100 dark:bg-yellow-900/30 dark:text-yellow-400'
      case 'error': return 'text-red-600 bg-red-100 dark:bg-red-900/30 dark:text-red-400'
      case 'debug': return 'text-gray-600 bg-gray-100 dark:bg-gray-700 dark:text-gray-400'
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.02 }}
      className="bg-white/60 dark:bg-gray-800/60 backdrop-blur-sm rounded-lg border border-gray-200/50 dark:border-gray-700/50 p-3 text-sm"
    >
      <div className="flex items-start justify-between mb-1">
        <div className="flex items-center space-x-2">
          <span className={`px-2 py-1 text-xs rounded font-medium ${getLevelColor(log.level)}`}>
            {log.level.toUpperCase()}
          </span>
          <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">
            {log.service}
          </span>
        </div>
        <span className="text-xs text-gray-400 dark:text-gray-500">
          {log.timestamp.toLocaleTimeString()}
        </span>
      </div>

      <p className="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
        {log.message}
      </p>

      {log.metadata && (
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400 font-mono bg-gray-50 dark:bg-gray-700/50 rounded p-2">
          {Object.entries(log.metadata).map(([key, value]) => (
            <div key={key}>
              <span className="text-gray-400">{key}:</span> {String(value)}
            </div>
          ))}
        </div>
      )}
    </motion.div>
  )
}
