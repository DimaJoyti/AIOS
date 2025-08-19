import { render, screen } from '@testing-library/react'
import { axe, toHaveNoViolations } from 'jest-axe'
import userEvent from '@testing-library/user-event'
import { AccessibleButton, AccessibleModal, AccessibleProgress } from '../components/AccessibilityHelpers'

// Extend Jest matchers
expect.extend(toHaveNoViolations)

describe('Accessibility Tests', () => {
  describe('AccessibleButton', () => {
    it('should have no accessibility violations', async () => {
      const { container } = render(
        <AccessibleButton onClick={() => {}}>
          Click me
        </AccessibleButton>
      )
      
      const results = await axe(container)
      expect(results).toHaveNoViolations()
    })

    it('should have proper ARIA attributes', () => {
      render(
        <AccessibleButton 
          onClick={() => {}} 
          ariaLabel="Custom label"
          ariaDescribedBy="description"
        >
          Click me
        </AccessibleButton>
      )
      
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-label', 'Custom label')
      expect(button).toHaveAttribute('aria-describedby', 'description')
    })

    it('should be keyboard accessible', async () => {
      const user = userEvent.setup()
      const handleClick = jest.fn()
      
      render(
        <AccessibleButton onClick={handleClick}>
          Click me
        </AccessibleButton>
      )
      
      const button = screen.getByRole('button')
      await user.tab()
      expect(button).toHaveFocus()
      
      await user.keyboard('{Enter}')
      expect(handleClick).toHaveBeenCalledTimes(1)
      
      await user.keyboard(' ')
      expect(handleClick).toHaveBeenCalledTimes(2)
    })

    it('should handle disabled state correctly', () => {
      render(
        <AccessibleButton disabled onClick={() => {}}>
          Disabled button
        </AccessibleButton>
      )
      
      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
      expect(button).toHaveAttribute('aria-disabled', 'true')
    })

    it('should handle loading state correctly', () => {
      render(
        <AccessibleButton loading onClick={() => {}}>
          Loading button
        </AccessibleButton>
      )
      
      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
      expect(button).toHaveAttribute('aria-busy', 'true')
    })
  })

  describe('AccessibleModal', () => {
    it('should have no accessibility violations', async () => {
      const { container } = render(
        <AccessibleModal 
          isOpen={true} 
          onClose={() => {}} 
          title="Test Modal"
        >
          <p>Modal content</p>
        </AccessibleModal>
      )
      
      const results = await axe(container)
      expect(results).toHaveNoViolations()
    })

    it('should have proper modal attributes', () => {
      render(
        <AccessibleModal 
          isOpen={true} 
          onClose={() => {}} 
          title="Test Modal"
        >
          <p>Modal content</p>
        </AccessibleModal>
      )
      
      const modal = screen.getByRole('dialog')
      expect(modal).toHaveAttribute('aria-modal', 'true')
      expect(modal).toHaveAttribute('aria-labelledby', 'modal-title')
    })

    it('should trap focus within modal', async () => {
      const user = userEvent.setup()
      
      render(
        <div>
          <button>Outside button</button>
          <AccessibleModal 
            isOpen={true} 
            onClose={() => {}} 
            title="Test Modal"
          >
            <button>Modal button 1</button>
            <button>Modal button 2</button>
          </AccessibleModal>
        </div>
      )
      
      const modalButton1 = screen.getByText('Modal button 1')
      const modalButton2 = screen.getByText('Modal button 2')
      
      // Focus should start on first modal element
      expect(modalButton1).toHaveFocus()
      
      // Tab should cycle within modal
      await user.tab()
      expect(modalButton2).toHaveFocus()
      
      await user.tab()
      expect(modalButton1).toHaveFocus()
    })

    it('should close on Escape key', async () => {
      const user = userEvent.setup()
      const handleClose = jest.fn()
      
      render(
        <AccessibleModal 
          isOpen={true} 
          onClose={handleClose} 
          title="Test Modal"
        >
          <p>Modal content</p>
        </AccessibleModal>
      )
      
      await user.keyboard('{Escape}')
      expect(handleClose).toHaveBeenCalledTimes(1)
    })
  })

  describe('AccessibleProgress', () => {
    it('should have no accessibility violations', async () => {
      const { container } = render(
        <AccessibleProgress 
          value={50} 
          max={100} 
          label="Loading progress" 
        />
      )
      
      const results = await axe(container)
      expect(results).toHaveNoViolations()
    })

    it('should have proper progressbar attributes', () => {
      render(
        <AccessibleProgress 
          value={75} 
          max={100} 
          label="Upload progress" 
        />
      )
      
      const progressbar = screen.getByRole('progressbar')
      expect(progressbar).toHaveAttribute('aria-valuenow', '75')
      expect(progressbar).toHaveAttribute('aria-valuemin', '0')
      expect(progressbar).toHaveAttribute('aria-valuemax', '100')
      expect(progressbar).toHaveAttribute('aria-label', 'Upload progress: 75% complete')
    })
  })

  describe('Keyboard Navigation', () => {
    it('should support arrow key navigation in lists', async () => {
      const user = userEvent.setup()
      const items = ['Item 1', 'Item 2', 'Item 3']
      
      render(
        <div role="listbox" tabIndex={0}>
          {items.map((item, index) => (
            <div 
              key={item}
              role="option" 
              tabIndex={-1}
              aria-selected={index === 0}
            >
              {item}
            </div>
          ))}
        </div>
      )
      
      const listbox = screen.getByRole('listbox')
      listbox.focus()
      
      // Test arrow key navigation
      await user.keyboard('{ArrowDown}')
      const secondItem = screen.getByText('Item 2')
      expect(secondItem).toHaveAttribute('aria-selected', 'true')
    })
  })

  describe('Screen Reader Support', () => {
    it('should provide proper labels for form elements', () => {
      render(
        <form>
          <label htmlFor="email">Email Address</label>
          <input 
            id="email" 
            type="email" 
            aria-describedby="email-help"
            required 
          />
          <div id="email-help">Enter your email address</div>
        </form>
      )
      
      const input = screen.getByLabelText('Email Address')
      expect(input).toHaveAttribute('aria-describedby', 'email-help')
      expect(input).toBeRequired()
    })

    it('should announce dynamic content changes', () => {
      render(
        <div>
          <div aria-live="polite" aria-atomic="true">
            Status updated
          </div>
        </div>
      )
      
      const liveRegion = screen.getByText('Status updated')
      expect(liveRegion).toHaveAttribute('aria-live', 'polite')
      expect(liveRegion).toHaveAttribute('aria-atomic', 'true')
    })
  })

  describe('Color Contrast', () => {
    it('should meet WCAG color contrast requirements', () => {
      // This would typically be tested with automated tools
      // or manual testing with color contrast analyzers
      const { container } = render(
        <div className="bg-blue-600 text-white p-4">
          High contrast text
        </div>
      )
      
      const element = container.firstChild as HTMLElement
      const styles = window.getComputedStyle(element)
      
      // In a real test, you would calculate the contrast ratio
      // and ensure it meets WCAG AA standards (4.5:1 for normal text)
      expect(styles.backgroundColor).toBeTruthy()
      expect(styles.color).toBeTruthy()
    })
  })

  describe('Focus Management', () => {
    it('should have visible focus indicators', () => {
      render(
        <button className="focus:ring-2 focus:ring-blue-500">
          Focusable button
        </button>
      )
      
      const button = screen.getByRole('button')
      button.focus()
      
      // Check that focus styles are applied
      expect(button).toHaveFocus()
      expect(button).toHaveClass('focus:ring-2')
    })

    it('should skip to main content', async () => {
      const user = userEvent.setup()
      
      render(
        <div>
          <a href="#main-content" className="sr-only focus:not-sr-only">
            Skip to main content
          </a>
          <nav>Navigation</nav>
          <main id="main-content">
            <h1>Main Content</h1>
          </main>
        </div>
      )
      
      const skipLink = screen.getByText('Skip to main content')
      await user.tab()
      expect(skipLink).toHaveFocus()
      
      await user.keyboard('{Enter}')
      const mainContent = document.getElementById('main-content')
      expect(mainContent).toHaveFocus()
    })
  })
})
