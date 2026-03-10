import axios from 'axios'
import {
  APIKey,
  Provider,
  Model,
  RequestLog,
  SystemLog,
  SystemSettings,
  User,
  PaginatedResult
} from '@/types'

// Allows overriding the API host at build time; defaults to same-origin /api
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

// Create axios instance with base configuration
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add token to requests if available
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Handle token expiration
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API service
class APIService {
  private paginationParams(params?: any) {
    return params || {}
  }

  // Authentication
  async login(username: string, password: string): Promise<{ token: string, user: User }> {
    const response = await apiClient.post('/login', { username, password })
    return response.data
  }

  async resetPassword(oldPassword: string, newPassword: string): Promise<void> {
    await apiClient.post('/reset-password', { oldPassword, newPassword })
  }

  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get('/user')
    return response.data
  }

  async updateCurrentUser(payload: Partial<User> & { password?: string }): Promise<User> {
    const response = await apiClient.put('/user', payload)
    return response.data
  }

  async getUsers(params?: { page?: number, pageSize?: number }): Promise<PaginatedResult<User>> {
    const response = await apiClient.get('/users', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<User>
  }

  async getAllUsers(): Promise<User[]> {
    const users: User[] = []
    const pageSize = 50
    let page = 1

    while (true) {
      const { items, total } = await this.getUsers({ page, pageSize })
      users.push(...items)
      if (users.length >= total || items.length === 0) break
      page += 1
    }

    return users
  }

  async createUser(data: Partial<User> & { password: string }): Promise<User> {
    const response = await apiClient.post('/users', data)
    return response.data
  }

  async updateUser(id: string | number, data: Partial<User> & { password?: string }): Promise<User> {
    const response = await apiClient.put(`/users/${id}`, data)
    return response.data
  }

  async deleteUser(id: string | number): Promise<void> {
    await apiClient.delete(`/users/${id}`)
  }

  // API Keys
  async getAPIKeys(params?: { page?: number, pageSize?: number }): Promise<PaginatedResult<APIKey>> {
    const response = await apiClient.get('/keys', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<APIKey>
  }

  async createAPIKey(data: Partial<APIKey>): Promise<APIKey> {
    const response = await apiClient.post('/keys', data)
    return response.data
  }

  async updateAPIKey(id: string, data: Partial<APIKey>): Promise<APIKey> {
    const response = await apiClient.put(`/keys/${id}`, data)
    return response.data
  }

  async deleteAPIKey(id: string): Promise<void> {
    await apiClient.delete(`/keys/${id}`)
  }

  // Providers
  async getProviders(params?: { page?: number, pageSize?: number }): Promise<PaginatedResult<Provider>> {
    const response = await apiClient.get('/providers', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<Provider>
  }

  async getAllProviders(): Promise<Provider[]> {
    const providers: Provider[] = []
    const pageSize = 50
    let page = 1

    while (true) {
      const { items, total } = await this.getProviders({ page, pageSize })
      providers.push(...items)
      if (providers.length >= total || items.length === 0) {
        break
      }
      page += 1
    }

    return providers
  }

  // Platforms
  async getPlatforms(): Promise<string[]> {
    const response = await apiClient.get('/platforms')
    return response.data
  }

  async createProvider(data: Partial<Provider>): Promise<Provider> {
    const response = await apiClient.post('/providers', data)
    return response.data
  }

  async updateProvider(id: string | number, data: Partial<Provider>): Promise<Provider> {
    const response = await apiClient.put(`/providers/${id}`, data)
    return response.data
  }

  async deleteProvider(id: string | number): Promise<void> {
    await apiClient.delete(`/providers/${id}`)
  }

  // Models
  async getModels(params?: { page?: number, pageSize?: number, providerId?: number }): Promise<PaginatedResult<Model>> {
    const response = await apiClient.get('/models', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<Model>
  }

  async getAllModels(params?: { providerId?: number }): Promise<Model[]> {
    const models: Model[] = []
    const pageSize = 50
    let page = 1

    while (true) {
      const { items, total } = await this.getModels({ page, pageSize, providerId: params?.providerId })
      models.push(...items)
      if (models.length >= total || items.length === 0) {
        break
      }
      page += 1
    }

    return models
  }

  async createModel(data: Partial<Model>): Promise<Model> {
    const response = await apiClient.post('/models', data)
    return response.data
  }

  async updateModel(id: string | number, data: Partial<Model>): Promise<Model> {
    const response = await apiClient.put(`/models/${id}`, data)
    return response.data
  }

  async deleteModel(id: string | number): Promise<void> {
    await apiClient.delete(`/models/${id}`)
  }

  // Requests
  async getRequests(params?: { page?: number, pageSize?: number }): Promise<PaginatedResult<RequestLog>> {
    const response = await apiClient.get('/requests', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<RequestLog>
  }

  // Logs
  async getLogs(params?: { page?: number, pageSize?: number }): Promise<PaginatedResult<SystemLog>> {
    const response = await apiClient.get('/logs', { params: this.paginationParams(params) })
    return response.data as PaginatedResult<SystemLog>
  }

  // Settings
  async getSettings(): Promise<SystemSettings> {
    const response = await apiClient.get('/settings')
    return response.data
  }

  async updateSettings(settings: SystemSettings): Promise<SystemSettings> {
    const response = await apiClient.put('/settings', settings)
    return response.data
  }

  // Dashboard stats
  async getDashboardStats(): Promise<{
    totalRequests: number
    activeAPIKeys: number
    activeProviders: number
    successRate: number
  }> {
    const response = await apiClient.get('/stats')
    return response.data
  }
}

export const apiService = new APIService()
