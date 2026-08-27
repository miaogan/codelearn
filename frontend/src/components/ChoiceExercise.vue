<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Exercise } from '@/types'

const props = defineProps<{ exercise: Exercise }>()
const emit = defineEmits<{
  answer: [value: string]
}>()

const selected = ref<string | null>(null)
const options = computed(() => {
  try { return JSON.parse(props.exercise.options) as string[] } catch { return [] }
})

function select(opt: string) {
  selected.value = opt
  emit('answer', opt)
}
</script>

<template>
  <div class="choice-exercise">
    <div class="options">
      <button
        v-for="opt in options"
        :key="opt"
        class="option-btn"
        :class="{ selected: selected === opt }"
        @click="select(opt)"
      >
        {{ opt }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.option-btn {
  background: white;
  border: 3px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 16px;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
  text-align: center;
  box-shadow: 0 3px 0 var(--border);
  transition: all 0.15s;
}

.option-btn:hover {
  border-color: var(--secondary);
  color: var(--secondary);
}

.option-btn.selected {
  border-color: var(--secondary);
  color: var(--secondary);
  background: var(--bg-blue);
  box-shadow: 0 3px 0 var(--secondary);
}
</style>
