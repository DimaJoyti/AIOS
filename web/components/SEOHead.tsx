import Head from 'next/head'

interface SEOProps {
  title?: string
  description?: string
  keywords?: string[]
  image?: string
  url?: string
  type?: 'website' | 'article' | 'profile'
  siteName?: string
  locale?: string
  author?: string
  publishedTime?: string
  modifiedTime?: string
  section?: string
  tags?: string[]
  noIndex?: boolean
  noFollow?: boolean
  canonical?: string
}

export function SEOHead({
  title = 'AIOS - Advanced AI Operating System',
  description = 'Experience the future of AI with AIOS - a comprehensive AI operating system featuring advanced chat interfaces, document management, project tracking, and real-time analytics.',
  keywords = [
    'AI', 'artificial intelligence', 'operating system', 'chat', 'GPT', 'Claude', 
    'document management', 'project management', 'analytics', 'dashboard', 'automation'
  ],
  image = '/og-image.png',
  url = 'https://aios.dev',
  type = 'website',
  siteName = 'AIOS',
  locale = 'en_US',
  author = 'AIOS Team',
  publishedTime,
  modifiedTime,
  section,
  tags,
  noIndex = false,
  noFollow = false,
  canonical
}: SEOProps) {
  const fullTitle = title.includes('AIOS') ? title : `${title} | AIOS`
  const fullUrl = url.startsWith('http') ? url : `https://aios.dev${url}`
  const imageUrl = image.startsWith('http') ? image : `https://aios.dev${image}`

  return (
    <Head>
      {/* Basic Meta Tags */}
      <title>{fullTitle}</title>
      <meta name="description" content={description} />
      <meta name="keywords" content={keywords.join(', ')} />
      <meta name="author" content={author} />
      
      {/* Robots */}
      <meta 
        name="robots" 
        content={`${noIndex ? 'noindex' : 'index'}, ${noFollow ? 'nofollow' : 'follow'}`} 
      />
      
      {/* Canonical URL */}
      {canonical && <link rel="canonical" href={canonical} />}
      
      {/* Viewport */}
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      
      {/* Language */}
      <meta httpEquiv="content-language" content={locale.replace('_', '-')} />
      
      {/* Open Graph */}
      <meta property="og:type" content={type} />
      <meta property="og:title" content={fullTitle} />
      <meta property="og:description" content={description} />
      <meta property="og:image" content={imageUrl} />
      <meta property="og:url" content={fullUrl} />
      <meta property="og:site_name" content={siteName} />
      <meta property="og:locale" content={locale} />
      
      {/* Article specific */}
      {type === 'article' && (
        <>
          {publishedTime && <meta property="article:published_time" content={publishedTime} />}
          {modifiedTime && <meta property="article:modified_time" content={modifiedTime} />}
          {author && <meta property="article:author" content={author} />}
          {section && <meta property="article:section" content={section} />}
          {tags && tags.map(tag => (
            <meta key={tag} property="article:tag" content={tag} />
          ))}
        </>
      )}
      
      {/* Twitter Card */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:title" content={fullTitle} />
      <meta name="twitter:description" content={description} />
      <meta name="twitter:image" content={imageUrl} />
      <meta name="twitter:site" content="@aios_dev" />
      <meta name="twitter:creator" content="@aios_dev" />
      
      {/* Additional Meta Tags */}
      <meta name="theme-color" content="#3b82f6" />
      <meta name="msapplication-TileColor" content="#3b82f6" />
      <meta name="application-name" content={siteName} />
      
      {/* Apple Meta Tags */}
      <meta name="apple-mobile-web-app-capable" content="yes" />
      <meta name="apple-mobile-web-app-status-bar-style" content="default" />
      <meta name="apple-mobile-web-app-title" content={siteName} />
      
      {/* Favicons */}
      <link rel="icon" type="image/x-icon" href="/favicon.ico" />
      <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
      <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
      <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
      <link rel="manifest" href="/site.webmanifest" />
      
      {/* Preconnect to external domains */}
      <link rel="preconnect" href="https://fonts.googleapis.com" />
      <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
      
      {/* DNS Prefetch */}
      <link rel="dns-prefetch" href="//api.openai.com" />
      <link rel="dns-prefetch" href="//api.anthropic.com" />
      
      {/* Structured Data */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "SoftwareApplication",
            "name": siteName,
            "description": description,
            "url": fullUrl,
            "image": imageUrl,
            "author": {
              "@type": "Organization",
              "name": author
            },
            "applicationCategory": "BusinessApplication",
            "operatingSystem": "Web",
            "offers": {
              "@type": "Offer",
              "price": "0",
              "priceCurrency": "USD"
            },
            "aggregateRating": {
              "@type": "AggregateRating",
              "ratingValue": "4.8",
              "ratingCount": "1250"
            }
          })
        }}
      />
    </Head>
  )
}

// Page-specific SEO components
export function DashboardSEO() {
  return (
    <SEOHead
      title="Dashboard - AIOS"
      description="Monitor your AI operations with real-time metrics, analytics, and system health indicators in the AIOS dashboard."
      keywords={['dashboard', 'metrics', 'analytics', 'monitoring', 'AI operations']}
      url="/dashboard"
    />
  )
}

export function ChatSEO() {
  return (
    <SEOHead
      title="AI Chat - AIOS"
      description="Engage with multiple AI models including GPT-4, Claude, and Gemini in an advanced chat interface with conversation management."
      keywords={['AI chat', 'GPT-4', 'Claude', 'conversation', 'artificial intelligence']}
      url="/dashboard/chat"
    />
  )
}

export function DocumentsSEO() {
  return (
    <SEOHead
      title="Document Management - AIOS"
      description="Upload, process, and manage documents with AI-powered analysis, OCR, and intelligent search capabilities."
      keywords={['document management', 'OCR', 'AI analysis', 'file upload', 'search']}
      url="/dashboard/documents"
    />
  )
}

export function ProjectsSEO() {
  return (
    <SEOHead
      title="Project Management - AIOS"
      description="Manage projects with Kanban boards, team collaboration, budget tracking, and progress analytics."
      keywords={['project management', 'kanban', 'team collaboration', 'budget tracking']}
      url="/dashboard/projects"
    />
  )
}

export function MonitoringSEO() {
  return (
    <SEOHead
      title="System Monitoring - AIOS"
      description="Real-time system monitoring with performance metrics, service health checks, and live log streaming."
      keywords={['system monitoring', 'performance metrics', 'health checks', 'logs']}
      url="/dashboard/monitoring"
    />
  )
}

export function AnalyticsSEO() {
  return (
    <SEOHead
      title="Analytics - AIOS"
      description="Comprehensive analytics dashboard with usage insights, cost tracking, and performance analysis."
      keywords={['analytics', 'usage insights', 'cost tracking', 'performance analysis']}
      url="/dashboard/analytics"
    />
  )
}

export function SettingsSEO() {
  return (
    <SEOHead
      title="Settings - AIOS"
      description="Customize your AIOS experience with user preferences, AI model settings, and integration management."
      keywords={['settings', 'preferences', 'AI models', 'integrations', 'customization']}
      url="/dashboard/settings"
    />
  )
}

// Blog/Article SEO
export function ArticleSEO({
  title,
  description,
  author,
  publishedTime,
  modifiedTime,
  tags,
  image,
  slug
}: {
  title: string
  description: string
  author: string
  publishedTime: string
  modifiedTime?: string
  tags: string[]
  image?: string
  slug: string
}) {
  return (
    <SEOHead
      title={title}
      description={description}
      type="article"
      author={author}
      publishedTime={publishedTime}
      modifiedTime={modifiedTime}
      tags={tags}
      image={image}
      url={`/blog/${slug}`}
      section="Technology"
    />
  )
}

export default SEOHead
