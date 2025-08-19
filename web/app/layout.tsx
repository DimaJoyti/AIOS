import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { SkipToMain } from '../components/AccessibilityHelpers'

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
})

export const metadata: Metadata = {
  title: 'AIOS - AI Operating System',
  description: 'Experience the future of computing with our AI-integrated operating system. Intelligent, adaptive, and designed for the next generation of digital interaction.',
  keywords: ['AI', 'Operating System', 'Artificial Intelligence', 'AIOS', 'Smart Computing'],
  authors: [{ name: 'AIOS Team' }],
  creator: 'AIOS Team',
  publisher: 'AIOS',
  robots: 'index, follow',
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://aios.ai',
    title: 'AIOS - AI Operating System',
    description: 'Experience the future of computing with our AI-integrated operating system.',
    siteName: 'AIOS',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'AIOS - AI Operating System',
    description: 'Experience the future of computing with our AI-integrated operating system.',
    creator: '@aios',
  },
}

export const viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={inter.variable}>
      <head>
        {/* Preload critical resources */}
        <link rel="preload" href="/fonts/inter-var.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />

        {/* Performance optimizations */}
        <link rel="dns-prefetch" href="//api.openai.com" />
        <link rel="dns-prefetch" href="//api.anthropic.com" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />

        {/* Security headers */}
        <meta httpEquiv="X-Content-Type-Options" content="nosniff" />
        <meta httpEquiv="X-Frame-Options" content="DENY" />
        <meta httpEquiv="X-XSS-Protection" content="1; mode=block" />
        <meta httpEquiv="Referrer-Policy" content="strict-origin-when-cross-origin" />

        {/* Theme and appearance */}
        <meta name="theme-color" content="#3b82f6" media="(prefers-color-scheme: light)" />
        <meta name="theme-color" content="#1e40af" media="(prefers-color-scheme: dark)" />
        <meta name="color-scheme" content="light dark" />
      </head>
      <body className={`${inter.className} antialiased`}>
        <SkipToMain />
        <ErrorBoundary>
          <div id="root">
            <main id="main-content">
              {children}
            </main>
          </div>
        </ErrorBoundary>

        {/* Performance monitoring script */}
        <script
          dangerouslySetInnerHTML={{
            __html: `
              // Performance monitoring
              if (typeof window !== 'undefined' && window.performance) {
                window.addEventListener('load', function() {
                  setTimeout(function() {
                    const perfData = window.performance.timing;
                    const loadTime = perfData.loadEventEnd - perfData.navigationStart;
                    console.log('Page load time:', loadTime + 'ms');

                    // Send to analytics in production
                    if (process.env.NODE_ENV === 'production') {
                      // Analytics tracking code would go here
                    }
                  }, 0);
                });
              }
            `
          }}
        />
      </body>
    </html>
  )
}
