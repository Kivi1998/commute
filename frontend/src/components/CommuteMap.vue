<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { loadAMap } from '@/lib/amap'
import type { HomeAddress } from '@/api/address'
import type { CompanyCommute } from '@/api/commute'

const props = defineProps<{
  home?: HomeAddress | null
  companies?: CompanyCommute[]
  highlightCompanyId?: number | null
  height?: string
}>()

const container = ref<HTMLElement>()
let AMap: any = null
let map: any = null
let homeMarker: any = null
const companyMarkers = new Map<number, any>()
const lines = new Map<number, any>()

function clearMapElements() {
  homeMarker?.setMap?.(null)
  homeMarker = null
  companyMarkers.forEach((m) => m.setMap(null))
  companyMarkers.clear()
  lines.forEach((l) => l.setMap(null))
  lines.clear()
}

function minDurationMin(cc: CompanyCommute): number {
  const work = cc.items.filter((it) => it.direction === 'to_work')
  if (work.length === 0) return 0
  return Math.min(...work.map((it) => it.duration_min))
}

function colorByDuration(min: number): string {
  if (min === 0) return '#94a3b8'
  if (min <= 30) return '#16a34a'
  if (min <= 60) return '#f59e0b'
  return '#dc2626'
}

function render() {
  if (!map || !AMap || !props.home) return
  clearMapElements()

  // 家标记（蓝色房屋）
  homeMarker = new AMap.Marker({
    position: [props.home.longitude, props.home.latitude],
    map,
    content: `<div style="background:#2563eb;color:white;padding:4px 10px;border-radius:6px;font-size:12px;font-weight:500;box-shadow:0 2px 6px rgba(0,0,0,0.2);white-space:nowrap">🏠 ${props.home.alias}</div>`,
    offset: new AMap.Pixel(-20, -36),
  })

  const positions: [number, number][] = [
    [props.home.longitude, props.home.latitude],
  ]

  props.companies?.forEach((cc) => {
    const dur = minDurationMin(cc)
    const color = colorByDuration(dur)
    const isHighlight = props.highlightCompanyId === cc.company_id
    positions.push([cc.company_longitude, cc.company_latitude])

    // 公司标记
    const marker = new AMap.Marker({
      position: [cc.company_longitude, cc.company_latitude],
      map,
      content: `<div style="background:${color};color:white;padding:4px 10px;border-radius:6px;font-size:12px;font-weight:500;box-shadow:0 2px 6px rgba(0,0,0,0.2);white-space:nowrap;${isHighlight ? 'outline:2px solid #0ea5e9;' : ''}">🏢 ${cc.company_name} · ${dur || '-'}min</div>`,
      offset: new AMap.Pixel(-30, -36),
    })
    companyMarkers.set(cc.company_id, marker)

    // 连线
    const polyline = new AMap.Polyline({
      path: [
        [props.home!.longitude, props.home!.latitude],
        [cc.company_longitude, cc.company_latitude],
      ],
      strokeColor: isHighlight ? '#0ea5e9' : color,
      strokeWeight: isHighlight ? 5 : 2.5,
      strokeOpacity: isHighlight ? 0.9 : 0.5,
      strokeStyle: isHighlight ? 'solid' : 'dashed',
      map,
    })
    lines.set(cc.company_id, polyline)
  })

  // 自动适配
  if (positions.length > 1) {
    map.setFitView(undefined, false, [40, 40, 40, 40])
  } else {
    map.setCenter(positions[0])
    map.setZoom(13)
  }
}

onMounted(async () => {
  try {
    AMap = await loadAMap()
    await nextTick()
    map = new AMap.Map(container.value!, {
      zoom: 11,
      center: [116.397428, 39.908722],
      viewMode: '2D',
    })
    map.addControl(new AMap.ToolBar({ position: 'RB' }))
    map.addControl(new AMap.Scale())
    render()
  } catch (e) {
    message.error('地图加载失败：' + (e as Error).message)
  }
})

onBeforeUnmount(() => {
  clearMapElements()
  map?.destroy?.()
})

watch(
  () => [props.home, props.companies, props.highlightCompanyId],
  () => {
    if (map && AMap) render()
  },
  { deep: true },
)
</script>

<template>
  <div
    ref="container"
    class="w-full border border-slate-200 rounded-md overflow-hidden"
    :style="{ height: props.height ?? '500px' }"
  />
</template>
