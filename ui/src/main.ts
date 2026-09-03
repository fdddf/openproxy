import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import { apiService } from '@/api'
import './assets/styles/main.css'

// Import routes
import routes from './router'

const app = createApp(App)
const pinia = createPinia()
const router = createRouter({
  history: createWebHistory(),
  routes,
})

// A fresh instance has no administrator yet. Checked once and cached, so normal
// navigation does not pay for an extra request on every route change.
let needsSetup: boolean | null = null

async function checkNeedsSetup(): Promise<boolean> {
  if (needsSetup !== null) return needsSetup
  try {
    const status = await apiService.getSetupStatus()
    needsSetup = status.needsSetup
  } catch {
    // If the check fails, fall through to the normal auth flow rather than
    // trapping the user on the setup screen.
    needsSetup = false
  }
  return needsSetup
}

// Global guards
router.beforeEach(async (to, _from, next) => {
  const isAuthenticated = localStorage.getItem('token')

  if (to.path === '/setup') {
    next()
    return
  }

  if (!isAuthenticated && await checkNeedsSetup()) {
    next('/setup')
    return
  }

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

app.use(pinia)
app.use(router)
app.mount('#app')