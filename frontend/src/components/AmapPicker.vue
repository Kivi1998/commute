<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { EnvironmentOutlined } from '@ant-design/icons-vue'
import { loadAMap, lnglatToPair } from '@/lib/amap'

export interface AmapPickerValue {
  address: string
  longitude: number
  latitude: number
  province?: string
  city?: string
  district?: string
}

const props = withDefaults(
  defineProps<{
    modelValue?: AmapPickerValue | null
    city?: string
    height?: string
    initialCenter?: [number, number]
  }>(),
  {
    height: '320px',
    initialCenter: () => [116.397428, 39.908722], // 天安门默认
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: AmapPickerValue): void
}>()

const mapContainer = ref<HTMLElement>()
const searchInput = ref<HTMLInputElement>()
const keyword = ref('')
const loading = ref(true)
const current = ref<AmapPickerValue | null>(props.modelValue ?? null)

let AMap: any = null
let map: any = null
let marker: any = null
let geocoder: any = null
let autoComplete: any = null

async function handleClick(lngLat: any) {
  const { lng, lat } = lnglatToPair(lngLat)
  if (marker) {
    marker.setPosition([lng, lat])
  } else {
    marker = new AMap.Marker({ position: [lng, lat], map })
  }
  // 逆地理编码
  await new Promise<void>((resolve) => {
    geocoder.getAddress([lng, lat], (status: string, result: any) => {
      if (status === 'complete' && result?.regeocode) {
        const rc = result.regeocode
        const ac = rc.addressComponent
        current.value = {
          longitude: Number(lng.toFixed(6)),
          latitude: Number(lat.toFixed(6)),
          address: rc.formattedAddress || '',
          province: ac?.province || undefined,
          city: ac?.city || ac?.province || undefined,
          district: ac?.district || undefined,
        }
        emit('update:modelValue', current.value)
      } else {
        current.value = {
          longitude: Number(lng.toFixed(6)),
          latitude: Number(lat.toFixed(6)),
          address: `${lng.toFixed(6)}, ${lat.toFixed(6)}`,
        }
        emit('update:modelValue', current.value)
      }
      resolve()
    })
  })
}

function attachAutoComplete() {
  if (!searchInput.value) return
  autoComplete = new AMap.AutoComplete({
    input: searchInput.value,
    city: props.city || '全国',
  })
  autoComplete.on('select', (e: any) => {
    const poi = e.poi
    if (!poi?.location) {
      message.warning('该结果无坐标，请重选')
      return
    }
    const { lng, lat } = lnglatToPair(poi.location)
    map.setCenter([lng, lat])
    map.setZoom(16)
    handleClick(poi.location)
  })
}

onMounted(async () => {
  try {
    AMap = await loadAMap()
    await nextTick()
    map = new AMap.Map(mapContainer.value!, {
      zoom: 11,
      center: current.value
        ? [current.value.longitude, current.value.latitude]
        : props.initialCenter,
      viewMode: '2D',
    })
    map.addControl(new AMap.ToolBar({ position: 'RB' }))
    map.addControl(new AMap.Scale())

    geocoder = new AMap.Geocoder({ city: props.city || '全国' })

    if (current.value) {
      marker = new AMap.Marker({
        position: [current.value.longitude, current.value.latitude],
        map,
      })
    }

    map.on('click', (e: any) => handleClick(e.lnglat))
    attachAutoComplete()
  } catch (e) {
    message.error('地图加载失败：' + (e as Error).message)
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  autoComplete?.destroy?.()
  marker?.setMap?.(null)
  map?.destroy?.()
})

// 外部设置新值时同步到地图
watch(
  () => props.modelValue,
  (v) => {
    if (!v || !map || !AMap) return
    // 只有当外部值跟当前不一致时才移动
    if (
      current.value &&
      Math.abs(v.longitude - current.value.longitude) < 0.000001 &&
      Math.abs(v.latitude - current.value.latitude) < 0.000001
    ) {
      return
    }
    current.value = v
    map.setCenter([v.longitude, v.latitude])
    map.setZoom(16)
    if (marker) {
      marker.setPosition([v.longitude, v.latitude])
    } else {
      marker = new AMap.Marker({ position: [v.longitude, v.latitude], map })
    }
  },
  { deep: true },
)
</script>

<template>
  <div class="amap-picker">
    <div class="mb-2 relative">
      <input
        ref="searchInput"
        v-model="keyword"
        type="text"
        placeholder="输入地址搜索，或直接在地图上点选"
        class="w-full px-3 py-2 border border-slate-300 rounded-md text-sm outline-none focus:border-blue-500"
      />
      <EnvironmentOutlined class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400" />
    </div>

    <a-spin :spinning="loading">
      <div
        ref="mapContainer"
        class="w-full border border-slate-200 rounded-md overflow-hidden"
        :style="{ height: props.height }"
      />
    </a-spin>

    <div v-if="current" class="mt-2 text-xs text-slate-600 bg-slate-50 px-3 py-2 rounded">
      <div class="truncate">📍 {{ current.address }}</div>
      <div class="font-mono mt-0.5 text-slate-400">
        {{ current.longitude.toFixed(6) }}, {{ current.latitude.toFixed(6) }}
        <span v-if="current.city" class="ml-2 text-slate-500">{{ current.city }}</span>
      </div>
    </div>
    <div v-else class="mt-2 text-xs text-slate-400">尚未选择位置</div>
  </div>
</template>

<style>
/* 高德自动补全下拉层需要较高 z-index 才能覆盖 Modal */
.amap-sug-result {
  z-index: 2000 !important;
}
</style>
