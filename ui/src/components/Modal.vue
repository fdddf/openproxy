<template>
  <div v-if="show" class="modal-backdrop">
    <div class="modal-content" @click.stop>
      <div class="p-6 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
          <button 
            class="text-gray-400 hover:text-gray-600"
            @click="close"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>
      <div class="p-6">
        <slot />
      </div>
      <div class="p-4 border-t border-gray-200 bg-gray-50 flex justify-end space-x-3">
        <button 
          class="btn-secondary"
          @click="close"
        >
          Cancel
        </button>
        <button 
          class="btn-primary"
          @click="handleSubmit"
          :disabled="loading"
        >
          <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
          {{ submitText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineEmits } from 'vue'

const emit = defineEmits(['submit', 'close'])

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
    default: false
  },
  title: {
    type: String,
    required: true
  },
  submitText: {
    type: String,
    default: 'Save'
  },
  closeOnSubmit: {
    type: Boolean,
    default: true
  }
})

const loading = ref(false)

const close = () => {
  emit('close')
}

const handleSubmit = async () => {
  loading.value = true
  try {
    emit('submit')
    if (props.closeOnSubmit) {
      close()
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
</style>
