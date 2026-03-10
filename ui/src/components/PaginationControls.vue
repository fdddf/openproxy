<template>
  <div class="flex flex-col md:flex-row md:items-center md:justify-between mt-4 space-y-3 md:space-y-0">
    <div class="text-sm text-gray-600">
      Showing {{ rangeStart }}-{{ rangeEnd }} of {{ total }}
    </div>
    <div class="flex items-center space-x-2">
      <button 
        class="btn-sm btn-secondary"
        :disabled="total === 0 || page <= 1"
        @click="changePage(page - 1)"
      >
        Previous
      </button>
      <span class="text-sm text-gray-700">
        Page {{ page }} of {{ totalPages }}
      </span>
      <button 
        class="btn-sm btn-primary"
        :disabled="total === 0 || page >= totalPages"
        @click="changePage(page + 1)"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  page: {
    type: Number,
    required: true
  },
  pageSize: {
    type: Number,
    required: true
  },
  total: {
    type: Number,
    required: true
  }
})

const emit = defineEmits<{
  (e: 'update:page', value: number): void
}>()

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(props.total / props.pageSize))
})

const rangeStart = computed(() => {
  if (props.total === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const rangeEnd = computed(() => {
  if (props.total === 0) return 0
  return Math.min(props.total, props.page * props.pageSize)
})

const changePage = (targetPage: number) => {
  const nextPage = Math.min(Math.max(targetPage, 1), totalPages.value)
  if (nextPage !== props.page) {
    emit('update:page', nextPage)
  }
}
</script>

<style scoped>
</style>
