<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4 py-10">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full mb-4">
          <i class="fas fa-rocket text-2xl text-primary"></i>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">Welcome to OpenProxy</h1>
        <p class="text-gray-600 mt-2">Create the administrator account to get started.</p>
      </div>

      <div class="bg-white rounded-xl shadow-lg p-8">
        <div v-if="checking" class="text-center text-slate-500 py-6">
          <i class="fas fa-spinner fa-spin mr-2"></i>
          Checking instance status…
        </div>

        <form v-else @submit.prevent="handleSetup">
          <div class="mb-4">
            <label class="form-label">Username</label>
            <input
              type="text"
              v-model="username"
              class="input-field"
              placeholder="admin"
              autocomplete="username"
              required
            >
          </div>

          <div class="mb-4">
            <label class="form-label">Password</label>
            <input
              type="password"
              v-model="password"
              class="input-field"
              placeholder="At least 8 characters"
              autocomplete="new-password"
              required
            >
          </div>

          <div class="mb-4">
            <label class="form-label">Confirm password</label>
            <input
              type="password"
              v-model="confirmPassword"
              class="input-field"
              placeholder="Re-enter the password"
              autocomplete="new-password"
              required
            >
          </div>

          <div class="mb-6">
            <label class="form-label">Email <span class="text-slate-400 font-normal">(optional)</span></label>
            <input
              type="email"
              v-model="email"
              class="input-field"
              placeholder="you@example.com"
              autocomplete="email"
            >
          </div>

          <p v-if="error" class="mb-4 text-sm text-red-600 bg-red-50 border border-red-100 rounded-lg px-3 py-2">
            {{ error }}
          </p>

          <button
            type="submit"
            class="btn-primary w-full"
            :disabled="submitting"
          >
            <i v-if="submitting" class="fas fa-spinner fa-spin mr-2"></i>
            Create administrator
          </button>
        </form>

        <p class="mt-6 text-xs text-slate-500 text-center">
          This account is a super user. It can manage providers, models, and other users.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '@/api'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('admin')
const password = ref('')
const confirmPassword = ref('')
const email = ref('')
const error = ref('')
const submitting = ref(false)
const checking = ref(true)

onMounted(async () => {
  try {
    const { needsSetup } = await apiService.getSetupStatus()
    if (!needsSetup) {
      // Already initialized: nothing to do here.
      router.replace('/login')
      return
    }
  } catch (e) {
    error.value = 'Could not reach the server. Is it running?'
  } finally {
    checking.value = false
  }
})

const handleSetup = async () => {
  error.value = ''

  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = 'The two passwords do not match.'
    return
  }

  submitting.value = true
  try {
    const { token, user } = await apiService.completeSetup({
      username: username.value.trim(),
      password: password.value,
      email: email.value.trim() || undefined,
    })
    authStore.setSession(token, user)
    router.replace('/')
  } catch (e: any) {
    error.value = e?.response?.data?.error || e?.response?.data?.message || 'Setup failed. Please try again.'
  } finally {
    submitting.value = false
  }
}
</script>
