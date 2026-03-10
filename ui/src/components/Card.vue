<template>
  <div class="card">
    <div class="card-header">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
        <slot name="header-actions"></slot>
      </div>
    </div>
    <div class="card-body">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  title: {
    type: String,
    required: false,
    default: ''
  }
})
</script>

<style scoped>
</style>