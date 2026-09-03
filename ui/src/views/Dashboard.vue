<template>
  <Layout>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
      <p class="text-gray-600 mt-1">System overview and statistics</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
      <StatCard
        label="Total Requests"
        :value="formatNumber(stats.totalRequests)"
        icon="fas fa-exchange-alt"
      />

      <StatCard
        label="Active API Keys"
        :value="formatNumber(stats.activeAPIKeys)"
        icon="fas fa-key"
      />

      <StatCard
        label="Active Providers"
        :value="formatNumber(stats.activeProviders)"
        icon="fas fa-server"
      />

      <StatCard
        label="Success Rate"
        :value="formatPercentage(stats.successRate)"
        icon="fas fa-chart-line"
      />
    </div>

    <!-- Recent Activity -->
    <Card title="Recent Activity">
      <div v-if="recent.length === 0" class="text-center text-gray-500 py-8">
        No requests have been proxied yet.
      </div>
      <div v-else class="space-y-4">
        <div
          v-for="item in recent"
          :key="item.id"
          class="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
        >
          <div class="flex items-center">
            <div
              class="p-2 rounded-full mr-3"
              :class="item.success ? 'bg-green-100' : 'bg-red-100'"
            >
              <i
                :class="item.success ? 'fas fa-check text-green-600' : 'fas fa-xmark text-red-600'"
              ></i>
            </div>
            <div>
              <p class="font-medium text-gray-900">
                {{ item.success ? 'Request succeeded' : `Request failed (${item.statusCode})` }}
              </p>
              <p class="text-sm text-gray-600">
                {{ getModelName(item.modelId) }} via {{ getProviderName(item.providerId) }}
                <span v-if="item.totalTokens"> · {{ formatNumber(item.totalTokens) }} tokens</span>
                <span v-if="item.responseTime"> · {{ item.responseTime }} ms</span>
              </p>
            </div>
          </div>
          <p class="text-sm text-gray-500">{{ timeAgo(item.requestTime) }}</p>
        </div>
      </div>
    </Card>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import StatCard from '@/components/StatCard.vue'
import Card from '@/components/Card.vue'
import { apiService } from '@/api'
import { Model, Provider, RequestLog } from '@/types'
import { onMounted, ref } from 'vue'

const stats = ref({
  totalRequests: 0,
  activeAPIKeys: 0,
  activeProviders: 0,
  successRate: 0
})

const recent = ref<RequestLog[]>([])
const models = ref<Model[]>([])
const providers = ref<Provider[]>([])

const formatNumber = (num: number): string => {
  return new Intl.NumberFormat().format(num)
}

const formatPercentage = (num: number): string => {
  return `${(num * 100).toFixed(1)}%`
}

const getProviderName = (providerId: string | number | undefined) => {
  if (!providerId) return 'Unknown'
  const provider = providers.value.find(p => String(p.id) === String(providerId))
  return provider ? provider.name : 'Unknown'
}

const getModelName = (modelId: string | number | undefined) => {
  if (!modelId) return 'Unknown'
  const model = models.value.find(m => String(m.id) === String(modelId))
  return model ? model.mappedName : 'Unknown'
}

const timeAgo = (value: string): string => {
  const then = new Date(value).getTime()
  if (Number.isNaN(then)) return ''

  const seconds = Math.max(0, Math.floor((Date.now() - then) / 1000))
  if (seconds < 60) return 'just now'

  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} minute${minutes === 1 ? '' : 's'} ago`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hour${hours === 1 ? '' : 's'} ago`

  const days = Math.floor(hours / 24)
  return `${days} day${days === 1 ? '' : 's'} ago`
}

onMounted(async () => {
  try {
    stats.value = await apiService.getDashboardStats()
  } catch (error) {
    console.error('Failed to fetch dashboard stats:', error)
  }

  // Model and provider names are resolved client-side, the way the Requests
  // view does it: the request records only carry ids.
  try {
    const [requests, modelList, providerList] = await Promise.all([
      apiService.getRequests({ page: 1, pageSize: 5 }),
      apiService.getModels({ page: 1, pageSize: 100 }),
      apiService.getProviders({ page: 1, pageSize: 100 }),
    ])
    recent.value = requests.items
    models.value = modelList.items
    providers.value = providerList.items
  } catch (error) {
    console.error('Failed to fetch recent activity:', error)
  }
})
</script>

<style scoped>
</style>
