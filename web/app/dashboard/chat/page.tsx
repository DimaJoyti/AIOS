'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import {
  PaperAirplaneIcon,
  SparklesIcon,
  UserIcon,
  CpuChipIcon,
  ClockIcon,
  CurrencyDollarIcon,
  DocumentTextIcon,
  PhotoIcon,
  MicrophoneIcon,
  PlusIcon,
  EllipsisVerticalIcon,
  TrashIcon,
  PencilIcon,
  BookmarkIcon,
  ShareIcon,
  ArrowDownIcon,
  StopIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  AdjustmentsHorizontalIcon,
  ChatBubbleLeftRightIcon,
  BoltIcon,
  BeakerIcon,
  CodeBracketIcon,
  LanguageIcon,
  LightBulbIcon,
  MagnifyingGlassIcon,
  XMarkIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  type?: 'text' | 'code' | 'image' | 'file' | 'error'
  metadata?: {
    model?: string
    tokens?: number
    cost?: number
    latency?: number
    temperature?: number
    reasoning?: string
  }
  attachments?: {
    type: 'image' | 'document' | 'code'
    name: string
    url: string
    size?: number
  }[]
  isStreaming?: boolean
  isBookmarked?: boolean
  reactions?: string[]
}

interface ChatSession {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
  updatedAt: Date
  model: string
  settings: {
    temperature: number
    maxTokens: number
    systemPrompt?: string
  }
  tags?: string[]
  isArchived?: boolean
  messageCount: number
  totalCost: number
}

interface AIModel {
  id: string
  name: string
  description: string
  provider: string
  capabilities: string[]
  costPer1kTokens: number
  maxTokens: number
  isAvailable: boolean
  icon: string
  color: string
}

interface ChatSettings {
  temperature: number
  maxTokens: number
  topP: number
  frequencyPenalty: number
  presencePenalty: number
  systemPrompt: string
}

export default function ChatPage() {
  const [mounted, setMounted] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'assistant',
      content: 'Welcome to AIOS Epic Chat! 🚀\n\nI\'m your advanced AI assistant with access to multiple cutting-edge models. I can help you with:\n\n• **Text Generation** - Creative writing, summaries, translations\n• **Code Development** - Programming, debugging, code review\n• **Document Analysis** - PDF processing, data extraction\n• **Image Understanding** - Visual analysis and description\n• **Research & Analysis** - Deep insights and explanations\n\nChoose your preferred AI model from the sidebar and let\'s get started! What would you like to explore today?',
      timestamp: new Date(),
      type: 'text',
      metadata: {
        model: 'gpt-4',
        tokens: 95,
        cost: 0.0038,
        latency: 187
      }
    }
  ])

  const [inputValue, setInputValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isStreaming, setIsStreaming] = useState(false)
  const [selectedModel, setSelectedModel] = useState('gpt-4')
  const [sessions, setSessions] = useState<ChatSession[]>([])
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null)
  const [showSettings, setShowSettings] = useState(false)
  const [showModelSelector, setShowModelSelector] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [copiedMessageId, setCopiedMessageId] = useState<string | null>(null)
  const [chatSettings, setChatSettings] = useState<ChatSettings>({
    temperature: 0.7,
    maxTokens: 2000,
    topP: 1,
    frequencyPenalty: 0,
    presencePenalty: 0,
    systemPrompt: ''
  })

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const models: AIModel[] = [
    {
      id: 'gpt-4',
      name: 'GPT-4',
      description: 'Most capable model for complex reasoning',
      provider: 'OpenAI',
      capabilities: ['text', 'code', 'analysis', 'reasoning'],
      costPer1kTokens: 0.03,
      maxTokens: 8192,
      isAvailable: true,
      icon: '🧠',
      color: 'blue'
    },
    {
      id: 'gpt-4-turbo',
      name: 'GPT-4 Turbo',
      description: 'Latest GPT-4 with improved speed and context',
      provider: 'OpenAI',
      capabilities: ['text', 'code', 'analysis', 'vision', 'reasoning'],
      costPer1kTokens: 0.01,
      maxTokens: 128000,
      isAvailable: true,
      icon: '⚡',
      color: 'purple'
    },
    {
      id: 'gpt-3.5-turbo',
      name: 'GPT-3.5 Turbo',
      description: 'Fast and efficient for most tasks',
      provider: 'OpenAI',
      capabilities: ['text', 'code', 'chat'],
      costPer1kTokens: 0.002,
      maxTokens: 4096,
      isAvailable: true,
      icon: '🚀',
      color: 'green'
    },
    {
      id: 'claude-3-opus',
      name: 'Claude 3 Opus',
      description: 'Anthropic\'s most powerful model',
      provider: 'Anthropic',
      capabilities: ['text', 'code', 'analysis', 'reasoning', 'safety'],
      costPer1kTokens: 0.015,
      maxTokens: 200000,
      isAvailable: true,
      icon: '🎭',
      color: 'orange'
    },
    {
      id: 'claude-3-sonnet',
      name: 'Claude 3 Sonnet',
      description: 'Balanced performance and speed',
      provider: 'Anthropic',
      capabilities: ['text', 'code', 'analysis'],
      costPer1kTokens: 0.003,
      maxTokens: 200000,
      isAvailable: true,
      icon: '🎵',
      color: 'teal'
    },
    {
      id: 'gemini-pro',
      name: 'Gemini Pro',
      description: 'Google\'s advanced multimodal model',
      provider: 'Google',
      capabilities: ['text', 'code', 'vision', 'multimodal'],
      costPer1kTokens: 0.0005,
      maxTokens: 32768,
      isAvailable: true,
      icon: '💎',
      color: 'indigo'
    }
  ]

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    // Auto-focus input when not loading
    if (!isLoading && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isLoading])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const getCurrentModel = () => models.find(m => m.id === selectedModel) || models[0]

  const copyToClipboard = useCallback(async (text: string, messageId: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedMessageId(messageId)
      setTimeout(() => setCopiedMessageId(null), 2000)
    } catch (err) {
      console.error('Failed to copy text: ', err)
    }
  }, [])

  const toggleBookmark = useCallback((messageId: string) => {
    setMessages(prev => prev.map(msg =>
      msg.id === messageId
        ? { ...msg, isBookmarked: !msg.isBookmarked }
        : msg
    ))
  }, [])

  useEffect(() => {
    setMounted(true)
  }, [])

  const deleteMessage = useCallback((messageId: string) => {
    setMessages(prev => prev.filter(msg => msg.id !== messageId))
  }, [])

  const regenerateResponse = useCallback(async (messageId: string) => {
    const messageIndex = messages.findIndex(msg => msg.id === messageId)
    if (messageIndex === -1) return

    // Remove the message and all subsequent messages
    const newMessages = messages.slice(0, messageIndex)
    setMessages(newMessages)

    // Find the last user message to regenerate from
    const lastUserMessage = newMessages.reverse().find(msg => msg.role === 'user')
    if (lastUserMessage) {
      setInputValue(lastUserMessage.content)
      handleSendMessage()
    }
  }, [messages])

  const handleSendMessage = async () => {
    if (!inputValue.trim() || isLoading) return

    const currentModel = getCurrentModel()
    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue.trim(),
      timestamp: new Date(),
      type: 'text'
    }

    setMessages(prev => [...prev, userMessage])
    setInputValue('')
    setIsLoading(true)
    setIsStreaming(true)

    // Create a placeholder assistant message for streaming
    const assistantMessageId = (Date.now() + 1).toString()
    const assistantMessage: Message = {
      id: assistantMessageId,
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      type: 'text',
      isStreaming: true,
      metadata: {
        model: selectedModel,
        temperature: chatSettings.temperature
      }
    }

    setMessages(prev => [...prev, assistantMessage])

    try {
      // Simulate streaming response
      const simulatedResponse = generateSimulatedResponse(userMessage.content, currentModel)

      // Simulate streaming by updating content gradually
      for (let i = 0; i <= simulatedResponse.length; i += 3) {
        await new Promise(resolve => setTimeout(resolve, 50))
        const partialContent = simulatedResponse.slice(0, i)

        setMessages(prev => prev.map(msg =>
          msg.id === assistantMessageId
            ? {
                ...msg,
                content: partialContent,
                isStreaming: i < simulatedResponse.length
              }
            : msg
        ))
      }

      // Final update with metadata
      const finalMetadata = {
        model: selectedModel,
        tokens: Math.floor(simulatedResponse.length / 4),
        cost: (Math.floor(simulatedResponse.length / 4) / 1000) * currentModel.costPer1kTokens,
        latency: Math.floor(Math.random() * 500) + 200,
        temperature: chatSettings.temperature
      }

      setMessages(prev => prev.map(msg =>
        msg.id === assistantMessageId
          ? {
              ...msg,
              isStreaming: false,
              metadata: finalMetadata
            }
          : msg
      ))

    } catch (error) {
      console.error('Error sending message:', error)

      setMessages(prev => prev.map(msg =>
        msg.id === assistantMessageId
          ? {
              ...msg,
              content: '❌ I apologize, but I encountered an error processing your request. Please try again.',
              type: 'error',
              isStreaming: false
            }
          : msg
      ))
    } finally {
      setIsLoading(false)
      setIsStreaming(false)
    }
  }

  const generateSimulatedResponse = (userInput: string, model: AIModel): string => {
    const responses = [
      `I understand you're asking about "${userInput}". Let me provide you with a comprehensive response.

This is a simulated response from ${model.name} (${model.provider}). In a real implementation, this would be connected to the actual AI model API.

Here are some key points to consider:

• **Analysis**: Your question touches on several important aspects that I can help clarify.
• **Context**: Based on the information provided, I can offer relevant insights.
• **Solutions**: Let me suggest some practical approaches you might consider.

Would you like me to elaborate on any specific aspect of this topic?`,

      `Great question! As ${model.name}, I can help you with that.

Here's my analysis:

\`\`\`
// Example code snippet
function processUserInput(input) {
  return {
    analysis: analyzeInput(input),
    suggestions: generateSuggestions(input),
    confidence: calculateConfidence(input)
  };
}
\`\`\`

This approach would allow you to:
1. Process the input systematically
2. Generate relevant suggestions
3. Provide confidence metrics

Is there a particular aspect you'd like me to focus on?`,

      `Excellent question about "${userInput}"!

Let me break this down for you:

**Understanding the Context**
Your inquiry relates to several interconnected concepts that are worth exploring in detail.

**Key Considerations**
- Technical feasibility and implementation approaches
- Best practices and industry standards
- Potential challenges and mitigation strategies

**Recommendations**
Based on my analysis using ${model.name}'s capabilities, I'd suggest starting with a systematic approach that considers both immediate needs and long-term scalability.

Would you like me to dive deeper into any of these areas?`
    ]

    return responses[Math.floor(Math.random() * responses.length)]
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  const startNewSession = () => {
    const currentModel = getCurrentModel()
    const newSession: ChatSession = {
      id: Date.now().toString(),
      title: `New Chat`,
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date(),
      model: selectedModel,
      settings: { ...chatSettings },
      messageCount: 0,
      totalCost: 0
    }

    setSessions(prev => [newSession, ...prev])
    setCurrentSessionId(newSession.id)
    setMessages([{
      id: '1',
      role: 'assistant',
      content: `Welcome to your new chat session! 🎉\n\nI'm ${currentModel.name} from ${currentModel.provider}, ready to assist you with:\n\n${currentModel.capabilities.map(cap => `• ${cap.charAt(0).toUpperCase() + cap.slice(1)}`).join('\n')}\n\nWhat would you like to explore today?`,
      timestamp: new Date(),
      type: 'text',
      metadata: {
        model: selectedModel,
        tokens: 45,
        cost: 0.0018,
        latency: 150
      }
    }])
  }

  const updateSessionTitle = (sessionId: string, title: string) => {
    setSessions(prev => prev.map(session =>
      session.id === sessionId ? { ...session, title } : session
    ))
  }

  const deleteSession = (sessionId: string) => {
    setSessions(prev => prev.filter(session => session.id !== sessionId))
    if (currentSessionId === sessionId) {
      setCurrentSessionId(null)
      setMessages([])
    }
  }

  const filteredSessions = sessions.filter(session =>
    session.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    session.messages.some(msg =>
      msg.content.toLowerCase().includes(searchQuery.toLowerCase())
    )
  )

  if (!mounted) {
    return null
  }

  return (
    <div className="h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-purple-50 dark:from-gray-900 dark:via-blue-900/20 dark:to-purple-900/20 flex">
      {/* Enhanced Sidebar */}
      <div className="w-80 bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-r border-gray-200/50 dark:border-gray-700/50 flex flex-col">
        {/* Header */}
        <div className="p-4 border-b border-gray-200/50 dark:border-gray-700/50">
          <motion.button
            onClick={startNewSession}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white px-4 py-3 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 flex items-center justify-center space-x-2 shadow-lg"
          >
            <PlusIcon className="w-5 h-5" />
            <span className="font-medium">New Epic Chat</span>
          </motion.button>
        </div>

        {/* Enhanced Model Selection */}
        <div className="p-4 border-b border-gray-200/50 dark:border-gray-700/50">
          <div className="flex items-center justify-between mb-3">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              AI Model
            </label>
            <motion.button
              onClick={() => setShowSettings(!showSettings)}
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              className="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            >
              <AdjustmentsHorizontalIcon className="w-4 h-4" />
            </motion.button>
          </div>

          <div className="relative">
            <motion.button
              onClick={() => setShowModelSelector(!showModelSelector)}
              className="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl px-4 py-3 text-left focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
              whileHover={{ scale: 1.01 }}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span className="text-2xl">{getCurrentModel().icon}</span>
                  <div>
                    <div className="font-medium text-gray-900 dark:text-white">
                      {getCurrentModel().name}
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {getCurrentModel().provider} • ${getCurrentModel().costPer1kTokens}/1k tokens
                    </div>
                  </div>
                </div>
                <ArrowDownIcon className="w-4 h-4 text-gray-400" />
              </div>
            </motion.button>

            <AnimatePresence>
              {showModelSelector && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-xl z-50 max-h-80 overflow-y-auto"
                >
                  {models.map((model) => (
                    <motion.button
                      key={model.id}
                      onClick={() => {
                        setSelectedModel(model.id)
                        setShowModelSelector(false)
                      }}
                      whileHover={{ backgroundColor: 'rgba(59, 130, 246, 0.1)' }}
                      className={`w-full p-4 text-left border-b border-gray-100 dark:border-gray-700 last:border-b-0 transition-colors ${
                        selectedModel === model.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                      }`}
                    >
                      <div className="flex items-center space-x-3">
                        <span className="text-2xl">{model.icon}</span>
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <div className="font-medium text-gray-900 dark:text-white">
                              {model.name}
                            </div>
                            <div className={`w-2 h-2 rounded-full ${
                              model.isAvailable ? 'bg-green-500' : 'bg-red-500'
                            }`} />
                          </div>
                          <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {model.description}
                          </div>
                          <div className="flex items-center space-x-4 mt-2">
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              {model.provider}
                            </span>
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              ${model.costPer1kTokens}/1k tokens
                            </span>
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              {model.maxTokens.toLocaleString()} max tokens
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-1 mt-2">
                            {model.capabilities.map((cap) => (
                              <span
                                key={cap}
                                className="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full"
                              >
                                {cap}
                              </span>
                            ))}
                          </div>
                        </div>
                      </div>
                    </motion.button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-gray-200/50 dark:border-gray-700/50">
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search conversations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
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
        </div>

        {/* Enhanced Chat Sessions */}
        <div className="flex-1 overflow-y-auto">
          <div className="p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Chat History ({filteredSessions.length})
              </h3>
              {sessions.length > 0 && (
                <button
                  onClick={() => setSessions([])}
                  className="text-xs text-gray-400 hover:text-red-500 transition-colors"
                >
                  Clear All
                </button>
              )}
            </div>

            <div className="space-y-2">
              <AnimatePresence>
                {filteredSessions.length === 0 ? (
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="text-center py-8"
                  >
                    <ChatBubbleLeftRightIcon className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-3" />
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {searchQuery ? 'No conversations found' : 'No conversations yet'}
                    </p>
                    <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                      {searchQuery ? 'Try a different search term' : 'Start a new chat to begin'}
                    </p>
                  </motion.div>
                ) : (
                  filteredSessions.map((session, index) => (
                    <motion.div
                      key={session.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: 20 }}
                      transition={{ delay: index * 0.05 }}
                      className="group relative"
                    >
                      <button
                        onClick={() => setCurrentSessionId(session.id)}
                        className={`w-full text-left p-3 rounded-xl transition-all duration-200 ${
                          currentSessionId === session.id
                            ? 'bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/30 dark:to-purple-900/30 border border-blue-200 dark:border-blue-700 shadow-md'
                            : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
                        }`}
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-sm text-gray-900 dark:text-white truncate">
                              {session.title}
                            </div>
                            <div className="flex items-center space-x-2 mt-1">
                              <span className="text-xs text-gray-500 dark:text-gray-400">
                                {session.updatedAt.toLocaleDateString()}
                              </span>
                              <span className="text-xs text-gray-400 dark:text-gray-500">•</span>
                              <span className="text-xs text-gray-500 dark:text-gray-400">
                                {session.messageCount} messages
                              </span>
                              {session.totalCost > 0 && (
                                <>
                                  <span className="text-xs text-gray-400 dark:text-gray-500">•</span>
                                  <span className="text-xs text-gray-500 dark:text-gray-400">
                                    ${session.totalCost.toFixed(4)}
                                  </span>
                                </>
                              )}
                            </div>
                            <div className="flex items-center space-x-1 mt-2">
                              <span className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full">
                                {models.find(m => m.id === session.model)?.name || session.model}
                              </span>
                            </div>
                          </div>

                          <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                // Edit session title
                              }}
                              className="p-1 text-gray-400 hover:text-blue-500 transition-colors"
                            >
                              <PencilIcon className="w-3 h-3" />
                            </button>
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                deleteSession(session.id)
                              }}
                              className="p-1 text-gray-400 hover:text-red-500 transition-colors"
                            >
                              <TrashIcon className="w-3 h-3" />
                            </button>
                          </div>
                        </div>
                      </button>
                    </motion.div>
                  ))
                )}
              </AnimatePresence>
            </div>
          </div>
        </div>
      </div>

      {/* Enhanced Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {/* Enhanced Header */}
        <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/50 dark:border-gray-700/50 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-3">
                <motion.div
                  className={`w-10 h-10 rounded-xl bg-gradient-to-r ${
                    getCurrentModel().color === 'blue' ? 'from-blue-500 to-blue-600' :
                    getCurrentModel().color === 'purple' ? 'from-purple-500 to-purple-600' :
                    getCurrentModel().color === 'green' ? 'from-green-500 to-green-600' :
                    getCurrentModel().color === 'orange' ? 'from-orange-500 to-orange-600' :
                    getCurrentModel().color === 'teal' ? 'from-teal-500 to-teal-600' :
                    'from-indigo-500 to-indigo-600'
                  } text-white flex items-center justify-center shadow-lg`}
                  animate={{ rotate: isStreaming ? 360 : 0 }}
                  transition={{ duration: 2, repeat: isStreaming ? Infinity : 0, ease: "linear" }}
                >
                  <span className="text-lg">{getCurrentModel().icon}</span>
                </motion.div>
                <div>
                  <h1 className="text-xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                    Epic AI Chat
                  </h1>
                  <div className="flex items-center space-x-2">
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {getCurrentModel().name} • {getCurrentModel().provider}
                    </p>
                    <div className={`w-2 h-2 rounded-full ${
                      isStreaming ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
                    }`} />
                  </div>
                </div>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              <div className="hidden md:flex items-center space-x-4 text-sm text-gray-500 dark:text-gray-400">
                <div className="flex items-center space-x-1">
                  <ChatBubbleLeftRightIcon className="w-4 h-4" />
                  <span>{messages.length} messages</span>
                </div>
                <div className="flex items-center space-x-1">
                  <CurrencyDollarIcon className="w-4 h-4" />
                  <span>
                    ${messages.reduce((total, msg) => total + (msg.metadata?.cost || 0), 0).toFixed(4)}
                  </span>
                </div>
                <div className="flex items-center space-x-1">
                  <ClockIcon className="w-4 h-4" />
                  <span>
                    {messages.length > 0 ?
                      `${Math.round(messages.reduce((total, msg) => total + (msg.metadata?.latency || 0), 0) / messages.filter(m => m.metadata?.latency).length)}ms avg`
                      : '0ms'
                    }
                  </span>
                </div>
              </div>

              <div className="flex items-center space-x-2">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => setShowSettings(!showSettings)}
                  className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                >
                  <AdjustmentsHorizontalIcon className="w-5 h-5" />
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                >
                  <ShareIcon className="w-5 h-5" />
                </motion.button>
              </div>
            </div>
          </div>
        </div>

        {/* Enhanced Messages */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 bg-gradient-to-b from-transparent to-gray-50/30 dark:to-gray-900/30">
          <AnimatePresence>
            {messages.map((message, index) => (
              <motion.div
                key={message.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                transition={{ delay: index * 0.05 }}
                className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'} group`}
              >
                <div className={`max-w-4xl w-full ${message.role === 'user' ? 'flex justify-end' : 'flex justify-start'}`}>
                  <div className={`flex items-start space-x-3 ${message.role === 'user' ? 'flex-row-reverse space-x-reverse' : ''}`}>
                    {/* Enhanced Avatar */}
                    <motion.div
                      className={`w-10 h-10 rounded-xl flex items-center justify-center shadow-lg ${
                        message.role === 'user'
                          ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white'
                          : `bg-gradient-to-r ${
                              getCurrentModel().color === 'blue' ? 'from-blue-500 to-blue-600' :
                              getCurrentModel().color === 'purple' ? 'from-purple-500 to-purple-600' :
                              getCurrentModel().color === 'green' ? 'from-green-500 to-green-600' :
                              getCurrentModel().color === 'orange' ? 'from-orange-500 to-orange-600' :
                              getCurrentModel().color === 'teal' ? 'from-teal-500 to-teal-600' :
                              'from-indigo-500 to-indigo-600'
                            } text-white`
                      }`}
                      whileHover={{ scale: 1.05 }}
                      animate={message.isStreaming ? { rotate: [0, 360] } : {}}
                      transition={message.isStreaming ? { duration: 2, repeat: Infinity, ease: "linear" } : {}}
                    >
                      {message.role === 'user' ? (
                        <UserIcon className="w-5 h-5" />
                      ) : (
                        <span className="text-lg">{getCurrentModel().icon}</span>
                      )}
                    </motion.div>

                    <div className={`flex-1 ${message.role === 'user' ? 'flex justify-end' : ''}`}>
                      {/* Message Content */}
                      <motion.div
                        className={`relative max-w-3xl ${
                          message.role === 'user'
                            ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg'
                            : message.type === 'error'
                            ? 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200'
                            : 'bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 shadow-lg'
                        } rounded-2xl px-6 py-4 transition-all duration-200`}
                        whileHover={{ y: -2 }}
                      >
                        {/* Message Actions */}
                        <div className={`absolute top-2 right-2 flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity ${
                          message.role === 'user' ? 'text-white/70' : 'text-gray-400'
                        }`}>
                          <motion.button
                            whileHover={{ scale: 1.1 }}
                            whileTap={{ scale: 0.9 }}
                            onClick={() => copyToClipboard(message.content, message.id)}
                            className="p-1 hover:bg-black/10 dark:hover:bg-white/10 rounded transition-colors"
                          >
                            {copiedMessageId === message.id ? (
                              <CheckIcon className="w-3 h-3" />
                            ) : (
                              <ClipboardDocumentIcon className="w-3 h-3" />
                            )}
                          </motion.button>

                          <motion.button
                            whileHover={{ scale: 1.1 }}
                            whileTap={{ scale: 0.9 }}
                            onClick={() => toggleBookmark(message.id)}
                            className={`p-1 hover:bg-black/10 dark:hover:bg-white/10 rounded transition-colors ${
                              message.isBookmarked ? 'text-yellow-500' : ''
                            }`}
                          >
                            <BookmarkIcon className={`w-3 h-3 ${message.isBookmarked ? 'fill-current' : ''}`} />
                          </motion.button>

                          {message.role === 'assistant' && (
                            <motion.button
                              whileHover={{ scale: 1.1 }}
                              whileTap={{ scale: 0.9 }}
                              onClick={() => regenerateResponse(message.id)}
                              className="p-1 hover:bg-black/10 dark:hover:bg-white/10 rounded transition-colors"
                            >
                              <ArrowDownIcon className="w-3 h-3 transform rotate-45" />
                            </motion.button>
                          )}
                        </div>

                        {/* Content */}
                        <div className={`${message.role === 'user' ? 'text-white' : 'text-gray-900 dark:text-gray-100'}`}>
                          <div className="prose prose-sm max-w-none">
                            <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed">
                              {message.content}
                              {message.isStreaming && (
                                <motion.span
                                  animate={{ opacity: [1, 0] }}
                                  transition={{ duration: 0.8, repeat: Infinity }}
                                  className="inline-block w-2 h-4 bg-current ml-1"
                                />
                              )}
                            </pre>
                          </div>
                        </div>
                      </motion.div>

                      {/* Enhanced Message Metadata */}
                      <div className={`flex items-center justify-between mt-3 text-xs ${
                        message.role === 'user' ? 'text-right' : 'text-left'
                      }`}>
                        <div className="flex items-center space-x-3 text-gray-500 dark:text-gray-400">
                          <span>{message.timestamp.toLocaleTimeString()}</span>
                          {message.metadata && (
                            <>
                              {message.metadata.model && (
                                <div className="flex items-center space-x-1">
                                  <CpuChipIcon className="w-3 h-3" />
                                  <span>{message.metadata.model}</span>
                                </div>
                              )}
                              {message.metadata.tokens && (
                                <div className="flex items-center space-x-1">
                                  <span>{message.metadata.tokens} tokens</span>
                                </div>
                              )}
                              {message.metadata.latency && (
                                <div className="flex items-center space-x-1">
                                  <ClockIcon className="w-3 h-3" />
                                  <span>{message.metadata.latency}ms</span>
                                </div>
                              )}
                              {message.metadata.cost && (
                                <div className="flex items-center space-x-1">
                                  <CurrencyDollarIcon className="w-3 h-3" />
                                  <span>${message.metadata.cost.toFixed(4)}</span>
                                </div>
                              )}
                              {message.metadata.temperature && (
                                <div className="flex items-center space-x-1">
                                  <span>temp: {message.metadata.temperature}</span>
                                </div>
                              )}
                            </>
                          )}
                        </div>

                        {message.isBookmarked && (
                          <BookmarkIcon className="w-3 h-3 text-yellow-500 fill-current" />
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>

          {/* Enhanced Loading Indicator */}
          {isLoading && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex justify-start"
            >
              <div className="flex items-start space-x-3">
                <motion.div
                  className={`w-10 h-10 rounded-xl bg-gradient-to-r ${
                    getCurrentModel().color === 'blue' ? 'from-blue-500 to-blue-600' :
                    getCurrentModel().color === 'purple' ? 'from-purple-500 to-purple-600' :
                    getCurrentModel().color === 'green' ? 'from-green-500 to-green-600' :
                    getCurrentModel().color === 'orange' ? 'from-orange-500 to-orange-600' :
                    getCurrentModel().color === 'teal' ? 'from-teal-500 to-teal-600' :
                    'from-indigo-500 to-indigo-600'
                  } text-white flex items-center justify-center shadow-lg`}
                  animate={{ rotate: 360 }}
                  transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
                >
                  <span className="text-lg">{getCurrentModel().icon}</span>
                </motion.div>
                <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 rounded-2xl px-6 py-4 shadow-lg">
                  <div className="flex items-center space-x-3">
                    <div className="flex space-x-1">
                      <motion.div
                        className="w-2 h-2 bg-blue-500 rounded-full"
                        animate={{ scale: [1, 1.2, 1], opacity: [1, 0.5, 1] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0 }}
                      />
                      <motion.div
                        className="w-2 h-2 bg-purple-500 rounded-full"
                        animate={{ scale: [1, 1.2, 1], opacity: [1, 0.5, 1] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0.2 }}
                      />
                      <motion.div
                        className="w-2 h-2 bg-green-500 rounded-full"
                        animate={{ scale: [1, 1.2, 1], opacity: [1, 0.5, 1] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0.4 }}
                      />
                    </div>
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {getCurrentModel().name} is thinking...
                    </span>
                  </div>
                </div>
              </div>
            </motion.div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* Enhanced Input Area */}
        <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-t border-gray-200/50 dark:border-gray-700/50 p-6">
          <div className="max-w-4xl mx-auto">
            {/* Quick Actions */}
            <div className="flex items-center space-x-2 mb-4">
              <div className="flex items-center space-x-2">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="px-3 py-1 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full hover:bg-blue-200 dark:hover:bg-blue-900/50 transition-colors"
                  onClick={() => setInputValue('Explain this concept in simple terms: ')}
                >
                  <LightBulbIcon className="w-3 h-3 mr-1" />
                  Explain
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="px-3 py-1 text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-full hover:bg-green-200 dark:hover:bg-green-900/50 transition-colors"
                  onClick={() => setInputValue('Write code for: ')}
                >
                  <CodeBracketIcon className="w-3 h-3 mr-1" />
                  Code
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="px-3 py-1 text-xs bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-full hover:bg-purple-200 dark:hover:bg-purple-900/50 transition-colors"
                  onClick={() => setInputValue('Translate this to: ')}
                >
                  <LanguageIcon className="w-3 h-3 mr-1" />
                  Translate
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="px-3 py-1 text-xs bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300 rounded-full hover:bg-orange-200 dark:hover:bg-orange-900/50 transition-colors"
                  onClick={() => setInputValue('Analyze this: ')}
                >
                  <BeakerIcon className="w-3 h-3 mr-1" />
                  Analyze
                </motion.button>
              </div>
            </div>

            <div className="flex items-end space-x-4">
              <div className="flex-1">
                <div className="relative">
                  <textarea
                    ref={inputRef}
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    onKeyPress={handleKeyPress}
                    placeholder={`Message ${getCurrentModel().name}... (Press Enter to send, Shift+Enter for new line)`}
                    className="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-2xl px-6 py-4 pr-12 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none transition-all duration-200 shadow-lg"
                    rows={Math.min(Math.max(inputValue.split('\n').length, 1), 6)}
                    disabled={isLoading}
                  />

                  {/* Character count */}
                  <div className="absolute bottom-2 right-12 text-xs text-gray-400 dark:text-gray-500">
                    {inputValue.length}/4000
                  </div>
                </div>
              </div>

              {/* Enhanced Action Buttons */}
              <div className="flex items-center space-x-2">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,.doc,.docx,.txt,.md"
                  className="hidden"
                  onChange={(e) => {
                    // Handle file upload
                    console.log('File selected:', e.target.files?.[0])
                  }}
                />

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="p-3 text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 transition-colors rounded-xl hover:bg-blue-50 dark:hover:bg-blue-900/20"
                  title="Upload document"
                >
                  <DocumentTextIcon className="w-5 h-5" />
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  type="button"
                  className="p-3 text-gray-400 hover:text-green-500 dark:hover:text-green-400 transition-colors rounded-xl hover:bg-green-50 dark:hover:bg-green-900/20"
                  title="Upload image"
                >
                  <PhotoIcon className="w-5 h-5" />
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  type="button"
                  className="p-3 text-gray-400 hover:text-purple-500 dark:hover:text-purple-400 transition-colors rounded-xl hover:bg-purple-50 dark:hover:bg-purple-900/20"
                  title="Voice input"
                >
                  <MicrophoneIcon className="w-5 h-5" />
                </motion.button>

                {isLoading ? (
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => {
                      setIsLoading(false)
                      setIsStreaming(false)
                    }}
                    className="bg-red-500 hover:bg-red-600 text-white p-3 rounded-xl transition-all duration-200 shadow-lg"
                    title="Stop generation"
                  >
                    <StopIcon className="w-5 h-5" />
                  </motion.button>
                ) : (
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={handleSendMessage}
                    disabled={!inputValue.trim()}
                    className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed text-white p-3 rounded-xl transition-all duration-200 shadow-lg"
                    title="Send message"
                  >
                    <PaperAirplaneIcon className="w-5 h-5" />
                  </motion.button>
                )}
              </div>
            </div>

            {/* Model Info */}
            <div className="flex items-center justify-between mt-4 text-xs text-gray-500 dark:text-gray-400">
              <div className="flex items-center space-x-4">
                <span>Model: {getCurrentModel().name}</span>
                <span>Temperature: {chatSettings.temperature}</span>
                <span>Max tokens: {chatSettings.maxTokens}</span>
              </div>
              <div className="flex items-center space-x-2">
                <span>Cost per 1k tokens: ${getCurrentModel().costPer1kTokens}</span>
                <div className={`w-2 h-2 rounded-full ${
                  getCurrentModel().isAvailable ? 'bg-green-500' : 'bg-red-500'
                }`} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
