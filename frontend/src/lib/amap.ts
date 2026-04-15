import AMapLoader from '@amap/amap-jsapi-loader'

/**
 * 高德 JS SDK 单例 loader。首次调用异步加载，后续直接返回同一份 AMap namespace。
 * 安全密钥（securityJsCode）必须在 loader 执行前挂到 window。
 */

declare global {
  interface Window {
    _AMapSecurityConfig?: { securityJsCode: string }
  }
}

type AMapNS = typeof AMap

let cache: Promise<AMapNS> | null = null

const DEFAULT_PLUGINS = [
  'AMap.PlaceSearch',
  'AMap.AutoComplete',
  'AMap.Geocoder',
  'AMap.ToolBar',
  'AMap.Scale',
]

export interface LoadOptions {
  extraPlugins?: string[]
}

export function loadAMap(opts: LoadOptions = {}): Promise<AMapNS> {
  if (cache) return cache

  const key = import.meta.env.VITE_AMAP_JS_KEY as string | undefined
  const security = import.meta.env.VITE_AMAP_JS_SECURITY as string | undefined

  if (!key || !security) {
    return Promise.reject(
      new Error('VITE_AMAP_JS_KEY 或 VITE_AMAP_JS_SECURITY 未配置'),
    )
  }

  window._AMapSecurityConfig = { securityJsCode: security }

  cache = AMapLoader.load({
    key,
    version: '2.0',
    plugins: [...DEFAULT_PLUGINS, ...(opts.extraPlugins ?? [])],
  }).then((AMap: AMapNS) => AMap)

  return cache
}

/**
 * 解析高德 LngLat 返回的经纬度（兼容数组或对象）
 */
export function lnglatToPair(ll: any): { lng: number; lat: number } {
  if (!ll) return { lng: 0, lat: 0 }
  if (Array.isArray(ll)) return { lng: ll[0], lat: ll[1] }
  if (typeof ll.getLng === 'function') return { lng: ll.getLng(), lat: ll.getLat() }
  return { lng: ll.lng ?? ll.longitude ?? 0, lat: ll.lat ?? ll.latitude ?? 0 }
}
