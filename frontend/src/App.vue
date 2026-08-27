<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/client'
import StatsBar from '@/components/StatsBar.vue'

const auth = useAuthStore()

onMounted(async () => {
  if (auth.isLoggedIn) {
    try {
      const res = await userApi.stats()
      auth.setStats(res.data)
    } catch (e) {
      // ignore
    }
  }
})
</script>

<template>
  <div class="app-wrapper">
    <StatsBar v-if="auth.isLoggedIn" />
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<style scoped>
.app-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
