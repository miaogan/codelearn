<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  answer: [value: string]
}>()

const inputVal = ref('')

function onInput(e: Event) {
  inputVal.value = (e.target as HTMLInputElement).value
  emit('answer', inputVal.value)
}
</script>

<template>
  <div class="fillblank-exercise">
    <input
      :value="inputVal"
      @input="onInput"
      type="text"
      placeholder="输入你的答案"
      class="fill-input"
    />
  </div>
</template>

<style scoped>
.fill-input {
  font-size: 18px;
  text-align: center;
  padding: 16px;
}
</style>
