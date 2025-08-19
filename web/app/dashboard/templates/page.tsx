'use client'

import { useState, useEffect } from 'react'
import { 
  SparklesIcon, 
  PlusIcon, 
  PencilIcon,
  TrashIcon,
  PlayIcon,
  DocumentDuplicateIcon,
  TagIcon,
  ClockIcon,
  UserIcon,
  CodeBracketIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'

interface PromptTemplate {
  id: string
  name: string
  description: string
  category: string
  template: string
  variables: TemplateVariable[]
  examples: TemplateExample[]
  config: TemplateConfig
  tags: string[]
  version: string
  createdAt: Date
  updatedAt: Date
  createdBy: string
  isActive: boolean
  usageCount: number
}

interface TemplateVariable {
  name: string
  type: 'string' | 'number' | 'boolean' | 'array' | 'object'
  description: string
  required: boolean
  defaultValue?: any
}

interface TemplateExample {
  name: string
  description: string
  variables: Record<string, any>
  expected: string
}

interface TemplateConfig {
  modelId: string
  temperature: number
  maxTokens: number
  topP: number
  systemPrompt?: string
}

export default function TemplatesPage() {
  const [mounted, setMounted] = useState(false)
  const [templates, setTemplates] = useState<PromptTemplate[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState<PromptTemplate | null>(null)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [isExecuteModalOpen, setIsExecuteModalOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [loading, setLoading] = useState(true)

  const categories = ['all', 'text_processing', 'translation', 'analysis', 'generation', 'custom']

  useEffect(() => {
    loadTemplates()
  }, [])

  useEffect(() => {
    setMounted(true)
  }, [])

  const loadTemplates = async () => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      const mockTemplates: PromptTemplate[] = [
        {
          id: '1',
          name: 'Text Summarization',
          description: 'Summarizes long text into key points',
          category: 'text_processing',
          template: 'Please summarize the following text in {{max_points}} key points:\n\n{{text}}',
          variables: [
            {
              name: 'text',
              type: 'string',
              description: 'Text to summarize',
              required: true
            },
            {
              name: 'max_points',
              type: 'number',
              description: 'Maximum number of summary points',
              required: false,
              defaultValue: 5
            }
          ],
          examples: [
            {
              name: 'Article Summary',
              description: 'Summarize a news article',
              variables: {
                text: 'Long article text here...',
                max_points: 3
              },
              expected: '1. Main point one\n2. Main point two\n3. Main point three'
            }
          ],
          config: {
            modelId: 'gpt-3.5-turbo',
            temperature: 0.3,
            maxTokens: 500,
            topP: 1.0
          },
          tags: ['summarization', 'text_processing'],
          version: '1.0',
          createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          createdBy: 'admin',
          isActive: true,
          usageCount: 156
        },
        {
          id: '2',
          name: 'Language Translation',
          description: 'Translates text between languages',
          category: 'translation',
          template: 'Translate the following text from {{source_language}} to {{target_language}}:\n\n{{text}}',
          variables: [
            {
              name: 'text',
              type: 'string',
              description: 'Text to translate',
              required: true
            },
            {
              name: 'source_language',
              type: 'string',
              description: 'Source language',
              required: true
            },
            {
              name: 'target_language',
              type: 'string',
              description: 'Target language',
              required: true
            }
          ],
          examples: [
            {
              name: 'English to Spanish',
              description: 'Translate English text to Spanish',
              variables: {
                text: 'Hello, how are you?',
                source_language: 'English',
                target_language: 'Spanish'
              },
              expected: 'Hola, ¿cómo estás?'
            }
          ],
          config: {
            modelId: 'gpt-3.5-turbo',
            temperature: 0.1,
            maxTokens: 1000,
            topP: 1.0
          },
          tags: ['translation', 'language'],
          version: '1.2',
          createdAt: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
          createdBy: 'admin',
          isActive: true,
          usageCount: 89
        },
        {
          id: '3',
          name: 'Code Review',
          description: 'Reviews code and provides feedback',
          category: 'analysis',
          template: 'Review the following {{language}} code and provide feedback on:\n- Code quality\n- Best practices\n- Potential improvements\n- Security concerns\n\nCode:\n```{{language}}\n{{code}}\n```',
          variables: [
            {
              name: 'code',
              type: 'string',
              description: 'Code to review',
              required: true
            },
            {
              name: 'language',
              type: 'string',
              description: 'Programming language',
              required: true
            }
          ],
          examples: [
            {
              name: 'JavaScript Function',
              description: 'Review a JavaScript function',
              variables: {
                code: 'function add(a, b) { return a + b; }',
                language: 'javascript'
              },
              expected: 'Code review feedback here...'
            }
          ],
          config: {
            modelId: 'gpt-4',
            temperature: 0.3,
            maxTokens: 1000,
            topP: 1.0,
            systemPrompt: 'You are an expert code reviewer with years of experience in software development.'
          },
          tags: ['code', 'review', 'analysis'],
          version: '1.0',
          createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          updatedAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          createdBy: 'developer',
          isActive: true,
          usageCount: 42
        }
      ]

      setTemplates(mockTemplates)
    } catch (error) {
      console.error('Failed to load templates:', error)
    } finally {
      setLoading(false)
    }
  }

  const filteredTemplates = templates.filter(template => {
    const matchesSearch = template.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         template.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         template.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
    
    const matchesCategory = selectedCategory === 'all' || template.category === selectedCategory
    
    return matchesSearch && matchesCategory && template.isActive
  })

  const executeTemplate = async (template: PromptTemplate, variables: Record<string, any>) => {
    try {
      const response = await fetch(`/api/ai/templates/${template.id}/execute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          variables,
          user_id: 'current_user'
        })
      })

      if (!response.ok) {
        throw new Error('Failed to execute template')
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Template execution error:', error)
      throw error
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading templates...</p>
        </div>
      </div>
    )
  }

  if (!mounted) {
    return null
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Prompt Templates</h1>
            <p className="text-sm text-gray-500">
              Create and manage reusable AI prompt templates
            </p>
          </div>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
          >
            <PlusIcon className="w-5 h-5" />
            <span>New Template</span>
          </button>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Filters */}
        <div className="flex items-center space-x-4 mb-6">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search templates..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="border border-gray-300 rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            {categories.map(category => (
              <option key={category} value={category}>
                {category === 'all' ? 'All Categories' : category.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
              </option>
            ))}
          </select>
        </div>

        {/* Templates Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <AnimatePresence>
            {filteredTemplates.map((template) => (
              <motion.div
                key={template.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                      {template.name}
                    </h3>
                    <p className="text-sm text-gray-600 mb-3">
                      {template.description}
                    </p>
                  </div>
                  <div className="flex items-center space-x-1">
                    <button
                      onClick={() => {
                        setSelectedTemplate(template)
                        setIsExecuteModalOpen(true)
                      }}
                      className="p-2 text-gray-400 hover:text-blue-600 transition-colors"
                      title="Execute template"
                    >
                      <PlayIcon className="w-4 h-4" />
                    </button>
                    <button
                      className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                      title="Edit template"
                    >
                      <PencilIcon className="w-4 h-4" />
                    </button>
                    <button
                      className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                      title="Delete template"
                    >
                      <TrashIcon className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                {/* Category and Tags */}
                <div className="flex flex-wrap gap-2 mb-4">
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                    {template.category.replace('_', ' ')}
                  </span>
                  {template.tags.slice(0, 2).map((tag) => (
                    <span
                      key={tag}
                      className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full"
                    >
                      {tag}
                    </span>
                  ))}
                  {template.tags.length > 2 && (
                    <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                      +{template.tags.length - 2}
                    </span>
                  )}
                </div>

                {/* Variables */}
                <div className="mb-4">
                  <div className="flex items-center space-x-2 mb-2">
                    <CodeBracketIcon className="w-4 h-4 text-gray-400" />
                    <span className="text-sm font-medium text-gray-700">
                      Variables ({template.variables.length})
                    </span>
                  </div>
                  <div className="space-y-1">
                    {template.variables.slice(0, 3).map((variable) => (
                      <div key={variable.name} className="flex items-center justify-between text-xs">
                        <span className="text-gray-600">{variable.name}</span>
                        <span className={`px-1 py-0.5 rounded text-xs ${
                          variable.required 
                            ? 'bg-red-100 text-red-600' 
                            : 'bg-green-100 text-green-600'
                        }`}>
                          {variable.required ? 'required' : 'optional'}
                        </span>
                      </div>
                    ))}
                    {template.variables.length > 3 && (
                      <div className="text-xs text-gray-400">
                        +{template.variables.length - 3} more
                      </div>
                    )}
                  </div>
                </div>

                {/* Stats */}
                <div className="flex items-center justify-between text-sm text-gray-500">
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-1">
                      <ClockIcon className="w-4 h-4" />
                      <span>v{template.version}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <UserIcon className="w-4 h-4" />
                      <span>{template.usageCount}</span>
                    </div>
                  </div>
                  <span>{template.updatedAt.toLocaleDateString()}</span>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        {filteredTemplates.length === 0 && (
          <div className="text-center py-12">
            <SparklesIcon className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              No templates found
            </h3>
            <p className="text-gray-500 mb-4">
              {searchQuery || selectedCategory !== 'all' 
                ? 'Try adjusting your search or filters'
                : 'Get started by creating your first template'
              }
            </p>
            {!searchQuery && selectedCategory === 'all' && (
              <button
                onClick={() => setIsCreateModalOpen(true)}
                className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
              >
                Create Template
              </button>
            )}
          </div>
        )}
      </div>

      {/* Execute Template Modal */}
      {isExecuteModalOpen && selectedTemplate && (
        <TemplateExecuteModal
          template={selectedTemplate}
          onClose={() => {
            setIsExecuteModalOpen(false)
            setSelectedTemplate(null)
          }}
          onExecute={executeTemplate}
        />
      )}
    </div>
  )
}

function TemplateExecuteModal({ 
  template, 
  onClose, 
  onExecute 
}: {
  template: PromptTemplate
  onClose: () => void
  onExecute: (template: PromptTemplate, variables: Record<string, any>) => Promise<any>
}) {
  const [variables, setVariables] = useState<Record<string, any>>({})
  const [isExecuting, setIsExecuting] = useState(false)
  const [result, setResult] = useState<any>(null)

  useEffect(() => {
    // Initialize variables with default values
    const initialVariables: Record<string, any> = {}
    template.variables.forEach(variable => {
      if (variable.defaultValue !== undefined) {
        initialVariables[variable.name] = variable.defaultValue
      }
    })
    setVariables(initialVariables)
  }, [template])

  const handleExecute = async () => {
    setIsExecuting(true)
    try {
      const result = await onExecute(template, variables)
      setResult(result)
    } catch (error) {
      console.error('Execution failed:', error)
    } finally {
      setIsExecuting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              Execute Template: {template.name}
            </h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              ×
            </button>
          </div>
        </div>

        <div className="p-6">
          {/* Variables Input */}
          <div className="space-y-4 mb-6">
            <h3 className="text-lg font-medium text-gray-900">Variables</h3>
            {template.variables.map((variable) => (
              <div key={variable.name}>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {variable.name}
                  {variable.required && <span className="text-red-500 ml-1">*</span>}
                </label>
                <p className="text-xs text-gray-500 mb-2">{variable.description}</p>
                {variable.type === 'string' ? (
                  <textarea
                    value={variables[variable.name] || ''}
                    onChange={(e) => setVariables(prev => ({
                      ...prev,
                      [variable.name]: e.target.value
                    }))}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    rows={3}
                    placeholder={`Enter ${variable.name}...`}
                  />
                ) : (
                  <input
                    type={variable.type === 'number' ? 'number' : 'text'}
                    value={variables[variable.name] || ''}
                    onChange={(e) => setVariables(prev => ({
                      ...prev,
                      [variable.name]: variable.type === 'number' 
                        ? parseFloat(e.target.value) || 0
                        : e.target.value
                    }))}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder={`Enter ${variable.name}...`}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Execute Button */}
          <div className="flex justify-end space-x-3 mb-6">
            <button
              onClick={onClose}
              className="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleExecute}
              disabled={isExecuting}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center space-x-2"
            >
              {isExecuting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  <span>Executing...</span>
                </>
              ) : (
                <>
                  <PlayIcon className="w-4 h-4" />
                  <span>Execute</span>
                </>
              )}
            </button>
          </div>

          {/* Result */}
          {result && (
            <div>
              <h3 className="text-lg font-medium text-gray-900 mb-3">Result</h3>
              <div className="bg-gray-50 rounded-lg p-4">
                <pre className="text-sm text-gray-700 whitespace-pre-wrap">
                  {result.text || JSON.stringify(result, null, 2)}
                </pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
