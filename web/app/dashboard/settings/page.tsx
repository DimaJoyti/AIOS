'use client'

import { useState } from 'react'
import { 
  Cog6ToothIcon,
  UserIcon,
  BellIcon,
  ShieldCheckIcon,
  CpuChipIcon,
  CloudIcon,
  KeyIcon,
  PaintBrushIcon,
  GlobeAltIcon,
  MoonIcon,
  SunIcon,
  ComputerDesktopIcon,
  CheckIcon,
  XMarkIcon,
  PlusIcon,
  TrashIcon,
  EyeIcon,
  EyeSlashIcon
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'

interface UserSettings {
  profile: {
    name: string
    email: string
    avatar: string
    timezone: string
    language: string
  }
  preferences: {
    theme: 'light' | 'dark' | 'system'
    notifications: {
      email: boolean
      push: boolean
      desktop: boolean
      marketing: boolean
    }
    privacy: {
      analytics: boolean
      crashReports: boolean
      usageData: boolean
    }
  }
  aiSettings: {
    defaultModel: string
    temperature: number
    maxTokens: number
    systemPrompt: string
    autoSave: boolean
    streamResponses: boolean
  }
  integrations: {
    openai: { enabled: boolean; apiKey: string }
    anthropic: { enabled: boolean; apiKey: string }
    google: { enabled: boolean; apiKey: string }
    github: { enabled: boolean; token: string }
    slack: { enabled: boolean; webhook: string }
  }
}

export default function SettingsPage() {
  const [mounted, setMounted] = useState(false)
  const [activeTab, setActiveTab] = useState<'profile' | 'preferences' | 'ai' | 'integrations' | 'security'>('profile')
  const [settings, setSettings] = useState<UserSettings>({
    profile: {
      name: 'Alex Chen',
      email: 'alex.chen@example.com',
      avatar: '/avatars/alex.jpg',
      timezone: 'America/New_York',
      language: 'en'
    },
    preferences: {
      theme: 'system',
      notifications: {
        email: true,
        push: true,
        desktop: false,
        marketing: false
      },
      privacy: {
        analytics: true,
        crashReports: true,
        usageData: false
      }
    },
    aiSettings: {
      defaultModel: 'gpt-4',
      temperature: 0.7,
      maxTokens: 2000,
      systemPrompt: 'You are a helpful AI assistant.',
      autoSave: true,
      streamResponses: true
    },
    integrations: {
      openai: { enabled: true, apiKey: 'sk-...' },
      anthropic: { enabled: false, apiKey: '' },
      google: { enabled: false, apiKey: '' },
      github: { enabled: true, token: 'ghp_...' },
      slack: { enabled: false, webhook: '' }
    }
  })

  const [showApiKeys, setShowApiKeys] = useState<Record<string, boolean>>({})
  const [unsavedChanges, setUnsavedChanges] = useState(false)

  const tabs = [
    { id: 'profile', label: 'Profile', icon: <UserIcon className="w-5 h-5" /> },
    { id: 'preferences', label: 'Preferences', icon: <PaintBrushIcon className="w-5 h-5" /> },
    { id: 'ai', label: 'AI Settings', icon: <CpuChipIcon className="w-5 h-5" /> },
    { id: 'integrations', label: 'Integrations', icon: <CloudIcon className="w-5 h-5" /> },
    { id: 'security', label: 'Security', icon: <ShieldCheckIcon className="w-5 h-5" /> }
  ]

  const updateSettings = (path: string[], value: any) => {
    setSettings(prev => {
      const newSettings = { ...prev }
      let current: any = newSettings
      
      for (let i = 0; i < path.length - 1; i++) {
        current = current[path[i]]
      }
      
      current[path[path.length - 1]] = value
      return newSettings
    })
    setUnsavedChanges(true)
  }

  const saveSettings = async () => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000))
    setUnsavedChanges(false)
  }

  const toggleApiKeyVisibility = (key: string) => {
    setShowApiKeys(prev => ({ ...prev, [key]: !prev[key] }))
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
              className="w-12 h-12 bg-gradient-to-r from-gray-500 to-gray-600 rounded-xl flex items-center justify-center shadow-lg"
              whileHover={{ scale: 1.05, rotate: 5 }}
              transition={{ type: "spring", stiffness: 400, damping: 10 }}
            >
              <Cog6ToothIcon className="w-7 h-7 text-white" />
            </motion.div>
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-300 bg-clip-text text-transparent">
                Settings
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Customize your AIOS experience and preferences
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            {unsavedChanges && (
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="flex items-center space-x-2 px-3 py-2 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400 rounded-lg text-sm"
              >
                <span>Unsaved changes</span>
              </motion.div>
            )}
            
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={saveSettings}
              disabled={!unsavedChanges}
              className={`px-6 py-2 rounded-xl transition-all duration-200 shadow-lg flex items-center space-x-2 ${
                unsavedChanges
                  ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white hover:from-blue-700 hover:to-purple-700'
                  : 'bg-gray-300 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed'
              }`}
            >
              <CheckIcon className="w-4 h-4" />
              <span>Save Changes</span>
            </motion.button>
          </div>
        </div>
      </div>

      <div className="flex-1 flex">
        {/* Sidebar Navigation */}
        <div className="w-64 bg-white/60 dark:bg-gray-900/60 backdrop-blur-sm border-r border-gray-200/50 dark:border-gray-700/50 p-6">
          <nav className="space-y-2">
            {tabs.map((tab) => (
              <motion.button
                key={tab.id}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => setActiveTab(tab.id as any)}
                className={`w-full flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 ${
                  activeTab === tab.id
                    ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
              >
                {tab.icon}
                <span className="font-medium">{tab.label}</span>
              </motion.button>
            ))}
          </nav>
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <AnimatePresence mode="wait">
            {activeTab === 'profile' && (
              <ProfileSettings 
                settings={settings.profile} 
                updateSettings={updateSettings}
              />
            )}
            {activeTab === 'preferences' && (
              <PreferencesSettings 
                settings={settings.preferences} 
                updateSettings={updateSettings}
              />
            )}
            {activeTab === 'ai' && (
              <AISettings 
                settings={settings.aiSettings} 
                updateSettings={updateSettings}
              />
            )}
            {activeTab === 'integrations' && (
              <IntegrationsSettings 
                settings={settings.integrations} 
                updateSettings={updateSettings}
                showApiKeys={showApiKeys}
                toggleApiKeyVisibility={toggleApiKeyVisibility}
              />
            )}
            {activeTab === 'security' && (
              <SecuritySettings />
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  )
}

// Profile Settings Component
function ProfileSettings({ settings, updateSettings }: {
  settings: UserSettings['profile']
  updateSettings: (path: string[], value: any) => void
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="space-y-6"
    >
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Profile Information</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Full Name
            </label>
            <input
              type="text"
              value={settings.name}
              onChange={(e) => updateSettings(['profile', 'name'], e.target.value)}
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Email Address
            </label>
            <input
              type="email"
              value={settings.email}
              onChange={(e) => updateSettings(['profile', 'email'], e.target.value)}
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Timezone
            </label>
            <select
              value={settings.timezone}
              onChange={(e) => updateSettings(['profile', 'timezone'], e.target.value)}
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            >
              <option value="America/New_York">Eastern Time</option>
              <option value="America/Chicago">Central Time</option>
              <option value="America/Denver">Mountain Time</option>
              <option value="America/Los_Angeles">Pacific Time</option>
              <option value="Europe/London">London</option>
              <option value="Europe/Paris">Paris</option>
              <option value="Asia/Tokyo">Tokyo</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Language
            </label>
            <select
              value={settings.language}
              onChange={(e) => updateSettings(['profile', 'language'], e.target.value)}
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            >
              <option value="en">English</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
              <option value="de">German</option>
              <option value="ja">Japanese</option>
              <option value="zh">Chinese</option>
            </select>
          </div>
        </div>
      </div>
    </motion.div>
  )
}

// Preferences Settings Component
function PreferencesSettings({ settings, updateSettings }: {
  settings: UserSettings['preferences']
  updateSettings: (path: string[], value: any) => void
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="space-y-6"
    >
      {/* Theme Settings */}
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Appearance</h2>

        <div className="space-y-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
            Theme
          </label>
          <div className="grid grid-cols-3 gap-4">
            {[
              { value: 'light', label: 'Light', icon: <SunIcon className="w-5 h-5" /> },
              { value: 'dark', label: 'Dark', icon: <MoonIcon className="w-5 h-5" /> },
              { value: 'system', label: 'System', icon: <ComputerDesktopIcon className="w-5 h-5" /> }
            ].map((theme) => (
              <motion.button
                key={theme.value}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => updateSettings(['preferences', 'theme'], theme.value)}
                className={`flex flex-col items-center space-y-2 p-4 rounded-xl border-2 transition-all duration-200 ${
                  settings.theme === theme.value
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                {theme.icon}
                <span className="text-sm font-medium">{theme.label}</span>
              </motion.button>
            ))}
          </div>
        </div>
      </div>

      {/* Notification Settings */}
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Notifications</h2>

        <div className="space-y-4">
          {Object.entries(settings.notifications).map(([key, value]) => (
            <div key={key} className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                  {key.charAt(0).toUpperCase() + key.slice(1)} Notifications
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Receive {key} notifications for important updates
                </p>
              </div>
              <ToggleSwitch
                enabled={value}
                onChange={(enabled) => updateSettings(['preferences', 'notifications', key], enabled)}
              />
            </div>
          ))}
        </div>
      </div>

      {/* Privacy Settings */}
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Privacy</h2>

        <div className="space-y-4">
          {Object.entries(settings.privacy).map(([key, value]) => (
            <div key={key} className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                  {key.charAt(0).toUpperCase() + key.slice(1).replace(/([A-Z])/g, ' $1')}
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Allow collection of {key.replace(/([A-Z])/g, ' $1').toLowerCase()} to improve the service
                </p>
              </div>
              <ToggleSwitch
                enabled={value}
                onChange={(enabled) => updateSettings(['preferences', 'privacy', key], enabled)}
              />
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  )
}

// AI Settings Component
function AISettings({ settings, updateSettings }: {
  settings: UserSettings['aiSettings']
  updateSettings: (path: string[], value: any) => void
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="space-y-6"
    >
      <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">AI Model Configuration</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Default Model
            </label>
            <select
              value={settings.defaultModel}
              onChange={(e) => updateSettings(['aiSettings', 'defaultModel'], e.target.value)}
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            >
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-4-turbo">GPT-4 Turbo</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
              <option value="claude-3-opus">Claude 3 Opus</option>
              <option value="claude-3-sonnet">Claude 3 Sonnet</option>
              <option value="gemini-pro">Gemini Pro</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Max Tokens
            </label>
            <input
              type="number"
              value={settings.maxTokens}
              onChange={(e) => updateSettings(['aiSettings', 'maxTokens'], parseInt(e.target.value))}
              min="100"
              max="8000"
              className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            />
          </div>
        </div>

        <div className="mt-6">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Temperature: {settings.temperature}
          </label>
          <input
            type="range"
            min="0"
            max="2"
            step="0.1"
            value={settings.temperature}
            onChange={(e) => updateSettings(['aiSettings', 'temperature'], parseFloat(e.target.value))}
            className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer dark:bg-gray-700"
          />
          <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
            <span>Conservative</span>
            <span>Balanced</span>
            <span>Creative</span>
          </div>
        </div>

        <div className="mt-6">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            System Prompt
          </label>
          <textarea
            value={settings.systemPrompt}
            onChange={(e) => updateSettings(['aiSettings', 'systemPrompt'], e.target.value)}
            rows={4}
            className="w-full px-4 py-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            placeholder="Enter a system prompt to customize AI behavior..."
          />
        </div>

        <div className="mt-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium text-gray-900 dark:text-white">Auto-save Conversations</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Automatically save chat conversations</p>
            </div>
            <ToggleSwitch
              enabled={settings.autoSave}
              onChange={(enabled) => updateSettings(['aiSettings', 'autoSave'], enabled)}
            />
          </div>

          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium text-gray-900 dark:text-white">Stream Responses</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Show AI responses as they are generated</p>
            </div>
            <ToggleSwitch
              enabled={settings.streamResponses}
              onChange={(enabled) => updateSettings(['aiSettings', 'streamResponses'], enabled)}
            />
          </div>
        </div>
      </div>
    </motion.div>
  )
}

// Toggle Switch Component
function ToggleSwitch({ enabled, onChange }: {
  enabled: boolean
  onChange: (enabled: boolean) => void
}) {
  return (
    <motion.button
      whileTap={{ scale: 0.95 }}
      onClick={() => onChange(!enabled)}
      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 ${
        enabled ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
      }`}
    >
      <motion.span
        layout
        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200 ${
          enabled ? 'translate-x-6' : 'translate-x-1'
        }`}
      />
    </motion.button>
  )
}

// Placeholder components for remaining tabs
function IntegrationsSettings({ settings, updateSettings, showApiKeys, toggleApiKeyVisibility }: any) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6"
    >
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">API Integrations</h2>
      <p className="text-gray-600 dark:text-gray-400">Configure your API keys and integrations here.</p>
    </motion.div>
  )
}

function SecuritySettings() {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 p-6"
    >
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Security Settings</h2>
      <p className="text-gray-600 dark:text-gray-400">Manage your security preferences and two-factor authentication.</p>
    </motion.div>
  )
}
