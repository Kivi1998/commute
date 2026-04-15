<script setup lang="ts">
import { computed } from 'vue'
import {
  CarOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import type { CompanyCommute, CommuteResultItem, TransportMode } from '@/api/commute'

const props = defineProps<{
  results: CompanyCommute[]
}>()

const emit = defineEmits<{
  (e: 'highlight', companyId: number | null): void
}>()

const modeLabel: Record<TransportMode, string> = {
  transit: '🚇 公交/地铁',
  driving: '🚗 驾车',
  cycling: '🚴 骑行',
  walking: '🚶 步行',
}

interface Row {
  mode: TransportMode
  morning?: CommuteResultItem
  evening?: CommuteResultItem
}

function buildRows(cc: CompanyCommute): Row[] {
  const modes = Array.from(new Set(cc.items.map((it) => it.transport_mode)))
  return modes.map((m) => ({
    mode: m,
    morning: cc.items.find((it) => it.transport_mode === m && it.direction === 'to_work'),
    evening: cc.items.find((it) => it.transport_mode === m && it.direction === 'to_home'),
  }))
}

function bestMorning(cc: CompanyCommute): CommuteResultItem | null {
  const items = cc.items.filter((it) => it.direction === 'to_work')
  if (!items.length) return null
  return items.reduce((a, b) => (a.duration_min <= b.duration_min ? a : b))
}

// 把公司按最优早通勤时长升序排列
const sortedResults = computed(() => {
  return [...props.results].sort((a, b) => {
    const da = bestMorning(a)?.duration_min ?? Infinity
    const db = bestMorning(b)?.duration_min ?? Infinity
    return da - db
  })
})

function durationColor(min: number): string {
  if (min <= 30) return 'green'
  if (min <= 60) return 'orange'
  return 'red'
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-for="cc in sortedResults"
      :key="cc.company_id"
      class="bg-white rounded-md shadow-sm p-4 border border-slate-100 hover:shadow-md transition-shadow"
      @mouseenter="emit('highlight', cc.company_id)"
      @mouseleave="emit('highlight', null)"
    >
      <div class="flex justify-between items-start mb-3">
        <div>
          <div class="text-lg font-semibold text-slate-800 flex items-center gap-2">
            <EnvironmentOutlined class="text-red-500" />
            {{ cc.company_name }}
          </div>
        </div>
        <div v-if="bestMorning(cc)" class="text-right">
          <div class="text-xs text-slate-400">最优早通勤</div>
          <div class="text-2xl font-bold">
            <span :class="{
              'text-green-600': bestMorning(cc)!.duration_min <= 30,
              'text-amber-600': bestMorning(cc)!.duration_min > 30 && bestMorning(cc)!.duration_min <= 60,
              'text-red-600': bestMorning(cc)!.duration_min > 60,
            }">
              {{ bestMorning(cc)!.duration_min }}
            </span>
            <span class="text-base text-slate-400 ml-1">min</span>
          </div>
        </div>
      </div>

      <a-table
        :columns="[
          { title: '方式', key: 'mode', width: 140 },
          { title: '早通勤（去）', key: 'morning' },
          { title: '晚通勤（回）', key: 'evening' },
        ]"
        :data-source="buildRows(cc)"
        :pagination="false"
        size="small"
        row-key="mode"
      >
        <template #bodyCell="slotProps">
          <template v-if="slotProps.column.key === 'mode'">
            <span class="font-medium">{{ modeLabel[slotProps.record.mode as TransportMode] }}</span>
          </template>

          <template v-else-if="slotProps.column.key === 'morning'">
            <div v-if="slotProps.record.morning">
              <div class="flex items-center gap-2">
                <a-tag :color="durationColor(slotProps.record.morning.duration_min)" class="!m-0">
                  {{ slotProps.record.morning.duration_min }} min
                </a-tag>
                <span class="text-sm text-slate-600">
                  {{ slotProps.record.morning.distance_km }} km
                </span>
                <ThunderboltOutlined v-if="!slotProps.record.morning.from_cache" class="text-amber-400 text-xs" title="实时计算" />
              </div>
              <div class="text-xs text-slate-400 mt-0.5">
                {{ slotProps.record.morning.depart_time }} → {{ slotProps.record.morning.arrive_time }}
                <span v-if="slotProps.record.morning.cost_yuan !== undefined && slotProps.record.morning.cost_yuan !== null" class="ml-2">
                  ￥{{ slotProps.record.morning.cost_yuan }}
                </span>
                <span v-if="slotProps.record.morning.transfer_count !== undefined && slotProps.record.morning.transfer_count !== null" class="ml-2">
                  换乘 {{ slotProps.record.morning.transfer_count }} 次
                </span>
              </div>
            </div>
            <span v-else class="text-slate-300">—</span>
          </template>

          <template v-else-if="slotProps.column.key === 'evening'">
            <div v-if="slotProps.record.evening">
              <div class="flex items-center gap-2">
                <a-tag :color="durationColor(slotProps.record.evening.duration_min)" class="!m-0">
                  {{ slotProps.record.evening.duration_min }} min
                </a-tag>
                <span class="text-sm text-slate-600">
                  {{ slotProps.record.evening.distance_km }} km
                </span>
                <ThunderboltOutlined v-if="!slotProps.record.evening.from_cache" class="text-amber-400 text-xs" title="实时计算" />
              </div>
              <div class="text-xs text-slate-400 mt-0.5">
                {{ slotProps.record.evening.depart_time }} → {{ slotProps.record.evening.arrive_time }}
                <span v-if="slotProps.record.evening.cost_yuan !== undefined && slotProps.record.evening.cost_yuan !== null" class="ml-2">
                  ￥{{ slotProps.record.evening.cost_yuan }}
                </span>
                <span v-if="slotProps.record.evening.transfer_count !== undefined && slotProps.record.evening.transfer_count !== null" class="ml-2">
                  换乘 {{ slotProps.record.evening.transfer_count }} 次
                </span>
              </div>
            </div>
            <span v-else class="text-slate-300">—</span>
          </template>
        </template>
      </a-table>

      <div v-if="cc.errors.length" class="mt-2">
        <a-alert
          v-for="err in cc.errors"
          :key="`${err.direction}-${err.transport_mode}`"
          type="warning"
          :message="`${err.direction === 'to_work' ? '去' : '回'} · ${modeLabel[err.transport_mode]}：${err.message}`"
          show-icon
          class="!mb-1 !text-xs"
        />
      </div>
    </div>

    <div v-if="!sortedResults.length" class="text-center text-slate-400 py-12">
      <CarOutlined class="text-4xl mb-2" />
      <div>填写上方参数后点「计算通勤」</div>
    </div>
  </div>
</template>
