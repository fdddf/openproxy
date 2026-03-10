<template>
  <div 
    v-if="show"
    class="fixed inset-0 bg-black bg-opacity-50 flex z-50"
    @click="close"
  >
    <div 
      class="bg-white w-full max-w-2xl max-h-full overflow-y-auto animate-slide-up"
      @click.stop
      :class="{ 'ml-auto': position === 'right', 'mr-auto': position === 'left' }"
    >
      <div class="p-6 border-b border-gray-200 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
        <button 
          @click="close"
          class="text-gray-400 hover:text-gray-600"
        >
          <i class="fa fa-times"></i>
        </button>
      </div>
      
      <div class="p-6">
        <slot />
      </div>
      
      <div class="p-6 border-t border-gray-200 flex justify-end space-x-3">
        <button 
          @click="close"
          class="btn-secondary"
        >
          Cancel
        </button>
        <button 
          @click="handleSubmit"
          class="btn-primary"
          :disabled="loading"
        >
          <span v-if="loading" class="inline-block animate-spin mr-2">
            <i class="fa fa-circle-o-notch"></i>
          </span>
          {{ submitText || 'Save' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineEmits, defineProps } from 'vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    required: true
  },
  position: {
    type: String,
    default: 'right',
    validator: (value: string) => ['left', 'right'].includes(value)
  },
  submitText: {
    type: String,
    default: 'Save'
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'submit'])

const close = () => {
  emit('close')
}

const handleSubmit = () => {
  emit('submit')
}
</script>