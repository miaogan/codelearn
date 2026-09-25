import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')

// PWA：注册 Service Worker（生产环境，桌面 WebView 下跳过）
const isDesktop = 'wails' in window
if ('serviceWorker' in navigator && import.meta.env.PROD && !isDesktop) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch((e) => {
      console.error('Service Worker 注册失败', e)
    })
  })
}
