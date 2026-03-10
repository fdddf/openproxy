<template>
  <Layout>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 bg-gradient-to-r from-indigo-600 to-cyan-500 bg-clip-text text-transparent">Providers</h1>
        <p class="text-slate-500 mt-1.5">Manage AI service providers</p>
      </div>
      <button 
        class="btn-primary flex items-center gap-2"
        @click="openCreateModal"
      >
        <i class="fas fa-plus"></i>
        <span>Add Provider</span>
      </button>
    </div>
    
    <!-- Providers Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="provider in providers" :key="provider.id" class="card group">
        <div class="card-body p-6 relative">
          <div class="flex items-start justify-between mb-5">
            <div>
              <h3 class="text-xl font-bold text-gray-900 mb-1">{{ provider.name }}</h3>
              <p class="text-sm font-medium text-slate-500 uppercase tracking-wide">{{ provider.type }}</p>
            </div>
            <span class="badge px-3 py-1.5 text-xs font-semibold" :class="provider.isActive ? 'badge-success' : 'badge-danger'">
              {{ provider.isActive ? 'Active' : 'Inactive' }}
            </span>
          </div>
          
          <div class="space-y-3 mb-0">
            <div class="flex items-center justify-between py-2 border-b border-slate-50">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">API Key</span>
              <span class="text-xs font-mono bg-slate-100 px-2 py-1 rounded">{{ maskAPIKey(provider.key) }}</span>
            </div>
            <div class="flex items-center justify-between py-2 border-b border-slate-50">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Health Check</span>
              <span class="text-xs font-semibold" :class="provider.healthCheckEnabled ? 'text-emerald-600' : 'text-slate-400'">
                {{ provider.healthCheckEnabled ? (provider.healthCheckStatus || 'Pending') : 'Disabled' }}
              </span>
            </div>
            <div class="flex items-center justify-between py-2 border-b border-slate-50" v-if="provider.healthCheckChecked">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Last Checked</span>
              <span class="text-xs text-slate-600">{{ formatDate(provider.healthCheckChecked) }}</span>
            </div>
            <div class="flex items-center justify-between py-2 border-b border-slate-50">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Base URL</span>
              <span class="text-xs text-slate-600 truncate max-w-[200px]">{{ provider.baseUrl }}</span>
            </div>
            <div class="flex items-center justify-between py-2 border-b border-slate-50" v-if="provider.proxyUrl">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Proxy URL</span>
              <span class="text-xs text-slate-600">{{ provider.proxyUrl }}</span>
            </div>
            <div 
              v-if="provider.type === 'codex'" 
              class="flex items-center justify-between py-2 border-b border-slate-50"
            >
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">ChatGPT Account ID</span>
              <span class="text-xs font-mono text-slate-600">{{ provider.accountId || '-' }}</span>
            </div>
            <div 
              v-if="provider.type === 'codex' || provider.type === 'antigravity'" 
              class="flex items-center justify-between py-2 border-b border-slate-50"
            >
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Token Expiry</span>
              <span class="text-xs font-semibold" :class="getTokenExpiryClass(provider.tokenExpiry)">
                {{ provider.tokenExpiry && provider.tokenExpiry.trim() ? formatDate(provider.tokenExpiry) : 'N/A' }}
              </span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-xs font-medium text-slate-500 uppercase tracking-wide">Models</span>
              <span class="text-xs font-semibold bg-indigo-100 text-indigo-700 px-2 py-1 rounded-full">{{ provider.models.length }}</span>
            </div>
          </div>
          
          <div class="flex flex-wrap gap-2.5 mt-5 pt-4 border-t border-slate-100">
            <button 
              class="btn-sm btn-primary flex items-center gap-1.5"
              @click="openModelsModal(provider)"
            >
              <i class="fas fa-cogs text-xs"></i>
              <span>Models</span>
            </button>
            <button 
              v-if="provider.type === 'codex'"
              class="btn-sm btn-primary flex items-center gap-1.5"
              @click="initiateOAuth2Flow(provider.id)"
              title="Connect to Codex"
            >
              <i class="fas fa-plug text-xs"></i>
              <span>Connect</span>
            </button>
            <!-- Refresh Token Button for OAuth providers -->
            <button 
              v-if="provider.type === 'codex' || provider.type === 'antigravity'"
              class="btn-sm btn-warning flex items-center gap-1.5"
              @click="refreshToken(provider)"
              title="Refresh Token"
            >
              <i class="fas fa-sync-alt text-xs"></i>
              <span>Refresh</span>
            </button>
            <button 
              class="btn-sm btn-secondary flex items-center justify-center w-9 h-9 p-0"
              @click="openEditModal(provider)"
              title="Edit"
            >
              <i class="fas fa-edit text-xs"></i>
            </button>
            <button 
              class="btn-sm flex items-center justify-center w-9 h-9 p-0"
              :class="provider.isActive ? 'btn-danger' : 'btn-success'"
              @click="toggleProviderStatus(provider)"
              title="Toggle Status"
            >
              <i :class="provider.isActive ? 'fas fa-toggle-off' : 'fas fa-toggle-on'"></i>
            </button>
            <button 
              class="btn-sm btn-danger flex items-center justify-center w-9 h-9 p-0"
              @click="confirmDelete(provider)"
              title="Delete"
            >
              <i class="fas fa-trash text-xs"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <PaginationControls
      :page="pagination.page"
      :page-size="pagination.pageSize"
      :total="pagination.total"
      @update:page="handlePageChange"
    />
    
    <!-- Create Modal -->
    <Modal 
      :show="showCreateModal"
      title="Add New Provider"
      @close="closeCreateModal"
      @submit="handleCreate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">Name</label>
          <input 
            type="text" 
            v-model="newProvider.name"
            class="input-field"
            placeholder="Enter provider name"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Platform</label>
          <select 
            v-model="newProvider.type"
            class="input-field"
            required
            @change="handlePlatformChange('new')"
          >
            <option value="">Select platform</option>
            <option v-for="platform in platforms" :key="platform" :value="platform">{{ platform }}</option>
          </select>
        </div>
        
        <div v-if="newProvider.type !== 'codex' && newProvider.type !== 'antigravity'">
          <label class="form-label">API Key</label>
          <input 
            type="text" 
            v-model="newProvider.key"
            class="input-field"
            placeholder="Enter API key"
            required
          >
        </div>
        
        <div v-if="newProvider.type === 'codex' || newProvider.type === 'antigravity'">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="form-label">Client ID</label>
              <input 
                type="text" 
                v-model="newProvider.clientId"
                class="input-field"
                placeholder="Enter client ID"
                required
              >
            </div>
            <div>
              <label class="form-label">Client Secret</label>
              <input 
                type="password" 
                v-model="newProvider.clientSecret"
                class="input-field"
                placeholder="Enter client secret"
                required
              >
            </div>
          </div>
          
          <div class="mt-4">
            <label class="form-label">Authorization URL</label>
            <input 
              type="text" 
              v-model="newProvider.authUrl"
              class="input-field"
              placeholder="Enter authorization URL"
              required
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Token URL</label>
            <input 
              type="text" 
              v-model="newProvider.tokenUrl"
              class="input-field"
              placeholder="Enter token URL"
              required
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Redirect URL</label>
            <input 
              type="text" 
              v-model="newProvider.redirectUrl"
              class="input-field"
              placeholder="Enter redirect URL"
              required
              :disabled="true"
              title="This is automatically set to the required value for Codex"
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Scopes (comma separated)</label>
            <input 
              type="text" 
              v-model="newProvider.scopes"
              class="input-field"
              placeholder="Enter scopes, e.g., read write"
            >
          </div>

          <div class="mt-4">
            <label class="form-label">ChatGPT Account ID</label>
            <input 
              type="text" 
              v-model="newProvider.accountId"
              class="input-field"
              placeholder="Will be set automatically after OAuth"
              readonly
              title="Codex requires chatgpt_account_id, it will be filled from the token exchange"
            >
          </div>
        </div>
        
        <div>
          <label class="form-label">Base URL</label>
          <input 
            type="text" 
            v-model="newProvider.baseUrl"
            class="input-field"
            placeholder="https://api.provider.com"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Proxy URL (optional)</label>
          <input 
            type="text" 
            v-model="newProvider.proxyUrl"
            class="input-field"
            placeholder="/api/proxy/provider"
          >
        </div>

        <div class="flex items-center space-x-2">
          <input
            type="checkbox"
            v-model="newProvider.healthCheckEnabled"
          >
          <label class="form-label !mb-0">Enable Health Check</label>
        </div>

        <div>
          <label class="flex items-center">
            <input 
              type="checkbox" 
              v-model="newProvider.isActive"
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
      title="Edit Provider"
      @close="closeEditModal"
      @submit="handleUpdate"
    >
      <div class="space-y-4">
        <div>
          <label class="form-label">Name</label>
          <input 
            type="text" 
            v-model="editingProvider.name"
            class="input-field"
            placeholder="Enter provider name"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Platform</label>
          <select 
            v-model="editingProvider.type"
            class="input-field"
            required
            @change="handlePlatformChange('edit')"
          >
            <option value="">Select platform</option>
            <option v-for="platform in platforms" :key="platform" :value="platform">{{ platform }}</option>
          </select>
        </div>
        
        <div v-if="editingProvider.type !== 'codex' && editingProvider.type !== 'antigravity'">
          <label class="form-label">API Key</label>
          <input 
            type="text" 
            v-model="editingProvider.key"
            class="input-field"
            placeholder="Enter API key"
            required
          >
        </div>
        
        <div v-if="editingProvider.type === 'codex' || editingProvider.type === 'antigravity'">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="form-label">Client ID</label>
              <input 
                type="text" 
                v-model="editingProvider.clientId"
                class="input-field"
                placeholder="Enter client ID"
                required
              >
            </div>
            <div>
              <label class="form-label">Client Secret</label>
              <input 
                type="password" 
                v-model="editingProvider.clientSecret"
                class="input-field"
                placeholder="Enter client secret"
                required
              >
            </div>
          </div>
          
          <div class="mt-4">
            <label class="form-label">Authorization URL</label>
            <input 
              type="text" 
              v-model="editingProvider.authUrl"
              class="input-field"
              placeholder="Enter authorization URL"
              required
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Token URL</label>
            <input 
              type="text" 
              v-model="editingProvider.tokenUrl"
              class="input-field"
              placeholder="Enter token URL"
              required
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Redirect URL</label>
            <input 
              type="text" 
              v-model="editingProvider.redirectUrl"
              class="input-field"
              placeholder="Enter redirect URL"
              required
              :disabled="true"
              title="This is automatically set to the required value for Codex"
            >
          </div>
          
          <div class="mt-4">
            <label class="form-label">Scopes (comma separated)</label>
            <input 
              type="text" 
              v-model="editingProvider.scopes"
              class="input-field"
              placeholder="Enter scopes, e.g., read write"
            >
          </div>

          <div class="mt-4">
            <label class="form-label">ChatGPT Account ID</label>
            <input 
              type="text" 
              v-model="editingProvider.accountId"
              class="input-field"
              placeholder="Will be set automatically after OAuth"
              readonly
              title="Codex requires chatgpt_account_id, it will be filled from the token exchange"
            >
          </div>
        </div>
        
        <div>
          <label class="form-label">Base URL</label>
          <input 
            type="text" 
            v-model="editingProvider.baseUrl"
            class="input-field"
            placeholder="https://api.provider.com"
            required
          >
        </div>
        
        <div>
          <label class="form-label">Proxy URL (optional)</label>
          <input 
            type="text" 
            v-model="editingProvider.proxyUrl"
            class="input-field"
            placeholder="/api/proxy/provider"
          >
        </div>

        <div class="flex items-center space-x-2">
          <input
            type="checkbox"
            v-model="editingProvider.healthCheckEnabled"
          >
          <label class="form-label !mb-0">Enable Health Check</label>
        </div>

        <div>
          <label class="flex items-center">
            <input 
              type="checkbox" 
              v-model="editingProvider.isActive"
              class="mr-2"
            >
            Active
          </label>
        </div>
      </div>
    </Modal>
    
    <!-- Models Modal -->
    <Modal 
      :show="showModelsModal"
      title="Manage Models"
      @close="closeModelsModal"
      width="800px"
    >
      <div class="space-y-4">
        <div class="flex justify-between items-center">
          <div>
            <p class="text-xs uppercase tracking-[0.2em] text-indigo-500 font-semibold">Mappings</p>
            <h4 class="font-medium text-gray-900">Models for {{ currentProvider?.name }}</h4>
          </div>
          <span class="text-xs text-gray-500">Provider ID: {{ currentProvider?.id }}</span>
        </div>

        <div class="p-4 bg-gray-50 border border-dashed border-gray-200 rounded-lg space-y-3">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="form-label">Original Name</label>
              <input 
                v-model="providerModelForm.name"
                type="text"
                class="input-field"
                placeholder="Provider model id"
              >
            </div>
            <div>
              <label class="form-label">Mapped Name</label>
              <input 
                v-model="providerModelForm.mappedName"
                type="text"
                class="input-field"
                placeholder="Client alias"
              >
            </div>
            <div class="flex items-center space-x-2 pt-6">
              <input 
                id="provider-model-active"
                v-model="providerModelForm.isActive"
                type="checkbox"
                class="mr-2"
              >
              <label for="provider-model-active" class="text-sm text-gray-700">Active</label>
            </div>
          </div>
          <div class="flex justify-end">
            <button 
              class="btn-primary"
              @click="addModelMapping"
            >
              <i class="fas fa-plus mr-2"></i>
              Add Mapping
            </button>
          </div>
        </div>
        
        <Table :columns="modelColumns" :items="currentProvider?.models || []">
          <template #isActive="{ value }">
            <span class="badge" :class="value ? 'badge-success' : 'badge-danger'">
              {{ value ? 'Active' : 'Inactive' }}
            </span>
          </template>
          
          <template #actions="{ item }">
            <div class="flex items-center space-x-2">
              <button 
                class="text-blue-500 hover:text-blue-700"
                @click="openModelEditor(item)"
                title="Edit Model"
              >
                <i class="fas fa-edit"></i>
              </button>
              <button 
                class="text-green-500 hover:text-green-700"
                @click="toggleModelStatus(item)"
                title="Toggle Model Status"
              >
                <i :class="item.isActive ? 'fas fa-toggle-on' : 'fas fa-toggle-off'"></i>
              </button>
              <button 
                class="text-red-500 hover:text-red-700"
                @click="deleteModel(item)"
                title="Delete Model"
              >
                <i class="fas fa-trash"></i>
              </button>
            </div>
          </template>
        </Table>
      </div>
    </Modal>

    <Modal 
      :show="showModelEditModal"
      title="Edit Mapping"
      @close="closeModelEditModal"
      @submit="saveModelEdit"
    >
      <div class="space-y-4" v-if="editingModel">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div>
            <label class="form-label">Original Name</label>
            <input 
              v-model="editingModel.name"
              class="input-field"
              placeholder="Provider model id"
            >
          </div>
          <div>
            <label class="form-label">Mapped Name</label>
            <input 
              v-model="editingModel.mappedName"
              class="input-field"
              placeholder="Client alias"
            >
          </div>
        </div>
        <label class="flex items-center">
          <input 
            v-model="editingModel.isActive"
            type="checkbox"
            class="mr-2"
          >
          Active
        </label>
      </div>
    </Modal>

    <!-- OAuth Manual Flow Modal -->
    <Modal 
      :show="showOAuthModal"
      title="Connect Codex Manually"
      :submit-text="oauthSubmitting ? 'Submitting...' : 'Submit Code'"
      :close-on-submit="false"
      @close="closeOAuthModal"
      @submit="submitOAuthCode"
    >
      <div class="space-y-4">
        <div class="p-4 bg-blue-50 border border-blue-100 rounded-lg text-sm text-blue-800">
          <p>1) 点击下方链接登录 Codex 并完成授权。</p>
          <p>2) 复制跳转后的 <code>http://localhost:1455/auth/callback</code> 地址（或其中的 code）。</p>
          <p>3) 粘贴到下方输入框，点击提交完成剩余流程。</p>
          <p class="mt-2 text-xs text-blue-700">State：{{ oauthFlow.state || 'N/A' }}</p>
        </div>

        <div>
          <label class="form-label">Authorization URL</label>
          <div class="flex space-x-2">
            <input 
              class="input-field flex-1"
              :value="oauthFlow.authUrl"
              readonly
            >
            <button 
              class="btn-secondary"
              type="button"
              @click="openAuthUrl"
            >
              打开
            </button>
            <button 
              class="btn-primary"
              type="button"
              @click="copyAuthUrl"
            >
              复制
            </button>
          </div>
        </div>

        <div>
          <label class="form-label">Callback URL / Code</label>
          <input 
            class="input-field"
            v-model="manualCallbackInput"
            placeholder="粘贴 http://localhost:1455/auth/callback?... 或仅填写 code"
          >
          <p class="text-xs text-gray-500 mt-1">系统会自动使用当前会话的 state 和 code_verifier。</p>
        </div>

        <p v-if="oauthError" class="text-sm text-red-600">{{ oauthError }}</p>
        <p v-if="oauthSuccess" class="text-sm text-green-600">{{ oauthSuccess }}</p>
      </div>
    </Modal>
    
    <!-- Confirm Delete Modal -->
    <Modal 
      :show="showDeleteModal"
      title="Confirm Delete"
      @close="closeDeleteModal"
      @submit="handleDelete"
    >
      <div class="text-center">
        <i class="fas fa-exclamation-triangle text-4xl text-warning mb-4"></i>
        <h3 class="text-lg font-semibold mb-2">Are you sure?</h3>
        <p class="text-gray-600 mb-4">
          You are about to delete <strong>{{ deletingProvider?.name }}</strong>. This action cannot be undone.
        </p>
        <div class="flex space-x-3 justify-center">
          <button 
            class="btn-secondary"
            @click="closeDeleteModal"
          >
            Cancel
          </button>
          <button 
            class="btn-danger"
            @click="handleDelete"
          >
            Delete
          </button>
        </div>
      </div>
    </Modal>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import Table from '@/components/Table.vue'
import Modal from '@/components/Modal.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import { apiService } from '@/api'
import { Provider, Model } from '@/types'
import { ref, onMounted, reactive } from 'vue'

const modelColumns = [
  { key: 'name', title: 'Original Name' },
  { key: 'mappedName', title: 'Mapped Name' },
  { key: 'isActive', title: 'Status' },
  { key: 'usageCount', title: 'Usage' }
]

const providers = ref<Provider[]>([])
const platforms = ref<string[]>([])
const pagination = reactive({
  page: 1,
  pageSize: 6,
  total: 0
})
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showModelsModal = ref(false)
const showDeleteModal = ref(false)
const currentProvider = ref<Provider | null>(null)
const deletingProvider = ref<Provider | null>(null)
const providerModelForm = ref({
  name: '',
  mappedName: '',
  isActive: true
})
const showModelEditModal = ref(false)
const editingModel = ref<Model | null>(null)
const showOAuthModal = ref(false)
const oauthFlow = ref({
  providerId: 0,
  sessionId: '',
  state: '',
  codeVerifier: '',
  authUrl: ''
})
const manualCallbackInput = ref('')
const oauthError = ref('')
const oauthSuccess = ref('')
const oauthSubmitting = ref(false)

const formatDate = (value: string | null | undefined) => {
  if (!value) return '-'
  const d = new Date(value)
  return d.toLocaleString()
}

const getTokenExpiryClass = (value: string | null | undefined) => {
  if (!value) return 'text-slate-400'
  const expiry = new Date(value)
  const now = new Date()
  const hoursUntilExpiry = (expiry.getTime() - now.getTime()) / (1000 * 60 * 60)

  if (hoursUntilExpiry < 0) return 'text-red-600'
  if (hoursUntilExpiry < 1) return 'text-orange-600'
  if (hoursUntilExpiry < 24) return 'text-amber-600'
  return 'text-emerald-600'
}

const newProvider = ref<Partial<Provider>>({
  name: '',
  type: '',
  key: '',
  baseUrl: '',
  proxyUrl: '',
  isActive: true,
  healthCheckEnabled: false,
  clientId: '',
  clientSecret: '',
  authUrl: '',
  tokenUrl: '',
  redirectUrl: '',
  scopes: '',
  accountId: ''
})

const editingProvider = ref<Provider>({
  id: 0,
  name: '',
  type: '',
  key: '',
  baseUrl: '',
  proxyUrl: null,
  isActive: true,
  healthCheckEnabled: false,
  createdAt: '',
  updatedAt: '',
  models: [],
  clientId: '',
  clientSecret: '',
  authUrl: '',
  tokenUrl: '',
  redirectUrl: '',
  scopes: '',
  accountId: ''
})

const handlePlatformChange = (formType: string) => {
  const isCodex = (type: string) => type === 'codex'
  const isAntigravity = (type: string) => type === 'antigravity'

  if (formType === 'new') {
    if (isCodex(newProvider.value.type)) {
      newProvider.value.key = ''
      newProvider.value.accountId = ''
      newProvider.value.authUrl = ''
      newProvider.value.tokenUrl = ''
      newProvider.value.redirectUrl = ''
      newProvider.value.baseUrl = ''
      newProvider.value.scopes = ''
    } else if (isAntigravity(newProvider.value.type)) {
      newProvider.value.key = ''
      newProvider.value.accountId = ''
      newProvider.value.authUrl = ''
      newProvider.value.tokenUrl = ''
      newProvider.value.redirectUrl = ''
      newProvider.value.baseUrl = ''
      newProvider.value.scopes = ''
    } else {
      newProvider.value.clientId = ''
      newProvider.value.clientSecret = ''
      newProvider.value.authUrl = ''
      newProvider.value.tokenUrl = ''
      newProvider.value.redirectUrl = ''
      newProvider.value.scopes = ''
      newProvider.value.baseUrl = ''
      newProvider.value.accountId = ''
    }
  } else if (formType === 'edit') {
    if (isCodex(editingProvider.value.type)) {
      editingProvider.value.key = ''
      editingProvider.value.authUrl = ''
      editingProvider.value.tokenUrl = ''
      editingProvider.value.redirectUrl = ''
      editingProvider.value.baseUrl = ''
      editingProvider.value.scopes = ''
      if (!editingProvider.value.accountId) {
        editingProvider.value.accountId = ''
      }
    } else if (isAntigravity(editingProvider.value.type)) {
      editingProvider.value.key = ''
      editingProvider.value.authUrl = ''
      editingProvider.value.tokenUrl = ''
      editingProvider.value.redirectUrl = ''
      editingProvider.value.baseUrl = ''
      editingProvider.value.scopes = ''
      editingProvider.value.accountId = ''
    } else {
      editingProvider.value.clientId = ''
      editingProvider.value.clientSecret = ''
      editingProvider.value.authUrl = ''
      editingProvider.value.tokenUrl = ''
      editingProvider.value.redirectUrl = ''
      editingProvider.value.scopes = ''
      editingProvider.value.accountId = ''
    }
  }
}

// Function to generate a random string for state and PKCE
function generateRandomString(length: number): string {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
}

// Function to generate code verifier for PKCE
function generateCodeVerifier(): string {
  return generateRandomString(64);
}

// Function to generate code challenge from code verifier (S256 method)
async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  const base64Digest = btoa(String.fromCharCode(...new Uint8Array(digest)));
  return base64Digest
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

const resetOAuthState = () => {
  oauthFlow.value = {
    providerId: 0,
    sessionId: '',
    state: '',
    codeVerifier: '',
    authUrl: ''
  }
  manualCallbackInput.value = ''
  oauthError.value = ''
  oauthSuccess.value = ''
  oauthSubmitting.value = false
}

const parseCallbackInput = (rawInput: string): { code: string, state: string } => {
  const trimmed = rawInput.trim()
  if (!trimmed) {
    return { code: '', state: '' }
  }

  try {
    const parsedUrl = new URL(trimmed)
    return {
      code: parsedUrl.searchParams.get('code') || '',
      state: parsedUrl.searchParams.get('state') || ''
    }
  } catch {
    // Not a full URL, try parsing as query string
    const params = new URLSearchParams(trimmed.startsWith('?') ? trimmed.slice(1) : trimmed)
    return {
      code: params.get('code') || trimmed,
      state: params.get('state') || ''
    }
  }
}

// Function to initiate the OAuth2 flow for a provider with PKCE
const initiateOAuth2Flow = async (providerId: number) => {
  try {
    resetOAuthState()

    // Get the provider details to construct the OAuth2 URL
    const provider = providers.value.find(p => p.id === providerId);
    if (!provider) {
      oauthError.value = 'Provider not found'
      showOAuthModal.value = true
      return
    }
    
    // Generate state and PKCE parameters
    const state = generateRandomString(32);
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    
    // Create OAuth session on the backend to store the state and code verifier
    const sessionData = {
      providerId: providerId,
      state: state,
      codeVerifier: codeVerifier,
      codeChallenge: codeChallenge,
      codeChallengeMethod: 'S256'
    };
    
    const response = await fetch('/api/oauth2/session', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}` // Assuming there's a token for auth
      },
      body: JSON.stringify(sessionData)
    });
    
    const sessionResult = await response.json();
    if (!response.ok || !sessionResult.sessionId) {
      throw new Error(sessionResult?.message || 'Failed to create OAuth2 session');
    }
    
    // Store the session ID and verifier locally so we can finish manually
    sessionStorage.setItem(`oauth_session_id`, sessionResult.sessionId);
    sessionStorage.setItem(`oauth_code_verifier_${sessionResult.sessionId}`, codeVerifier);
    sessionStorage.setItem(`oauth_state_${sessionResult.sessionId}`, state);
    
    // Construct the OAuth2 authorization URL
    const redirectUri = `http://localhost:1455/auth/callback`;
    const authUrl = new URL(provider.authUrl!);
    authUrl.searchParams.set('id_token_add_organizations', 'false');
    if (!authUrl.searchParams.get('codex_cli_simplified_flow')) {
      authUrl.searchParams.set('codex_cli_simplified_flow', 'true');
    }
    if (!authUrl.searchParams.get('originator')) {
      authUrl.searchParams.set('originator', 'codex_vscode');
    }
    
    authUrl.searchParams.set('client_id', provider.clientId || '');
    authUrl.searchParams.set('redirect_uri', redirectUri);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('state', state);
    authUrl.searchParams.set('code_challenge', codeChallenge);
    authUrl.searchParams.set('code_challenge_method', 'S256');
    
    // Normalize scope to avoid "+" encoding surprises
    const scopeParam = (provider.scopes || '').replace(/\+/g, ' ').trim();
    if (scopeParam) {
      authUrl.searchParams.set('scope', scopeParam);
    }
    const authUrlString = authUrl.toString().replace(/\+/g, '%20');

    oauthFlow.value = {
      providerId: providerId,
      sessionId: sessionResult.sessionId,
      state: state,
      codeVerifier: codeVerifier,
      authUrl: authUrlString
    }
    oauthError.value = ''
    oauthSuccess.value = ''
    showOAuthModal.value = true
  } catch (error) {
    console.error('Failed to initiate OAuth2 flow:', error);
    oauthError.value = (error as Error)?.message || 'Failed to initiate OAuth2 flow'
    showOAuthModal.value = true
  }
}

const copyAuthUrl = async () => {
  if (!oauthFlow.value.authUrl) return
  try {
    await navigator.clipboard.writeText(oauthFlow.value.authUrl)
    oauthSuccess.value = '已复制授权链接'
  } catch (error) {
    console.error('Failed to copy URL', error)
    oauthError.value = '无法复制链接，请手动选择后复制'
  }
}

const openAuthUrl = () => {
  if (!oauthFlow.value.authUrl) return
  window.open(oauthFlow.value.authUrl, '_blank', 'noopener,noreferrer')
}

const submitOAuthCode = async () => {
  if (oauthSubmitting.value) return
  oauthError.value = ''
  oauthSuccess.value = ''

  const { code, state: parsedState } = parseCallbackInput(manualCallbackInput.value)
  const state = parsedState || oauthFlow.value.state

  if (!code) {
    oauthError.value = '请填写或粘贴回调中的 code'
    return
  }
  if (!state) {
    oauthError.value = '缺少 state，无法完成校验'
    return
  }

  oauthSubmitting.value = true
  try {
    const payload = {
      code,
      state,
      code_verifier: oauthFlow.value.codeVerifier
    }

    const response = await fetch('/api/oauth2/callback', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(payload)
    })

    const result = await response.json().catch(() => ({}))
    if (!response.ok) {
      throw new Error(result?.message || result?.error || '兑换授权码失败')
    }

    oauthSuccess.value = result?.message || '授权完成，令牌已保存'
    await fetchProviders()
  } catch (error) {
    console.error('Failed to finish OAuth2 flow:', error)
    oauthError.value = (error as Error)?.message || 'Failed to finish OAuth2 flow'
  } finally {
    oauthSubmitting.value = false
  }
}

const fetchProviders = async () => {
  try {
    const data = await apiService.getProviders({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    providers.value = data.items.map(normalizeProvider)
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to fetch providers:', error)
  }
}
const fetchPlatforms = async () => { 
  try {
    const data = await apiService.getPlatforms()
    platforms.value = data
  } catch (error) {
    console.error('Failed to fetch platforms:', error)
  }
}

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchProviders()
}

onMounted(() => {
  fetchProviders()
  fetchPlatforms()
})

const maskAPIKey = (apiKey: string) => {
  if (apiKey.length <= 8) return apiKey
  return apiKey.substring(0, 4) + '...' + apiKey.substring(apiKey.length - 4)
}

const openCreateModal = () => {
  newProvider.value = {
    name: '',
    type: '',
    key: '',
    baseUrl: '',
    proxyUrl: '',
    isActive: true,
    accountId: ''
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const openEditModal = (provider: Provider) => {
  editingProvider.value = { ...provider }
  // For codex providers, ensure redirect URL is always the required value
  if (editingProvider.value.type === 'codex') {
    editingProvider.value.redirectUrl = 'http://localhost:1455/auth/callback';
  }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const closeOAuthModal = () => {
  showOAuthModal.value = false
  resetOAuthState()
}

const openModelsModal = (provider: Provider) => {
  currentProvider.value = normalizeProvider(provider)
  providerModelForm.value = {
    name: '',
    mappedName: '',
    isActive: true
  }
  showModelsModal.value = true
}

const closeModelsModal = () => {
  showModelsModal.value = false
  currentProvider.value = null
}

const confirmDelete = (provider: Provider) => {
  deletingProvider.value = provider
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deletingProvider.value = null
}

const handleCreate = async () => {
  try {
  await apiService.createProvider({...newProvider.value, healthCheckEnabled: newProvider.value.healthCheckEnabled ?? false})
    pagination.page = 1
    await fetchProviders()
    closeCreateModal()
  } catch (error) {
    console.error('Failed to create provider:', error)
  }
}

const handleUpdate = async () => {
  try {
    const updatedProvider = await apiService.updateProvider(editingProvider.value.id, editingProvider.value)
    replaceProvider(normalizeProvider(updatedProvider))
    closeEditModal()
  } catch (error) {
    console.error('Failed to update provider:', error)
  }
}

const handleDelete = async () => {
  if (!deletingProvider.value) return
  
  try {
    await apiService.deleteProvider(deletingProvider.value.id)
    const isLastItemOnPage = providers.value.length === 1 && pagination.page > 1
    if (isLastItemOnPage) {
      pagination.page -= 1
    }
    await fetchProviders()
    closeDeleteModal()
  } catch (error) {
    console.error('Failed to delete provider:', error)
  }
}

const toggleProviderStatus = async (provider: Provider) => {
  try {
    const updatedProvider = await apiService.updateProvider(provider.id, {
      ...provider,
      isActive: !provider.isActive
    })
    replaceProvider(normalizeProvider(updatedProvider))
    // Show success message
    alert(`Provider ${updatedProvider.name} has been ${updatedProvider.isActive ? 'enabled' : 'disabled'}`)
  } catch (error) {
    console.error('Failed to toggle provider status:', error)
    // Show error message
    alert(`Failed to update provider status: ${(error as Error).message || 'Unknown error'}`)
  }
}

const addModelMapping = async () => {
  if (!currentProvider.value) return
  if (!providerModelForm.value.name || !providerModelForm.value.mappedName) return

  try {
    const payload = {
      providerId: Number(currentProvider.value.id),
      name: providerModelForm.value.name,
      mappedName: providerModelForm.value.mappedName,
      isActive: providerModelForm.value.isActive
    }
    const created = await apiService.createModel(payload)
    const normalized = normalizeModel(created)
    currentProvider.value.models.unshift(normalized)
    replaceProvider(currentProvider.value)
    providerModelForm.value = { name: '', mappedName: '', isActive: true }
  } catch (error) {
    console.error('Failed to add model mapping:', error)
  }
}

const openModelEditor = (model: Model | Record<string, any>) => {
  editingModel.value = { ...(model as Model) }
  showModelEditModal.value = true
}

const closeModelEditModal = () => {
  showModelEditModal.value = false
  editingModel.value = null
}

const saveModelEdit = async () => {
  if (!editingModel.value || !currentProvider.value) return

  try {
    const payload = {
      providerId: Number(currentProvider.value.id),
      name: editingModel.value.name,
      mappedName: editingModel.value.mappedName,
      isActive: editingModel.value.isActive
    }
    const updated = await apiService.updateModel(editingModel.value.id, payload)
    const normalized = normalizeModel(updated)

    const idx = currentProvider.value.models.findIndex(m => m.id === normalized.id)
    if (idx !== -1) currentProvider.value.models[idx] = normalized
    replaceProvider(currentProvider.value)
    closeModelEditModal()
  } catch (error) {
    console.error('Failed to update model:', error)
  }
}

const toggleModelStatus = async (model: Model | Record<string, any>) => {
  if (!currentProvider.value) return
  
  try {
    const typedModel = model as Model
    const updatedModel = await apiService.updateModel(typedModel.id, {
      providerId: Number(currentProvider.value.id),
      name: typedModel.name,
      mappedName: typedModel.mappedName,
      isActive: !typedModel.isActive
    })
    
    // Update in current provider
    const index = currentProvider.value.models.findIndex(m => m.id === updatedModel.id)
    if (index !== -1) {
      currentProvider.value.models[index] = normalizeModel(updatedModel)
    }
    
    // Update in main providers list
    replaceProvider(currentProvider.value)
  } catch (error) {
    console.error('Failed to toggle model status:', error)
  }
}

const deleteModel = async (model: Model | Record<string, any>) => {
  if (!currentProvider.value) return
  const typed = model as Model
  if (!confirm(`Delete mapping "${typed.mappedName}"?`)) return
  try {
    await apiService.deleteModel(typed.id)
    currentProvider.value.models = currentProvider.value.models.filter(m => m.id !== typed.id)
    replaceProvider(currentProvider.value)
  } catch (error) {
    console.error('Failed to delete model:', error)
  }
}

const refreshToken = async (provider: Provider) => {
  if (!confirm(`Refresh token for ${provider.name}? This will update the access token using the refresh token.`)) {
    return;
  }

  try {
    const response = await fetch(`/api/providers/${provider.id}/refresh-token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });

    if (response.ok) {
      const result = await response.json();
      alert(`Token refreshed successfully for ${provider.name}`);
      // Update the provider in the list with the new information
      replaceProvider(normalizeProvider(result));
    } else {
      const error = await response.json();
      alert(`Failed to refresh token: ${error.message || 'Unknown error'}`);
    }
  } catch (error) {
    console.error('Failed to refresh token:', error);
    alert(`Failed to refresh token: ${(error as Error).message || 'Unknown error'}`);
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

const replaceProvider = (provider: Provider) => {
  const index = providers.value.findIndex(p => p.id === provider.id)
  if (index !== -1) {
    providers.value[index] = { ...provider }
  }
}
</script>

<style scoped>
</style>
