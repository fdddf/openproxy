<template>
  <Layout>
    <div class="flex items-center justify-between mb-6">
      <div>
        <p class="text-xs uppercase tracking-[0.2em] text-indigo-500 font-semibold mb-1">Model Directory</p>
        <h1 class="text-2xl font-bold text-gray-900">Mappings & Routing</h1>
        <p class="text-gray-600 mt-1">Control how external provider models are exposed to your clients.</p>
      </div>
      <div class="flex space-x-3">
        <select 
          v-model="providerFilter"
          class="input-field w-40"
          @change="handleProviderChange"
        >
          <option value="">All Providers</option>
          <option v-for="provider in providers" :key="provider.id" :value="provider.id">
            {{ provider.name }}
          </option>
        </select>
        <button 
          class="btn-primary"
          @click="openCreateModal"
        >
          <i class="fas fa-plus mr-2"></i>
          Add Mapping
        </button>
      </div>
    </div>
    
    <Card>
      <Table :columns="columns" :items="filteredModels">
        <template #providerName="{ item }">
          <span class="text-sm font-medium">
            {{ item.providerName }}
          </span>
        </template>
        
        <template #isActive="{ value }">
          <span class="badge" :class="value ? 'badge-success' : 'badge-danger'">
            {{ value ? 'Active' : 'Inactive' }}
          </span>
        </template>
        
        <template #usageCount="{ value }">
          <span class="text-sm">
            {{ value }} requests
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
              class="text-green-500 hover:text-green-700"
              @click="toggleModelStatus(item)"
            >
              <i :class="item.isActive ? 'fas fa-toggle-on' : 'fas fa-toggle-off'"></i>
            </button>
            <button 
              class="text-red-500 hover:text-red-700"
              @click="removeModel(item)"
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
      title="Add Model Mapping"
      @close="closeCreateModal"
      @submit="handleCreate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">Provider</label>
          <select 
            v-model="newModel.providerId"
            class="input-field"
            required
          >
            <option value="">Select provider</option>
            <option v-for="provider in providers" :key="provider.id" :value="provider.id">
              {{ provider.name }}
            </option>
          </select>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="form-label">Original Name</label>
            <input 
              type="text" 
              v-model="newModel.name"
              class="input-field"
              placeholder="Upstream model id (e.g., gpt-4o-mini)"
              required
            >
          </div>
          <div>
            <label class="form-label">Client Alias</label>
            <input 
              type="text" 
              v-model="newModel.mappedName"
              class="input-field"
              placeholder="Public name exposed to clients"
              required
            >
          </div>
        </div>

        <div class="flex items-center">
          <input 
            id="new-model-active"
            type="checkbox" 
            v-model="newModel.isActive"
            class="mr-2"
          >
          <label for="new-model-active" class="text-sm text-gray-700">Mark as active</label>
        </div>
      </div>
    </Modal>

    <!-- Edit Modal -->
    <Modal 
      :show="showEditModal"
      title="Edit Model Mapping"
      @close="closeEditModal"
      @submit="handleUpdate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">Provider</label>
          <input 
            type="text" 
            :value="getProviderName(editingModel.providerId)"
            class="input-field"
            readonly
          >
        </div>
        
        <div>
          <label class="form-label">Original Name</label>
          <input 
            type="text" 
            v-model="editingModel.name"
            class="input-field"
            placeholder="Upstream model id"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Mapped Name</label>
          <input 
            type="text" 
            v-model="editingModel.mappedName"
            class="input-field"
            placeholder="e.g., gpt-4, gpt-3.5-turbo"
            required
          >
          <p class="text-xs text-gray-500 mt-1">
            This name will be used in API requests
          </p>
        </div>
        
        <div>
          <label class="flex items-center">
            <input 
              type="checkbox" 
              v-model="editingModel.isActive"
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
import { Model, Provider } from '@/types'
import { ref, computed, onMounted, reactive } from 'vue'

const columns = [
  { key: 'providerName', title: 'Provider' },
  { key: 'name', title: 'Original Name' },
  { key: 'mappedName', title: 'Mapped Name' },
  { key: 'isActive', title: 'Status' },
  { key: 'usageCount', title: 'Usage' }
]

const models = ref<Model[]>([])
const providers = ref<Provider[]>([])
const providerFilter = ref('')
const showEditModal = ref(false)
const showCreateModal = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const editingModel = ref<Model>({
  id: 0,
  providerId: 0,
  name: '',
  mappedName: '',
  isActive: true,
  createdAt: '',
  updatedAt: '',
  usageCount: 0
})

const newModel = ref<Model>({
  id: 0,
  providerId: 0,
  name: '',
  mappedName: '',
  isActive: true,
  createdAt: '',
  updatedAt: '',
  usageCount: 0
})

const filteredModels = computed(() => {
  // Add providerName to each model for proper sorting
  return models.value.map(model => ({
    ...model,
    providerName: getProviderName(model.providerId)
  }))
})

const fetchModels = async () => {
  try {
    const data = await apiService.getModels({
      page: pagination.page,
      pageSize: pagination.pageSize,
      providerId: providerFilter.value ? Number(providerFilter.value) : undefined
    })
    models.value = data.items.map(normalizeModel)
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to fetch models:', error)
  }
}

const fetchProviders = async () => {
  try {
    const data = await apiService.getAllProviders()
    providers.value = data.map(normalizeProvider)
  } catch (error) {
    console.error('Failed to fetch providers:', error)
  }
}

const getProviderName = (providerId: number) => {
  const provider = providers.value.find(p => p.id === providerId)
  return provider ? provider.name : 'Unknown'
}

const handleProviderChange = () => {
  pagination.page = 1
  fetchModels()
}

onMounted(async () => {
  await Promise.all([
    fetchModels(),
    fetchProviders()
  ])
})

const openEditModal = (model: Model | Record<string, any>) => {
  editingModel.value = { ...(model as Model) }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const openCreateModal = () => {
  newModel.value = {
    id: 0,
    providerId: 0,
    name: '',
    mappedName: '',
    isActive: true,
    createdAt: '',
    updatedAt: '',
    usageCount: 0
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const handleUpdate = async () => {
  try {
    const payload = {
      providerId: Number(editingModel.value.providerId),
      name: editingModel.value.name,
      mappedName: editingModel.value.mappedName,
      isActive: editingModel.value.isActive
    }
    await apiService.updateModel(editingModel.value.id, payload)
    await fetchModels()
    closeEditModal()
  } catch (error) {
    console.error('Failed to update model:', error)
  }
}

const toggleModelStatus = async (model: Model | Record<string, any>) => {
  try {
    const typedModel = model as Model
    await apiService.updateModel(typedModel.id, {
      ...typedModel,
      isActive: !typedModel.isActive
    })
    await fetchModels()
  } catch (error) {
    console.error('Failed to toggle model status:', error)
  }
}

const handleCreate = async () => {
  try {
    const payload = {
      providerId: Number(newModel.value.providerId),
      name: newModel.value.name,
      mappedName: newModel.value.mappedName,
      isActive: newModel.value.isActive
    }
    await apiService.createModel(payload)
    pagination.page = 1
    await fetchModels()
    closeCreateModal()
  } catch (error) {
    console.error('Failed to create model:', error)
  }
}

const removeModel = async (model: Model | Record<string, any>) => {
  const typed = model as Model
  if (!confirm(`Delete mapping "${typed.mappedName}"? This cannot be undone.`)) return
  try {
    await apiService.deleteModel(typed.id)
    const isLastOnPage = models.value.length === 1 && pagination.page > 1
    if (isLastOnPage) {
      pagination.page -= 1
    }
    await fetchModels()
  } catch (error) {
    console.error('Failed to delete model:', error)
  }
}

const normalizeModel = (model: Model): Model => ({
  ...model,
  id: Number(model.id),
  providerId: Number(model.providerId)
})

const normalizeProvider = (provider: Provider): Provider => ({
  ...provider,
  models: provider.models?.map(normalizeModel) || []
})

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchModels()
}
</script>

<style scoped>
</style>
