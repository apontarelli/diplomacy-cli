import 'htmx.org'
import './styles/main.css'

// Initialize HTMX with modern configuration
document.addEventListener('DOMContentLoaded', () => {
  // Configure HTMX for modern usage
  window.htmx.config.globalViewTransitions = true
  window.htmx.config.scrollBehavior = 'smooth'
  window.htmx.config.defaultSwapStyle = 'innerHTML'
  window.htmx.config.defaultSwapDelay = 0
  window.htmx.config.defaultSettleDelay = 20
  
  // Enable logging in development
  if (import.meta.env.DEV) {
    window.htmx.logAll()
  }
  
  // Load WebSocket extension
  const wsScript = document.createElement('script')
  wsScript.src = '/js/ws.js'
  wsScript.onload = () => {
    console.log('HTMX WebSocket extension loaded')
  }
  document.head.appendChild(wsScript)
  
  // Global error handling for HTMX requests
  document.addEventListener('htmx:responseError', (event: Event) => {
    const detail = (event as CustomEvent).detail
    console.error('HTMX Response Error:', detail)
    
    // Show user-friendly error message
    const errorDiv = document.createElement('div')
    errorDiv.className = 'error-toast'
    errorDiv.textContent = 'Something went wrong. Please try again.'
    document.body.appendChild(errorDiv)
    
    setTimeout(() => errorDiv.remove(), 5000)
  })
  
  // Handle WebSocket connection events
  document.addEventListener('htmx:wsOpen', (event: Event) => {
    console.log('WebSocket connected:', (event as CustomEvent).detail)
    updateConnectionStatus('connected')
  })
  
  document.addEventListener('htmx:wsClose', (event: Event) => {
    console.log('WebSocket disconnected:', (event as CustomEvent).detail)
    updateConnectionStatus('disconnected')
  })
  
  document.addEventListener('htmx:wsError', (event: Event) => {
    console.error('WebSocket error:', (event as CustomEvent).detail)
    updateConnectionStatus('error')
  })
})

// Utility functions
function updateConnectionStatus(status: 'connected' | 'disconnected' | 'error') {
  const statusElement = document.getElementById('connection-status')
  if (statusElement) {
    statusElement.className = `connection-status ${status}`
    statusElement.textContent = status.charAt(0).toUpperCase() + status.slice(1)
  }
}

// Game-specific utilities
export class GameBoard {
  private selectedProvince: string | null = null
  
  selectProvince(provinceId: string) {
    // Clear previous selection
    document.querySelectorAll('.province.selected').forEach(el => {
      el.classList.remove('selected')
    })
    
    // Select new province
    const province = document.getElementById(`province-${provinceId}`)
    if (province) {
      province.classList.add('selected')
      this.selectedProvince = provinceId
      
      // Store in session for form population
      sessionStorage.setItem('selectedProvince', provinceId)
      
      // Trigger HTMX request for province details
      window.htmx.trigger(province, 'provinceSelected', { provinceId })
    }
  }
  
  getSelectedProvince(): string | null {
    return this.selectedProvince || sessionStorage.getItem('selectedProvince')
  }
  
  clearSelection() {
    document.querySelectorAll('.province.selected').forEach(el => {
      el.classList.remove('selected')
    })
    this.selectedProvince = null
    sessionStorage.removeItem('selectedProvince')
  }
}

// Export global game board instance
export const gameBoard = new GameBoard()

// Make it available globally for HTMX attributes
;(window as any).gameBoard = gameBoard
