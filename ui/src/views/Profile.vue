<template>
  <Layout>
    <div class="mb-6">
      <p class="text-xs uppercase tracking-[0.2em] text-indigo-500 font-semibold mb-1">Account</p>
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Your Profile</h1>
          <p class="text-gray-600 mt-1">Update contact info, avatar, and credentials.</p>
        </div>
        <div class="flex items-center space-x-3">
          <div class="px-3 py-2 bg-white rounded-lg shadow-sm border border-gray-200">
            <p class="text-xs text-gray-500">Role</p>
            <p class="text-sm font-semibold text-gray-800">{{ user?.isSuper ? 'Super Admin' : 'Member' }}</p>
          </div>
          <div class="px-3 py-2 bg-white rounded-lg shadow-sm border border-gray-200">
            <p class="text-xs text-gray-500">Joined</p>
            <p class="text-sm font-semibold text-gray-800">{{ formattedDate(user?.createdAt) }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 space-y-6">
        <div class="card">
          <div class="card-header flex items-center justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900">Profile Details</h2>
              <p class="text-sm text-gray-500">These details are used across the admin console.</p>
            </div>
            <button class="btn-secondary" @click="resetProfileForm">
              <i class="fas fa-undo mr-2"></i>
              Reset
            </button>
          </div>
          <div class="card-body space-y-4">
            <div class="flex items-center space-x-4">
              <img 
                :src="profileForm.avatarUrl || fallbackAvatar(profileForm.username)"
                class="w-16 h-16 rounded-full border-2 border-indigo-100 object-cover"
                alt="avatar"
              >
              <div class="flex-1">
                <label class="form-label">Avatar URL</label>
                <input 
                  v-model="profileForm.avatarUrl"
                  type="url"
                  class="input-field"
                  placeholder="https://images.example/avatar.png"
                >
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="form-label">Display Name</label>
                <input 
                  v-model="profileForm.displayName"
                  type="text"
                  class="input-field"
                  placeholder="Your name"
                >
              </div>
              <div>
                <label class="form-label">Username</label>
                <input 
                  v-model="profileForm.username"
                  type="text"
                  class="input-field"
                  disabled
                >
              </div>
            </div>

            <div>
              <label class="form-label">Email</label>
              <input 
                v-model="profileForm.email"
                type="email"
                class="input-field"
                placeholder="name@company.com"
              >
            </div>

            <div>
              <label class="form-label">Bio</label>
              <textarea 
                v-model="profileForm.bio"
                rows="2"
                class="input-field"
                placeholder="What do you focus on?"
              ></textarea>
            </div>

            <div class="flex justify-end">
              <button class="btn-primary" :disabled="savingProfile" @click="saveProfile">
                <i v-if="savingProfile" class="fas fa-spinner fa-spin mr-2"></i>
                Save Changes
              </button>
            </div>

            <p v-if="profileMessage" class="text-sm" :class="profileError ? 'text-red-600' : 'text-green-600'">
              {{ profileMessage }}
            </p>
          </div>
        </div>
      </div>

      <div class="space-y-6">
        <div class="card">
          <div class="card-header">
            <h2 class="text-lg font-semibold text-gray-900">Reset Password</h2>
            <p class="text-sm text-gray-500">Choose a strong password to secure your account.</p>
          </div>
          <div class="card-body space-y-3">
            <div>
              <label class="form-label">Current Password</label>
              <input 
                type="password" 
                v-model="currentPassword"
                class="input-field"
                placeholder="Enter your current password"
              >
            </div>
            <div>
              <label class="form-label">New Password</label>
              <input 
                type="password" 
                v-model="newPassword"
                class="input-field"
                placeholder="Enter a new password"
              >
            </div>
            <div>
              <label class="form-label">Confirm New Password</label>
              <input 
                type="password" 
                v-model="confirmPassword"
                class="input-field"
                placeholder="Re-enter the new password"
              >
            </div>
            <button 
              class="btn-primary w-full"
              :disabled="resetLoading"
              @click="handleReset"
            >
              <i v-if="resetLoading" class="fas fa-spinner fa-spin mr-2"></i>
              Update Password
            </button>
            <p v-if="message" class="text-sm" :class="error ? 'text-red-600' : 'text-green-600'">
              {{ message }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Layout from '@/components/Layout.vue'
import { useAuthStore } from '@/store/auth'
import { apiService } from '@/api'
import { User } from '@/types'

const authStore = useAuthStore()
const user = ref<User | null>(authStore.user)

const profileForm = ref({
  username: '',
  displayName: '',
  email: '',
  avatarUrl: '',
  bio: ''
})
const savingProfile = ref(false)
const profileMessage = ref('')
const profileError = ref(false)

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const resetLoading = ref(false)
const message = ref('')
const error = ref(false)

const loadProfile = async () => {
  try {
    const current = await apiService.getCurrentUser()
    user.value = current
    authStore.user = current
    localStorage.setItem('user', JSON.stringify(current))
    profileForm.value = {
      username: current.username,
      displayName: current.displayName,
      email: current.email,
      avatarUrl: current.avatarUrl,
      bio: current.bio
    }
  } catch (e) {
    console.error('Failed to load profile:', e)
  }
}

onMounted(loadProfile)

const resetProfileForm = () => {
  if (!user.value) return
  profileForm.value = {
    username: user.value.username,
    displayName: user.value.displayName,
    email: user.value.email,
    avatarUrl: user.value.avatarUrl,
    bio: user.value.bio
  }
}

const saveProfile = async () => {
  savingProfile.value = true
  profileMessage.value = ''
  profileError.value = false

  try {
    const updated = await apiService.updateCurrentUser(profileForm.value)
    user.value = updated
    authStore.user = updated
    localStorage.setItem('user', JSON.stringify(updated))
    profileMessage.value = 'Profile updated successfully.'
  } catch (e) {
    console.error('Failed to update profile:', e)
    profileError.value = true
    profileMessage.value = 'Unable to save profile. Please try again.'
  } finally {
    savingProfile.value = false
  }
}

const handleReset = async () => {
  resetLoading.value = true
  message.value = ''
  error.value = false

  if (newPassword.value !== confirmPassword.value) {
    resetLoading.value = false
    error.value = true
    message.value = 'New password and confirmation do not match.'
    return
  }

  try {
    await apiService.resetPassword(currentPassword.value, newPassword.value)
    message.value = 'Password updated successfully.'
    error.value = false
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e) {
    console.error('Password reset failed:', e)
    error.value = true
    message.value = 'Password reset failed. Please verify your current password and try again.'
  } finally {
    resetLoading.value = false
  }
}

const formattedDate = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleDateString()
}

const fallbackAvatar = (username: string) => {
  return `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(username || 'user')}`
}
</script>

<style scoped>
</style>
