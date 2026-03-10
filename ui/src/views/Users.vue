<template>
  <Layout>
    <div class="flex items-center justify-between mb-6">
      <div>
        <p class="text-xs uppercase tracking-[0.2em] text-indigo-500 font-semibold mb-1">Team Access</p>
        <h1 class="text-2xl font-bold text-gray-900">User Directory</h1>
        <p class="text-gray-600 mt-1">Manage admin accounts and access levels.</p>
      </div>
      <div class="flex items-center space-x-3">
        <input 
          v-model="search"
          type="text"
          class="input-field w-48"
          placeholder="Search users"
        >
        <button 
          class="btn-primary"
          @click="openCreateModal"
        >
          <i class="fas fa-user-plus mr-2"></i>
          Add User
        </button>
      </div>
    </div>

    <Card>
      <Table :columns="columns" :items="filteredUsers">
        <template #username="{ item }">
          <div class="flex items-center space-x-3">
            <img 
              :src="item.avatarUrl || fallbackAvatar(item.username)"
              class="w-9 h-9 rounded-full border border-gray-200 object-cover"
              alt="avatar"
            >
            <div>
              <p class="font-medium text-gray-900">{{ item.displayName || item.username }}</p>
              <p class="text-xs text-gray-500">@{{ item.username }}</p>
            </div>
          </div>
        </template>

        <template #createdAt="{ value }">
          <span class="text-xs text-gray-500">{{ formatDate(value) }}</span>
        </template>

        <template #email="{ value }">
          <span class="text-sm text-gray-700">{{ value || 'Not set' }}</span>
        </template>

        <template #isSuper="{ value }">
          <span class="badge" :class="value ? 'badge-primary' : 'badge-warning'">
            {{ value ? 'Super Admin' : 'Member' }}
          </span>
        </template>

        <template #actions="{ item }">
          <div class="flex justify-end space-x-2">
            <button class="text-blue-600 hover:text-blue-800" @click="openEditModal(item)">
              <i class="fas fa-edit"></i>
            </button>
            <button class="text-red-500 hover:text-red-700" @click="confirmDelete(item)">
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

    <Modal 
      :show="showFormModal"
      :title="formMode === 'create' ? 'Create User' : 'Edit User'"
      @close="closeFormModal"
      @submit="saveUser"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="form-label">Username</label>
            <input 
              v-model="form.username"
              :disabled="formMode === 'edit'"
              type="text"
              class="input-field"
              placeholder="Unique username"
              required
            >
          </div>
          <div>
            <label class="form-label">Display Name</label>
            <input 
              v-model="form.displayName"
              type="text"
              class="input-field"
              placeholder="Name shown in UI"
            >
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="form-label">Email</label>
            <input 
              v-model="form.email"
              type="email"
              class="input-field"
              placeholder="name@company.com"
            >
          </div>
          <div>
            <label class="form-label">Avatar URL</label>
            <input 
              v-model="form.avatarUrl"
              type="url"
              class="input-field"
              placeholder="https://images.example/avatar.png"
            >
          </div>
        </div>

        <div>
          <label class="form-label">Bio</label>
          <textarea 
            v-model="form.bio"
            rows="2"
            class="input-field"
            placeholder="Short description"
          ></textarea>
        </div>

        <div>
          <label class="form-label">Password</label>
          <input 
            v-model="form.password"
            :required="formMode === 'create'"
            type="password"
            class="input-field"
            :placeholder="formMode === 'create' ? 'Set an initial password' : 'Leave blank to keep current'"
          >
        </div>

        <label class="flex items-center space-x-2">
          <input 
            v-model="form.isSuper"
            type="checkbox"
            class="mr-2"
          >
          <span class="text-sm text-gray-700">Grant super admin privileges</span>
        </label>
      </div>
    </Modal>

    <Modal 
      :show="showDeleteModal"
      title="Remove User"
      @close="closeDeleteModal"
      @submit="deleteUser"
    >
      <div class="space-y-3 text-sm">
        <p class="text-gray-700">
          Delete <strong>{{ selectedUser?.displayName || selectedUser?.username }}</strong>? This action cannot be undone.
        </p>
        <p class="text-gray-500">Their API keys and logs will remain linked for auditing.</p>
      </div>
    </Modal>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, reactive } from 'vue'
import Layout from '@/components/Layout.vue'
import Card from '@/components/Card.vue'
import Table from '@/components/Table.vue'
import Modal from '@/components/Modal.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import { apiService } from '@/api'
import { User } from '@/types'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()

const columns = [
  { key: 'username', title: 'User' },
  { key: 'email', title: 'Email' },
  { key: 'isSuper', title: 'Role' },
  { key: 'createdAt', title: 'Created' }
]

const users = ref<User[]>([])
const search = ref('')
const showFormModal = ref(false)
const showDeleteModal = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedUser = ref<User | null>(null)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = ref({
  username: '',
  displayName: '',
  email: '',
  avatarUrl: '',
  bio: '',
  password: '',
  isSuper: false
})

const filteredUsers = computed(() => {
  const term = search.value.toLowerCase()
  if (!term) return users.value
  return users.value.filter(user =>
    user.username.toLowerCase().includes(term) ||
    (user.displayName || '').toLowerCase().includes(term) ||
    (user.email || '').toLowerCase().includes(term)
  )
})

const fetchUsers = async () => {
  try {
    const data = await apiService.getUsers({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    users.value = data.items
    pagination.total = data.total
    pagination.page = data.page
    pagination.pageSize = data.pageSize
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

onMounted(fetchUsers)

const openCreateModal = () => {
  formMode.value = 'create'
  selectedUser.value = null
  form.value = {
    username: '',
    displayName: '',
    email: '',
    avatarUrl: '',
    bio: '',
    password: '',
    isSuper: false
  }
  showFormModal.value = true
}

const openEditModal = (user: User | Record<string, any>) => {
  const typed = user as User
  formMode.value = 'edit'
  selectedUser.value = typed
  form.value = {
    username: typed.username,
    displayName: typed.displayName,
    email: typed.email,
    avatarUrl: typed.avatarUrl,
    bio: typed.bio,
    password: '',
    isSuper: typed.isSuper
  }
  showFormModal.value = true
}

const closeFormModal = () => {
  showFormModal.value = false
}

const saveUser = async () => {
  try {
    if (formMode.value === 'create') {
      const created = await apiService.createUser({
        ...form.value,
        password: form.value.password
      })
      pagination.page = 1
      await fetchUsers()
    } else if (selectedUser.value) {
      const payload = { ...form.value }
      if (!payload.password) {
        delete (payload as any).password
      }
      const updated = await apiService.updateUser(selectedUser.value.id, payload)
      await fetchUsers()
      if (authStore.user?.id === updated.id) {
        localStorage.setItem('user', JSON.stringify(updated))
        authStore.user = updated
      }
    }
    closeFormModal()
  } catch (error) {
    console.error('Failed to save user:', error)
  }
}

const confirmDelete = (user: User | Record<string, any>) => {
  selectedUser.value = user as User
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  selectedUser.value = null
}

const deleteUser = async () => {
  if (!selectedUser.value) return
  try {
    await apiService.deleteUser(selectedUser.value.id)
    const isLastOnPage = users.value.length === 1 && pagination.page > 1
    if (isLastOnPage) {
      pagination.page -= 1
    }
    await fetchUsers()
    closeDeleteModal()
  } catch (error) {
    console.error('Failed to delete user:', error)
  }
}

const formatDate = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleDateString()
}

const fallbackAvatar = (username: string) => {
  return `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(username)}`
}

const handlePageChange = (newPage: number) => {
  pagination.page = newPage
  fetchUsers()
}
</script>

<style scoped>
</style>
