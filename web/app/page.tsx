'use client'

import { useState, useEffect } from 'react'
import { ChevronRightIcon, SparklesIcon, CpuChipIcon, CommandLineIcon, ComputerDesktopIcon } from '@heroicons/react/24/outline'

export default function Home() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return null
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      {/* Animated background */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -inset-10 opacity-50">
          <div className="absolute top-1/4 left-1/4 w-72 h-72 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse-slow"></div>
          <div className="absolute top-1/3 right-1/4 w-72 h-72 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse-slow animation-delay-2000"></div>
          <div className="absolute bottom-1/4 left-1/3 w-72 h-72 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse-slow animation-delay-4000"></div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="relative z-10 flex items-center justify-between p-6 lg:px-8">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 bg-gradient-to-r from-purple-500 to-blue-500 rounded-lg flex items-center justify-center">
            <SparklesIcon className="w-5 h-5 text-white" />
          </div>
          <span className="text-xl font-bold text-white">AIOS</span>
        </div>
        <div className="hidden md:flex items-center space-x-8">
          <a href="#" className="text-gray-300 hover:text-white transition-colors">Features</a>
          <a href="#" className="text-gray-300 hover:text-white transition-colors">Documentation</a>
          <a href="#" className="text-gray-300 hover:text-white transition-colors">Support</a>
          <button className="bg-gradient-to-r from-purple-500 to-blue-500 text-white px-4 py-2 rounded-lg hover:from-purple-600 hover:to-blue-600 transition-all">
            Get Started
          </button>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="relative z-10 flex flex-col items-center justify-center min-h-[80vh] px-6 text-center">
        <div className="max-w-4xl mx-auto">
          <div className="mb-8">
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-purple-500/10 text-purple-300 border border-purple-500/20">
              <SparklesIcon className="w-4 h-4 mr-2" />
              AI-Powered Operating System
            </span>
          </div>

          <h1 className="text-5xl md:text-7xl font-bold text-white mb-6 leading-tight">
            Welcome to{' '}
            <span className="text-gradient bg-gradient-to-r from-purple-400 via-pink-400 to-blue-400 bg-clip-text text-transparent">
              AIOS
            </span>
          </h1>

          <p className="text-xl md:text-2xl text-gray-300 mb-12 max-w-3xl mx-auto leading-relaxed">
            Experience the future of computing with our AI-integrated operating system.
            Intelligent, adaptive, and designed for the next generation of digital interaction.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
            <a href="/dashboard" className="group bg-gradient-to-r from-purple-500 to-blue-500 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:from-purple-600 hover:to-blue-600 transition-all transform hover:scale-105 shadow-lg hover:shadow-xl inline-flex items-center justify-center">
              Launch AIOS
              <ChevronRightIcon className="w-5 h-5 inline-block ml-2 group-hover:translate-x-1 transition-transform" />
            </a>
            <button className="border border-gray-600 text-gray-300 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-gray-800 hover:text-white transition-all">
              Learn More
            </button>
          </div>
        </div>
      </main>

      {/* Features Grid */}
      <section className="relative z-10 max-w-7xl mx-auto px-6 pb-20">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <FeatureCard
            icon={<SparklesIcon className="w-8 h-8" />}
            title="AI Assistant"
            description="Intelligent assistant that learns from your workflow and anticipates your needs."
            gradient="from-purple-500 to-pink-500"
          />
          <FeatureCard
            icon={<ComputerDesktopIcon className="w-8 h-8" />}
            title="Smart Desktop"
            description="Adaptive desktop environment that organizes itself based on your usage patterns."
            gradient="from-blue-500 to-cyan-500"
          />
          <FeatureCard
            icon={<CommandLineIcon className="w-8 h-8" />}
            title="AI Tools"
            description="Powerful suite of AI-enhanced productivity tools and utilities."
            gradient="from-green-500 to-emerald-500"
          />
          <FeatureCard
            icon={<CpuChipIcon className="w-8 h-8" />}
            title="System Intelligence"
            description="Real-time system optimization and predictive maintenance powered by AI."
            gradient="from-orange-500 to-red-500"
          />
        </div>
      </section>
    </div>
  )
}

function FeatureCard({ icon, title, description, gradient }: {
  icon: React.ReactNode
  title: string
  description: string
  gradient: string
}) {
  return (
    <div className="group relative">
      <div className="absolute -inset-0.5 bg-gradient-to-r opacity-75 rounded-2xl blur group-hover:opacity-100 transition duration-1000 group-hover:duration-200 animate-tilt"></div>
      <div className="relative bg-gray-900/50 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6 hover:bg-gray-800/50 transition-all duration-300">
        <div className={`inline-flex p-3 rounded-xl bg-gradient-to-r ${gradient} mb-4`}>
          <div className="text-white">
            {icon}
          </div>
        </div>
        <h3 className="text-xl font-semibold text-white mb-3">{title}</h3>
        <p className="text-gray-400 leading-relaxed">{description}</p>
      </div>
    </div>
  )
}
