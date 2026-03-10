<template>
  <div class="flex h-screen bg-slate-100">
    <!-- Sidebar -->
    <aside class="w-[17rem] flex-shrink-0 bg-gradient-to-b from-slate-900 via-slate-900 to-slate-950 text-slate-100 shadow-2xl relative overflow-hidden">
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(circle_at_20%_20%,#6366f1,transparent_25%),radial-gradient(circle_at_80%_0%,#22d3ee,transparent_20%)]"></div>
      <div class="relative z-10">
        <div class="h-16 flex items-center justify-between px-5 border-b border-white/10">
          <div class="flex items-center space-x-3">
            <div class="w-9 h-9 rounded-xl bg-indigo-500 flex items-center justify-center text-white shadow-lg shadow-indigo-500/40">
              <i class="fas fa-diagram-project"></i>
            </div>
            <div>
              <p class="text-xs uppercase tracking-[0.25em] text-indigo-200">GPT Proxy</p>
              <p class="text-sm font-semibold">Control Plane</p>
            </div>
          </div>
          <span class="text-[10px] px-2 py-1 rounded-full bg-white/10 border border-white/10">v2</span>
        </div>

        <nav class="p-4 space-y-1">
          <RouterLink 
            v-for="item in navItems" 
            :key="item.path" 
            :to="item.path" 
            class="nav-link"
            :class="{ active: currentRoute === item.path }"
          >
            <i :class="item.icon + ' mr-3'"></i>
            {{ item.label }}
            <span v-if="item.path === '/users'" class="ml-auto text-[10px] px-2 py-1 rounded-full bg-indigo-100 text-indigo-700">new</span>
          </RouterLink>
        </nav>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Header -->
      <header class="h-16 bg-white/80 backdrop-blur-sm border-b border-slate-200 flex items-center justify-between px-6 sticky top-0 z-20">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
            <i class="fas fa-wave-square"></i>
          </div>
          <h1 class="text-xl font-semibold text-slate-900">{{ pageTitle }}</h1>
        </div>
        <div class="flex items-center space-x-4">
          <div class="hidden lg:flex items-center space-x-3 px-3 py-2 rounded-lg bg-slate-50 border border-slate-200 text-xs text-slate-600">
            <i class="fas fa-satellite-dish text-indigo-500"></i>
            <span>Realtime Proxy</span>
          </div>
          <div class="relative" ref="menuRef">
            <button class="flex items-center space-x-2 focus:outline-none rounded-full px-2 py-1 hover:bg-slate-100" @click="toggleUserMenu">
              <img 
                :src="avatarUrl"
                alt="User Avatar" 
                class="w-9 h-9 rounded-full border-2 border-white shadow-sm object-cover"
              >
              <div class="hidden md:block text-left">
                <p class="text-sm font-semibold text-slate-800 leading-tight">{{ displayName }}</p>
                <p class="text-[11px] text-slate-500 leading-tight">{{ authStore.user?.email || 'admin@local' }}</p>
              </div>
              <i class="fas fa-chevron-down text-xs text-slate-500"></i>
            </button>
            
            <!-- Dropdown Menu -->
            <div 
              v-if="showUserMenu"
              class="absolute right-0 mt-2 w-52 bg-white rounded-xl shadow-xl border border-gray-100 z-50 overflow-hidden"
            >
              <div class="py-2">
                <RouterLink 
                  to="/profile" 
                  class="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors duration-150"
                  @click="showUserMenu = false"
                >
                  <i class="fas fa-user mr-2 text-indigo-500"></i>
                  Profile
                </RouterLink>
                <RouterLink 
                  to="/settings" 
                  class="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors duration-150"
                  @click="showUserMenu = false"
                >
                  <i class="fas fa-sliders-h mr-2 text-indigo-500"></i>
                  Settings
                </RouterLink>
                <button 
                  @click="logout"
                  class="w-full flex items-center text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors duration-150"
                >
                  <i class="fas fa-sign-out-alt mr-2"></i>
                  Logout
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>
      
      <!-- Page Content -->
      <main class="flex-1 overflow-y-auto p-6 bg-gradient-to-br from-slate-50 via-white to-indigo-50/40">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

const navItems = [
  { path: '/', label: 'Dashboard', icon: 'fas fa-gauge-high' },
  { path: '/users', label: 'Users', icon: 'fas fa-users' },
  { path: '/profile', label: 'Profile', icon: 'fas fa-id-card' },
  { path: '/keys', label: 'API Keys', icon: 'fas fa-key' },
  { path: '/providers', label: 'Providers', icon: 'fas fa-server' },
  { path: '/models', label: 'Models', icon: 'fas fa-cubes' },
  { path: '/requests', label: 'Requests', icon: 'fas fa-wave-square' },
  { path: '/logs', label: 'Logs', icon: 'fas fa-list' },
  { path: '/settings', label: 'Settings', icon: 'fas fa-sliders-h' }
]

const currentRoute = computed(() => router.currentRoute.value.path)
const showUserMenu = ref(false)
const menuRef = ref<HTMLElement | null>(null)

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': 'Dashboard',
    '/profile': 'Profile',
    '/keys': 'API Keys',
    '/providers': 'Providers',
    '/models': 'Models',
    '/requests': 'Requests',
    '/logs': 'Logs',
    '/settings': 'Settings',
    '/users': 'Users'
  }
  return titles[currentRoute.value] || 'GPT Proxy Admin'
})

const displayName = computed(() => authStore.user?.displayName || authStore.user?.username || 'Admin')
const avatarUrl = computed(() => authStore.user?.avatarUrl || fallbackAvatar(displayName.value))

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
}

const logout = () => {
  authStore.logout()
  router.push('/login')
}

const handleClickOutside = (event: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
    showUserMenu.value = false
  }
}

const fallbackAvatar = (seed: string) => {
  return `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(seed)}`
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.nav-link {
  display: flex;
  align-items: center;
  padding: 12px 14px;
  border-radius: 10px;
  color: #cbd5e1;
  text-decoration: none;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s ease;
  position: relative;
}

.nav-link:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.nav-link.active {
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.4), rgba(34, 211, 238, 0.2));
  color: #fff;
  box-shadow: 0 10px 30px rgba(99, 102, 241, 0.35);
}

.nav-link i {
  width: 20px;
  text-align: center;
}
</style>
