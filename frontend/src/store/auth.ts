import { defineStore } from 'pinia'
import { login as apiLogin, fetchMe, type LoginInput, type User } from '@/api/auth'

const TOKEN_KEY = 'commute.token'
const USER_KEY = 'commute.user'

function loadUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    user: loadUser(),
    loading: false,
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
    displayName: (s) => s.user?.name || s.user?.email || '',
  },
  actions: {
    async login(payload: LoginInput) {
      this.loading = true
      try {
        const r = await apiLogin(payload)
        this.token = r.token
        this.user = r.user
        localStorage.setItem(TOKEN_KEY, r.token)
        localStorage.setItem(USER_KEY, JSON.stringify(r.user))
        return r
      } finally {
        this.loading = false
      }
    },
    async refreshMe() {
      if (!this.token) return
      try {
        const u = await fetchMe()
        this.user = u
        localStorage.setItem(USER_KEY, JSON.stringify(u))
      } catch {
        // 401 会被拦截器处理（clear + redirect），这里不需重复
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    },
  },
})
