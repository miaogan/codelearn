<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { certificateApi } from '@/api/client'
import type { Certificate } from '@/types'

const route = useRoute()
const router = useRouter()
const certNo = ref(String(route.query.no || ''))
const result = ref<{ valid: boolean; certificate?: Certificate; message?: string } | null>(null)
const checking = ref(false)

onMounted(() => {
  if (certNo.value.trim()) {
    verify()
  }
})

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

async function verify() {
  const no = certNo.value.trim()
  if (!no) return
  checking.value = true
  result.value = null
  try {
    const res = await certificateApi.verify(no)
    result.value = res.data
  } catch (e) {
    result.value = { valid: false, message: '验证服务暂不可用，请稍后再试' }
  }
  checking.value = false
}

function levelText(level: string) {
  const map: Record<string, string> = { beginner: '入门', intermediate: '进阶', advanced: '熟练' }
  return map[level] || level
}

function fmtDate(iso: string) {
  return iso ? iso.slice(0, 10) : ''
}
</script>

<template>
  <div class="container verify-page">
    <h1 class="page-title">🔍 证书在线验证</h1>
    <p class="page-sub">输入证书编号，验证真伪与证书详情（公开查询，无需登录）</p>

    <div class="verify-form card">
      <input
        class="verify-input"
        v-model="certNo"
        placeholder="请输入证书编号，如 CL-20260925-1A2B3C4D"
        @keyup.enter="verify"
      />
      <button class="btn-primary" :disabled="checking || !certNo.trim()" @click="verify">
        {{ checking ? '验证中...' : '开始验证' }}
      </button>
    </div>

    <div v-if="result && result.valid && result.certificate" class="card result-card valid">
      <div class="result-icon">✅</div>
      <h2 class="result-title">验证通过 · 真实证书</h2>
      <div class="cert-info">
        <div class="info-row"><span class="info-label">持有人</span><span class="info-value">{{ result.certificate.user_name }}</span></div>
        <div class="info-row"><span class="info-label">课程</span><span class="info-value">{{ result.certificate.course_title }}</span></div>
        <div class="info-row">
          <span class="info-label">等级</span>
          <span class="info-value level-badge" :class="'lv-' + result.certificate.level">{{ levelText(result.certificate.level) }}</span>
        </div>
        <div class="info-row"><span class="info-label">成绩</span><span class="info-value">{{ result.certificate.score }} 分</span></div>
        <div class="info-row"><span class="info-label">证书编号</span><span class="info-value mono">{{ result.certificate.cert_no }}</span></div>
        <div class="info-row"><span class="info-label">颁发日期</span><span class="info-value">{{ fmtDate(result.certificate.issued_at) }}</span></div>
      </div>
    </div>

    <div v-else-if="result" class="card result-card invalid">
      <div class="result-icon">❌</div>
      <h2 class="result-title">验证失败</h2>
      <p class="result-msg">{{ result.message || '未找到该证书，请核对编号是否正确' }}</p>
    </div>

    <button class="btn-ghost back-btn" @click="goBack">← 返回</button>
  </div>
</template>

<style scoped>
.verify-page { padding-top: 24px; padding-bottom: 40px; }
.page-title { font-size: 24px; margin-bottom: 4px; }
.page-sub { color: var(--text-light); font-size: 14px; margin-bottom: 20px; }

.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 24px;
  box-shadow: 0 4px 0 #e5e5e5;
}

.verify-form { display: flex; gap: 12px; margin-bottom: 20px; }
.verify-input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 15px;
  outline: none;
  font-family: 'SF Mono', Consolas, monospace;
}
.verify-input:focus { border-color: var(--primary); }
.btn-primary, .btn-ghost {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 4px 0 var(--primary-dark); white-space: nowrap; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-ghost { background: transparent; color: var(--text-light); }

.result-card { text-align: center; margin-bottom: 20px; }
.result-card.valid { border-color: #86efac; }
.result-card.invalid { border-color: #fca5a5; }
.result-icon { font-size: 48px; display: block; margin-bottom: 6px; }
.result-title { font-size: 20px; margin-bottom: 14px; }
.valid .result-title { color: var(--secondary); }
.invalid .result-title { color: var(--danger); }
.result-msg { color: var(--text-light); font-size: 14px; }

.cert-info { text-align: left; }
.info-row {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px dashed var(--border);
}
.info-row:last-child { border-bottom: none; }
.info-label { width: 90px; color: var(--text-light); font-size: 14px; flex-shrink: 0; }
.info-value { font-weight: 700; font-size: 14px; }
.info-value.mono { font-family: 'SF Mono', Consolas, monospace; }
.level-badge { padding: 2px 12px; border-radius: 12px; font-size: 13px; color: white; }
.lv-beginner { background: #10b981; }
.lv-intermediate { background: #3b82f6; }
.lv-advanced { background: #f59e0b; }

.back-btn { width: 100%; }
</style>
