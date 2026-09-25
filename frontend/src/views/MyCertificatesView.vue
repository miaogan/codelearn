<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { certificateApi } from '@/api/client'
import type { Certificate } from '@/types'

const router = useRouter()
const certificates = ref<Certificate[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await certificateApi.list()
    certificates.value = res.data.certificates
  } catch (e) {
    // ignore
  }
  loading.value = false
})

function levelText(level: string) {
  const map: Record<string, string> = { beginner: '入门', intermediate: '进阶', advanced: '熟练' }
  return map[level] || level
}

function levelClass(level: string) {
  const map: Record<string, string> = { beginner: 'lv-beginner', intermediate: 'lv-intermediate', advanced: 'lv-advanced' }
  return map[level] || ''
}

function fmtDate(iso: string) {
  return iso ? iso.slice(0, 10) : ''
}
</script>

<template>
  <div class="container certs-page">
    <h1 class="page-title">🎖️ 我的证书</h1>
    <p class="page-sub">完成课程认证考试即可获得能力认证证书</p>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="certificates.length === 0" class="empty card">
      <div class="empty-icon">🏅</div>
      <p class="empty-text">还没有证书</p>
      <p class="empty-sub">先完成一门课程的全部单元考试，再通过课程认证考试，即可获得证书</p>
      <button class="btn-primary" @click="router.push('/')">去学习</button>
    </div>

    <div v-else class="cert-list">
      <div v-for="cert in certificates" :key="cert.id" class="cert-card">
        <div class="cert-card-head">
          <span class="cert-medal">🎖️</span>
          <div class="cert-card-info">
            <div class="cert-course">{{ cert.course_title }}</div>
            <div class="cert-meta">编号 {{ cert.cert_no }}</div>
          </div>
          <span class="level-badge" :class="levelClass(cert.level)">{{ levelText(cert.level) }}</span>
        </div>
        <div class="cert-card-foot">
          <div class="cert-meta">颁发于 {{ fmtDate(cert.issued_at) }} · 成绩 {{ cert.score }} 分</div>
          <div class="cert-actions">
            <button class="btn-secondary small" @click="router.push(`/certificate/${cert.cert_no}`)">查看证书</button>
            <button class="btn-secondary small" @click="router.push(`/verify?no=${cert.cert_no}`)">验证</button>
          </div>
        </div>
      </div>
    </div>

    <button class="btn-ghost back-btn" @click="router.push('/profile')">← 返回个人中心</button>
  </div>
</template>

<style scoped>
.certs-page { padding-top: 24px; padding-bottom: 40px; }
.page-title { font-size: 24px; margin-bottom: 4px; }
.page-sub { color: var(--text-light); font-size: 14px; margin-bottom: 20px; }
.loading { text-align: center; padding: 40px; color: var(--text-light); }

.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 32px 24px;
  box-shadow: 0 4px 0 #e5e5e5;
}
.empty { text-align: center; }
.empty-icon { font-size: 56px; display: block; margin-bottom: 8px; }
.empty-text { font-size: 18px; font-weight: 800; margin-bottom: 4px; }
.empty-sub { font-size: 13px; color: var(--text-light); margin-bottom: 20px; }

.cert-list { display: flex; flex-direction: column; gap: 14px; }
.cert-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 18px 20px;
  box-shadow: 0 4px 0 #e5e5e5;
}
.cert-card-head { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.cert-medal { font-size: 34px; }
.cert-card-info { flex: 1; min-width: 0; }
.cert-course { font-size: 17px; font-weight: 900; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cert-meta { font-size: 12px; color: var(--text-light); margin-top: 2px; }

.level-badge {
  padding: 4px 12px;
  border-radius: 14px;
  font-size: 13px;
  font-weight: 800;
  color: white;
  flex-shrink: 0;
}
.lv-beginner { background: #10b981; }
.lv-intermediate { background: #3b82f6; }
.lv-advanced { background: #f59e0b; }

.cert-card-foot { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.cert-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary, .btn-ghost {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 4px 0 var(--primary-dark); }
.btn-secondary {
  background: white;
  color: var(--text);
  border: 2px solid var(--border);
  box-shadow: 0 4px 0 #e5e5e5;
}
.btn-ghost { background: transparent; color: var(--text-light); }
.small { padding: 8px 14px; font-size: 13px; box-shadow: 0 3px 0 #e5e5e5; }
.back-btn { margin-top: 24px; width: 100%; }
</style>
