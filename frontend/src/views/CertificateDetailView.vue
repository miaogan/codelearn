<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QRCode from 'qrcode'
import { certificateApi } from '@/api/client'
import type { Certificate } from '@/types'

const route = useRoute()
const router = useRouter()
const cert = ref<Certificate | null>(null)
const loading = ref(true)
const errorMsg = ref('')
const qrCanvas = ref<HTMLCanvasElement | null>(null)
const sharing = ref(false)

const certNo = String(route.params.no || '')

onMounted(async () => {
  try {
    const res = await certificateApi.verify(certNo)
    if (res.data.valid && res.data.certificate) {
      cert.value = res.data.certificate
      await renderQR()
    } else {
      errorMsg.value = '未找到该证书'
    }
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || '加载证书失败'
  }
  loading.value = false
})

const verifyUrl = () => `${location.origin}/verify?no=${certNo}`

async function renderQR() {
  if (!qrCanvas.value) return
  await QRCode.toCanvas(qrCanvas.value, verifyUrl(), { width: 140, margin: 1, color: { dark: '#78350f', light: '#ffffff' } })
}

function levelText(level: string) {
  const map: Record<string, string> = { beginner: '入门', intermediate: '进阶', advanced: '熟练' }
  return map[level] || level
}

function fmtDate(iso: string) {
  return iso ? iso.slice(0, 10) : ''
}

// 生成分享图：将证书信息绘制到 Canvas 并下载
async function downloadShare() {
  if (!cert.value) return
  sharing.value = true
  try {
    const W = 840
    const H = 560
    const canvas = document.createElement('canvas')
    canvas.width = W
    canvas.height = H
    const ctx = canvas.getContext('2d')!
    const c = cert.value

    const grad = ctx.createLinearGradient(0, 0, W, H)
    grad.addColorStop(0, '#fff8e1')
    grad.addColorStop(1, '#ffe0b2')
    ctx.fillStyle = grad
    ctx.fillRect(0, 0, W, H)

    ctx.strokeStyle = '#d97706'
    ctx.lineWidth = 8
    ctx.strokeRect(20, 20, W - 40, H - 40)
    ctx.strokeStyle = '#f59e0b'
    ctx.lineWidth = 2
    ctx.strokeRect(36, 36, W - 72, H - 72)

    ctx.textAlign = 'center'
    ctx.fillStyle = '#92400e'
    ctx.font = 'bold 46px sans-serif'
    ctx.fillText('能力认证证书', W / 2, 118)

    ctx.fillStyle = '#78350f'
    ctx.font = '22px sans-serif'
    ctx.fillText(`兹证明 ${c.user_name} 已完成本课程学习并通过认证考试`, W / 2, 178)

    ctx.fillStyle = '#92400e'
    ctx.font = 'bold 34px sans-serif'
    ctx.fillText(c.course_title, W / 2, 240)

    ctx.fillStyle = '#78350f'
    ctx.font = '22px sans-serif'
    ctx.fillText(`认证等级：${levelText(c.level)} · 成绩 ${c.score} 分`, W / 2, 296)

    ctx.fillStyle = '#92400e'
    ctx.font = '20px monospace'
    ctx.fillText(`证书编号：${c.cert_no}`, W / 2, 342)
    ctx.fillText(`颁发日期：${fmtDate(c.issued_at)}`, W / 2, 378)

    const qr = document.createElement('canvas')
    await QRCode.toCanvas(qr, verifyUrl(), { width: 130, margin: 1, color: { dark: '#78350f', light: '#ffffff' } })
    ctx.drawImage(qr, W / 2 - 65, 400, 130, 130)

    const a = document.createElement('a')
    a.href = canvas.toDataURL('image/png')
    a.download = `证书-${c.cert_no}.png`
    a.click()
  } finally {
    sharing.value = false
  }
}
</script>

<template>
  <div class="container cert-page">
    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="errorMsg || !cert" class="card error-card">
      <h2>😢 {{ errorMsg || '证书不存在' }}</h2>
      <button class="btn-primary" @click="router.push('/certificates')">返回我的证书</button>
    </div>

    <template v-else>
      <div class="certificate-card">
        <div class="cert-border">
          <div class="cert-inner">
            <div class="cert-seal">🎖️</div>
            <h1 class="cert-h1">能力认证证书</h1>
            <p class="cert-line">兹证明 <strong>{{ cert.user_name }}</strong> 已完成本课程学习并通过认证考试</p>
            <div class="cert-course">{{ cert.course_title }}</div>
            <div class="cert-details">
              <span class="level-badge" :class="'lv-' + cert.level">{{ levelText(cert.level) }}</span>
              <span class="cert-score">成绩 {{ cert.score }} 分</span>
            </div>
            <div class="cert-no">编号 {{ cert.cert_no }}</div>
            <div class="cert-bottom">
              <div class="cert-date">颁发日期<br />{{ fmtDate(cert.issued_at) }}</div>
              <div class="cert-qr">
                <canvas ref="qrCanvas"></canvas>
                <p class="qr-tip">扫码验证真伪</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="cert-actions">
        <button class="btn-primary" :disabled="sharing" @click="downloadShare">
          {{ sharing ? '生成中...' : '📤 生成分享图' }}
        </button>
        <button class="btn-secondary" @click="router.push(`/verify?no=${cert!.cert_no}`)">🔍 在线验证</button>
      </div>
      <button class="btn-ghost back-btn" @click="router.push('/certificates')">← 返回我的证书</button>
    </template>
  </div>
</template>

<style scoped>
.cert-page { padding-top: 24px; padding-bottom: 40px; }
.loading { text-align: center; padding: 40px; color: var(--text-light); }
.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 32px 24px;
  box-shadow: 0 4px 0 #e5e5e5;
  text-align: center;
}
.error-card h2 { margin-bottom: 20px; font-size: 18px; }

.certificate-card {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}
.cert-border {
  background: linear-gradient(135deg, #f59e0b, #d97706, #f59e0b);
  padding: 10px;
  border-radius: 14px;
  box-shadow: 0 8px 0 #b45309;
  width: 100%;
  max-width: 560px;
}
.cert-inner {
  background: linear-gradient(160deg, #fffbeb 0%, #fff7ed 60%, #ffedd5 100%);
  border: 2px solid #fbbf24;
  border-radius: 8px;
  padding: 32px 28px;
  text-align: center;
  position: relative;
}
.cert-seal { font-size: 52px; margin-bottom: 6px; }
.cert-h1 {
  font-size: 32px;
  letter-spacing: 8px;
  color: #92400e;
  margin-bottom: 14px;
}
.cert-line { font-size: 14px; color: #78350f; margin-bottom: 16px; }
.cert-line strong { font-size: 17px; }
.cert-course {
  font-size: 24px;
  font-weight: 900;
  color: #92400e;
  margin-bottom: 14px;
}
.cert-details {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 12px;
}
.level-badge {
  padding: 4px 14px;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 800;
  color: white;
}
.lv-beginner { background: #10b981; }
.lv-intermediate { background: #3b82f6; }
.lv-advanced { background: #f59e0b; }
.cert-score { font-size: 14px; color: #78350f; font-weight: 700; }
.cert-no {
  font-size: 14px;
  font-family: 'SF Mono', Consolas, monospace;
  color: #92400e;
  margin-bottom: 14px;
}
.cert-bottom {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 0 8px;
}
.cert-date { font-size: 13px; color: #78350f; text-align: left; line-height: 1.6; }
.cert-qr { text-align: center; }
.cert-qr canvas { border: 2px solid #fcd34d; border-radius: 6px; }
.qr-tip { font-size: 11px; color: #92400e; margin-top: 4px; }

.cert-actions { display: flex; gap: 12px; }
.cert-actions > * { flex: 1; }
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
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary {
  background: white;
  color: var(--text);
  border: 2px solid var(--border);
  box-shadow: 0 4px 0 #e5e5e5;
}
.btn-ghost { background: transparent; color: var(--text-light); }
.back-btn { margin-top: 24px; width: 100%; }
</style>
