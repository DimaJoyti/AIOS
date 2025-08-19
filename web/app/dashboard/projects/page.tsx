'use client'

import { useState, useCallback, useMemo } from 'react'
import { 
  PlusIcon,
  RocketLaunchIcon,
  UserGroupIcon,
  CalendarIcon,
  ClockIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  EllipsisVerticalIcon,
  PencilIcon,
  TrashIcon,
  ShareIcon,
  StarIcon,
  FlagIcon,
  ChatBubbleLeftRightIcon,
  DocumentTextIcon,
  ChartBarIcon,
  Cog6ToothIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  Squares2X2Icon,
  ListBulletIcon,
  ArrowUpIcon,
  ArrowDownIcon,
  PlayIcon,
  PauseIcon,
  StopIcon,
  BoltIcon,
  SparklesIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'

interface Task {
  id: string
  title: string
  description: string
  status: 'todo' | 'in-progress' | 'review' | 'done'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  assignee?: {
    id: string
    name: string
    avatar: string
  }
  dueDate?: Date
  createdAt: Date
  updatedAt: Date
  tags: string[]
  comments: number
  attachments: number
  estimatedHours?: number
  actualHours?: number
  dependencies?: string[]
}

interface Project {
  id: string
  name: string
  description: string
  status: 'planning' | 'active' | 'on-hold' | 'completed' | 'cancelled'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  progress: number
  startDate: Date
  endDate?: Date
  dueDate?: Date
  budget?: number
  spent?: number
  team: {
    id: string
    name: string
    avatar: string
    role: string
  }[]
  tasks: Task[]
  tags: string[]
  color: string
  isStarred: boolean
  createdAt: Date
  updatedAt: Date
}

export default function ProjectsPage() {
  const [mounted, setMounted] = useState(false)
  const [projects, setProjects] = useState<Project[]>([
    {
      id: '1',
      name: 'AIOS Core Development',
      description: 'Core AI operating system development with advanced features and integrations.',
      status: 'active',
      priority: 'high',
      progress: 75,
      startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
      dueDate: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000),
      budget: 50000,
      spent: 37500,
      team: [
        { id: '1', name: 'Alex Chen', avatar: '/avatars/alex.jpg', role: 'Lead Developer' },
        { id: '2', name: 'Sarah Kim', avatar: '/avatars/sarah.jpg', role: 'AI Specialist' },
        { id: '3', name: 'Mike Johnson', avatar: '/avatars/mike.jpg', role: 'Backend Developer' }
      ],
      tasks: [
        {
          id: '1',
          title: 'Implement AI Chat Interface',
          description: 'Create advanced chat interface with multiple AI models',
          status: 'done',
          priority: 'high',
          assignee: { id: '1', name: 'Alex Chen', avatar: '/avatars/alex.jpg' },
          dueDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          createdAt: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          tags: ['frontend', 'ai', 'chat'],
          comments: 8,
          attachments: 3,
          estimatedHours: 40,
          actualHours: 38
        },
        {
          id: '2',
          title: 'Knowledge Management System',
          description: 'Build document upload and AI processing pipeline',
          status: 'done',
          priority: 'high',
          assignee: { id: '2', name: 'Sarah Kim', avatar: '/avatars/sarah.jpg' },
          dueDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          createdAt: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
          tags: ['ai', 'documents', 'processing'],
          comments: 12,
          attachments: 5,
          estimatedHours: 50,
          actualHours: 52
        },
        {
          id: '3',
          title: 'Real-time Dashboard',
          description: 'Create interactive dashboard with live metrics',
          status: 'in-progress',
          priority: 'medium',
          assignee: { id: '3', name: 'Mike Johnson', avatar: '/avatars/mike.jpg' },
          dueDate: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000),
          createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 1 * 60 * 60 * 1000),
          tags: ['dashboard', 'metrics', 'realtime'],
          comments: 6,
          attachments: 2,
          estimatedHours: 30,
          actualHours: 18
        },
        {
          id: '4',
          title: 'API Integration Layer',
          description: 'Integrate with external AI services and APIs',
          status: 'todo',
          priority: 'medium',
          assignee: { id: '1', name: 'Alex Chen', avatar: '/avatars/alex.jpg' },
          dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
          createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
          tags: ['api', 'integration', 'backend'],
          comments: 2,
          attachments: 1,
          estimatedHours: 35
        }
      ],
      tags: ['ai', 'core', 'development'],
      color: 'blue',
      isStarred: true,
      createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 1 * 60 * 60 * 1000)
    },
    {
      id: '2',
      name: 'Mobile App Development',
      description: 'Cross-platform mobile application for AIOS with native features.',
      status: 'planning',
      priority: 'medium',
      progress: 15,
      startDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      dueDate: new Date(Date.now() + 60 * 24 * 60 * 60 * 1000),
      budget: 30000,
      spent: 4500,
      team: [
        { id: '4', name: 'Emma Davis', avatar: '/avatars/emma.jpg', role: 'Mobile Developer' },
        { id: '5', name: 'James Wilson', avatar: '/avatars/james.jpg', role: 'UI/UX Designer' }
      ],
      tasks: [
        {
          id: '5',
          title: 'Mobile UI Design System',
          description: 'Create consistent design system for mobile app',
          status: 'in-progress',
          priority: 'high',
          assignee: { id: '5', name: 'James Wilson', avatar: '/avatars/james.jpg' },
          dueDate: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000),
          createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
          tags: ['design', 'mobile', 'ui'],
          comments: 4,
          attachments: 8,
          estimatedHours: 25,
          actualHours: 12
        }
      ],
      tags: ['mobile', 'app', 'cross-platform'],
      color: 'green',
      isStarred: false,
      createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000)
    }
  ])

  const [viewMode, setViewMode] = useState<'grid' | 'list' | 'kanban'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterPriority, setFilterPriority] = useState<string>('all')
  const [showFilters, setShowFilters] = useState(false)
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [showNewProjectModal, setShowNewProjectModal] = useState(false)

  // Computed values
  const filteredProjects = useMemo(() => {
    return projects.filter(project => {
      const matchesSearch = project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        project.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        project.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
      
      const matchesStatus = filterStatus === 'all' || project.status === filterStatus
      const matchesPriority = filterPriority === 'all' || project.priority === filterPriority
      
      return matchesSearch && matchesStatus && matchesPriority
    })
  }, [projects, searchQuery, filterStatus, filterPriority])

  const projectStats = useMemo(() => {
    const total = projects.length
    const active = projects.filter(p => p.status === 'active').length
    const completed = projects.filter(p => p.status === 'completed').length
    const overdue = projects.filter(p => 
      p.dueDate && new Date() > p.dueDate && p.status !== 'completed'
    ).length
    
    const totalBudget = projects.reduce((sum, p) => sum + (p.budget || 0), 0)
    const totalSpent = projects.reduce((sum, p) => sum + (p.spent || 0), 0)
    
    const totalTasks = projects.reduce((sum, p) => sum + p.tasks.length, 0)
    const completedTasks = projects.reduce((sum, p) => 
      sum + p.tasks.filter(t => t.status === 'done').length, 0
    )
    
    return {
      total,
      active,
      completed,
      overdue,
      totalBudget,
      totalSpent,
      totalTasks,
      completedTasks,
      avgProgress: total > 0 ? Math.round(projects.reduce((sum, p) => sum + p.progress, 0) / total) : 0
    }
  }, [projects])

  const toggleProjectStar = useCallback((projectId: string) => {
    setProjects(prev => prev.map(project => 
      project.id === projectId ? { ...project, isStarred: !project.isStarred } : project
    ))
  }, [])

  const deleteProject = useCallback((projectId: string) => {
    setProjects(prev => prev.filter(project => project.id !== projectId))
  }, [])

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
              className="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-600 rounded-xl flex items-center justify-center shadow-lg"
              whileHover={{ scale: 1.05, rotate: 5 }}
              transition={{ type: "spring", stiffness: 400, damping: 10 }}
            >
              <RocketLaunchIcon className="w-7 h-7 text-white" />
            </motion.div>
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                Epic Project Hub
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Advanced project management with AI-powered insights
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            {/* Enhanced Search */}
            <div className="relative">
              <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search projects..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 pr-4 py-3 w-80 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all duration-200 shadow-lg"
              />
            </div>
            
            {/* View Mode Toggle */}
            <div className="flex items-center bg-white dark:bg-gray-800 rounded-xl border border-gray-300 dark:border-gray-600 p-1 shadow-lg">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-lg transition-all duration-200 ${
                  viewMode === 'grid' 
                    ? 'bg-purple-500 text-white shadow-md' 
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                }`}
              >
                <Squares2X2Icon className="w-4 h-4" />
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setViewMode('list')}
                className={`p-2 rounded-lg transition-all duration-200 ${
                  viewMode === 'list' 
                    ? 'bg-purple-500 text-white shadow-md' 
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                }`}
              >
                <ListBulletIcon className="w-4 h-4" />
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setViewMode('kanban')}
                className={`p-2 rounded-lg transition-all duration-200 ${
                  viewMode === 'kanban' 
                    ? 'bg-purple-500 text-white shadow-md' 
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                }`}
              >
                <ChartBarIcon className="w-4 h-4" />
              </motion.button>
            </div>
            
            {/* Filters Toggle */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setShowFilters(!showFilters)}
              className={`p-3 rounded-xl transition-all duration-200 shadow-lg ${
                showFilters 
                  ? 'bg-purple-500 text-white' 
                  : 'bg-white dark:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 border border-gray-300 dark:border-gray-600'
              }`}
            >
              <FunnelIcon className="w-5 h-5" />
            </motion.button>
            
            {/* New Project Button */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setShowNewProjectModal(true)}
              className="px-6 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl hover:from-purple-700 hover:to-pink-700 transition-all duration-200 shadow-lg flex items-center space-x-2"
            >
              <PlusIcon className="w-5 h-5" />
              <span className="font-medium">New Project</span>
            </motion.button>
          </div>
        </div>
      </div>

      {/* Enhanced Stats Bar */}
      <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
        <div className="grid grid-cols-2 md:grid-cols-8 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">{projectStats.total}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Total Projects</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">{projectStats.active}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Active</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{projectStats.completed}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Completed</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">{projectStats.overdue}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Overdue</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">{projectStats.avgProgress}%</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Avg Progress</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-teal-600 dark:text-teal-400">{projectStats.totalTasks}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Total Tasks</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
              ${(projectStats.totalBudget / 1000).toFixed(0)}k
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Budget</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-pink-600 dark:text-pink-400">
              ${(projectStats.totalSpent / 1000).toFixed(0)}k
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400">Spent</div>
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
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Status:</label>
                <select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                  className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                  <option value="all">All Status</option>
                  <option value="planning">Planning</option>
                  <option value="active">Active</option>
                  <option value="on-hold">On Hold</option>
                  <option value="completed">Completed</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>

              <div className="flex items-center space-x-2">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Priority:</label>
                <select
                  value={filterPriority}
                  onChange={(e) => setFilterPriority(e.target.value)}
                  className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                  <option value="all">All Priorities</option>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="urgent">Urgent</option>
                </select>
              </div>

              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Showing {filteredProjects.length} of {projects.length} projects
                </span>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        {viewMode === 'grid' && (
          <div className="h-full overflow-y-auto p-6">
            {filteredProjects.length === 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-center py-16"
              >
                <RocketLaunchIcon className="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                  {searchQuery ? 'No projects found' : 'No projects yet'}
                </h3>
                <p className="text-gray-500 dark:text-gray-400 mb-6">
                  {searchQuery
                    ? 'Try adjusting your search terms or filters'
                    : 'Create your first project to get started'
                  }
                </p>
                {!searchQuery && (
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setShowNewProjectModal(true)}
                    className="px-6 py-3 bg-purple-500 text-white rounded-xl hover:bg-purple-600 transition-colors shadow-lg"
                  >
                    Create Project
                  </motion.button>
                )}
              </motion.div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <AnimatePresence>
                  {filteredProjects.map((project, index) => (
                    <ProjectCard
                      key={project.id}
                      project={project}
                      index={index}
                      onStar={() => toggleProjectStar(project.id)}
                      onDelete={() => deleteProject(project.id)}
                      onView={() => setSelectedProject(project)}
                    />
                  ))}
                </AnimatePresence>
              </div>
            )}
          </div>
        )}

        {viewMode === 'list' && (
          <div className="h-full overflow-y-auto p-6">
            {filteredProjects.length === 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-center py-16"
              >
                <RocketLaunchIcon className="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">No projects found</h3>
                <p className="text-gray-500 dark:text-gray-400">Try adjusting your search terms or filters</p>
              </motion.div>
            ) : (
              <div className="space-y-4">
                <AnimatePresence>
                  {filteredProjects.map((project, index) => (
                    <ProjectListItem
                      key={project.id}
                      project={project}
                      index={index}
                      onStar={() => toggleProjectStar(project.id)}
                      onDelete={() => deleteProject(project.id)}
                      onView={() => setSelectedProject(project)}
                    />
                  ))}
                </AnimatePresence>
              </div>
            )}
          </div>
        )}

        {viewMode === 'kanban' && (
          <KanbanBoard projects={filteredProjects} />
        )}
      </div>
    </div>
  )
}

// Project Card Component
function ProjectCard({ project, index, onStar, onDelete, onView }: {
  project: Project
  index: number
  onStar: () => void
  onDelete: () => void
  onView: () => void
}) {
  const getStatusColor = (status: Project['status']) => {
    switch (status) {
      case 'planning': return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      case 'active': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      case 'on-hold': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
      case 'completed': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
      case 'cancelled': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    }
  }

  const getPriorityColor = (priority: Project['priority']) => {
    switch (priority) {
      case 'low': return 'text-gray-500'
      case 'medium': return 'text-blue-500'
      case 'high': return 'text-orange-500'
      case 'urgent': return 'text-red-500'
    }
  }

  const isOverdue = project.dueDate && new Date() > project.dueDate && project.status !== 'completed'

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={{ delay: index * 0.05 }}
      className={`group relative bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border transition-all duration-200 hover:shadow-xl hover:-translate-y-1 ${
        project.color === 'blue' ? 'border-blue-200 dark:border-blue-800' :
        project.color === 'green' ? 'border-green-200 dark:border-green-800' :
        project.color === 'purple' ? 'border-purple-200 dark:border-purple-800' :
        project.color === 'orange' ? 'border-orange-200 dark:border-orange-800' :
        'border-gray-200 dark:border-gray-700'
      }`}
    >
      {/* Header */}
      <div className="p-6 pb-4">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center space-x-3">
            <div className={`w-3 h-3 rounded-full ${
              project.color === 'blue' ? 'bg-blue-500' :
              project.color === 'green' ? 'bg-green-500' :
              project.color === 'purple' ? 'bg-purple-500' :
              project.color === 'orange' ? 'bg-orange-500' :
              'bg-gray-500'
            }`} />
            <span className={`px-2 py-1 text-xs rounded-full font-medium ${getStatusColor(project.status)}`}>
              {project.status.replace('-', ' ')}
            </span>
          </div>

          <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={(e) => {
                e.stopPropagation()
                onStar()
              }}
              className={`p-1 rounded transition-colors ${
                project.isStarred
                  ? 'text-yellow-500'
                  : 'text-gray-400 hover:text-yellow-500'
              }`}
            >
              <StarIcon className={`w-4 h-4 ${project.isStarred ? 'fill-current' : ''}`} />
            </motion.button>

            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={(e) => {
                e.stopPropagation()
                onDelete()
              }}
              className="p-1 text-gray-400 hover:text-red-500 transition-colors"
            >
              <TrashIcon className="w-4 h-4" />
            </motion.button>
          </div>
        </div>

        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2 line-clamp-2">
          {project.name}
        </h3>

        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4 line-clamp-2">
          {project.description}
        </p>

        {/* Progress */}
        <div className="mb-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Progress</span>
            <span className="text-sm text-gray-500 dark:text-gray-400">{project.progress}%</span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <motion.div
              className={`h-2 rounded-full ${
                project.color === 'blue' ? 'bg-blue-500' :
                project.color === 'green' ? 'bg-green-500' :
                project.color === 'purple' ? 'bg-purple-500' :
                project.color === 'orange' ? 'bg-orange-500' :
                'bg-gray-500'
              }`}
              initial={{ width: 0 }}
              animate={{ width: `${project.progress}%` }}
              transition={{ duration: 1, ease: "easeOut" }}
            />
          </div>
        </div>

        {/* Team */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <UserGroupIcon className="w-4 h-4 text-gray-400" />
            <span className="text-sm text-gray-500 dark:text-gray-400">{project.team.length} members</span>
          </div>

          <div className="flex items-center space-x-1">
            <FlagIcon className={`w-4 h-4 ${getPriorityColor(project.priority)}`} />
            <span className={`text-sm font-medium ${getPriorityColor(project.priority)}`}>
              {project.priority}
            </span>
          </div>
        </div>

        {/* Due Date */}
        {project.dueDate && (
          <div className={`flex items-center space-x-2 mb-4 ${isOverdue ? 'text-red-500' : 'text-gray-500 dark:text-gray-400'}`}>
            <CalendarIcon className="w-4 h-4" />
            <span className="text-sm">
              Due {project.dueDate.toLocaleDateString()}
              {isOverdue && ' (Overdue)'}
            </span>
          </div>
        )}

        {/* Tags */}
        {project.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-4">
            {project.tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded-full"
              >
                {tag}
              </span>
            ))}
            {project.tags.length > 3 && (
              <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded-full">
                +{project.tags.length - 3}
              </span>
            )}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="px-6 pb-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4 text-sm text-gray-500 dark:text-gray-400">
            <div className="flex items-center space-x-1">
              <CheckCircleIcon className="w-4 h-4" />
              <span>{project.tasks.filter(t => t.status === 'done').length}/{project.tasks.length}</span>
            </div>
            {project.budget && (
              <div className="flex items-center space-x-1">
                <span>${(project.spent || 0).toLocaleString()}</span>
                <span>/</span>
                <span>${project.budget.toLocaleString()}</span>
              </div>
            )}
          </div>

          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onView}
            className="px-4 py-2 bg-purple-500 text-white text-sm rounded-lg hover:bg-purple-600 transition-colors"
          >
            View Details
          </motion.button>
        </div>
      </div>
    </motion.div>
  )
}

// Project List Item Component
function ProjectListItem({ project, index, onStar, onDelete, onView }: {
  project: Project
  index: number
  onStar: () => void
  onDelete: () => void
  onView: () => void
}) {
  const getStatusColor = (status: Project['status']) => {
    switch (status) {
      case 'planning': return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      case 'active': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      case 'on-hold': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
      case 'completed': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
      case 'cancelled': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    }
  }

  const isOverdue = project.dueDate && new Date() > project.dueDate && project.status !== 'completed'

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 20 }}
      transition={{ delay: index * 0.02 }}
      className="group flex items-center space-x-6 p-6 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl border border-gray-200/50 dark:border-gray-700/50 transition-all duration-200 hover:shadow-lg"
    >
      {/* Project Color & Status */}
      <div className="flex items-center space-x-3">
        <div className={`w-4 h-4 rounded-full ${
          project.color === 'blue' ? 'bg-blue-500' :
          project.color === 'green' ? 'bg-green-500' :
          project.color === 'purple' ? 'bg-purple-500' :
          project.color === 'orange' ? 'bg-orange-500' :
          'bg-gray-500'
        }`} />
        <span className={`px-3 py-1 text-xs rounded-full font-medium ${getStatusColor(project.status)}`}>
          {project.status.replace('-', ' ')}
        </span>
      </div>

      {/* Project Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center space-x-2 mb-1">
          <h3 className="font-semibold text-gray-900 dark:text-white truncate">
            {project.name}
          </h3>
          {project.isStarred && (
            <StarIcon className="w-4 h-4 text-yellow-500 fill-current" />
          )}
        </div>
        <p className="text-sm text-gray-600 dark:text-gray-400 truncate mb-2">
          {project.description}
        </p>
        <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
          <span>{project.team.length} members</span>
          <span>{project.tasks.filter(t => t.status === 'done').length}/{project.tasks.length} tasks</span>
          {project.dueDate && (
            <span className={isOverdue ? 'text-red-500' : ''}>
              Due {project.dueDate.toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      {/* Progress */}
      <div className="w-32">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs text-gray-500 dark:text-gray-400">Progress</span>
          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{project.progress}%</span>
        </div>
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <motion.div
            className={`h-2 rounded-full ${
              project.color === 'blue' ? 'bg-blue-500' :
              project.color === 'green' ? 'bg-green-500' :
              project.color === 'purple' ? 'bg-purple-500' :
              project.color === 'orange' ? 'bg-orange-500' :
              'bg-gray-500'
            }`}
            initial={{ width: 0 }}
            animate={{ width: `${project.progress}%` }}
            transition={{ duration: 1, ease: "easeOut" }}
          />
        </div>
      </div>

      {/* Budget */}
      {project.budget && (
        <div className="text-right">
          <div className="text-sm font-medium text-gray-900 dark:text-white">
            ${(project.spent || 0).toLocaleString()}
          </div>
          <div className="text-xs text-gray-500 dark:text-gray-400">
            of ${project.budget.toLocaleString()}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center space-x-2">
        <motion.button
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onView}
          className="px-3 py-1 bg-purple-500 text-white text-sm rounded-lg hover:bg-purple-600 transition-colors"
        >
          View
        </motion.button>

        <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            onClick={onStar}
            className={`p-1 rounded transition-colors ${
              project.isStarred
                ? 'text-yellow-500'
                : 'text-gray-400 hover:text-yellow-500'
            }`}
          >
            <StarIcon className={`w-4 h-4 ${project.isStarred ? 'fill-current' : ''}`} />
          </motion.button>

          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            onClick={onDelete}
            className="p-1 text-gray-400 hover:text-red-500 transition-colors"
          >
            <TrashIcon className="w-4 h-4" />
          </motion.button>
        </div>
      </div>
    </motion.div>
  )
}

// Kanban Board Component
function KanbanBoard({ projects }: { projects: Project[] }) {
  const columns = [
    { id: 'planning', title: 'Planning', color: 'gray' },
    { id: 'active', title: 'Active', color: 'green' },
    { id: 'on-hold', title: 'On Hold', color: 'yellow' },
    { id: 'completed', title: 'Completed', color: 'blue' }
  ]

  return (
    <div className="h-full p-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 h-full">
        {columns.map((column) => {
          const columnProjects = projects.filter(p => p.status === column.id)

          return (
            <div key={column.id} className="flex flex-col">
              <div className={`flex items-center space-x-2 mb-4 p-3 rounded-xl ${
                column.color === 'gray' ? 'bg-gray-100 dark:bg-gray-800' :
                column.color === 'green' ? 'bg-green-100 dark:bg-green-900/30' :
                column.color === 'yellow' ? 'bg-yellow-100 dark:bg-yellow-900/30' :
                'bg-blue-100 dark:bg-blue-900/30'
              }`}>
                <h3 className={`font-semibold ${
                  column.color === 'gray' ? 'text-gray-800 dark:text-gray-300' :
                  column.color === 'green' ? 'text-green-800 dark:text-green-400' :
                  column.color === 'yellow' ? 'text-yellow-800 dark:text-yellow-400' :
                  'text-blue-800 dark:text-blue-400'
                }`}>
                  {column.title}
                </h3>
                <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                  column.color === 'gray' ? 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-300' :
                  column.color === 'green' ? 'bg-green-200 text-green-700 dark:bg-green-800 dark:text-green-300' :
                  column.color === 'yellow' ? 'bg-yellow-200 text-yellow-700 dark:bg-yellow-800 dark:text-yellow-300' :
                  'bg-blue-200 text-blue-700 dark:bg-blue-800 dark:text-blue-300'
                }`}>
                  {columnProjects.length}
                </span>
              </div>

              <div className="flex-1 space-y-3 overflow-y-auto">
                <AnimatePresence>
                  {columnProjects.map((project, index) => (
                    <motion.div
                      key={project.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -20 }}
                      transition={{ delay: index * 0.05 }}
                      className="p-4 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 hover:shadow-md transition-shadow cursor-pointer"
                    >
                      <div className="flex items-start justify-between mb-3">
                        <h4 className="font-medium text-gray-900 dark:text-white text-sm line-clamp-2">
                          {project.name}
                        </h4>
                        {project.isStarred && (
                          <StarIcon className="w-4 h-4 text-yellow-500 fill-current flex-shrink-0 ml-2" />
                        )}
                      </div>

                      <p className="text-xs text-gray-600 dark:text-gray-400 mb-3 line-clamp-2">
                        {project.description}
                      </p>

                      <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-3">
                        <span>{project.tasks.filter(t => t.status === 'done').length}/{project.tasks.length} tasks</span>
                        <span>{project.progress}%</span>
                      </div>

                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1 mb-3">
                        <div
                          className={`h-1 rounded-full ${
                            project.color === 'blue' ? 'bg-blue-500' :
                            project.color === 'green' ? 'bg-green-500' :
                            project.color === 'purple' ? 'bg-purple-500' :
                            project.color === 'orange' ? 'bg-orange-500' :
                            'bg-gray-500'
                          }`}
                          style={{ width: `${project.progress}%` }}
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-1">
                          <UserGroupIcon className="w-3 h-3 text-gray-400" />
                          <span className="text-xs text-gray-500 dark:text-gray-400">{project.team.length}</span>
                        </div>

                        {project.dueDate && (
                          <div className="flex items-center space-x-1">
                            <CalendarIcon className="w-3 h-3 text-gray-400" />
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              {project.dueDate.toLocaleDateString()}
                            </span>
                          </div>
                        )}
                      </div>
                    </motion.div>
                  ))}
                </AnimatePresence>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
