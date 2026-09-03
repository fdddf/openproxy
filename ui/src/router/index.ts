import { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/Setup.vue'),
    meta: {
      requiresAuth: false,
      layout: 'empty'
    }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: {
      requiresAuth: false,
      layout: 'empty'
    }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/keys',
    name: 'API Keys',
    component: () => import('@/views/APIKeys.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/providers',
    name: 'Providers',
    component: () => import('@/views/Providers.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/models',
    name: 'Models',
    component: () => import('@/views/Models.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/Users.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/requests',
    name: 'Requests',
    component: () => import('@/views/Requests.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('@/views/Logs.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/Profile.vue'),
    meta: {
      requiresAuth: true,
      layout: 'main'
    }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: {
      requiresAuth: false,
      layout: 'empty'
    }
  }
]

export default routes
