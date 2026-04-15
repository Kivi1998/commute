<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { RocketOutlined, ReloadOutlined, CheckCircleFilled, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import {
  recommendCompanies,
  type AIRecommendInput,
  type AIRecommendedCompany,
} from '@/api/ai'
import { fetchProfile, type CompanyType } from '@/api/profile'
import { fetchEnums, type EnumItem } from '@/api/meta'
import { batchCreateCompanies, type CompanyCreateInput } from '@/api/company'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'imported'): void
}>()

const form = reactive<AIRecommendInput>({
  city: '北京',
  position: '后台开发',
  experience_years: undefined,
  company_types: ['big_tech', 'mid_tech'],
  count: 20,
  force_refresh: false,
})

const categoryOptions = ref<EnumItem[]>([])
const recommending = ref(false)
const importing = ref(false)
const result = ref<AIRecommendedCompany[]>([])
const summary = ref<{ fromCache: boolean; tokenIn: number; tokenOut: number } | null>(null)
const selectedNames = ref<string[]>([])

const cityOptions = [
  { value: '北京', label: '北京' },
  { value: '上海', label: '上海' },
  { value: '广州', label: '广州' },
  { value: '深圳', label: '深圳' },
  { value: '杭州', label: '杭州' },
  { value: '成都', label: '成都' },
  { value: '南京', label: '南京' },
  { value: '武汉', label: '武汉' },
  { value: '西安', label: '西安' },
  { value: '苏州', label: '苏州' },
]

const categoryColorMap: Record<string, string> = {
  big_tech: 'red',
  mid_tech: 'blue',
  startup: 'purple',
  foreign: 'gold',
  other: 'default',
}

function categoryLabel(v: string) {
  return categoryOptions.value.find((o) => o.value === v)?.label ?? v
}

watch(
  () => props.open,
  async (open) => {
    if (!open) return
    // 首次打开加载字典 + 尝试从 profile 回填
    const enums = await fetchEnums()
    categoryOptions.value = enums.company_type
    try {
      const p = await fetchProfile()
      if (p) {
        form.city = p.current_city || form.city
        form.position = p.target_position || form.position
        form.experience_years = p.experience_years
        if (p.preferred_company_types?.length) {
          form.company_types = p.preferred_company_types as CompanyType[]
        }
      }
    } catch {
      // 无画像也 OK
    }
  },
  { immediate: true },
)

async function handleRecommend() {
  if (!form.city.trim() || !form.position.trim()) {
    message.warning('请填写城市和岗位')
    return
  }
  recommending.value = true
  result.value = []
  selectedNames.value = []
  try {
    const r = await recommendCompanies({
      city: form.city,
      position: form.position,
      experience_years: form.experience_years,
      company_types: form.company_types,
      count: form.count,
      force_refresh: form.force_refresh,
    })
    result.value = r.companies
    summary.value = {
      fromCache: r.from_cache,
      tokenIn: r.token_input ?? 0,
      tokenOut: r.token_output ?? 0,
    }
    // 默认勾选所有"坐标已匹配"的公司
    selectedNames.value = r.companies.filter((c) => c.location_confident).map((c) => c.name)
    if (r.from_cache) {
      message.info('返回结果来自 24h 缓存')
    } else {
      message.success(`AI 推荐完成，共 ${r.companies.length} 家`)
    }
  } finally {
    recommending.value = false
  }
}

const selectedCompanies = computed(() =>
  result.value.filter((c) => selectedNames.value.includes(c.name)),
)

const selectableCount = computed(
  () => result.value.filter((c) => c.location_confident).length,
)

function toggleSelectAll() {
  if (selectedNames.value.length === selectableCount.value) {
    selectedNames.value = []
  } else {
    selectedNames.value = result.value.filter((c) => c.location_confident).map((c) => c.name)
  }
}

function toggleSelected(name: string, checked: boolean) {
  if (checked) {
    if (!selectedNames.value.includes(name)) selectedNames.value.push(name)
  } else {
    selectedNames.value = selectedNames.value.filter((n) => n !== name)
  }
}

async function handleImport() {
  const list = selectedCompanies.value
  if (!list.length) {
    message.warning('请先勾选至少一家公司')
    return
  }
  importing.value = true
  try {
    const payload: CompanyCreateInput[] = list
      .filter((c) => c.location_confident && c.resolved_longitude && c.resolved_latitude)
      .map((c) => ({
        name: c.name,
        address: c.resolved_address || c.address_hint,
        province: c.resolved_province,
        city: c.resolved_city,
        district: c.resolved_district,
        longitude: c.resolved_longitude!,
        latitude: c.resolved_latitude!,
        category: c.category as CompanyType,
        industry: c.industry,
        status: 'watching',
        source: 'ai_recommend',
        ai_reason: c.reason,
      }))

    const r = await batchCreateCompanies(payload)
    const created = r.created.length
    const skipped = r.skipped.length
    if (skipped > 0) {
      message.success(`已加入 ${created} 家，跳过 ${skipped} 家（已存在）`)
    } else {
      message.success(`已加入 ${created} 家公司`)
    }
    emit('imported')
    emit('update:open', false)
  } finally {
    importing.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    title="🤖 AI 推荐公司"
    width="900px"
    :footer="null"
    destroy-on-close
    @cancel="emit('update:open', false)"
    @update:open="(v: boolean) => emit('update:open', v)"
  >
    <!-- 配置区 -->
    <a-card size="small" :bordered="false" class="!bg-slate-50">
      <a-form layout="vertical" :colon="false" class="!mt-0">
        <div class="grid grid-cols-2 gap-x-3">
          <a-form-item label="城市" required class="!mb-2">
            <a-select
              v-model:value="form.city"
              :options="cityOptions"
              show-search
              placeholder="选择城市"
            />
          </a-form-item>
          <a-form-item label="求职岗位" required class="!mb-2">
            <a-input v-model:value="form.position" placeholder="如：后台开发" />
          </a-form-item>
        </div>

        <div class="grid grid-cols-3 gap-x-3">
          <a-form-item label="工作经验（年）" class="!mb-2">
            <a-input-number
              v-model:value="form.experience_years"
              :min="0"
              :max="30"
              class="!w-full"
            />
          </a-form-item>
          <a-form-item label="推荐数量" class="!mb-2">
            <a-input-number v-model:value="form.count" :min="5" :max="50" class="!w-full" />
          </a-form-item>
          <a-form-item label="偏好类型" class="!mb-2">
            <a-select
              v-model:value="form.company_types"
              mode="multiple"
              :options="categoryOptions.map((o) => ({ value: o.value, label: o.label }))"
              :max-tag-count="3"
              placeholder="可选"
            />
          </a-form-item>
        </div>

        <div class="flex items-center gap-3">
          <a-button type="primary" :loading="recommending" @click="handleRecommend">
            <template #icon><RocketOutlined /></template>
            开始推荐
          </a-button>
          <a-checkbox v-model:checked="form.force_refresh">
            强制刷新（跳过 24h 缓存）
          </a-checkbox>
          <span v-if="summary" class="text-xs text-slate-500 ml-auto">
            <a-tag v-if="summary.fromCache" color="blue">来自缓存</a-tag>
            <a-tag v-else color="green">实时生成</a-tag>
            <span v-if="!summary.fromCache">
              tokens: 入 {{ summary.tokenIn }} · 出 {{ summary.tokenOut }}
            </span>
          </span>
        </div>
      </a-form>
    </a-card>

    <!-- 结果区 -->
    <div v-if="result.length" class="mt-4">
      <div class="flex justify-between items-center mb-3">
        <div>
          <a-checkbox
            :checked="selectedNames.length === selectableCount && selectableCount > 0"
            :indeterminate="selectedNames.length > 0 && selectedNames.length < selectableCount"
            @change="toggleSelectAll"
          >
            全选（仅坐标已匹配）
          </a-checkbox>
          <span class="ml-3 text-sm text-slate-500">
            已选 {{ selectedNames.length }} / {{ selectableCount }} 家
          </span>
        </div>
        <a-button
          type="primary"
          :loading="importing"
          :disabled="!selectedNames.length"
          @click="handleImport"
        >
          加入我的公司列表（{{ selectedNames.length }}）
        </a-button>
      </div>

      <div class="max-h-96 overflow-y-auto space-y-2">
        <div
          v-for="c in result"
          :key="c.name"
          class="p-3 border rounded-md flex items-start gap-3 transition-colors"
          :class="[
            !c.location_confident
              ? 'bg-amber-50 border-amber-200'
              : 'bg-white hover:bg-slate-50 border-slate-200',
          ]"
        >
          <a-checkbox
            :checked="selectedNames.includes(c.name)"
            :disabled="!c.location_confident"
            class="!mt-1"
            @change="(e: any) => toggleSelected(c.name, e.target.checked)"
          />
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-1">
              <span class="font-semibold text-slate-800">{{ c.name }}</span>
              <a-tag :color="categoryColorMap[c.category] || 'default'" class="!mr-0">
                {{ categoryLabel(c.category) }}
              </a-tag>
              <span class="text-xs text-slate-400">{{ c.industry }}</span>
              <CheckCircleFilled
                v-if="c.location_confident"
                class="!text-green-500"
                title="坐标已匹配"
              />
              <ExclamationCircleOutlined
                v-else
                class="!text-amber-500"
                title="坐标未匹配，无法加入"
              />
            </div>
            <div class="text-sm text-slate-600">{{ c.reason }}</div>
            <div class="text-xs text-slate-400 mt-1">
              📍
              <template v-if="c.location_confident">
                {{ c.resolved_address }}
                <span class="ml-2 font-mono text-slate-300">
                  {{ c.resolved_longitude?.toFixed(6) }},
                  {{ c.resolved_latitude?.toFixed(6) }}
                </span>
              </template>
              <span v-else class="text-amber-600">
                AI 提示：{{ c.address_hint }} · 高德未匹配到精确坐标
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="recommending" class="text-center py-12">
      <a-spin size="large" />
      <div class="mt-3 text-slate-500">豆包 AI 正在思考，通常需要 20-40 秒...</div>
    </div>

    <div v-else class="text-center py-12 text-slate-400">
      <ReloadOutlined class="text-3xl mb-2" />
      <div>填写上方参数后点「开始推荐」</div>
    </div>
  </a-modal>
</template>
