<script setup lang="ts">
import { ref, onMounted } from 'vue'

const deferredPrompt = ref<any>(null)
const showInstall = ref(false)
const notifyOn = ref(false)
// 桌面版（Wails WebView）无需 PWA 安装/通知横幅
const isDesktop = 'wails' in window

onMounted(async () => {
  if (isDesktop) return
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault()
    deferredPrompt.value = e
    showInstall.value = true
  })
  window.addEventListener('appinstalled', () => {
    showInstall.value = false
  })

  if ('Notification' in window) {
    notifyOn.value = Notification.permission === 'granted'
  }
  if (notifyOn.value) {
    tryDailyReminder()
  }
})

async function install() {
  if (!deferredPrompt.value) return
  deferredPrompt.value.prompt()
  const choice = await deferredPrompt.value.userChoice
  if (choice.outcome === 'accepted') showInstall.value = false
  deferredPrompt.value = null
}

async function enableReminder() {
  if (!('Notification' in window) || !navigator.serviceWorker) {
    alert('当前浏览器不支持通知功能')
    return
  }
  const perm = await Notification.requestPermission()
  notifyOn.value = perm === 'granted'
  if (notifyOn.value) {
    localStorage.setItem('learnReminder', '1')
    sendReminder()
    alert('已开启学习提醒！每次打开应用时都会收到打卡提醒。')
  }
}

async function sendReminder() {
  const reg = await navigator.serviceWorker.getRegistration()
  if (!reg) return
  await reg.showNotification('CodeLearn 📚', {
    body: '今天也要坚持学习哦！完成每日目标即可打卡 🔥',
    icon: '/icons/icon.svg',
    badge: '/icons/icon.svg',
    data: { url: '/' },
  })
  localStorage.setItem('lastReminderDate', new Date().toDateString())
}

async function tryDailyReminder() {
  if (!localStorage.getItem('learnReminder')) return
  const last = localStorage.getItem('lastReminderDate')
  const today = new Date().toDateString()
  if (last !== today) {
    await sendReminder()
  }
}
</script>

<template>
  <div v-if="!isDesktop && showInstall" class="pwa-banner">
    <span class="pwa-icon">📚</span>
    <div class="pwa-body">
      <div class="pwa-title">安装 CodeLearn</div>
      <div class="pwa-desc">添加到主屏幕，随时开始学习</div>
    </div>
    <button class="pwa-btn" @click="install">安装</button>
  </div>

  <div v-if="!isDesktop && !notifyOn" class="pwa-banner reminder">
    <span class="pwa-icon">🔔</span>
    <div class="pwa-body">
      <div class="pwa-title">开启学习提醒</div>
      <div class="pwa-desc">每天打开时提醒你坚持打卡</div>
    </div>
    <button class="pwa-btn" @click="enableReminder">开启</button>
  </div>
</template>

<style scoped>
.pwa-banner {
  position: fixed;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  display: flex;
  align-items: center;
  gap: 12px;
  width: calc(100% - 32px);
  max-width: 460px;
  background: white;
  border: 2px solid var(--primary);
  border-radius: var(--radius);
  padding: 12px 16px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.15);
}
.pwa-banner.reminder {
  bottom: 76px;
  border-color: #f59e0b;
}
.pwa-icon { font-size: 28px; flex-shrink: 0; }
.pwa-body { flex: 1; min-width: 0; }
.pwa-title { font-size: 14px; font-weight: 800; }
.pwa-desc { font-size: 12px; color: var(--text-light); }
.pwa-btn {
  border: none;
  background: var(--primary);
  color: white;
  font-size: 13px;
  font-weight: 800;
  padding: 8px 16px;
  border-radius: 12px;
  cursor: pointer;
  flex-shrink: 0;
}
.reminder .pwa-btn { background: #f59e0b; }
</style>
