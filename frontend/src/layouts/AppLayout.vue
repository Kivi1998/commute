<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Modal } from 'ant-design-vue'
import {
  HomeOutlined,
  BankOutlined,
  CarOutlined,
  HistoryOutlined,
  SettingOutlined,
  LogoutOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/store/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const menuItems = [
  { key: 'home', label: '首页', icon: HomeOutlined, path: '/' },
  { key: 'companies', label: '公司', icon: BankOutlined, path: '/companies' },
  { key: 'commute', label: '通勤', icon: CarOutlined, path: '/commute' },
  { key: 'history', label: '历史', icon: HistoryOutlined, path: '/history' },
  { key: 'settings', label: '设置', icon: SettingOutlined, path: '/settings' },
]

const selectedKeys = computed(() => [(route.name as string) || 'home'])

function handleSelect(key: string) {
  const item = menuItems.find((m) => m.key === key)
  if (item) router.push(item.path)
}

function handleLogout() {
  Modal.confirm({
    title: '确认退出登录？',
    content: '下次使用需要重新登录。',
    okType: 'danger',
    okText: '退出',
    cancelText: '取消',
    onOk: () => {
      auth.logout()
      router.push('/login')
    },
  })
}

onMounted(() => {
  // 启动时拉最新用户信息（token 过期会触发 401 → 拦截器跳登录）
  auth.refreshMe()
})
</script>

<template>
  <a-layout class="min-h-screen">
    <a-layout-header class="!bg-white !px-6 shadow-sm flex items-center">
      <div class="text-lg font-semibold text-slate-800 mr-8">
        🚇 通勤查询
      </div>
      <a-menu
        mode="horizontal"
        :selected-keys="selectedKeys"
        class="flex-1 border-b-0"
        @select="(info: { key: string | number }) => handleSelect(String(info.key))"
      >
        <a-menu-item v-for="item in menuItems" :key="item.key">
          <component :is="item.icon" />
          <span>{{ item.label }}</span>
        </a-menu-item>
      </a-menu>

      <a-dropdown v-if="auth.user">
        <div class="flex items-center gap-2 cursor-pointer px-3 py-1 rounded hover:bg-slate-50">
          <a-avatar size="small" class="!bg-blue-500">
            <template #icon><UserOutlined /></template>
          </a-avatar>
          <span class="text-sm text-slate-700">{{ auth.displayName }}</span>
        </div>
        <template #overlay>
          <a-menu>
            <a-menu-item disabled>
              <div class="text-xs text-slate-400">{{ auth.user?.email }}</div>
            </a-menu-item>
            <a-menu-divider />
            <a-menu-item key="settings" @click="router.push('/settings')">
              <SettingOutlined /> 个人设置
            </a-menu-item>
            <a-menu-item key="logout" danger @click="handleLogout">
              <LogoutOutlined /> 退出登录
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </a-layout-header>

    <a-layout-content class="p-6 bg-slate-50">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </a-layout-content>

    <a-layout-footer class="!text-center !bg-white text-slate-500 text-sm">
      Commute © 2026 · 基于高德地图与豆包 AI
    </a-layout-footer>
  </a-layout>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
