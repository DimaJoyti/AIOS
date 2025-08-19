'use client'

import { useState, useCallback, useEffect, useMemo } from 'react'
import {
  DocumentTextIcon,
  CloudArrowUpIcon,
  MagnifyingGlassIcon,
  EyeIcon,
  TrashIcon,
  DocumentArrowDownIcon,
  FolderIcon,
  CalendarIcon,
  UserIcon,
  TagIcon,
  PlusIcon,
  XMarkIcon,
  FunnelIcon,
  Squares2X2Icon,
  ListBulletIcon,
  StarIcon,
  ShareIcon,
  PencilIcon,
  DocumentDuplicateIcon,
  ArchiveBoxIcon,
  ChartBarIcon,
  ClockIcon,
  CpuChipIcon,
  SparklesIcon,
  BookmarkIcon,
  AdjustmentsHorizontalIcon,
  ArrowUpTrayIcon,
  ArrowDownTrayIcon,
  PhotoIcon,
  FilmIcon,
  MusicalNoteIcon,
  CodeBracketIcon,
  PresentationChartLineIcon,
  TableCellsIcon,
  DocumentChartBarIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useDropzone } from 'react-dropzone'

interface Document {
  id: string
  name: string
  type: string
  size: number
  uploadedAt: Date
  updatedAt: Date
  status: 'uploading' | 'processing' | 'completed' | 'error' | 'archived'
  tags: string[]
  category: 'document' | 'image' | 'video' | 'audio' | 'code' | 'data' | 'presentation'
  summary?: string
  extractedText?: string
  thumbnail?: string
  isStarred: boolean
  isShared: boolean
  folderId?: string
  metadata?: {
    pages?: number
    wordCount?: number
    language?: string
    author?: string
    createdDate?: Date
    modifiedDate?: Date
    version?: string
    fileFormat?: string
    encoding?: string
    dimensions?: { width: number; height: number }
    duration?: number
    bitrate?: number
    resolution?: string
  }
  aiAnalysis?: {
    sentiment?: 'positive' | 'negative' | 'neutral'
    topics?: string[]
    entities?: string[]
    keyPhrases?: string[]
    readabilityScore?: number
    complexity?: 'low' | 'medium' | 'high'
    confidence?: number
  }
  processingProgress?: number
  downloadUrl?: string
  previewUrl?: string
}

interface Folder {
  id: string
  name: string
  color: string
  createdAt: Date
  documentCount: number
  isShared: boolean
}

interface KnowledgeStats {
  totalDocuments: number
  totalSize: number
  processingQueue: number
  categoryCounts: Record<string, number>
  recentActivity: number
  storageUsed: number
  storageLimit: number
}

export default function DocumentsPage() {
  const [mounted, setMounted] = useState(false)
  const [documents, setDocuments] = useState<Document[]>([
    {
      id: '1',
      name: 'AI Healthcare Research Paper.pdf',
      type: 'application/pdf',
      size: 2048576,
      uploadedAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
      status: 'completed',
      category: 'document',
      tags: ['research', 'healthcare', 'ai', 'machine-learning'],
      isStarred: true,
      isShared: false,
      summary: 'A comprehensive study on the applications of artificial intelligence in modern healthcare systems, covering diagnostic tools, treatment optimization, and patient care automation.',
      thumbnail: '/api/thumbnails/1.jpg',
      metadata: {
        pages: 24,
        wordCount: 8500,
        language: 'en',
        author: 'Dr. Sarah Johnson',
        createdDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
        fileFormat: 'PDF/A-1b'
      },
      aiAnalysis: {
        sentiment: 'positive',
        topics: ['artificial intelligence', 'healthcare', 'diagnostics', 'automation'],
        entities: ['hospitals', 'patients', 'doctors', 'medical devices'],
        keyPhrases: ['machine learning algorithms', 'patient outcomes', 'diagnostic accuracy'],
        readabilityScore: 85,
        complexity: 'high',
        confidence: 0.92
      }
    },
    {
      id: '2',
      name: 'Project Proposal - AI Automation.docx',
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      size: 512000,
      uploadedAt: new Date(Date.now() - 5 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 3 * 60 * 60 * 1000),
      status: 'completed',
      category: 'document',
      tags: ['project', 'proposal', 'automation', 'business'],
      isStarred: false,
      isShared: true,
      summary: 'Detailed project proposal for implementing AI-driven automation in business processes, including timeline, budget, and expected ROI.',
      metadata: {
        pages: 12,
        wordCount: 3200,
        language: 'en',
        author: 'Michael Chen',
        version: '2.1'
      },
      aiAnalysis: {
        sentiment: 'neutral',
        topics: ['automation', 'business process', 'ROI', 'implementation'],
        complexity: 'medium',
        confidence: 0.88
      }
    },
    {
      id: '3',
      name: 'Team Meeting Notes - Q4 Planning.txt',
      type: 'text/plain',
      size: 8192,
      uploadedAt: new Date(Date.now() - 30 * 60 * 1000),
      updatedAt: new Date(Date.now() - 30 * 60 * 1000),
      status: 'processing',
      category: 'document',
      tags: ['meeting', 'notes', 'planning', 'q4'],
      isStarred: false,
      isShared: false,
      processingProgress: 65
    },
    {
      id: '4',
      name: 'Product Demo Video.mp4',
      type: 'video/mp4',
      size: 15728640,
      uploadedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
      status: 'completed',
      category: 'video',
      tags: ['demo', 'product', 'presentation'],
      isStarred: true,
      isShared: true,
      summary: 'Product demonstration video showcasing key features and user interface.',
      thumbnail: '/api/thumbnails/4.jpg',
      metadata: {
        duration: 180,
        resolution: '1920x1080',
        bitrate: 5000
      }
    },
    {
      id: '5',
      name: 'Data Analysis Script.py',
      type: 'text/x-python',
      size: 4096,
      uploadedAt: new Date(Date.now() - 6 * 60 * 60 * 1000),
      updatedAt: new Date(Date.now() - 4 * 60 * 60 * 1000),
      status: 'completed',
      category: 'code',
      tags: ['python', 'data-analysis', 'script'],
      isStarred: false,
      isShared: false,
      summary: 'Python script for automated data analysis and visualization.',
      metadata: {
        language: 'python',
        encoding: 'UTF-8'
      }
    }
  ])

  const [folders, setFolders] = useState<Folder[]>([
    { id: '1', name: 'Research Papers', color: 'blue', createdAt: new Date(), documentCount: 5, isShared: false },
    { id: '2', name: 'Project Documents', color: 'green', createdAt: new Date(), documentCount: 8, isShared: true },
    { id: '3', name: 'Meeting Notes', color: 'purple', createdAt: new Date(), documentCount: 12, isShared: false },
    { id: '4', name: 'Media Files', color: 'orange', createdAt: new Date(), documentCount: 3, isShared: false }
  ])

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [filterCategory, setFilterCategory] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [sortBy, setSortBy] = useState<'name' | 'date' | 'size' | 'type'>('date')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [selectedFolder, setSelectedFolder] = useState<string | null>(null)
  const [showFilters, setShowFilters] = useState(false)
  const [selectedDocuments, setSelectedDocuments] = useState<string[]>([])
  const [showNewFolderModal, setShowNewFolderModal] = useState(false)

  // Computed values
  const stats: KnowledgeStats = useMemo(() => ({
    totalDocuments: documents.length,
    totalSize: documents.reduce((sum, doc) => sum + doc.size, 0),
    processingQueue: documents.filter(doc => doc.status === 'processing' || doc.status === 'uploading').length,
    categoryCounts: documents.reduce((counts, doc) => {
      counts[doc.category] = (counts[doc.category] || 0) + 1
      return counts
    }, {} as Record<string, number>),
    recentActivity: documents.filter(doc =>
      new Date().getTime() - doc.updatedAt.getTime() < 24 * 60 * 60 * 1000
    ).length,
    storageUsed: documents.reduce((sum, doc) => sum + doc.size, 0),
    storageLimit: 10 * 1024 * 1024 * 1024 // 10GB
  }), [documents])

  const filteredAndSortedDocuments = useMemo(() => {
    let filtered = documents.filter(doc => {
      const matchesSearch = doc.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        doc.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (doc.summary && doc.summary.toLowerCase().includes(searchQuery.toLowerCase()))

      const matchesCategory = filterCategory === 'all' || doc.category === filterCategory
      const matchesStatus = filterStatus === 'all' || doc.status === filterStatus
      const matchesFolder = !selectedFolder || doc.folderId === selectedFolder

      return matchesSearch && matchesCategory && matchesStatus && matchesFolder
    })

    // Sort documents
    filtered.sort((a, b) => {
      let comparison = 0
      switch (sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'date':
          comparison = a.updatedAt.getTime() - b.updatedAt.getTime()
          break
        case 'size':
          comparison = a.size - b.size
          break
        case 'type':
          comparison = a.type.localeCompare(b.type)
          break
      }
      return sortOrder === 'asc' ? comparison : -comparison
    })

    return filtered
  }, [documents, searchQuery, filterCategory, filterStatus, selectedFolder, sortBy, sortOrder])

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    setIsUploading(true)

    for (const file of acceptedFiles) {
      const category = getFileCategory(file.type)
      const newDocument: Document = {
        id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
        name: file.name,
        type: file.type,
        size: file.size,
        uploadedAt: new Date(),
        updatedAt: new Date(),
        status: 'uploading',
        category,
        tags: generateInitialTags(file.name, category),
        isStarred: false,
        isShared: false,
        processingProgress: 0
      }

      setDocuments(prev => [newDocument, ...prev])

      try {
        // Simulate progressive upload and processing
        for (let progress = 0; progress <= 100; progress += 10) {
          await new Promise(resolve => setTimeout(resolve, 100))
          setDocuments(prev => prev.map(doc =>
            doc.id === newDocument.id
              ? { ...doc, processingProgress: progress, status: progress < 100 ? 'uploading' : 'processing' }
              : doc
          ))
        }

        // Simulate AI processing
        await new Promise(resolve => setTimeout(resolve, 2000))

        const mockResult = generateMockProcessingResult(file, category)

        // Update document with processing results
        setDocuments(prev => prev.map(doc =>
          doc.id === newDocument.id
            ? {
                ...doc,
                status: 'completed',
                summary: mockResult.summary,
                extractedText: mockResult.extractedText,
                metadata: mockResult.metadata,
                aiAnalysis: mockResult.aiAnalysis,
                tags: [...doc.tags, ...mockResult.suggestedTags],
                thumbnail: mockResult.thumbnail,
                processingProgress: 100
              }
            : doc
        ))
      } catch (error) {
        console.error('Upload error:', error)
        setDocuments(prev => prev.map(doc =>
          doc.id === newDocument.id
            ? { ...doc, status: 'error' }
            : doc
        ))
      }
    }

    setIsUploading(false)
  }, [])

  useEffect(() => {
    setMounted(true)
  }, [])

  // Utility functions
  const getFileCategory = (mimeType: string): Document['category'] => {
    if (mimeType.startsWith('image/')) return 'image'
    if (mimeType.startsWith('video/')) return 'video'
    if (mimeType.startsWith('audio/')) return 'audio'
    if (mimeType.includes('pdf') || mimeType.includes('document') || mimeType.includes('text')) return 'document'
    if (mimeType.includes('presentation')) return 'presentation'
    if (mimeType.includes('spreadsheet') || mimeType.includes('csv')) return 'data'
    if (mimeType.includes('javascript') || mimeType.includes('python') || mimeType.includes('code')) return 'code'
    return 'document'
  }

  const generateInitialTags = (filename: string, category: Document['category']): string[] => {
    const tags = [category]
    const name = filename.toLowerCase()

    if (name.includes('meeting')) tags.push('meeting')
    if (name.includes('report')) tags.push('report')
    if (name.includes('proposal')) tags.push('proposal')
    if (name.includes('research')) tags.push('research')
    if (name.includes('analysis')) tags.push('analysis')
    if (name.includes('demo')) tags.push('demo')

    return tags
  }

  const generateMockProcessingResult = (file: File, category: Document['category']) => {
    const mockResults = {
      document: {
        summary: `This document contains important information about ${file.name.split('.')[0]}. The content has been analyzed and processed for easy retrieval and reference.`,
        extractedText: `Sample extracted text from ${file.name}. This would contain the actual text content extracted from the document using OCR or text parsing technologies.`,
        suggestedTags: ['important', 'processed'],
        aiAnalysis: {
          sentiment: 'neutral' as const,
          topics: ['business', 'documentation'],
          complexity: 'medium' as const,
          confidence: 0.85
        }
      },
      image: {
        summary: `Image file containing visual content. Analyzed for objects, text, and visual elements.`,
        suggestedTags: ['visual', 'media'],
        aiAnalysis: {
          topics: ['visual content', 'media'],
          confidence: 0.90
        }
      },
      video: {
        summary: `Video content with duration and visual analysis. Contains multimedia information.`,
        suggestedTags: ['multimedia', 'video'],
        metadata: {
          duration: 120,
          resolution: '1920x1080'
        }
      },
      code: {
        summary: `Code file containing programming logic and implementation details.`,
        extractedText: `// Sample code content\nfunction example() {\n  return "processed";\n}`,
        suggestedTags: ['programming', 'development'],
        aiAnalysis: {
          topics: ['programming', 'software development'],
          complexity: 'high' as const,
          confidence: 0.95
        }
      }
    }

    return {
      ...mockResults[category] || mockResults.document,
      metadata: {
        fileFormat: file.type,
        encoding: 'UTF-8',
        ...mockResults[category]?.metadata
      },
      thumbnail: category === 'image' || category === 'video' ? `/api/thumbnails/${Date.now()}.jpg` : undefined
    }
  }

  // Document management functions
  const toggleStar = useCallback((documentId: string) => {
    setDocuments(prev => prev.map(doc =>
      doc.id === documentId ? { ...doc, isStarred: !doc.isStarred } : doc
    ))
  }, [])

  const toggleShare = useCallback((documentId: string) => {
    setDocuments(prev => prev.map(doc =>
      doc.id === documentId ? { ...doc, isShared: !doc.isShared } : doc
    ))
  }, [])

  const deleteDocument = useCallback((documentId: string) => {
    setDocuments(prev => prev.filter(doc => doc.id !== documentId))
    setSelectedDocuments(prev => prev.filter(id => id !== documentId))
  }, [])

  const archiveDocument = useCallback((documentId: string) => {
    setDocuments(prev => prev.map(doc =>
      doc.id === documentId ? { ...doc, status: 'archived' } : doc
    ))
  }, [])

  const duplicateDocument = useCallback((documentId: string) => {
    const original = documents.find(doc => doc.id === documentId)
    if (original) {
      const duplicate: Document = {
        ...original,
        id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
        name: `${original.name} (Copy)`,
        uploadedAt: new Date(),
        updatedAt: new Date(),
        isShared: false
      }
      setDocuments(prev => [duplicate, ...prev])
    }
  }, [documents])

  const getFileIcon = (type: string, category: Document['category']) => {
    switch (category) {
      case 'image': return <PhotoIcon className="w-5 h-5" />
      case 'video': return <FilmIcon className="w-5 h-5" />
      case 'audio': return <MusicalNoteIcon className="w-5 h-5" />
      case 'code': return <CodeBracketIcon className="w-5 h-5" />
      case 'presentation': return <PresentationChartLineIcon className="w-5 h-5" />
      case 'data': return <TableCellsIcon className="w-5 h-5" />
      default: return <DocumentTextIcon className="w-5 h-5" />
    }
  }

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf'],
      'application/msword': ['.doc'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'text/plain': ['.txt'],
      'text/markdown': ['.md'],
      'application/json': ['.json'],
      'text/csv': ['.csv'],
      'image/*': ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg'],
      'video/*': ['.mp4', '.avi', '.mov', '.wmv', '.webm'],
      'audio/*': ['.mp3', '.wav', '.flac', '.aac', '.ogg'],
      'text/x-python': ['.py'],
      'text/javascript': ['.js'],
      'text/typescript': ['.ts'],
      'application/vnd.ms-powerpoint': ['.ppt'],
      'application/vnd.openxmlformats-officedocument.presentationml.presentation': ['.pptx'],
      'application/vnd.ms-excel': ['.xls'],
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx']
    },
    multiple: true,
    maxSize: 100 * 1024 * 1024 // 100MB
  })

  const filteredDocuments = documents.filter(doc =>
    doc.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    doc.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
  )

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  if (!mounted) {
    return null
  }

  return (
    <div className="h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex">
      {/* Enhanced Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Enhanced Header */}
        <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <motion.div
                className="w-12 h-12 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg"
                whileHover={{ scale: 1.05, rotate: 5 }}
                transition={{ type: "spring", stiffness: 400, damping: 10 }}
              >
                <DocumentTextIcon className="w-7 h-7 text-white" />
              </motion.div>
              <div>
                <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                  Epic Knowledge Hub
                </h1>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  AI-powered document management and intelligent search
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {/* Enhanced Search */}
              <div className="relative">
                <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search documents, content, tags..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 pr-4 py-3 w-80 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 shadow-lg"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  >
                    <XMarkIcon className="w-4 h-4" />
                  </button>
                )}
              </div>

              {/* View Mode Toggle */}
              <div className="flex items-center bg-white dark:bg-gray-800 rounded-xl border border-gray-300 dark:border-gray-600 p-1 shadow-lg">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => setViewMode('grid')}
                  className={`p-2 rounded-lg transition-all duration-200 ${
                    viewMode === 'grid'
                      ? 'bg-blue-500 text-white shadow-md'
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
                      ? 'bg-blue-500 text-white shadow-md'
                      : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                  }`}
                >
                  <ListBulletIcon className="w-4 h-4" />
                </motion.button>
              </div>

              {/* Filters Toggle */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setShowFilters(!showFilters)}
                className={`p-3 rounded-xl transition-all duration-200 shadow-lg ${
                  showFilters
                    ? 'bg-blue-500 text-white'
                    : 'bg-white dark:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 border border-gray-300 dark:border-gray-600'
                }`}
              >
                <FunnelIcon className="w-5 h-5" />
              </motion.button>
            </div>
          </div>
        </div>

        {/* Enhanced Stats Bar */}
        <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
          <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.totalDocuments}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Total Files</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">{formatFileSize(stats.totalSize)}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Storage Used</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">{stats.processingQueue}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Processing</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">{stats.recentActivity}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Recent</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-teal-600 dark:text-teal-400">{Object.keys(stats.categoryCounts).length}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Categories</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
                {Math.round((stats.storageUsed / stats.storageLimit) * 100)}%
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-400">Storage</div>
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
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Category:</label>
                  <select
                    value={filterCategory}
                    onChange={(e) => setFilterCategory(e.target.value)}
                    className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="all">All Categories</option>
                    <option value="document">Documents</option>
                    <option value="image">Images</option>
                    <option value="video">Videos</option>
                    <option value="audio">Audio</option>
                    <option value="code">Code</option>
                    <option value="data">Data</option>
                    <option value="presentation">Presentations</option>
                  </select>
                </div>

                <div className="flex items-center space-x-2">
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Status:</label>
                  <select
                    value={filterStatus}
                    onChange={(e) => setFilterStatus(e.target.value)}
                    className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="all">All Status</option>
                    <option value="completed">Completed</option>
                    <option value="processing">Processing</option>
                    <option value="uploading">Uploading</option>
                    <option value="error">Error</option>
                    <option value="archived">Archived</option>
                  </select>
                </div>

                <div className="flex items-center space-x-2">
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Sort by:</label>
                  <select
                    value={sortBy}
                    onChange={(e) => setSortBy(e.target.value as any)}
                    className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="date">Date Modified</option>
                    <option value="name">Name</option>
                    <option value="size">Size</option>
                    <option value="type">Type</option>
                  </select>
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
                    className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                  >
                    {sortOrder === 'asc' ? '↑' : '↓'}
                  </motion.button>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="flex-1 flex">
          {/* Enhanced Upload Area */}
          <div className="w-80 border-r border-gray-200/50 dark:border-gray-700/50 flex flex-col bg-white/30 dark:bg-gray-900/30 backdrop-blur-sm">
            <div className="p-6 border-b border-gray-200/50 dark:border-gray-700/50">
              <motion.div
                {...getRootProps()}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className={`border-2 border-dashed rounded-2xl p-8 text-center cursor-pointer transition-all duration-300 ${
                  isDragActive
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 shadow-lg'
                    : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                }`}
              >
                <input {...getInputProps()} />
                <motion.div
                  animate={isDragActive ? { scale: [1, 1.1, 1] } : {}}
                  transition={{ duration: 0.5, repeat: isDragActive ? Infinity : 0 }}
                >
                  <CloudArrowUpIcon className="w-16 h-16 text-blue-500 mx-auto mb-4" />
                </motion.div>
                <p className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  {isDragActive ? 'Drop files here!' : 'Upload Files'}
                </p>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                  Drag and drop files here, or click to browse
                </p>
                <div className="flex flex-wrap gap-1 justify-center">
                  {['PDF', 'DOC', 'IMG', 'VIDEO', 'CODE'].map((type) => (
                    <span
                      key={type}
                      className="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full"
                    >
                      {type}
                    </span>
                  ))}
                </div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-3">
                  Max file size: 100MB
                </p>
              </motion.div>
            </div>

            {/* Enhanced Folders Section */}
            <div className="p-4 border-b border-gray-200/50 dark:border-gray-700/50">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Folders</h3>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => setShowNewFolderModal(true)}
                  className="p-1 text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 transition-colors"
                >
                  <PlusIcon className="w-4 h-4" />
                </motion.button>
              </div>
              <div className="space-y-2">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  onClick={() => setSelectedFolder(null)}
                  className={`w-full text-left p-2 rounded-lg transition-all duration-200 ${
                    !selectedFolder
                      ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                      : 'hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300'
                  }`}
                >
                  <div className="flex items-center space-x-2">
                    <DocumentTextIcon className="w-4 h-4" />
                    <span className="text-sm">All Documents</span>
                    <span className="text-xs text-gray-500 dark:text-gray-400 ml-auto">
                      {documents.length}
                    </span>
                  </div>
                </motion.button>

                {folders.map((folder) => (
                  <motion.button
                    key={folder.id}
                    whileHover={{ scale: 1.02 }}
                    onClick={() => setSelectedFolder(folder.id)}
                    className={`w-full text-left p-2 rounded-lg transition-all duration-200 ${
                      selectedFolder === folder.id
                        ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300'
                    }`}
                  >
                    <div className="flex items-center space-x-2">
                      <FolderIcon className={`w-4 h-4 text-${folder.color}-500`} />
                      <span className="text-sm">{folder.name}</span>
                      <span className="text-xs text-gray-500 dark:text-gray-400 ml-auto">
                        {folder.documentCount}
                      </span>
                      {folder.isShared && (
                        <ShareIcon className="w-3 h-3 text-gray-400" />
                      )}
                    </div>
                  </motion.button>
                ))}
              </div>
            </div>
          </div>

          {/* Enhanced Main Documents Area */}
          <div className="flex-1 flex flex-col">
            {/* Documents Header */}
            <div className="bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                    {selectedFolder
                      ? folders.find(f => f.id === selectedFolder)?.name || 'Documents'
                      : 'All Documents'
                    }
                  </h2>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {filteredAndSortedDocuments.length} files
                    {selectedDocuments.length > 0 && ` • ${selectedDocuments.length} selected`}
                  </p>
                </div>

                {selectedDocuments.length > 0 && (
                  <div className="flex items-center space-x-2">
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="px-3 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
                      onClick={() => {
                        selectedDocuments.forEach(deleteDocument)
                        setSelectedDocuments([])
                      }}
                    >
                      Delete Selected
                    </motion.button>
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="px-3 py-2 text-sm bg-gray-500 text-white rounded-lg hover:bg-gray-600 transition-colors"
                      onClick={() => {
                        selectedDocuments.forEach(archiveDocument)
                        setSelectedDocuments([])
                      }}
                    >
                      Archive Selected
                    </motion.button>
                  </div>
                )}
              </div>
            </div>

            {/* Enhanced Documents Grid/List */}
            <div className="flex-1 overflow-y-auto p-6">
              {filteredAndSortedDocuments.length === 0 ? (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="text-center py-16"
                >
                  <DocumentTextIcon className="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                    {searchQuery ? 'No documents found' : 'No documents yet'}
                  </h3>
                  <p className="text-gray-500 dark:text-gray-400 mb-6">
                    {searchQuery
                      ? 'Try adjusting your search terms or filters'
                      : 'Upload your first document to get started'
                    }
                  </p>
                  {!searchQuery && (
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      onClick={() => document.querySelector('input[type="file"]')?.click()}
                      className="px-6 py-3 bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-colors shadow-lg"
                    >
                      Upload Documents
                    </motion.button>
                  )}
                </motion.div>
              ) : (
                <div className={viewMode === 'grid'
                  ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'
                  : 'space-y-3'
                }>
                  <AnimatePresence>
                    {filteredAndSortedDocuments.map((document, index) => (
                      viewMode === 'grid' ? (
                        <DocumentCard
                          key={document.id}
                          document={document}
                          index={index}
                          isSelected={selectedDocuments.includes(document.id)}
                          onSelect={() => {
                            setSelectedDocuments(prev =>
                              prev.includes(document.id)
                                ? prev.filter(id => id !== document.id)
                                : [...prev, document.id]
                            )
                          }}
                          onView={() => setSelectedDocument(document)}
                          onStar={() => toggleStar(document.id)}
                          onShare={() => toggleShare(document.id)}
                          onDelete={() => deleteDocument(document.id)}
                          onDuplicate={() => duplicateDocument(document.id)}
                          getFileIcon={getFileIcon}
                          formatFileSize={formatFileSize}
                        />
                      ) : (
                        <DocumentListItem
                          key={document.id}
                          document={document}
                          index={index}
                          isSelected={selectedDocuments.includes(document.id)}
                          onSelect={() => {
                            setSelectedDocuments(prev =>
                              prev.includes(document.id)
                                ? prev.filter(id => id !== document.id)
                                : [...prev, document.id]
                            )
                          }}
                          onView={() => setSelectedDocument(document)}
                          onStar={() => toggleStar(document.id)}
                          onShare={() => toggleShare(document.id)}
                          onDelete={() => deleteDocument(document.id)}
                          onDuplicate={() => duplicateDocument(document.id)}
                          getFileIcon={getFileIcon}
                          formatFileSize={formatFileSize}
                        />
                      )
                    ))}
                  </AnimatePresence>
                </div>
              )}
            </div>
        </div>
      </div>
    </div>
  )
}

// Enhanced Document Card Component
function DocumentCard({
  document,
  index,
  isSelected,
  onSelect,
  onView,
  onStar,
  onShare,
  onDelete,
  onDuplicate,
  getFileIcon,
  formatFileSize
}: {
  document: Document
  index: number
  isSelected: boolean
  onSelect: () => void
  onView: () => void
  onStar: () => void
  onShare: () => void
  onDelete: () => void
  onDuplicate: () => void
  getFileIcon: (type: string, category: Document['category']) => React.ReactNode
  formatFileSize: (bytes: number) => string
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={{ delay: index * 0.05 }}
      className={`group relative bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border transition-all duration-200 hover:shadow-xl hover:-translate-y-1 ${
        isSelected
          ? 'border-blue-500 ring-2 ring-blue-200 dark:ring-blue-800'
          : 'border-gray-200/50 dark:border-gray-700/50 hover:border-blue-300 dark:hover:border-blue-600'
      }`}
    >
      {/* Selection Checkbox */}
      <div className="absolute top-3 left-3 z-10">
        <motion.button
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
          onClick={(e) => {
            e.stopPropagation()
            onSelect()
          }}
          className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-all duration-200 ${
            isSelected
              ? 'bg-blue-500 border-blue-500 text-white'
              : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 bg-white dark:bg-gray-800'
          }`}
        >
          {isSelected && <CheckIcon className="w-3 h-3" />}
        </motion.button>
      </div>

      {/* Star Button */}
      <div className="absolute top-3 right-3 z-10">
        <motion.button
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
          onClick={(e) => {
            e.stopPropagation()
            onStar()
          }}
          className={`p-1 rounded-full transition-all duration-200 ${
            document.isStarred
              ? 'text-yellow-500 bg-yellow-100 dark:bg-yellow-900/30'
              : 'text-gray-400 hover:text-yellow-500 hover:bg-yellow-50 dark:hover:bg-yellow-900/20'
          }`}
        >
          <StarIcon className={`w-4 h-4 ${document.isStarred ? 'fill-current' : ''}`} />
        </motion.button>
      </div>

      {/* Document Thumbnail/Icon */}
      <div className="p-6 pb-4">
        <div className="flex items-center justify-center h-20 mb-4">
          {document.thumbnail ? (
            <img
              src={document.thumbnail}
              alt={document.name}
              className="w-full h-full object-cover rounded-lg"
            />
          ) : (
            <div className={`w-16 h-16 rounded-xl flex items-center justify-center ${
              document.category === 'image' ? 'bg-pink-100 text-pink-600' :
              document.category === 'video' ? 'bg-purple-100 text-purple-600' :
              document.category === 'audio' ? 'bg-green-100 text-green-600' :
              document.category === 'code' ? 'bg-blue-100 text-blue-600' :
              document.category === 'data' ? 'bg-orange-100 text-orange-600' :
              document.category === 'presentation' ? 'bg-red-100 text-red-600' :
              'bg-gray-100 text-gray-600'
            }`}>
              {getFileIcon(document.type, document.category)}
            </div>
          )}
        </div>

        {/* Document Info */}
        <div className="space-y-2">
          <h3 className="font-semibold text-gray-900 dark:text-white truncate text-sm">
            {document.name}
          </h3>

          <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
            <span>{formatFileSize(document.size)}</span>
            <span>{document.updatedAt.toLocaleDateString()}</span>
          </div>

          {/* Status Indicator */}
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${
              document.status === 'completed' ? 'bg-green-500' :
              document.status === 'processing' ? 'bg-yellow-500' :
              document.status === 'uploading' ? 'bg-blue-500' :
              document.status === 'error' ? 'bg-red-500' :
              'bg-gray-400'
            }`} />
            <span className="text-xs text-gray-500 dark:text-gray-400 capitalize">
              {document.status}
            </span>
            {document.processingProgress !== undefined && document.status === 'processing' && (
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {document.processingProgress}%
              </span>
            )}
          </div>

          {/* Tags */}
          {document.tags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {document.tags.slice(0, 2).map((tag) => (
                <span
                  key={tag}
                  className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs rounded-full"
                >
                  {tag}
                </span>
              ))}
              {document.tags.length > 2 && (
                <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded-full">
                  +{document.tags.length - 2}
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="px-6 pb-6">
        <div className="flex items-center justify-between">
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onView}
            className="px-4 py-2 bg-blue-500 text-white text-sm rounded-lg hover:bg-blue-600 transition-colors"
          >
            View
          </motion.button>

          <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
            {document.isShared && (
              <ShareIcon className="w-4 h-4 text-green-500" />
            )}

            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={(e) => {
                e.stopPropagation()
                onShare()
              }}
              className="p-1 text-gray-400 hover:text-blue-500 transition-colors"
            >
              <ShareIcon className="w-4 h-4" />
            </motion.button>

            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={(e) => {
                e.stopPropagation()
                onDuplicate()
              }}
              className="p-1 text-gray-400 hover:text-green-500 transition-colors"
            >
              <DocumentDuplicateIcon className="w-4 h-4" />
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
      </div>

      {/* Processing Progress Bar */}
      {document.status === 'processing' && document.processingProgress !== undefined && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-gray-200 dark:bg-gray-700 rounded-b-2xl overflow-hidden">
          <motion.div
            className="h-full bg-blue-500"
            initial={{ width: 0 }}
            animate={{ width: `${document.processingProgress}%` }}
            transition={{ duration: 0.5 }}
          />
        </div>
      )}
    </motion.div>
  )
}

// Enhanced Document List Item Component
function DocumentListItem({
  document,
  index,
  isSelected,
  onSelect,
  onView,
  onStar,
  onShare,
  onDelete,
  onDuplicate,
  getFileIcon,
  formatFileSize
}: {
  document: Document
  index: number
  isSelected: boolean
  onSelect: () => void
  onView: () => void
  onStar: () => void
  onShare: () => void
  onDelete: () => void
  onDuplicate: () => void
  getFileIcon: (type: string, category: Document['category']) => React.ReactNode
  formatFileSize: (bytes: number) => string
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 20 }}
      transition={{ delay: index * 0.02 }}
      className={`group flex items-center space-x-4 p-4 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl border transition-all duration-200 hover:shadow-lg ${
        isSelected
          ? 'border-blue-500 ring-2 ring-blue-200 dark:ring-blue-800'
          : 'border-gray-200/50 dark:border-gray-700/50 hover:border-blue-300 dark:hover:border-blue-600'
      }`}
    >
      {/* Selection Checkbox */}
      <motion.button
        whileHover={{ scale: 1.1 }}
        whileTap={{ scale: 0.9 }}
        onClick={onSelect}
        className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-all duration-200 ${
          isSelected
            ? 'bg-blue-500 border-blue-500 text-white'
            : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 bg-white dark:bg-gray-800'
        }`}
      >
        {isSelected && <CheckIcon className="w-3 h-3" />}
      </motion.button>

      {/* File Icon */}
      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
        document.category === 'image' ? 'bg-pink-100 text-pink-600' :
        document.category === 'video' ? 'bg-purple-100 text-purple-600' :
        document.category === 'audio' ? 'bg-green-100 text-green-600' :
        document.category === 'code' ? 'bg-blue-100 text-blue-600' :
        document.category === 'data' ? 'bg-orange-100 text-orange-600' :
        document.category === 'presentation' ? 'bg-red-100 text-red-600' :
        'bg-gray-100 text-gray-600'
      }`}>
        {getFileIcon(document.type, document.category)}
      </div>

      {/* Document Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center space-x-2">
          <h3 className="font-medium text-gray-900 dark:text-white truncate">
            {document.name}
          </h3>
          {document.isStarred && (
            <StarIcon className="w-4 h-4 text-yellow-500 fill-current" />
          )}
          {document.isShared && (
            <ShareIcon className="w-4 h-4 text-green-500" />
          )}
        </div>

        <div className="flex items-center space-x-4 mt-1 text-sm text-gray-500 dark:text-gray-400">
          <span>{formatFileSize(document.size)}</span>
          <span>{document.updatedAt.toLocaleDateString()}</span>
          <div className="flex items-center space-x-1">
            <div className={`w-2 h-2 rounded-full ${
              document.status === 'completed' ? 'bg-green-500' :
              document.status === 'processing' ? 'bg-yellow-500' :
              document.status === 'uploading' ? 'bg-blue-500' :
              document.status === 'error' ? 'bg-red-500' :
              'bg-gray-400'
            }`} />
            <span className="capitalize">{document.status}</span>
          </div>
        </div>

        {/* Tags */}
        {document.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mt-2">
            {document.tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs rounded-full"
              >
                {tag}
              </span>
            ))}
            {document.tags.length > 3 && (
              <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded-full">
                +{document.tags.length - 3}
              </span>
            )}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center space-x-2">
        <motion.button
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onView}
          className="px-3 py-1 bg-blue-500 text-white text-sm rounded-lg hover:bg-blue-600 transition-colors"
        >
          View
        </motion.button>

        <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            onClick={onStar}
            className={`p-1 rounded transition-colors ${
              document.isStarred
                ? 'text-yellow-500'
                : 'text-gray-400 hover:text-yellow-500'
            }`}
          >
            <StarIcon className={`w-4 h-4 ${document.isStarred ? 'fill-current' : ''}`} />
          </motion.button>

          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            onClick={onShare}
            className="p-1 text-gray-400 hover:text-blue-500 transition-colors"
          >
            <ShareIcon className="w-4 h-4" />
          </motion.button>

          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            onClick={onDuplicate}
            className="p-1 text-gray-400 hover:text-green-500 transition-colors"
          >
            <DocumentDuplicateIcon className="w-4 h-4" />
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
