<script setup lang="ts">
import { ref } from 'vue'
import { codeApi, exerciseApi } from '@/api/client'
import type { Exercise, JudgeResult, RunResult } from '@/types'

const props = defineProps<{ exercise: Exercise }>()
const emit = defineEmits<{
  passed: []
  code: [value: string]
}>()

const code = ref(props.exercise.code_template || '')
const runResult = ref<RunResult | null>(null)
const judgeResult = ref<JudgeResult | null>(null)
const running = ref(false)
const judging = ref(false)
const showResult = ref(false)

async function runCode() {
  running.value = true
  showResult.value = true
  judgeResult.value = null
  try {
    const res = await codeApi.run(props.exercise.type === 'code' ? inferLanguage() : 'python', code.value)
    runResult.value = res.data
  } catch (e: any) {
    runResult.value = { output: '', error: e.response?.data?.error || '运行失败', exit_code: 1, time_ms: 0 }
  } finally {
    running.value = false
  }
}

async function judgeCode() {
  judging.value = true
  showResult.value = true
  runResult.value = null
  emit('code', code.value)
  try {
    const lang = inferLanguage()
    const res = await codeApi.judge(props.exercise.id, lang, code.value)
    judgeResult.value = res.data
    if (res.data.all_pass) {
      emit('passed')
    }
  } catch (e: any) {
    judgeResult.value = {
      results: [],
      all_pass: false,
      pass_count: 0,
      total_count: 0,
    }
  } finally {
    judging.value = false
  }
}

function inferLanguage(): string {
  if (props.exercise.code_template?.includes('package main') || props.exercise.code_template?.includes('func main')) return 'go'
  if (props.exercise.code_template?.includes('def ') || props.exercise.code_template?.includes('print(')) return 'python'
  return 'python'
}
</script>

<template>
  <div class="code-exercise">
    <div class="code-editor-wrap">
      <textarea
        v-model="code"
        class="code-editor"
        spellcheck="false"
        placeholder="在这里写代码..."
      ></textarea>
    </div>

    <div class="code-actions">
      <button class="btn-ghost" @click="runCode" :disabled="running || !code.trim()">
        {{ running ? '运行中...' : '▶ 运行' }}
      </button>
      <button class="btn-primary" @click="judgeCode" :disabled="judging || !code.trim()">
        {{ judging ? '评判中...' : '✓ 提交评判' }}
      </button>
    </div>

    <div v-if="showResult && runResult" class="result-panel">
      <div class="result-title">运行结果</div>
      <pre class="result-output">{{ runResult.output || '(无输出)' }}</pre>
      <div v-if="runResult.error" class="result-error">{{ runResult.error }}</div>
      <div class="result-meta">耗时: {{ runResult.time_ms }}ms</div>
    </div>

    <div v-if="showResult && judgeResult" class="result-panel">
      <div class="result-title">
        {{ judgeResult.all_pass ? '🎉 全部通过！' : `通过 ${judgeResult.pass_count}/${judgeResult.total_count}` }}
      </div>
      <div v-for="(tc, i) in judgeResult.results" :key="i" class="test-case" :class="{ pass: tc.pass, fail: !tc.pass }">
        <span class="tc-status">{{ tc.pass ? '✓' : '✗' }}</span>
        <div class="tc-detail">
          <div><strong>输入:</strong> {{ tc.input || '(空)' }}</div>
          <div><strong>期望:</strong> {{ tc.expected }}</div>
          <div v-if="!tc.pass"><strong>实际:</strong> {{ tc.actual }}</div>
          <div v-if="tc.error" class="tc-error">{{ tc.error }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.code-editor-wrap {
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.code-editor {
  width: 100%;
  min-height: 200px;
  padding: 12px;
  font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  font-weight: 500;
  border: none;
  outline: none;
  resize: vertical;
  background: #1e1e1e;
  color: #d4d4d4;
  line-height: 1.6;
}

.code-actions {
  display: flex;
  gap: 12px;
  margin-top: 12px;
}

.code-actions button { flex: 1; }

.result-panel {
  margin-top: 16px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px;
  background: var(--bg-gray);
}

.result-title { font-size: 16px; font-weight: 800; margin-bottom: 8px; }
.result-output {
  background: #1e1e1e;
  color: #4ec9b0;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
}
.result-error { color: var(--danger); margin-top: 4px; font-size: 13px; }
.result-meta { color: var(--text-light); font-size: 12px; margin-top: 4px; }

.test-case {
  display: flex;
  gap: 8px;
  padding: 8px;
  border-radius: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.test-case.pass { background: var(--primary-light); }
.test-case.fail { background: #ffe5e5; }

.tc-status { font-size: 16px; font-weight: 900; }
.tc-detail { flex: 1; }
.tc-error { color: var(--danger); margin-top: 4px; }
</style>
