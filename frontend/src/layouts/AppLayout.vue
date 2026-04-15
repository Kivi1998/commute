<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  HomeOutlined,
  BankOutlined,
  CarOutlined,
  HistoryOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()

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
