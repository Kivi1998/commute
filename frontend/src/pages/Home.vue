<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchHealth, type HealthResponse } from '@/api/health'

const loading = ref(true)
const health = ref<HealthResponse | null>(null)
const error = ref<string>('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    health.value = await fetchHealth()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

function depStatusColor(status: string) {
  if (status === 'ok') return 'green'
  if (status === 'configured') return 'blue'
  if (status === 'not_configured') return 'orange'
  return 'red'
}
</script>

<template>
  <div class="max-w-4xl mx-auto space-y-4">
    <a-card title="欢迎使用通勤查询" :bordered="false">
      <p class="text-slate-600">
        本工具帮助你基于<b>家庭住址</b>与<b>目标公司</b>，计算真实的周一早晚高峰通勤时间，
        辅助求职与租房决策。
      </p>
      <div class="mt-4 flex gap-2">
        <a-button type="primary" @click="$router.push('/settings')">设置我的信息</a-button>
        <a-button @click="$router.push('/commute')">开始对比通勤</a-button>
      </div>
    </a-card>

    <a-card title="后端健康状态" :bordered="false">
      <a-spin :spinning="loading">
        <a-alert v-if="error" type="error" :message="error" class="mb-3" />
        <a-descriptions v-if="health" bordered size="small" :column="1">
          <a-descriptions-item label="总体状态">
            <a-tag :color="health.status === 'ok' ? 'green' : 'orange'">
              {{ health.status.toUpperCase() }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="版本">{{ health.version }}</a-descriptions-item>
          <a-descriptions-item label="运行时长">{{ health.uptime_seconds }} 秒</a-descriptions-item>
          <a-descriptions-item label="依赖">
            <div class="flex flex-wrap gap-2">
              <a-tag
                v-for="(v, k) in health.dependencies"
                :key="k"
                :color="depStatusColor(v)"
              >
                {{ k }}: {{ v }}
              </a-tag>
            </div>
          </a-descriptions-item>
        </a-descriptions>
        <a-button class="mt-3" size="small" @click="load">刷新</a-button>
      </a-spin>
    </a-card>
  </div>
</template>
