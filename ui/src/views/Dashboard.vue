<template>
  <Layout>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
      <p class="text-gray-600 mt-1">System overview and statistics</p>
    </div>
    
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
      <StatCard 
        label="Total Requests" 
        :value="formatNumber(stats.totalRequests)" 
        icon="fas fa-exchange-alt"
        :trend="getTrend('requests')"
        :trendText="getTrendText('requests')"
      />
      
      <StatCard 
        label="Active API Keys" 
        :value="formatNumber(stats.activeAPIKeys)" 
        icon="fas fa-key"
        :trend="getTrend('keys')"
        :trendText="getTrendText('keys')"
      />
      
      <StatCard 
        label="Active Providers" 
        :value="formatNumber(stats.activeProviders)" 
        icon="fas fa-server"
        :trend="getTrend('providers')"
        :trendText="getTrendText('providers')"
      />
      
      <StatCard 
        label="Success Rate" 
        :value="formatPercentage(stats.successRate)" 
        icon="fas fa-chart-line"
        :trend="getTrend('success')"
        :trendText="getTrendText('success')"
      />
    </div>
    
    <!-- Recent Activity -->
    <Card title="Recent Activity">
      <div class="space-y-4">
        <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
          <div class="flex items-center">
            <div class="p-2 bg-green-100 rounded-full mr-3">
              <i class="fas fa-check text-green-600"></i>
            </div>
            <div>
              <p class="font-medium text-gray-900">API request successful</p>
              <p class="text-sm text-gray-600">qwen-plus model used</p>
            </div>
          </div>
          <p class="text-sm text-gray-500">2 minutes ago</p>
        </div>
        
        <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
          <div class="flex items-center">
            <div class="p-2 bg-blue-100 rounded-full mr-3">
              <i class="fas fa-user text-blue-600"></i>
            </div>
            <div>
              <p class="font-medium text-gray-900">User logged in</p>
              <p class="text-sm text-gray-600">Admin user session started</p>
            </div>
          </div>
          <p class="text-sm text-gray-500">15 minutes ago</p>
        </div>
        
        <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
          <div class="flex items-center">
            <div class="p-2 bg-yellow-100 rounded-full mr-3">
              <i class="fas fa-exclamation-triangle text-yellow-600"></i>
            </div>
            <div>
              <p class="font-medium text-gray-900">API key approaching limit</p>
              <p class="text-sm text-gray-600">90% of quota used</p>
            </div>
          </div>
          <p class="text-sm text-gray-500">30 minutes ago</p>
        </div>
      </div>
    </Card>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import StatCard from '@/components/StatCard.vue'
import Card from '@/components/Card.vue'
import { apiService } from '@/api'
import { onMounted, ref } from 'vue'

const stats = ref({
  totalRequests: 0,
  activeAPIKeys: 0,
  activeProviders: 0,
  successRate: 0
})

// Helper functions for formatting and trends
const formatNumber = (num: number): string => {
  return new Intl.NumberFormat().format(num)
}

const formatPercentage = (num: number): string => {
  return `${(num * 100).toFixed(1)}%`
}

const getTrend = (type: string): 'up' | 'down' | 'same' => {
  // In a real implementation, this would compare with previous period
  // For now, return a default trend based on type
  return 'up'
}

const getTrendText = (type: string): string => {
  // In a real implementation, this would calculate actual trend
  // For now, return generic trend text
  switch (type) {
    case 'requests':
      return '+12% from last week'
    case 'keys':
      return 'No change'
    case 'providers':
      return '+1 from last week'
    case 'success':
      return '+0.5% from last week'
    default:
      return 'No change'
  }
}

onMounted(async () => {
  try {
    const data = await apiService.getDashboardStats()
    stats.value = data
  } catch (error) {
    console.error('Failed to fetch dashboard stats:', error)
  }
})
</script>

<style scoped>
</style>
