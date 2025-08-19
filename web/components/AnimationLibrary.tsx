'use client'

import { motion, Variants, Transition } from 'framer-motion'
import { useReducedMotion } from './AccessibilityHelpers'

// Common animation variants
export const fadeInUp: Variants = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -20 }
}

export const fadeInDown: Variants = {
  initial: { opacity: 0, y: -20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: 20 }
}

export const fadeInLeft: Variants = {
  initial: { opacity: 0, x: -20 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: 20 }
}

export const fadeInRight: Variants = {
  initial: { opacity: 0, x: 20 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -20 }
}

export const scaleIn: Variants = {
  initial: { opacity: 0, scale: 0.9 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.9 }
}

export const slideInUp: Variants = {
  initial: { y: '100%' },
  animate: { y: 0 },
  exit: { y: '100%' }
}

export const slideInDown: Variants = {
  initial: { y: '-100%' },
  animate: { y: 0 },
  exit: { y: '-100%' }
}

export const slideInLeft: Variants = {
  initial: { x: '-100%' },
  animate: { x: 0 },
  exit: { x: '-100%' }
}

export const slideInRight: Variants = {
  initial: { x: '100%' },
  animate: { x: 0 },
  exit: { x: '100%' }
}

export const rotateIn: Variants = {
  initial: { opacity: 0, rotate: -180 },
  animate: { opacity: 1, rotate: 0 },
  exit: { opacity: 0, rotate: 180 }
}

export const bounceIn: Variants = {
  initial: { opacity: 0, scale: 0.3 },
  animate: { 
    opacity: 1, 
    scale: 1,
    transition: {
      type: "spring",
      stiffness: 400,
      damping: 10
    }
  },
  exit: { opacity: 0, scale: 0.3 }
}

// Stagger animations
export const staggerContainer: Variants = {
  initial: {},
  animate: {
    transition: {
      staggerChildren: 0.1
    }
  },
  exit: {
    transition: {
      staggerChildren: 0.05,
      staggerDirection: -1
    }
  }
}

export const staggerItem: Variants = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -20 }
}

// Common transitions
export const smoothTransition: Transition = {
  type: "tween",
  duration: 0.3,
  ease: "easeInOut"
}

export const springTransition: Transition = {
  type: "spring",
  stiffness: 300,
  damping: 30
}

export const bounceTransition: Transition = {
  type: "spring",
  stiffness: 400,
  damping: 10
}

// Animation wrapper components
export function FadeIn({ 
  children, 
  direction = 'up',
  delay = 0,
  duration = 0.3,
  className = ''
}: {
  children: React.ReactNode
  direction?: 'up' | 'down' | 'left' | 'right'
  delay?: number
  duration?: number
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()
  
  const variants = {
    up: fadeInUp,
    down: fadeInDown,
    left: fadeInLeft,
    right: fadeInRight
  }

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      variants={variants[direction]}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ ...smoothTransition, delay, duration }}
    >
      {children}
    </motion.div>
  )
}

export function ScaleIn({ 
  children, 
  delay = 0,
  className = ''
}: {
  children: React.ReactNode
  delay?: number
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      variants={scaleIn}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ ...smoothTransition, delay }}
    >
      {children}
    </motion.div>
  )
}

export function SlideIn({ 
  children, 
  direction = 'up',
  delay = 0,
  className = ''
}: {
  children: React.ReactNode
  direction?: 'up' | 'down' | 'left' | 'right'
  delay?: number
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()
  
  const variants = {
    up: slideInUp,
    down: slideInDown,
    left: slideInLeft,
    right: slideInRight
  }

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      variants={variants[direction]}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ ...smoothTransition, delay }}
    >
      {children}
    </motion.div>
  )
}

export function StaggeredList({ 
  children, 
  className = '',
  staggerDelay = 0.1
}: {
  children: React.ReactNode
  className?: string
  staggerDelay?: number
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      variants={staggerContainer}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ staggerChildren: staggerDelay }}
    >
      {children}
    </motion.div>
  )
}

export function StaggeredItem({ 
  children, 
  className = ''
}: {
  children: React.ReactNode
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      variants={staggerItem}
      transition={smoothTransition}
    >
      {children}
    </motion.div>
  )
}

// Interactive animations
export function HoverScale({ 
  children, 
  scale = 1.05,
  className = ''
}: {
  children: React.ReactNode
  scale?: number
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      whileHover={{ scale }}
      whileTap={{ scale: scale * 0.95 }}
      transition={springTransition}
    >
      {children}
    </motion.div>
  )
}

export function HoverLift({ 
  children, 
  lift = -4,
  className = ''
}: {
  children: React.ReactNode
  lift?: number
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      whileHover={{ y: lift }}
      transition={springTransition}
    >
      {children}
    </motion.div>
  )
}

export function PulseAnimation({ 
  children, 
  className = ''
}: {
  children: React.ReactNode
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      animate={{ scale: [1, 1.05, 1] }}
      transition={{ 
        duration: 2, 
        repeat: Infinity, 
        ease: "easeInOut" 
      }}
    >
      {children}
    </motion.div>
  )
}

export function FloatingAnimation({ 
  children, 
  className = ''
}: {
  children: React.ReactNode
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      animate={{ y: [-5, 5, -5] }}
      transition={{ 
        duration: 3, 
        repeat: Infinity, 
        ease: "easeInOut" 
      }}
    >
      {children}
    </motion.div>
  )
}

// Page transition wrapper
export function PageTransition({ 
  children, 
  className = ''
}: {
  children: React.ReactNode
  className?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>
  }

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={{ duration: 0.3, ease: "easeInOut" }}
    >
      {children}
    </motion.div>
  )
}

// Loading animation
export function LoadingDots({ 
  className = '',
  color = 'bg-blue-500'
}: {
  className?: string
  color?: string
}) {
  const prefersReducedMotion = useReducedMotion()

  if (prefersReducedMotion) {
    return (
      <div className={`flex space-x-1 ${className}`}>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className={`w-2 h-2 rounded-full ${color}`} />
        ))}
      </div>
    )
  }

  return (
    <div className={`flex space-x-1 ${className}`}>
      {Array.from({ length: 3 }).map((_, i) => (
        <motion.div
          key={i}
          className={`w-2 h-2 rounded-full ${color}`}
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{
            duration: 1.5,
            repeat: Infinity,
            delay: i * 0.2,
            ease: "easeInOut"
          }}
        />
      ))}
    </div>
  )
}

export default {
  FadeIn,
  ScaleIn,
  SlideIn,
  StaggeredList,
  StaggeredItem,
  HoverScale,
  HoverLift,
  PulseAnimation,
  FloatingAnimation,
  PageTransition,
  LoadingDots
}
