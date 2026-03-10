<template>
  <Layout>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">API Keys</h1>
        <p class="text-gray-600 mt-1">Manage your API keys for accessing the proxy service</p>
      </div>
      <button 
        class="btn-primary"
        @click="openCreateModal"
      >
        <i class="fas fa-plus mr-2"></i>
        Create API Key
      </button>
    </div>
    
    <Card>
      <div class="mb-4">
        <input 
          type="text" 
          v-model="searchQuery"
          placeholder="Search API keys..."
          class="input-field"
        >
      </div>
      
      <Table :columns="columns" :items="filteredKeys">
        <template #isActive="{ value }">
          <span class="badge" :class="value ? 'badge-success' : 'badge-danger'">
            {{ value ? 'Active' : 'Inactive' }}
          </span>
        </template>
        
        <template #expiresAt="{ value }">
          <span class="text-sm">
            {{ value ? new Date(value).toLocaleDateString() : 'Never' }}
          </span>
        </template>
        
        <template #actions="{ item }">
          <div class="flex items-center space-x-2">
            <button 
              class="text-blue-500 hover:text-blue-700"
              @click="openEditModal(item)"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button 
              class="text-red-500 hover:text-red-700"
              @click="handleDelete(item.id)"
            >
              <i class="fas fa-trash"></i>
            </button>
          </div>
        </template>
      </Table>
      <PaginationControls
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="handlePageChange"
      />
    </Card>
    
    <!-- Create Modal -->
    <Modal 
      :show="showCreateModal"
      title="Create New API Key"
      @close="closeCreateModal"
      @submit="handleCreate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">Description</label>
          <input 
            type="text" 
            v-model="newKey.description"
            class="input-field"
            placeholder="Enter key description"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Expiration Date</label>
          <input 
            type="date" 
            v-model="newKey.expiresAt"
            class="input-field"
          >
        </div>
        
        <div>
          <label class="form-label">Max Usage (optional)</label>
          <input 
            type="number" 
            v-model="newKey.maxUsage"
            class="input-field"
            placeholder="Unlimited"
            min="1"
          >
        </div>
        
        <div>
          <label class="flex items-center">
            <input 
              type="checkbox" 
              v-model="newKey.isActive"
              class="mr-2"
            >
            Active
          </label>
        </div>
      </div>
    </Modal>
    
    <!-- Edit Modal -->
    <Modal 
      :show="showEditModal"
      title="Edit API Key"
      @close="closeEditModal"
      @submit="handleUpdate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">API Key</label>
          <input 
            type="text" 
            :value="editingKey.key"
            class="input-field"
            readonly
          >
        </div>
        
        <div>
          <label class="form-label">Description</label>
          <input 
            type="text" 
            v-model="editingKey.description"
            class="input-field"
            placeholder="Enter key description"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Expiration Date</label>
          <input 
            type="date" 
            v-model="editingKey.expiresAt"
            class="input-field"
          >
        </div>
        
        <div>
          <label class="form-label">Max Usage (optional)</label>
          <input 
            type="number" 
            v-model="editingKey.maxUsage"
            class="input-field"
            placeholder="Unlimited"
            min="1"
          >
        </div>
        
        <div>
          <label class="flex items-center">
            <input 
              type="checkbox" 
              v-model="editingKey.isActive"
              class="mr-2"
            >
            Active
          </label>
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
import { APIKey } from '@/types'
import { ref, computed, onMounted, reactive } from 'vue'

const columns = [
  { key: 'key', title: 'API Key' },
  { key: 'description', title: 'Description' },
  { key: 'isActive', title: 'Status' },
  { key: 'usageCount', title: 'Usage' },
  { key: 'maxUsage', title: 'Max Usage' },
  { key: 'expiresAt', title: 'Expires At' }
]

const apiKeys = ref<APIKey[]>([])
const searchQuery = ref('')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const newKey = ref<Partial<APIKey>>({
  description: '',
  isActive: true,
  maxUsage: null
})

const editingKey = ref<APIKey>({
  id: '',
  key: '',
  description: '',
  expiresAt: null,
  createdAt: '',
  updatedAt: '',
  isActive: true,
  usageCount: 0,
  maxUsage: null
})

const filteredKeys = computed(() => {
  return apiKeys.value.filter(key => 
    key.key.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    key.description.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const fetchAPIKeys = async () => {
  try {
    const data = await apiService.getAPIKeys({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    apiKeys.value = data.items.map(normalizeKey)
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to fetch API keys:', error)
  }
}

onMounted(() => {
  fetchAPIKeys()
})

const openCreateModal = () => {
  newKey.value = {
    description: '',
    isActive: true,
    maxUsage: null
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const openEditModal = (key: APIKey | Record<string, any>) => {
  editingKey.value = { ...(key as APIKey) }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const handleCreate = async () => {
  try {
    const key = await apiService.createAPIKey(newKey.value)
    pagination.page = 1
    await fetchAPIKeys()
    closeCreateModal()
  } catch (error) {
    console.error('Failed to create API key:', error)
  }
}

const handleUpdate = async () => {
  try {
    const updatedKey = await apiService.updateAPIKey(editingKey.value.id, editingKey.value)
    const index = apiKeys.value.findIndex(k => k.id === updatedKey.id)
    if (index !== -1) {
      apiKeys.value[index] = normalizeKey(updatedKey)
    }
    closeEditModal()
  } catch (error) {
    console.error('Failed to update API key:', error)
  }
}

const handleDelete = async (id: string | number) => {
  if (confirm('Are you sure you want to delete this API key?')) {
    try {
      const keyId = String(id)
      await apiService.deleteAPIKey(keyId)
      const isLastOnPage = apiKeys.value.length === 1 && pagination.page > 1
      if (isLastOnPage) {
        pagination.page -= 1
      }
      await fetchAPIKeys()
    } catch (error) {
      console.error('Failed to delete API key:', error)
    }
  }
}

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchAPIKeys()
}

const normalizeKey = (key: APIKey): APIKey => ({
  ...key,
  id: String(key.id)
})
</script>

<style scoped>
</style>
