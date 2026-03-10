<template>
  <Layout>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Request History</h1>
        <p class="text-gray-600 mt-1">View detailed history of API requests</p>
      </div>
      <div class="flex space-x-3">
        <select 
          v-model="providerFilter"
          class="input-field"
        >
          <option value="">All Providers</option>
          <option v-for="provider in providers" :key="provider.id" :value="provider.id">
            {{ provider.name }}
          </option>
        </select>
        <select 
          v-model="statusFilter"
          class="input-field"
        >
          <option value="">All Status</option>
          <option value="success">Success</option>
          <option value="failed">Failed</option>
        </select>
      </div>
    </div>
    
    <Card>
      <Table :columns="columns" :items="filteredRequests">
        <template #providerName="{ item }">
          <span class="text-sm font-medium">
            {{ getProviderName(item.providerId) }}
          </span>
        </template>
        
        <template #modelName="{ item }">
          <span class="text-sm">
            {{ getModelName(item.modelId) }}
          </span>
        </template>
        
        <template #statusCode="{ item }">
          <span class="badge" :class="getStatusClass(item.statusCode)">
            {{ item.statusCode }}
          </span>
        </template>
        
        <template #success="{ value }">
          <span class="badge" :class="value ? 'badge-success' : 'badge-danger'">
            {{ value ? 'Success' : 'Failed' }}
          </span>
        </template>
        
        <template #cost="{ value }">
          <span class="text-sm">
            ${{ value.toFixed(4) }}
          </span>
        </template>
        
        <template #actions="{ item }">
          <button 
            class="text-blue-500 hover:text-blue-700"
            @click="openDetailsModal(item)"
          >
            <i class="fas fa-eye"></i>
          </button>
        </template>
      </Table>
      <PaginationControls
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="handlePageChange"
      />
    </Card>
    
    <!-- Details Modal -->
    <Modal 
      :show="showDetailsModal"
      title="Request Details"
      @close="closeDetailsModal"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <h4 class="font-medium text-sm text-gray-600">Request ID</h4>
            <p class="text-sm">{{ currentRequest?.id }}</p>
          </div>
          <div>
            <h4 class="font-medium text-sm text-gray-600">Provider</h4>
            <p class="text-sm">{{ getProviderName(currentRequest?.providerId) }}</p>
          </div>
          <div>
            <h4 class="font-medium text-sm text-gray-600">Model</h4>
            <p class="text-sm">{{ getModelName(currentRequest?.modelId) }}</p>
          </div>
          <div>
            <h4 class="font-medium text-sm text-gray-600">Status</h4>
            <p class="text-sm">
              <span class="badge" :class="currentRequest ? getStatusClass(currentRequest.statusCode) : ''">
                {{ currentRequest?.statusCode }} {{ currentRequest?.success ? 'Success' : 'Failed' }}
              </span>
            </p>
          </div>
          <div>
            <h4 class="font-medium text-sm text-gray-600">Response Time</h4>
            <p class="text-sm">{{ currentRequest?.responseTime }}ms</p>
          </div>
          <div>
            <h4 class="font-medium text-sm text-gray-600">Cost</h4>
            <p class="text-sm">${{ currentRequest ? currentRequest.cost.toFixed(4) : '0.0000' }}</p>
          </div>
        </div>
        
        <div>
          <h4 class="font-medium text-sm text-gray-600">Tokens</h4>
          <div class="grid grid-cols-3 gap-4 mt-2">
            <div class="text-center p-2 bg-gray-50 rounded">
              <p class="text-xs text-gray-500">Prompt</p>
              <p class="font-medium">{{ currentRequest?.promptTokens }}</p>
            </div>
            <div class="text-center p-2 bg-gray-50 rounded">
              <p class="text-xs text-gray-500">Completion</p>
              <p class="font-medium">{{ currentRequest?.completionTokens }}</p>
            </div>
            <div class="text-center p-2 bg-gray-50 rounded">
              <p class="text-xs text-gray-500">Total</p>
              <p class="font-medium">{{ currentRequest?.totalTokens }}</p>
            </div>
          </div>
        </div>
        
        <div>
          <h4 class="font-medium text-sm text-gray-600">Request Time</h4>
          <p class="text-sm">{{ currentRequest ? new Date(currentRequest.requestTime).toLocaleString() : '' }}</p>
        </div>
        
        <div v-if="currentRequest?.requestBody">
          <h4 class="font-medium text-sm text-gray-600">Request Body</h4>
          <pre class="text-sm bg-gray-50 p-2 rounded mt-1 overflow-auto max-h-40">{{ JSON.stringify(JSON.parse(currentRequest.requestBody), null, 2) }}</pre>
        </div>
        
        <div v-if="currentRequest?.responseBody">
          <h4 class="font-medium text-sm text-gray-600">Response Body</h4>
          <pre class="text-sm bg-gray-50 p-2 rounded mt-1 overflow-auto max-h-40">{{ JSON.stringify(JSON.parse(currentRequest.responseBody), null, 2) }}</pre>
        </div>
        
        <div v-if="currentRequest?.requestHeaders">
          <h4 class="font-medium text-sm text-gray-600">Request Headers</h4>
          <pre class="text-sm bg-gray-50 p-2 rounded mt-1 overflow-auto max-h-40">{{ JSON.stringify(JSON.parse(currentRequest.requestHeaders), null, 2) }}</pre>
        </div>
        
        <div v-if="currentRequest?.responseHeaders">
          <h4 class="font-medium text-sm text-gray-600">Response Headers</h4>
          <pre class="text-sm bg-gray-50 p-2 rounded mt-1 overflow-auto max-h-40">{{ JSON.stringify(JSON.parse(currentRequest.responseHeaders), null, 2) }}</pre>
        </div>
      </div>
    </Modal>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import Card from '@/components/Card.vue'
import Table from '@/components/Table.vue'
import Modal from '@/components/Modal.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import { apiService } from '@/api'
import { RequestLog, Provider, Model } from '@/types'
import { ref, computed, onMounted, reactive } from 'vue'

const columns = [
  { key: 'id', title: 'Request ID' },
  { key: 'providerName', title: 'Provider' },
  { key: 'modelName', title: 'Model' },
  { key: 'statusCode', title: 'Status' },
  { key: 'success', title: 'Result' },
  { key: 'responseTime', title: 'Time (ms)' },
  { key: 'totalTokens', title: 'Tokens' },
  { key: 'cost', title: 'Cost' }
]

const requests = ref<RequestLog[]>([])
const providers = ref<Provider[]>([])
const models = ref<Model[]>([])
const providerFilter = ref('')
const statusFilter = ref('')
const showDetailsModal = ref(false)
const currentRequest = ref<RequestLog | null>(null)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const filteredRequests = computed(() => {
  return requests.value.filter(request => {
    const matchesProvider = !providerFilter.value || String(request.providerId) === providerFilter.value
    const matchesStatus = !statusFilter.value || 
      (statusFilter.value === 'success' && request.success) ||
      (statusFilter.value === 'failed' && !request.success)
    return matchesProvider && matchesStatus
  })
})

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

const getStatusClass = (statusCode: number | undefined) => {
  if (!statusCode) return ''
  if (statusCode >= 200 && statusCode < 300) return 'badge-success'
  if (statusCode >= 400 && statusCode < 500) return 'badge-warning'
  if (statusCode >= 500) return 'badge-danger'
  return ''
}

const fetchRequests = async () => {
  try {
    const data = await apiService.getRequests({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    requests.value = data.items.map(normalizeRequest)
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to fetch requests:', error)
  }
}

const fetchProvidersAndModels = async () => {
  try {
    const [providersData, modelsData] = await Promise.all([
      apiService.getAllProviders(),
      apiService.getAllModels()
    ])
    providers.value = providersData
    models.value = modelsData
  } catch (error) {
    console.error('Failed to fetch data:', error)
  }
}

onMounted(() => {
  Promise.all([fetchProvidersAndModels(), fetchRequests()])
})

const openDetailsModal = (request: RequestLog | Record<string, any>) => {
  currentRequest.value = request as RequestLog
  showDetailsModal.value = true
}

const closeDetailsModal = () => {
  showDetailsModal.value = false
  currentRequest.value = null
}

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchRequests()
}

const normalizeRequest = (request: RequestLog): RequestLog => ({
  ...request,
  id: String(request.id),
  providerId: request.providerId ? String(request.providerId) : '',
  modelId: request.modelId ? String(request.modelId) : '',
  apiKeyId: request.apiKeyId ? String(request.apiKeyId) : ''
})
</script>

<style scoped>
</style>
