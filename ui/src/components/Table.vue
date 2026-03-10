<template>
  <div class="table-container">
    <table class="table">
      <thead>
        <tr>
          <th v-for="column in columns" :key="column.key" @click="sortBy(column.key)">
            <div class="flex items-center cursor-pointer">
              {{ column.title }}
              <i v-if="sortKey === column.key" :class="sortDirection === 'asc' ? 'fas fa-chevron-up ml-1' : 'fas fa-chevron-down ml-1'"></i>
              <i v-else class="fas fa-sort ml-1 text-gray-400"></i>
            </div>
          </th>
          <th v-if="$slots.actions" class="text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(item, index) in sortedItems" :key="item.id || index">
          <td v-for="column in columns" :key="column.key">
            <slot :name="column.key" :item="item" :value="item[column.key]">
              {{ item[column.key] }}
            </slot>
          </td>
          <td v-if="$slots.actions" class="text-right">
            <slot name="actions" :item="item"></slot>
          </td>
        </tr>
        <tr v-if="sortedItems.length === 0">
          <td :colspan="columns.length + ($slots.actions ? 1 : 0)" class="text-center py-8 text-gray-500">
            No data available
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Column {
  key: string
  title: string
}

interface Item {
  id?: string | number
  [key: string]: any
}

const props = defineProps({
  columns: {
    type: Array as () => Column[],
    required: true,
    validator: (value: Column[]) => {
      return value.every(col => col.key && col.title)
    }
  },
  items: {
    type: Array as () => Item[],
    required: true,
    default: () => []
  }
})

const sortKey = ref<string | null>(null)
const sortDirection = ref<'asc' | 'desc'>('asc')

const sortedItems = computed(() => {
  if (!sortKey.value) return props.items
  
  return [...props.items].sort((a, b) => {
    const aValue = a[sortKey.value!]
    const bValue = b[sortKey.value!]
    
    if (aValue < bValue) return sortDirection.value === 'asc' ? -1 : 1
    if (aValue > bValue) return sortDirection.value === 'asc' ? 1 : -1
    return 0
  })
})

const sortBy = (key: string) => {
  if (sortKey.value === key) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDirection.value = 'asc'
  }
}
</script>

<style scoped>
</style>