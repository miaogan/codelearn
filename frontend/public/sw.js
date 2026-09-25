const CACHE_NAME = 'codelearn-v1'
const APP_SHELL = [
  '/',
  '/manifest.webmanifest',
  '/icons/icon.svg',
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(APP_SHELL)).then(() => self.skipWaiting())
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k)))
    ).then(() => self.clients.claim())
  )
})

// 静态资源：缓存优先
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)

  if (event.request.method !== 'GET') return
  if (url.origin !== self.location.origin) return

  // API 请求不缓存，走网络
  if (url.pathname.startsWith('/api/')) return

  // 页面导航：网络优先，离线回退到缓存首页
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request)
        .then((res) => {
          const copy = res.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, copy))
          return res
        })
        .catch(() => caches.match('/'))
    )
    return
  }

  // 其他同源静态资源：缓存优先，网络回填
  event.respondWith(
    caches.match(event.request).then((cached) => {
      if (cached) return cached
      return fetch(event.request).then((res) => {
        if (res.ok) {
          const copy = res.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, copy))
        }
        return res
      })
    })
  )
})

// 学习提醒（本地通知）：支持每日打卡提醒
self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clients) => {
      for (const client of clients) {
        if ('focus' in client) {
          client.focus()
          client.navigate(event.notification.data?.url || '/')
          return
        }
      }
      return self.clients.openWindow(event.notification.data?.url || '/')
    })
  )
})

// 服务端推送预留
self.addEventListener('push', (event) => {
  let data = { title: 'CodeLearn', body: '今天也要坚持学习哦 💪' }
  try {
    if (event.data) data = event.data.json()
  } catch (e) {
    // ignore
  }
  event.waitUntil(
    self.registration.showNotification(data.title || 'CodeLearn', {
      body: data.body || '今天也要坚持学习哦 💪',
      icon: '/icons/icon.svg',
      badge: '/icons/icon.svg',
      data: { url: data.url || '/' },
    })
  )
})
