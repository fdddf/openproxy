export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface APIKey {
  id: string
  key: string
  description: string
  expiresAt: string | null
  createdAt: string
  updatedAt: string
  isActive: boolean
  usageCount: number
  maxUsage: number | null
}

export interface Provider {
  id: number
  name: string
  type: string
  key: string
  baseUrl: string
  proxyUrl: string | null
  isActive: boolean
  healthCheckEnabled: boolean
  healthCheckStatus?: string
  healthCheckChecked?: string | null
  createdAt: string
  updatedAt: string
  models: Model[]
  clientId?: string
  clientSecret?: string
  accessToken?: string
  refreshToken?: string
  tokenExpiry?: string
  authUrl?: string
  tokenUrl?: string
  redirectUrl?: string
  scopes?: string
  accountId?: string
}

export interface Model {
  id: number
  providerId: number
  name: string
  mappedName: string
  isActive: boolean
  createdAt: string
  updatedAt: string
  usageCount: number
}

export interface User {
  id: number
  username: string
  email: string
  displayName: string
  avatarUrl: string
  bio: string
  isSuper: boolean
  createdAt: string
  updatedAt: string
}

export interface RequestLog {
  id: string
  apiKeyId: string
  providerId: string
  modelId: string
  requestTime: string
  responseTime: number
  statusCode: number
  promptTokens: number
  completionTokens: number
  totalTokens: number
  cost: number
  success: boolean
  requestBody?: string
  responseBody?: string
  requestHeaders?: string
  responseHeaders?: string
}

export interface SystemLog {
  id: string
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  details: string
  timestamp: string
}

export interface SystemSettings {
  defaultProvider: string
  rateLimit: number
  concurrentLimit: number
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  cacheTTL: number
  maintenanceMode: boolean
  healthCheckEnabled: boolean
}
