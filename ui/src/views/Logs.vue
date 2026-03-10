<template>
  <Layout>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">System Logs</h1>
        <p class="text-gray-600 mt-1">View system activity and error logs</p>
      </div>
      <div class="flex space-x-3">
        <select 
          v-model="levelFilter"
          class="input-field"
        >
          <option value="">All Levels</option>
          <option value="debug">Debug</option>
          <option value="info">Info</option>
          <option value="warn">Warning</option>
          <option value="error">Error</option>
        </select>
        <input 
          type="text" 
          v-model="searchQuery"
          placeholder="Search logs..."
          class="input-field"
        >
      </div>
    </div>
    
    <Card>
      <Table :columns="columns" :items="filteredLogs">
        <template #level="{ value }">
          <span class="badge" :class="getLevelClass(value)">
            {{ value.toUpperCase() }}
          </span>
        </template>
        
        <template #timestamp="{ value }">
          <span class="text-sm">
            {{ new Date(value).toLocaleString() }}
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
      title="Log Details"
      @close="closeDetailsModal"
    >
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="font-medium">
              <span class="badge" :class="currentLog ? getLevelClass(currentLog.level) : ''">
                {{ currentLog ? currentLog.level.toUpperCase() : '' }}
              </span>
            </h4>
          </div>
          <p class="text-sm text-gray-500">
            {{ currentLog ? new Date(currentLog.timestamp).toLocaleString() : '' }}
          </p>
        </div>
        
        <div>
          <h4 class="font-medium text-sm text-gray-600">Message</h4>
          <p class="text-sm mt-1">{{ currentLog?.message }}</p>
        </div>
        
        <div v-if="currentLog?.details">
          <h4 class="font-medium text-sm text-gray-600">Details</h4>
          <pre class="text-xs bg-gray-50 p-3 rounded mt-1 overflow-x-auto">
{{ currentLog?.details }}
          </pre>
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
import { SystemLog } from '@/types'
import { ref, computed, onMounted, reactive } from 'vue'

const columns = [
  { key: 'level', title: 'Level' },
  { key: 'message', title: 'Message' },
  { key: 'timestamp', title: 'Time' }
]

const logs = ref<SystemLog[]>([])
const levelFilter = ref('')
const searchQuery = ref('')
const showDetailsModal = ref(false)
const currentLog = ref<SystemLog | null>(null)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    const matchesLevel = !levelFilter.value || log.level === levelFilter.value
    const matchesSearch = !searchQuery.value || 
      log.message.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (log.details && log.details.toLowerCase().includes(searchQuery.value.toLowerCase()))
    return matchesLevel && matchesSearch
  })
})

const getLevelClass = (level: string | undefined) => {
  switch (level) {
    case 'debug':
      return 'badge-primary'
    case 'info':
      return 'badge-success'
    case 'warn':
      return 'badge-warning'
    case 'error':
      return 'badge-danger'
    default:
      return ''
  }
}

const fetchLogs = async () => {
  try {
    const data = await apiService.getLogs({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    logs.value = data.items.map(normalizeLog)
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to fetch logs:', error)
  }
}

onMounted(() => {
  fetchLogs()
})

const openDetailsModal = (log: SystemLog | Record<string, any>) => {
  currentLog.value = log as SystemLog
  showDetailsModal.value = true
}

const closeDetailsModal = () => {
  showDetailsModal.value = false
  currentLog.value = null
}

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchLogs()
}

const normalizeLog = (log: SystemLog): SystemLog => ({
  ...log,
  id: String(log.id)
})
</script>

<style scoped>
</style>
