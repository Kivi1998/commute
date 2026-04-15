<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import CommuteConfigPanel from '@/components/CommuteConfigPanel.vue'
import CommuteMap from '@/components/CommuteMap.vue'
import CommuteResultsTable from '@/components/CommuteResultsTable.vue'
import CommuteCompareView from '@/components/CommuteCompareView.vue'
import {
  calculateCommute,
  getCommuteQuery,
  listResultsByQuery,
  type CommuteCalculateInput,
  type CommuteCalculateResponse,
  type TransportMode,
} from '@/api/commute'

const route = useRoute()
const router = useRouter()

const configRef = ref<InstanceType<typeof CommuteConfigPanel> | null>(null)
const response = ref<CommuteCalculateResponse | null>(null)
const highlightId = ref<number | null>(null)
const calculating = ref(false)
const viewMode = ref<'list' | 'bar' | 'scatter' | 'radar'>('list')

async function handleCalculate(payload: CommuteCalculateInput) {
  calculating.value = true
  configRef.value?.setLoading(true)
  try {
    const raw = await calculateCommute(payload)
    const r: CommuteCalculateResponse =
      'result' in raw ? raw.result : (raw as CommuteCalculateResponse)
    if ('warning' in raw && raw.warning) {
      message.warning('已计算，建议公司数量不超过 20 家')
    }
    response.value = r
    const s = r.summary
    message.success(
      `计算完成：${s.total_companies} 家公司（缓存命中 ${s.cache_hits}，失败 ${s.failures}）`,
    )
  } finally {
    calculating.value = false
    configRef.value?.setLoading(false)
  }
}

// 从 /commute?from_query=N 触发历史恢复
async function restoreFromQuery(queryId: number) {
  try {
    const [q, results] = await Promise.all([
      getCommuteQuery(queryId),
      listResultsByQuery(queryId),
    ])
    // 反推 company_ids
    const companyIds = Array.from(new Set(results.map((r) => r.company_id)))
    if (!companyIds.length) {
      message.warning('该查询没有保存的结果，请重新计算')
    }
    // 等 ConfigPanel 的 onMounted 加载完数据再 apply（setTimeout 零延迟让事件环空一下）
    setTimeout(() => {
      configRef.value?.applyQuery({
        home_id: q.home_id,
        transport_modes: q.transport_modes as TransportMode[],
        morning_time: q.morning_time,
        evening_time: q.evening_time,
        buffer_minutes: q.buffer_minutes,
        company_ids: companyIds,
      })
      message.success('已加载历史参数，点「计算通勤」即可重算')
    }, 500)
  } catch {
    // 拦截器已提示
  }
}

onMounted(() => {
  const id = Number(route.query.from_query)
  if (id) {
    restoreFromQuery(id)
    // 清掉 query string 避免刷新重复触发
    router.replace({ path: '/commute' })
  }
})

watch(
  () => route.query.from_query,
  (v) => {
    const id = Number(v)
    if (id) {
      restoreFromQuery(id)
      router.replace({ path: '/commute' })
    }
  },
)
</script>

<template>
  <div class="max-w-7xl mx-auto space-y-4">
    <CommuteConfigPanel ref="configRef" @calculate="handleCalculate" />

    <div v-if="response" class="grid grid-cols-1 lg:grid-cols-5 gap-4">
      <div class="lg:col-span-3">
        <a-card title="地图总览" :bordered="false" size="small">
          <template #extra>
            <a-space size="small" class="text-xs text-slate-400">
              <span class="flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-blue-600"></span>家
              </span>
              <span class="flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-green-600"></span>≤30min
              </span>
              <span class="flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-amber-500"></span>≤60min
              </span>
              <span class="flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-red-600"></span>&gt;60min
              </span>
            </a-space>
          </template>
          <CommuteMap
            :home="response.home"
            :companies="response.results"
            :highlight-company-id="highlightId"
            height="560px"
          />
        </a-card>
      </div>

      <div class="lg:col-span-2">
        <a-card :bordered="false" size="small">
          <template #title>
            <a-radio-group v-model:value="viewMode" size="small" button-style="solid">
              <a-radio-button value="list">📋 列表</a-radio-button>
              <a-radio-button value="bar">📊 柱状</a-radio-button>
              <a-radio-button value="scatter">🔵 散点</a-radio-button>
              <a-radio-button value="radar">🕸 雷达</a-radio-button>
            </a-radio-group>
          </template>
          <template #extra>
            <span v-if="viewMode === 'list'" class="text-xs text-slate-400">
              悬停看地图高亮
            </span>
          </template>
          <CommuteResultsTable
            v-if="viewMode === 'list'"
            :results="response.results"
            @highlight="(id) => (highlightId = id)"
          />
          <CommuteCompareView
            v-else
            :results="response.results"
            :mode="viewMode"
          />
        </a-card>
      </div>
    </div>

    <a-spin
      v-else-if="calculating"
      tip="正在调用高德 API 批量计算..."
      size="large"
      class="block py-20"
    />
  </div>
</template>
