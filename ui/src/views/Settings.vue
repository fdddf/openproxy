<template>
  <Layout>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">System Settings</h1>
      <p class="text-gray-600 mt-1">Configure global system settings</p>
    </div>
    
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- General Settings -->
      <Card title="General Settings">
        <div class="space-y-4">
          <div>
            <label class="form-label">Default Provider</label>
            <select 
              v-model="settings.defaultProvider"
              class="input-field"
            >
              <option value="">Select default provider</option>
              <option v-for="provider in providers" :key="provider.id" :value="provider.id">
                {{ provider.name }}
              </option>
            </select>
          </div>
          
          <div>
            <label class="form-label">Rate Limit (requests per minute)</label>
            <input 
              type="number" 
              v-model="settings.rateLimit"
              class="input-field"
              min="1"
              required
            >
          </div>
          
          <div>
            <label class="form-label">Concurrent Requests Limit</label>
            <input 
              type="number" 
              v-model="settings.concurrentLimit"
              class="input-field"
              min="1"
              required
            >
          </div>
        </div>
      </Card>
      
      <!-- Advanced Settings -->
      <Card title="Advanced Settings">
        <div class="space-y-4">
          <div>
            <label class="form-label">Log Level</label>
            <select 
              v-model="settings.logLevel"
              class="input-field"
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warning</option>
              <option value="error">Error</option>
            </select>
          </div>
          
          <div>
            <label class="form-label">Cache TTL (seconds)</label>
            <input 
              type="number" 
              v-model="settings.cacheTTL"
              class="input-field"
              min="0"
              required
            >
          </div>
        </div>
      </Card>
      
      <!-- System Status -->
      <Card title="System Status">
        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <h4 class="font-medium text-sm text-gray-600">Server Time</h4>
              <p class="text-sm">{{ serverTime }}</p>
            </div>
            <div>
              <h4 class="font-medium text-sm text-gray-600">Version</h4>
              <p class="text-sm">v1.0.0</p>
            </div>
          </div>
          
          <div>
            <label class="flex items-center">
              <input 
                type="checkbox" 
                v-model="settings.maintenanceMode"
                class="mr-2"
              >
              Maintenance Mode
              <span class="text-xs text-gray-500 ml-2">(Disables API access)</span>
            </label>
          </div>

          <div>
            <label class="flex items-center">
              <input 
                type="checkbox" 
                v-model="settings.healthCheckEnabled"
                class="mr-2"
              >
              Provider Health Check
              <span class="text-xs text-gray-500 ml-2">(Runs hourly)</span>
            </label>
          </div>
        </div>
      </Card>
      
      <!-- Save Button -->
      <div class="lg:col-span-2 flex justify-end">
        <button 
          class="btn-primary"
          @click="saveSettings"
          :disabled="loading"
        >
          <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
          Save Settings
        </button>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import Card from '@/components/Card.vue'
import { apiService } from '@/api'
import { SystemSettings, Provider } from '@/types'
import { ref, onMounted, computed } from 'vue'

const settings = ref<SystemSettings>({
  defaultProvider: '',
  rateLimit: 100,
  concurrentLimit: 10,
  logLevel: 'info',
  cacheTTL: 3600,
  maintenanceMode: false,
  healthCheckEnabled: false
})

const providers = ref<Provider[]>([])
const loading = ref(false)
const serverTime = computed(() => new Date().toLocaleString())

const fetchSettings = async () => {
  try {
    const data = await apiService.getSettings()
    settings.value = data
  } catch (error) {
    console.error('Failed to fetch settings:', error)
  }
}

const fetchProviders = async () => {
  try {
    const data = await apiService.getAllProviders()
    providers.value = data
  } catch (error) {
    console.error('Failed to fetch providers:', error)
  }
}

onMounted(async () => {
  await Promise.all([
    fetchSettings(),
    fetchProviders()
  ])
})

const saveSettings = async () => {
  loading.value = true
  try {
    await apiService.updateSettings(settings.value)
    alert('Settings saved successfully!')
  } catch (error) {
    console.error('Failed to save settings:', error)
    alert('Failed to save settings. Please try again.')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
</style>
