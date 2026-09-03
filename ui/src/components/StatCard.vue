<template>
  <div class="card hover:shadow-card-hover transition-all duration-300">
    <div class="card-body p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-gray-600">{{ label }}</p>
          <p class="text-2xl font-bold text-gray-900 mt-1">{{ value }}</p>
          <p v-if="trendText" class="text-xs text-gray-500 mt-2">
            <i :class="trendIcon" class="mr-1"></i>
            {{ trendText }}
          </p>
        </div>
        <div class="p-3 rounded-full bg-primary/10 text-primary">
          <i :class="icon" class="text-xl"></i>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  icon: {
    type: String,
    required: true
  },
  trend: {
    type: String,
    default: 'up',
    validator: (value: string) => ['up', 'down', 'same'].includes(value)
  },
  trendText: {
    type: String,
    default: ''
  }
})

const trendIcon = computed(() => {
  switch (props.trend) {
    case 'up':
      return 'fas fa-arrow-up text-green-500'
    case 'down':
      return 'fas fa-arrow-down text-red-500'
    default:
      return 'fas fa-minus text-gray-500'
  }
})
</script>

<style scoped>
</style>