import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

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
  viewport: {
    width: 'device-width',
    initialScale: 1,
    maximumScale: 1,
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={inter.variable}>
      <body className={`${inter.className} antialiased`}>
        <div id="root">
          {children}
        </div>
      </body>
    </html>
  )
}
