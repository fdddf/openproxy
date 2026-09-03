import { defineStore } from 'pinia'
import { apiService } from '@/api'
import { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('token'),
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    isAuthenticated: !!localStorage.getItem('token')
  }),
  
  actions: {
    // Persists a freshly issued token. Shared by login and first-run setup,
    // which both hand back a token/user pair.
    setSession(token: string, user: User) {
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))

      this.token = token
      this.user = user
      this.isAuthenticated = true
    },

    async login(username: string, password: string) {
      try {
        const { token, user } = await apiService.login(username, password)
        this.setSession(token, user)
        return true
      } catch (error) {
        console.error('Login failed:', error)
        return false
      }
    },
    
    async refreshUser() {
      if (!this.token) return
      try {
        const user = await apiService.getCurrentUser()
        this.user = user
        localStorage.setItem('user', JSON.stringify(user))
      } catch (error) {
        console.error('Failed to refresh user profile:', error)
      }
    },
    
    async logout() {
      try {
        // Note: logout endpoint may not exist in the API service
      } catch (error) {
        console.error('Logout failed:', error)
      }
      
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      
      this.token = null
      this.user = null
      this.isAuthenticated = false
    },
    
    checkAuth() {
      const token = localStorage.getItem('token')
      if (token) {
        this.token = token
        this.isAuthenticated = true
        try {
          const user = JSON.parse(localStorage.getItem('user') || 'null')
          this.user = user
        } catch {
          this.user = null
        }
      }
    }
  }
})
