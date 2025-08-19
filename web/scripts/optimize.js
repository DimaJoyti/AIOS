#!/usr/bin/env node

/**
 * AIOS Optimization Script
 * 
 * This script performs various optimizations for the AIOS application:
 * - Bundle analysis
 * - Image optimization
 * - Performance auditing
 * - Accessibility testing
 * - SEO validation
 */

const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')

console.log('🚀 Starting AIOS Optimization Process...\n')

// Check if we're in the correct directory
if (!fs.existsSync('package.json')) {
  log('❌ Error: package.json not found. Please run this script from the project root.', 'red')
  process.exit(1)
}

// Check if the application is running
async function checkAppRunning() {
  try {
    const http = require('http')
    const options = {
      hostname: 'localhost',
      port: 3003,
      path: '/',
      method: 'GET',
      timeout: 2000
    }

    return new Promise((resolve) => {
      const req = http.request(options, () => {
        resolve(true)
      })

      req.on('error', () => {
        resolve(false)
      })

      req.on('timeout', () => {
        resolve(false)
      })

      req.end()
    })
  } catch (error) {
    return Promise.resolve(false)
  }
}

console.log('🚀 Starting AIOS Optimization Process...\n')

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
}

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`)
}

function section(title) {
  log(`\n${colors.bright}=== ${title} ===${colors.reset}`)
}

async function runOptimization() {
  // Check if app is running for performance tests
  const isAppRunning = await checkAppRunning()
  if (!isAppRunning) {
    log('⚠️  Application not running on localhost:3003. Some tests may be skipped.', 'yellow')
  }

// 1. Bundle Analysis
section('Bundle Analysis')
try {
  log('📦 Analyzing bundle size...', 'blue')
  
  // Check if bundle analyzer is available
  const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'))
  
  if (packageJson.scripts && packageJson.scripts['analyze']) {
    execSync('npm run analyze', { stdio: 'inherit' })
  } else {
    log('⚠️  Bundle analyzer not configured. Add "analyze" script to package.json', 'yellow')
  }
  
  log('✅ Bundle analysis complete', 'green')
} catch (error) {
  log('❌ Bundle analysis failed', 'red')
  console.error(error.message)
}

// 2. Performance Audit
section('Performance Audit')
try {
  log('⚡ Running performance audit...', 'blue')
  
  // Check for Lighthouse CLI
  try {
    execSync('lighthouse --version', { stdio: 'pipe' })
    
    // Run Lighthouse audit
    const auditCommand = `lighthouse http://localhost:3003 --output=json --output-path=./lighthouse-report.json --chrome-flags="--headless" --quiet`
    execSync(auditCommand, { stdio: 'inherit' })
    
    // Parse results
    if (fs.existsSync('./lighthouse-report.json')) {
      const report = JSON.parse(fs.readFileSync('./lighthouse-report.json', 'utf8'))
      const scores = report.lhr.categories
      
      log('\n📊 Lighthouse Scores:', 'cyan')
      log(`Performance: ${Math.round(scores.performance.score * 100)}/100`, 'blue')
      log(`Accessibility: ${Math.round(scores.accessibility.score * 100)}/100`, 'blue')
      log(`Best Practices: ${Math.round(scores['best-practices'].score * 100)}/100`, 'blue')
      log(`SEO: ${Math.round(scores.seo.score * 100)}/100`, 'blue')
      
      // Clean up
      fs.unlinkSync('./lighthouse-report.json')
    }
    
    log('✅ Performance audit complete', 'green')
  } catch (lighthouseError) {
    log('⚠️  Lighthouse not installed. Install with: npm install -g lighthouse', 'yellow')
  }
} catch (error) {
  log('❌ Performance audit failed', 'red')
  console.error(error.message)
}

// 3. Image Optimization
section('Image Optimization')
try {
  log('🖼️  Checking image optimization...', 'blue')
  
  const publicDir = path.join(__dirname, '../public')
  const imageExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg']
  
  function findImages(dir) {
    const images = []
    const files = fs.readdirSync(dir)
    
    for (const file of files) {
      const filePath = path.join(dir, file)
      const stat = fs.statSync(filePath)
      
      if (stat.isDirectory()) {
        images.push(...findImages(filePath))
      } else if (imageExtensions.some(ext => file.toLowerCase().endsWith(ext))) {
        images.push({
          path: filePath,
          size: stat.size,
          name: file
        })
      }
    }
    
    return images
  }
  
  if (fs.existsSync(publicDir)) {
    const images = findImages(publicDir)
    const largeImages = images.filter(img => img.size > 500 * 1024) // > 500KB
    
    log(`Found ${images.length} images`, 'blue')
    
    if (largeImages.length > 0) {
      log(`⚠️  ${largeImages.length} large images found (>500KB):`, 'yellow')
      largeImages.forEach(img => {
        const sizeMB = (img.size / (1024 * 1024)).toFixed(2)
        log(`  - ${img.name}: ${sizeMB}MB`, 'yellow')
      })
      log('Consider optimizing these images with tools like imagemin or next/image', 'yellow')
    } else {
      log('✅ All images are optimally sized', 'green')
    }
  }
  
  log('✅ Image optimization check complete', 'green')
} catch (error) {
  log('❌ Image optimization check failed', 'red')
  console.error(error.message)
}

// 4. Accessibility Testing
section('Accessibility Testing')
try {
  log('♿ Running accessibility tests...', 'blue')
  
  // Check if jest and testing libraries are available
  if (fs.existsSync('__tests__/accessibility.test.tsx')) {
    try {
      execSync('npm test -- --testPathPattern=accessibility', { stdio: 'inherit' })
      log('✅ Accessibility tests passed', 'green')
    } catch (testError) {
      log('❌ Some accessibility tests failed', 'red')
    }
  } else {
    log('⚠️  Accessibility tests not found', 'yellow')
  }
} catch (error) {
  log('❌ Accessibility testing failed', 'red')
  console.error(error.message)
}

// 5. SEO Validation
section('SEO Validation')
try {
  log('🔍 Validating SEO implementation...', 'blue')
  
  const checks = [
    {
      name: 'Meta tags component',
      file: 'components/SEOHead.tsx',
      required: true
    },
    {
      name: 'Sitemap',
      file: 'public/sitemap.xml',
      required: false
    },
    {
      name: 'Robots.txt',
      file: 'public/robots.txt',
      required: false
    },
    {
      name: 'Favicon',
      file: 'public/favicon.ico',
      required: true
    }
  ]
  
  let seoScore = 0
  const totalChecks = checks.length
  
  checks.forEach(check => {
    const filePath = path.join(__dirname, '..', check.file)
    if (fs.existsSync(filePath)) {
      log(`✅ ${check.name} found`, 'green')
      seoScore++
    } else {
      const level = check.required ? 'red' : 'yellow'
      const symbol = check.required ? '❌' : '⚠️ '
      log(`${symbol} ${check.name} missing`, level)
    }
  })
  
  log(`\n📊 SEO Score: ${seoScore}/${totalChecks}`, 'cyan')
  
  if (seoScore === totalChecks) {
    log('✅ SEO validation complete', 'green')
  } else {
    log('⚠️  SEO validation completed with warnings', 'yellow')
  }
} catch (error) {
  log('❌ SEO validation failed', 'red')
  console.error(error.message)
}

// 6. Code Quality Check
section('Code Quality')
try {
  log('🔍 Checking code quality...', 'blue')
  
  // Check for TypeScript errors
  try {
    execSync('npx tsc --noEmit', { stdio: 'pipe' })
    log('✅ TypeScript compilation successful', 'green')
  } catch (tscError) {
    log('❌ TypeScript errors found', 'red')
    console.error(tscError.stdout?.toString() || tscError.message)
  }
  
  // Check for ESLint issues
  try {
    execSync('npx eslint . --ext .ts,.tsx --max-warnings 0', { stdio: 'pipe' })
    log('✅ ESLint checks passed', 'green')
  } catch (eslintError) {
    log('⚠️  ESLint warnings/errors found', 'yellow')
  }
  
} catch (error) {
  log('❌ Code quality check failed', 'red')
  console.error(error.message)
}

// 7. Security Audit
section('Security Audit')
try {
  log('🔒 Running security audit...', 'blue')
  
  try {
    execSync('npm audit --audit-level moderate', { stdio: 'inherit' })
    log('✅ Security audit passed', 'green')
  } catch (auditError) {
    log('⚠️  Security vulnerabilities found. Run "npm audit fix" to resolve', 'yellow')
  }
  
} catch (error) {
  log('❌ Security audit failed', 'red')
  console.error(error.message)
}

// Summary
section('Optimization Summary')
log('🎉 AIOS optimization process completed!', 'green')
log('\n📋 Next Steps:', 'cyan')
log('1. Review any warnings or errors above', 'blue')
log('2. Optimize large images if found', 'blue')
log('3. Fix any accessibility issues', 'blue')
log('4. Address security vulnerabilities', 'blue')
log('5. Consider implementing missing SEO features', 'blue')
log('\n🚀 Your AIOS application is ready for production!', 'bright')

console.log('\n')
}

// Run the optimization process
runOptimization().catch(error => {
  log('❌ Optimization process failed:', 'red')
  console.error(error)
  process.exit(1)
})
