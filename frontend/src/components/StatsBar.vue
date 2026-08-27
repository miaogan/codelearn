<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { wrongApi } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const wrongCount = ref(0)

async function loadWrongCount() {
  try {
    const res = await wrongApi.count()
    wrongCount.value = res.data.count
  } catch {
    // ignore
  }
}

onMounted(loadWrongCount)
</script>

<template>
  <header class="stats-bar">
    <div class="stats-content">
      <router-link to="/" class="logo">📚 CodeLearn</router-link>
      <div class="stats-items">
        <div class="stat-item ai-tutor" title="AI 编程导师" @click="router.push('/tutor')">
          <span class="icon">🤖</span>
        </div>
        <div class="stat-item knowledge" title="知识点问答" @click="router.push('/knowledge')">
          <span class="icon">💡</span>
        </div>
        <div class="stat-item adaptive" title="自适应学习" @click="router.push('/adaptive')">
          <span class="icon">📊</span>
        </div>
        <div class="stat-item streak" title="连续打卡">
          <span class="icon">🔥</span>
          <span class="value">{{ auth.stats.streak_days }}</span>
        </div>
        <div class="stat-item xp" title="经验值" @click="router.push('/profile')">
          <span class="icon">💎</span>
          <span class="value">{{ auth.stats.xp }}</span>
        </div>
        <div class="stat-item hearts" title="心数">
          <span class="icon">❤️</span>
          <span class="value">{{ auth.stats.hearts }}/{{ auth.stats.max_hearts }}</span>
        </div>
        <div class="stat-item wrong" title="错题本" @click="router.push('/wrong-exercises')">
          <span class="icon">📝</span>
          <span class="value" :class="{ 'has-wrong': wrongCount > 0 }">{{ wrongCount }}</span>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.stats-bar {
  background: white;
  border-bottom: 2px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.stats-content {
  max-width: 700px;
  margin: 0 auto;
  padding: 12px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 22px;
  font-weight: 900;
  color: var(--primary);
}

.stats-items {
  display: flex;
  gap: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 18px;
}

.stat-item .icon {
  font-size: 20px;
}

.stat-item .value {
  font-weight: 800;
  color: var(--text);
}

.stat-item.ai-tutor, .stat-item.knowledge, .stat-item.adaptive {
  cursor: pointer;
  font-size: 20px;
}

.stat-item.xp .icon {
  filter: hue-rotate(60deg);
}

.stat-item.wrong {
  cursor: pointer;
}

.stat-item.wrong .value.has-wrong {
  color: #ef4444;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.15); }
}
</style>
