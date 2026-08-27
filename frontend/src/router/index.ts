import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
    },
    {
      path: '/course/:id',
      name: 'path',
      component: () => import('@/views/PathView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/lesson/:id',
      name: 'lesson',
      component: () => import('@/views/LessonView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/course/:id/exam',
      name: 'exam',
      component: () => import('@/views/ExamView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/wrong-exercises',
      name: 'wrong-exercises',
      component: () => import('@/views/WrongExercisesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/adaptive',
      name: 'adaptive',
      component: () => import('@/views/AdaptiveView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/tutor',
      name: 'tutor',
      component: () => import('@/views/TutorView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/knowledge',
      name: 'knowledge',
      component: () => import('@/views/KnowledgeView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { name: 'login' }
  }
  if ((to.name === 'login' || to.name === 'register') && auth.isLoggedIn) {
    return { name: 'home' }
  }
})

export default router
