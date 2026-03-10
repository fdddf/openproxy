<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full mb-4">
          <i class="fas fa-bolt text-2xl text-primary"></i>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">OpenProxy Admin</h1>
        <p class="text-gray-600 mt-2">Sign in to your account</p>
      </div>
      
      <form @submit.prevent="handleLogin">
        <div class="mb-4">
          <label class="form-label">Username</label>
          <input 
            type="text" 
            v-model="username"
            class="input-field"
            placeholder="Enter your username"
            required
          >
        </div>
        
        <div class="mb-6">
          <label class="form-label">Password</label>
          <input 
            type="password" 
            v-model="password"
            class="input-field"
            placeholder="Enter your password"
            required
          >
        </div>
        
        <button 
          type="submit" 
          class="btn-primary w-full"
          :disabled="loading"
        >
          <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
          Sign In
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/store/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const username = ref('')
const password = ref('')
const loading = ref(false)

const handleLogin = async () => {
  loading.value = true
  try {
    const success = await authStore.login(username.value, password.value)
    if (success) {
      router.push('/')
    } else {
      alert('Invalid username or password')
    }
  } catch (error) {
    alert('Login failed. Please try again.')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
</style>