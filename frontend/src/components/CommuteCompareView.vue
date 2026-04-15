<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, RadarChart, ScatterChart } from 'echarts/charts'
import {
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  ToolboxComponent,
  MarkLineComponent,
} from 'echarts/components'
import type { CompanyCommute, CommuteResultItem, TransportMode } from '@/api/commute'

use([
  CanvasRenderer,
  BarChart,
  RadarChart,
  ScatterChart,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  ToolboxComponent,
  MarkLineComponent,
])

const props = defineProps<{
  results: CompanyCommute[]
  mode: 'bar' | 'scatter' | 'radar'
}>()

const modeLabel: Record<TransportMode, string> = {
  transit: '公交/地铁',
  driving: '驾车',
  cycling: '骑行',
  walking: '步行',
}

const modeColor: Record<TransportMode, string> = {
  transit: '#2563eb',
  driving: '#f59e0b',
  cycling: '#16a34a',
  walking: '#a855f7',
}

function bestItem(cc: CompanyCommute, direction: 'to_work' | 'to_home'): CommuteResultItem | undefined {
  const items = cc.items.filter((it) => it.direction === direction)
  if (!items.length) return undefined
  return items.reduce((a, b) => (a.duration_min <= b.duration_min ? a : b))
}

// --- 柱状图：按公司的早/晚最优通勤对比 ---
const barOption = computed(() => {
  const sorted = [...props.results].sort((a, b) => {
    const da = bestItem(a, 'to_work')?.duration_min ?? Infinity
    const db = bestItem(b, 'to_work')?.duration_min ?? Infinity
    return da - db
  })
  const names = sorted.map((r) => r.company_name)
  const morning = sorted.map((r) => bestItem(r, 'to_work')?.duration_min ?? 0)
  const evening = sorted.map((r) => bestItem(r, 'to_home')?.duration_min ?? 0)

  return {
    title: { text: '各公司早晚通勤最优耗时（分钟）', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: ['早通勤 去', '晚通勤 回'], top: 28 },
    grid: { left: 60, right: 30, top: 70, bottom: 60 },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: { interval: 0, rotate: names.length > 6 ? 30 : 0 },
    },
    yAxis: { type: 'value', name: '分钟' },
    series: [
      {
        name: '早通勤 去',
        type: 'bar',
        data: morning,
        itemStyle: { color: '#2563eb' },
        markLine: {
          silent: true,
          symbol: 'none',
          lineStyle: { color: '#16a34a', type: 'dashed' },
          data: [{ yAxis: 30, label: { formatter: '30min 优秀' } }],
        },
      },
      {
        name: '晚通勤 回',
        type: 'bar',
        data: evening,
        itemStyle: { color: '#f59e0b' },
      },
    ],
  }
})

// --- 散点图：距离 vs 时长，按出行方式着色 ---
const scatterOption = computed(() => {
  const series: any[] = []
  const modesInUse: TransportMode[] = Array.from(
    new Set(props.results.flatMap((r) => r.items.map((it) => it.transport_mode))),
  )

  for (const mode of modesInUse) {
    const data: [number, number, string][] = []
    for (const r of props.results) {
      for (const it of r.items) {
        if (it.transport_mode !== mode) continue
        const label = `${r.company_name} · ${it.direction === 'to_work' ? '去' : '回'}`
        data.push([it.distance_km, it.duration_min, label])
      }
    }
    series.push({
      name: modeLabel[mode],
      type: 'scatter',
      symbolSize: 12,
      data,
      itemStyle: { color: modeColor[mode] },
    })
  }

  return {
    title: { text: '距离 vs 耗时（所有方式）', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: {
      trigger: 'item',
      formatter: (p: any) =>
        `${p.data[2]}<br/>${p.seriesName}<br/>距离 ${p.data[0]} km · 耗时 ${p.data[1]} min`,
    },
    legend: { top: 28 },
    grid: { left: 60, right: 30, top: 70, bottom: 50 },
    xAxis: { type: 'value', name: '距离 (km)', nameLocation: 'middle', nameGap: 28 },
    yAxis: { type: 'value', name: '耗时 (min)', nameLocation: 'middle', nameGap: 40 },
    series,
  }
})

// --- 雷达图：前 N 家公司多维归一化对比 ---
const radarOption = computed(() => {
  const top = [...props.results]
    .sort((a, b) => {
      const da = bestItem(a, 'to_work')?.duration_min ?? Infinity
      const db = bestItem(b, 'to_work')?.duration_min ?? Infinity
      return da - db
    })
    .slice(0, 6)

  const dims = [
    { name: '早通勤耗时', getter: (r: CompanyCommute) => bestItem(r, 'to_work')?.duration_min ?? 0, invert: true },
    { name: '晚通勤耗时', getter: (r: CompanyCommute) => bestItem(r, 'to_home')?.duration_min ?? 0, invert: true },
    { name: '距离', getter: (r: CompanyCommute) => bestItem(r, 'to_work')?.distance_km ?? 0, invert: true },
    {
      name: '换乘次数',
      getter: (r: CompanyCommute) => {
        const it = bestItem(r, 'to_work')
        return it?.transfer_count ?? 0
      },
      invert: true,
    },
    {
      name: '单程费用',
      getter: (r: CompanyCommute) => {
        const it = bestItem(r, 'to_work')
        return it?.cost_yuan ?? 0
      },
      invert: true,
    },
  ]

  // 归一化：越小越好的维度，值越小得分越高
  const maxes = dims.map((d) => Math.max(1, ...top.map((r) => d.getter(r))))

  const indicators = dims.map((d) => ({ name: d.name, max: 100 }))

  const series = [
    {
      type: 'radar',
      data: top.map((r) => ({
        name: r.company_name,
        value: dims.map((d, i) => {
          const raw = d.getter(r)
          if (maxes[i] === 0) return 100
          const ratio = raw / maxes[i]
          // invert: 越小越好 → 得分 = 100 * (1 - ratio / 1.2) 让最差 ≈ 16 分，最好接近 100
          return Math.round(100 * (1 - ratio * 0.83))
        }),
      })),
      symbol: 'circle',
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.15 },
    },
  ]

  return {
    title: {
      text: `前 ${top.length} 家公司综合得分（数值越大越好）`,
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: {},
    legend: { bottom: 0, data: top.map((r) => r.company_name) },
    radar: { indicator: indicators, splitNumber: 4 },
    series,
  }
})

const chartOption = computed(() => {
  if (props.mode === 'bar') return barOption.value
  if (props.mode === 'scatter') return scatterOption.value
  return radarOption.value
})

</script>

<template>
  <div v-if="results.length" class="w-full">
    <v-chart
      :option="chartOption"
      :autoresize="true"
      style="height: 520px; width: 100%"
    />
  </div>
  <a-empty v-else description="暂无数据" class="py-16" />
</template>
