'use client'

import React, { useState, useMemo } from 'react'
import { 
  ChartBarIcon,
  CurrencyDollarIcon,
  ClockIcon,
  UserGroupIcon,
  DocumentTextIcon,
  ChatBubbleLeftRightIcon,
  CpuChipIcon,
  TrendingUpIcon,
  TrendingDownIcon,
  EyeIcon,
  CalendarIcon,
  ArrowUpIcon,
  ArrowDownIcon,
  SparklesIcon,
  BoltIcon,
  BeakerIcon,
  PresentationChartLineIcon,
  TableCellsIcon
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
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

interface AnalyticsData {
  date: string
  users: number
  sessions: number
  requests: number
  cost: number
  latency: number
  errors: number
  documents: number
  chatMessages: number
}

interface ModelUsage {
  name: string
  requests: number
  cost: number
  avgLatency: number
  errorRate: number
  color: string
}

interface UserMetrics {
  totalUsers: number
  activeUsers: number
  newUsers: number
  retentionRate: number
  avgSessionDuration: number
  bounceRate: number
}

export default function AnalyticsPage() {
  const [mounted, setMounted] = useState(false)
  const [selectedTimeRange, setSelectedTimeRange] = useState<'7d' | '30d' | '90d' | '1y'>('30d')
  const [selectedMetric, setSelectedMetric] = useState<'users' | 'requests' | 'cost' | 'latency'>('users')

  // Mock analytics data
  const analyticsData: AnalyticsData[] = useMemo(() => {
    const days = selectedTimeRange === '7d' ? 7 : selectedTimeRange === '30d' ? 30 : selectedTimeRange === '90d' ? 90 : 365
    return Array.from({ length: days }, (_, i) => {
      const date = new Date()
      date.setDate(date.getDate() - (days - i - 1))
      
      return {
        date: date.toISOString().split('T')[0],
        users: Math.floor(Math.random() * 500) + 200 + Math.sin(i / 7) * 100,
        sessions: Math.floor(Math.random() * 800) + 400 + Math.sin(i / 7) * 150,
        requests: Math.floor(Math.random() * 2000) + 1000 + Math.sin(i / 5) * 500,
        cost: Math.random() * 100 + 50 + Math.sin(i / 10) * 30,
        latency: Math.random() * 200 + 150 + Math.sin(i / 12) * 50,
        errors: Math.floor(Math.random() * 20) + 5,
        documents: Math.floor(Math.random() * 50) + 20,
        chatMessages: Math.floor(Math.random() * 300) + 150
      }
    })
  }, [selectedTimeRange])

  const modelUsage: ModelUsage[] = [
    { name: 'GPT-4', requests: 15420, cost: 1247.50, avgLatency: 1200, errorRate: 0.8, color: '#3b82f6' },
    { name: 'GPT-3.5', requests: 8930, cost: 234.80, avgLatency: 800, errorRate: 1.2, color: '#8b5cf6' },
    { name: 'Claude-3', requests: 5670, cost: 445.20, avgLatency: 950, errorRate: 0.6, color: '#10b981' },
    { name: 'Gemini', requests: 3240, cost: 89.40, avgLatency: 750, errorRate: 1.5, color: '#f59e0b' }
  ]

  const userMetrics: UserMetrics = {
    totalUsers: 12847,
    activeUsers: 8934,
    newUsers: 1247,
    retentionRate: 78.5,
    avgSessionDuration: 24.5,
    bounceRate: 23.8
  }

  const currentPeriodData = useMemo(() => {
    const total = analyticsData.reduce((acc, day) => ({
      users: acc.users + day.users,
      sessions: acc.sessions + day.sessions,
      requests: acc.requests + day.requests,
      cost: acc.cost + day.cost,
      latency: acc.latency + day.latency,
      errors: acc.errors + day.errors,
      documents: acc.documents + day.documents,
      chatMessages: acc.chatMessages + day.chatMessages
    }), {
      users: 0, sessions: 0, requests: 0, cost: 0, latency: 0, errors: 0, documents: 0, chatMessages: 0
    })

    return {
      ...total,
      avgLatency: total.latency / analyticsData.length,
      errorRate: (total.errors / total.requests) * 100
    }
  }, [analyticsData])

  const getMetricChange = (metric: keyof AnalyticsData) => {
    if (analyticsData.length < 2) return 0
    const recent = analyticsData.slice(-7).reduce((sum, day) => sum + day[metric], 0) / 7
    const previous = analyticsData.slice(-14, -7).reduce((sum, day) => sum + day[metric], 0) / 7
    return ((recent - previous) / previous) * 100
  }

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
              className="w-12 h-12 bg-gradient-to-r from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg"
              whileHover={{ scale: 1.05, rotate: 5 }}
              transition={{ type: "spring", stiffness: 400, damping: 10 }}
            >
              <PresentationChartLineIcon className="w-7 h-7 text-white" />
            </motion.div>
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                Analytics Dashboard
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Comprehensive usage analytics and performance insights
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            {/* Time Range Selector */}
            <select 
              value={selectedTimeRange}
              onChange={(e) => setSelectedTimeRange(e.target.value as any)}
              className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent shadow-lg"
            >
              <option value="7d">Last 7 Days</option>
              <option value="30d">Last 30 Days</option>
              <option value="90d">Last 90 Days</option>
              <option value="1y">Last Year</option>
            </select>

            {/* Export Button */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-6 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg flex items-center space-x-2"
            >
              <TableCellsIcon className="w-4 h-4" />
              <span>Export</span>
            </motion.button>
          </div>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-6">
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
          <AnalyticsCard
            title="Total Users"
            value={userMetrics.totalUsers.toLocaleString()}
            change={getMetricChange('users')}
            icon={<UserGroupIcon className="w-5 h-5" />}
            color="blue"
          />
          <AnalyticsCard
            title="Active Users"
            value={userMetrics.activeUsers.toLocaleString()}
            change={12.5}
            icon={<EyeIcon className="w-5 h-5" />}
            color="green"
          />
          <AnalyticsCard
            title="Total Requests"
            value={currentPeriodData.requests.toLocaleString()}
            change={getMetricChange('requests')}
            icon={<BoltIcon className="w-5 h-5" />}
            color="purple"
          />
          <AnalyticsCard
            title="Total Cost"
            value={`$${currentPeriodData.cost.toFixed(2)}`}
            change={getMetricChange('cost')}
            icon={<CurrencyDollarIcon className="w-5 h-5" />}
            color="orange"
          />
          <AnalyticsCard
            title="Avg Latency"
            value={`${currentPeriodData.avgLatency.toFixed(0)}ms`}
            change={-getMetricChange('latency')}
            icon={<ClockIcon className="w-5 h-5" />}
            color="teal"
          />
          <AnalyticsCard
            title="Error Rate"
            value={`${currentPeriodData.errorRate.toFixed(2)}%`}
            change={-8.3}
            icon={<SparklesIcon className="w-5 h-5" />}
            color="red"
          />
          <AnalyticsCard
            title="Documents"
            value={currentPeriodData.documents.toLocaleString()}
            change={getMetricChange('documents')}
            icon={<DocumentTextIcon className="w-5 h-5" />}
            color="indigo"
          />
          <AnalyticsCard
            title="Chat Messages"
            value={currentPeriodData.chatMessages.toLocaleString()}
            change={getMetricChange('chatMessages')}
            icon={<ChatBubbleLeftRightIcon className="w-5 h-5" />}
            color="pink"
          />
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Usage Trends */}
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Usage Trends</h3>
              <select 
                value={selectedMetric}
                onChange={(e) => setSelectedMetric(e.target.value as any)}
                className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option value="users">Users</option>
                <option value="requests">Requests</option>
                <option value="cost">Cost</option>
                <option value="latency">Latency</option>
              </select>
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={analyticsData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis 
                  dataKey="date" 
                  stroke="#6b7280" 
                  fontSize={12}
                  tickFormatter={(value) => new Date(value).toLocaleDateString()}
                />
                <YAxis stroke="#6b7280" fontSize={12} />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: 'rgba(255, 255, 255, 0.95)', 
                    border: 'none', 
                    borderRadius: '12px',
                    boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                  }}
                  labelFormatter={(value) => new Date(value).toLocaleDateString()}
                />
                <Area 
                  type="monotone" 
                  dataKey={selectedMetric} 
                  stroke="#6366f1" 
                  fill="#6366f1"
                  fillOpacity={0.6}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>

        {/* Model Performance Table */}
        <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">Model Performance</h3>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-700">
                  <th className="text-left py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Model</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Requests</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Cost</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Avg Latency</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Error Rate</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-700 dark:text-gray-300">Performance</th>
                </tr>
              </thead>
              <tbody>
                {modelUsage.map((model, index) => (
                  <motion.tr
                    key={model.name}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.1 }}
                    className="border-b border-gray-100 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="py-4 px-4">
                      <div className="flex items-center space-x-3">
                        <div
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: model.color }}
                        />
                        <span className="font-medium text-gray-900 dark:text-white">{model.name}</span>
                      </div>
                    </td>
                    <td className="py-4 px-4 text-right text-gray-700 dark:text-gray-300">
                      {model.requests.toLocaleString()}
                    </td>
                    <td className="py-4 px-4 text-right text-gray-700 dark:text-gray-300">
                      ${model.cost.toFixed(2)}
                    </td>
                    <td className="py-4 px-4 text-right text-gray-700 dark:text-gray-300">
                      {model.avgLatency}ms
                    </td>
                    <td className="py-4 px-4 text-right">
                      <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                        model.errorRate < 1 ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' :
                        model.errorRate < 2 ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400' :
                        'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                      }`}>
                        {model.errorRate.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-4 px-4 text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <div className="w-16 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className="h-2 rounded-full"
                            style={{
                              width: `${Math.min(100, (model.requests / Math.max(...modelUsage.map(m => m.requests))) * 100)}%`,
                              backgroundColor: model.color
                            }}
                          />
                        </div>
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          {Math.round((model.requests / modelUsage.reduce((sum, m) => sum + m.requests, 0)) * 100)}%
                        </span>
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* User Engagement Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">User Engagement</h3>
            <div className="space-y-4">
              <EngagementMetric
                label="Retention Rate"
                value={userMetrics.retentionRate}
                unit="%"
                color="green"
                target={80}
              />
              <EngagementMetric
                label="Avg Session Duration"
                value={userMetrics.avgSessionDuration}
                unit="min"
                color="blue"
                target={30}
              />
              <EngagementMetric
                label="Bounce Rate"
                value={userMetrics.bounceRate}
                unit="%"
                color="red"
                target={20}
                inverse
              />
            </div>
          </div>

          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">Cost Breakdown</h3>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={modelUsage}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" stroke="#6b7280" fontSize={12} />
                <YAxis stroke="#6b7280" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(255, 255, 255, 0.95)',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1)'
                  }}
                />
                <Bar dataKey="cost" fill="#6366f1" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  )
}

// Analytics Card Component
function AnalyticsCard({ title, value, change, icon, color }: {
  title: string
  value: string
  change: number
  icon: React.ReactNode
  color: string
}) {
  const colorClasses = {
    blue: 'from-blue-500 to-blue-600',
    green: 'from-green-500 to-green-600',
    purple: 'from-purple-500 to-purple-600',
    orange: 'from-orange-500 to-orange-600',
    teal: 'from-teal-500 to-teal-600',
    red: 'from-red-500 to-red-600',
    indigo: 'from-indigo-500 to-indigo-600',
    pink: 'from-pink-500 to-pink-600'
  }

  return (
    <motion.div
      whileHover={{ y: -2 }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-4"
    >
      <div className="flex items-center justify-between mb-3">
        <div className={`p-2 rounded-lg bg-gradient-to-r ${colorClasses[color as keyof typeof colorClasses]} text-white`}>
          {icon}
        </div>
        <div className={`flex items-center space-x-1 text-sm font-medium ${
          change >= 0 ? 'text-green-600' : 'text-red-600'
        }`}>
          {change >= 0 ? <ArrowUpIcon className="w-3 h-3" /> : <ArrowDownIcon className="w-3 h-3" />}
          <span>{Math.abs(change).toFixed(1)}%</span>
        </div>
      </div>
      <div className="text-2xl font-bold text-gray-900 dark:text-white mb-1">
        {value}
      </div>
      <div className="text-sm text-gray-500 dark:text-gray-400">{title}</div>
    </motion.div>
  )
}

// Engagement Metric Component
function EngagementMetric({ label, value, unit, color, target, inverse = false }: {
  label: string
  value: number
  unit: string
  color: string
  target: number
  inverse?: boolean
}) {
  const percentage = inverse ?
    Math.max(0, Math.min(100, ((target - value) / target) * 100)) :
    Math.max(0, Math.min(100, (value / target) * 100))

  const colorClasses = {
    green: 'bg-green-500',
    blue: 'bg-blue-500',
    red: 'bg-red-500'
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{label}</span>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {value.toFixed(1)}{unit} / {target}{unit}
        </span>
      </div>
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
        <motion.div
          className={`h-2 rounded-full ${colorClasses[color as keyof typeof colorClasses]}`}
          initial={{ width: 0 }}
          animate={{ width: `${percentage}%` }}
          transition={{ duration: 1, ease: "easeOut" }}
        />
      </div>
    </div>
  )
}

          {/* Model Usage Distribution */}
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">AI Model Usage</h3>
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
          </div>
        </div>
