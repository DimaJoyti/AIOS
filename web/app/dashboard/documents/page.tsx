'use client'

import { useState, useCallback } from 'react'
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
  TagIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useDropzone } from 'react-dropzone'

interface Document {
  id: string
  name: string
  type: string
  size: number
  uploadedAt: Date
  status: 'processing' | 'completed' | 'error'
  tags: string[]
  summary?: string
  extractedText?: string
  metadata?: {
    pages?: number
    wordCount?: number
    language?: string
  }
}

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<Document[]>([
    {
      id: '1',
      name: 'Research Paper - AI in Healthcare.pdf',
      type: 'application/pdf',
      size: 2048576,
      uploadedAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
      status: 'completed',
      tags: ['research', 'healthcare', 'ai'],
      summary: 'A comprehensive study on the applications of artificial intelligence in modern healthcare systems.',
      metadata: {
        pages: 24,
        wordCount: 8500,
        language: 'en'
      }
    },
    {
      id: '2',
      name: 'Project Proposal.docx',
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      size: 512000,
      uploadedAt: new Date(Date.now() - 5 * 60 * 60 * 1000),
      status: 'completed',
      tags: ['project', 'proposal'],
      summary: 'Detailed project proposal for implementing AI-driven automation in business processes.',
      metadata: {
        pages: 12,
        wordCount: 3200,
        language: 'en'
      }
    },
    {
      id: '3',
      name: 'Meeting Notes.txt',
      type: 'text/plain',
      size: 8192,
      uploadedAt: new Date(Date.now() - 30 * 60 * 1000),
      status: 'processing',
      tags: ['meeting', 'notes']
    }
  ])

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    setIsUploading(true)
    
    for (const file of acceptedFiles) {
      const newDocument: Document = {
        id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
        name: file.name,
        type: file.type,
        size: file.size,
        uploadedAt: new Date(),
        status: 'processing',
        tags: []
      }

      setDocuments(prev => [newDocument, ...prev])

      try {
        // Simulate file upload and processing
        const formData = new FormData()
        formData.append('file', file)

        const response = await fetch('/api/documents/upload', {
          method: 'POST',
          body: formData
        })

        if (!response.ok) {
          throw new Error('Upload failed')
        }

        const result = await response.json()

        // Update document with processing results
        setDocuments(prev => prev.map(doc => 
          doc.id === newDocument.id 
            ? {
                ...doc,
                status: 'completed',
                summary: result.summary,
                extractedText: result.extractedText,
                metadata: result.metadata,
                tags: result.suggestedTags || []
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

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf'],
      'application/msword': ['.doc'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'text/plain': ['.txt'],
      'text/markdown': ['.md'],
      'application/rtf': ['.rtf']
    },
    multiple: true
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

  const getFileIcon = (type: string) => {
    if (type.includes('pdf')) return '📄'
    if (type.includes('word') || type.includes('document')) return '📝'
    if (type.includes('text')) return '📃'
    return '📄'
  }

  const deleteDocument = async (documentId: string) => {
    try {
      await fetch(`/api/documents/${documentId}`, {
        method: 'DELETE'
      })
      setDocuments(prev => prev.filter(doc => doc.id !== documentId))
      if (selectedDocument?.id === documentId) {
        setSelectedDocument(null)
      }
    } catch (error) {
      console.error('Delete error:', error)
    }
  }

  return (
    <div className="h-screen bg-gray-50 flex">
      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="bg-white border-b border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Documents</h1>
              <p className="text-sm text-gray-500">
                Upload and manage your documents for AI processing
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <div className="relative">
                <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search documents..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="flex-1 flex">
          {/* Documents List */}
          <div className="w-1/2 border-r border-gray-200 flex flex-col">
            {/* Upload Area */}
            <div className="p-6 border-b border-gray-200">
              <div
                {...getRootProps()}
                className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
                  isDragActive
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
              >
                <input {...getInputProps()} />
                <CloudArrowUpIcon className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-lg font-medium text-gray-900 mb-2">
                  {isDragActive ? 'Drop files here' : 'Upload documents'}
                </p>
                <p className="text-sm text-gray-500">
                  Drag and drop files here, or click to select files
                </p>
                <p className="text-xs text-gray-400 mt-2">
                  Supports PDF, DOC, DOCX, TXT, MD, RTF
                </p>
              </div>
            </div>

            {/* Documents List */}
            <div className="flex-1 overflow-y-auto">
              <div className="p-4 space-y-3">
                <AnimatePresence>
                  {filteredDocuments.map((document) => (
                    <motion.div
                      key={document.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -20 }}
                      onClick={() => setSelectedDocument(document)}
                      className={`p-4 rounded-lg border cursor-pointer transition-all ${
                        selectedDocument?.id === document.id
                          ? 'border-blue-500 bg-blue-50'
                          : 'border-gray-200 hover:border-gray-300 bg-white'
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-start space-x-3 flex-1">
                          <div className="text-2xl">
                            {getFileIcon(document.type)}
                          </div>
                          <div className="flex-1 min-w-0">
                            <h3 className="font-medium text-gray-900 truncate">
                              {document.name}
                            </h3>
                            <div className="flex items-center space-x-4 mt-1 text-sm text-gray-500">
                              <span>{formatFileSize(document.size)}</span>
                              <span>{document.uploadedAt.toLocaleDateString()}</span>
                            </div>
                            {document.tags.length > 0 && (
                              <div className="flex flex-wrap gap-1 mt-2">
                                {document.tags.map((tag) => (
                                  <span
                                    key={tag}
                                    className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full"
                                  >
                                    {tag}
                                  </span>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <div className={`w-2 h-2 rounded-full ${
                            document.status === 'completed' 
                              ? 'bg-green-500' 
                              : document.status === 'processing'
                              ? 'bg-yellow-500'
                              : 'bg-red-500'
                          }`}></div>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              deleteDocument(document.id)
                            }}
                            className="p-1 text-gray-400 hover:text-red-500 transition-colors"
                          >
                            <TrashIcon className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </AnimatePresence>
              </div>
            </div>
          </div>

          {/* Document Details */}
          <div className="w-1/2 flex flex-col">
            {selectedDocument ? (
              <div className="flex-1 overflow-y-auto">
                <div className="p-6">
                  <div className="flex items-start justify-between mb-6">
                    <div>
                      <h2 className="text-xl font-bold text-gray-900 mb-2">
                        {selectedDocument.name}
                      </h2>
                      <div className="flex items-center space-x-4 text-sm text-gray-500">
                        <div className="flex items-center space-x-1">
                          <CalendarIcon className="w-4 h-4" />
                          <span>{selectedDocument.uploadedAt.toLocaleDateString()}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <DocumentTextIcon className="w-4 h-4" />
                          <span>{formatFileSize(selectedDocument.size)}</span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <button className="p-2 text-gray-400 hover:text-gray-600 transition-colors">
                        <EyeIcon className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-gray-400 hover:text-gray-600 transition-colors">
                        <DocumentArrowDownIcon className="w-5 h-5" />
                      </button>
                    </div>
                  </div>

                  {/* Status */}
                  <div className="mb-6">
                    <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                      selectedDocument.status === 'completed'
                        ? 'bg-green-100 text-green-800'
                        : selectedDocument.status === 'processing'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      <div className={`w-2 h-2 rounded-full mr-2 ${
                        selectedDocument.status === 'completed'
                          ? 'bg-green-500'
                          : selectedDocument.status === 'processing'
                          ? 'bg-yellow-500'
                          : 'bg-red-500'
                      }`}></div>
                      {selectedDocument.status === 'completed' && 'Processing Complete'}
                      {selectedDocument.status === 'processing' && 'Processing...'}
                      {selectedDocument.status === 'error' && 'Processing Error'}
                    </div>
                  </div>

                  {/* Metadata */}
                  {selectedDocument.metadata && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-3">Document Info</h3>
                      <div className="grid grid-cols-2 gap-4">
                        {selectedDocument.metadata.pages && (
                          <div>
                            <span className="text-sm text-gray-500">Pages</span>
                            <p className="font-medium">{selectedDocument.metadata.pages}</p>
                          </div>
                        )}
                        {selectedDocument.metadata.wordCount && (
                          <div>
                            <span className="text-sm text-gray-500">Words</span>
                            <p className="font-medium">{selectedDocument.metadata.wordCount.toLocaleString()}</p>
                          </div>
                        )}
                        {selectedDocument.metadata.language && (
                          <div>
                            <span className="text-sm text-gray-500">Language</span>
                            <p className="font-medium">{selectedDocument.metadata.language.toUpperCase()}</p>
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  {/* Tags */}
                  {selectedDocument.tags.length > 0 && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-3">Tags</h3>
                      <div className="flex flex-wrap gap-2">
                        {selectedDocument.tags.map((tag) => (
                          <span
                            key={tag}
                            className="px-3 py-1 bg-blue-100 text-blue-800 text-sm rounded-full"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Summary */}
                  {selectedDocument.summary && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-3">Summary</h3>
                      <div className="bg-gray-50 rounded-lg p-4">
                        <p className="text-gray-700 leading-relaxed">
                          {selectedDocument.summary}
                        </p>
                      </div>
                    </div>
                  )}

                  {/* Extracted Text Preview */}
                  {selectedDocument.extractedText && (
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 mb-3">Content Preview</h3>
                      <div className="bg-gray-50 rounded-lg p-4 max-h-96 overflow-y-auto">
                        <p className="text-gray-700 text-sm leading-relaxed whitespace-pre-wrap">
                          {selectedDocument.extractedText.substring(0, 1000)}
                          {selectedDocument.extractedText.length > 1000 && '...'}
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <DocumentTextIcon className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    Select a document
                  </h3>
                  <p className="text-gray-500">
                    Choose a document from the list to view its details
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
