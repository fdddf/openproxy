<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="text-center">
      <div class="inline-flex items-center justify-center w-32 h-32 bg-primary/10 rounded-full mb-8">
        <i class="fas fa-exclamation-triangle text-5xl text-primary"></i>
      </div>
      <h1 class="text-4xl font-bold text-gray-900 mb-4">404</h1>
      <h2 class="text-xl font-semibold text-gray-900 mb-2">Page Not Found</h2>
      <p class="text-gray-600 mb-8">The page you're looking for doesn't exist or has been moved.</p>
      <button 
        class="btn-primary"
        @click="goHome"
      >
        <i class="fas fa-home mr-2"></i>
        Go to Dashboard
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
</style>