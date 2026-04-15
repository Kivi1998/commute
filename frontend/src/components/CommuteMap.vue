<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { loadAMap } from '@/lib/amap'
import type { HomeAddress } from '@/api/address'
import type { CompanyCommute, CommuteResultItem } from '@/api/commute'

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
let highlightRoute: any = null // 高亮的真实路线

function clearStatic() {
  homeMarker?.setMap?.(null)
  homeMarker = null
  companyMarkers.forEach((m) => m.setMap(null))
  companyMarkers.clear()
  lines.forEach((l) => l.setMap(null))
  lines.clear()
}

function clearHighlightRoute() {
  highlightRoute?.setMap?.(null)
  highlightRoute = null
}

function minDurationMin(cc: CompanyCommute): number {
  const work = cc.items.filter((it) => it.direction === 'to_work')
  if (work.length === 0) return 0
  return Math.min(...work.map((it) => it.duration_min))
}

function bestItemWithPolyline(cc: CompanyCommute): CommuteResultItem | undefined {
  const work = cc.items.filter((it) => it.direction === 'to_work' && it.polyline)
  if (!work.length) return undefined
  return work.reduce((a, b) => (a.duration_min <= b.duration_min ? a : b))
}

function colorByDuration(min: number): string {
  if (min === 0) return '#94a3b8'
  if (min <= 30) return '#16a34a'
  if (min <= 60) return '#f59e0b'
  return '#dc2626'
}

function parsePolyline(p: string): [number, number][] {
  return p
    .split(';')
    .map((pair) => {
      const [lng, lat] = pair.split(',').map(Number)
      return [lng, lat] as [number, number]
    })
    .filter(([lng, lat]) => Number.isFinite(lng) && Number.isFinite(lat))
}

function renderStatic() {
  if (!map || !AMap || !props.home) return
  clearStatic()

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
    positions.push([cc.company_longitude, cc.company_latitude])

    const marker = new AMap.Marker({
      position: [cc.company_longitude, cc.company_latitude],
      map,
      content: `<div style="background:${color};color:white;padding:4px 10px;border-radius:6px;font-size:12px;font-weight:500;box-shadow:0 2px 6px rgba(0,0,0,0.2);white-space:nowrap">🏢 ${cc.company_name} · ${dur || '-'}min</div>`,
      offset: new AMap.Pixel(-30, -36),
    })
    companyMarkers.set(cc.company_id, marker)

    // 基础虚线（所有公司，细）
    const line = new AMap.Polyline({
      path: [
        [props.home!.longitude, props.home!.latitude],
        [cc.company_longitude, cc.company_latitude],
      ],
      strokeColor: color,
      strokeWeight: 2,
      strokeOpacity: 0.3,
      strokeStyle: 'dashed',
      map,
    })
    lines.set(cc.company_id, line)
  })

  if (positions.length > 1) {
    map.setFitView(undefined, false, [40, 40, 40, 40])
  } else {
    map.setCenter(positions[0])
    map.setZoom(13)
  }
}

function renderHighlight() {
  clearHighlightRoute()
  if (!map || !AMap || !props.highlightCompanyId) return

  const cc = props.companies?.find((c) => c.company_id === props.highlightCompanyId)
  if (!cc) return

  const best = bestItemWithPolyline(cc)
  if (!best || !best.polyline) {
    // 没有真实路线数据，高亮直线代替
    const line = lines.get(cc.company_id)
    if (line) {
      line.setOptions({
        strokeColor: '#0ea5e9',
        strokeWeight: 4,
        strokeOpacity: 0.9,
      })
    }
    return
  }

  const path = parsePolyline(best.polyline)
  if (path.length < 2) return

  highlightRoute = new AMap.Polyline({
    path,
    strokeColor: '#0ea5e9',
    strokeWeight: 6,
    strokeOpacity: 0.85,
    lineJoin: 'round',
    lineCap: 'round',
    showDir: true,
    map,
  })

  // 把高亮路线所在区域框入视野
  map.setFitView([highlightRoute, homeMarker, companyMarkers.get(cc.company_id)], false, [
    60, 60, 60, 60,
  ])
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
    renderStatic()
    renderHighlight()
  } catch (e) {
    message.error('地图加载失败：' + (e as Error).message)
  }
})

onBeforeUnmount(() => {
  clearStatic()
  clearHighlightRoute()
  map?.destroy?.()
})

// home/companies 变化 → 重绘静态 + 高亮
watch(
  () => [props.home, props.companies],
  () => {
    if (map && AMap) {
      renderStatic()
      renderHighlight()
    }
  },
  { deep: true },
)

// 高亮变化 → 只重绘高亮
watch(
  () => props.highlightCompanyId,
  () => {
    if (map && AMap) {
      // 恢复所有基础线（取消上次高亮的加粗直线）
      lines.forEach((line, id) => {
        const cc = props.companies?.find((c) => c.company_id === id)
        if (!cc) return
        const color = colorByDuration(minDurationMin(cc))
        line.setOptions({
          strokeColor: color,
          strokeWeight: 2,
          strokeOpacity: 0.3,
          strokeStyle: 'dashed',
        })
      })
      renderHighlight()
    }
  },
)
</script>

<template>
  <div
    ref="container"
    class="w-full border border-slate-200 rounded-md overflow-hidden"
    :style="{ height: props.height ?? '500px' }"
  />
</template>
