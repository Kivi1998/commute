<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { RocketOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import type { HomeAddress } from '@/api/address'
import { listAddresses } from '@/api/address'
import type { Company } from '@/api/company'
import { listCompanies } from '@/api/company'
import type { EnumItem } from '@/api/meta'
import { fetchEnums } from '@/api/meta'
import type {
  CommuteCalculateInput,
  TransportMode,
} from '@/api/commute'

const emit = defineEmits<{
  (e: 'calculate', payload: CommuteCalculateInput): void
}>()

const addresses = ref<HomeAddress[]>([])
const companies = ref<Company[]>([])
const transportOptions = ref<EnumItem[]>([])

const form = reactive<{
  home_id?: number
  company_ids: number[]
  transport_modes: TransportMode[]
  morning: Dayjs
  evening: Dayjs
  buffer_minutes: number
  force_refresh: boolean
}>({
  home_id: undefined,
  company_ids: [],
  transport_modes: ['transit', 'driving'],
  morning: dayjs('08:00', 'HH:mm'),
  evening: dayjs('17:30', 'HH:mm'),
  buffer_minutes: 5,
  force_refresh: false,
})

const loading = ref(false)

const homeOptions = computed(() =>
  addresses.value.map((a) => ({
    value: a.id,
    label: `${a.alias} · ${a.address}`,
  })),
)

const companyOptions = computed(() =>
  companies.value.map((c) => ({
    value: c.id,
    label: c.name,
    address: c.address,
  })),
)

const canCalc = computed(
  () =>
    !!form.home_id &&
    form.company_ids.length > 0 &&
    form.transport_modes.length > 0,
)

async function loadData() {
  const [addrList, comps, enums] = await Promise.all([
    listAddresses(),
    listCompanies({ page_size: 100 }),
    fetchEnums(),
  ])
  addresses.value = addrList
  companies.value = comps.list
  transportOptions.value = enums.transport_mode

  // 默认选中默认住址
  const def = addrList.find((a) => a.is_default) ?? addrList[0]
  if (def) form.home_id = def.id
  // 默认选全部公司
  form.company_ids = comps.list.map((c) => c.id)
}

onMounted(loadData)

function handleCalculate() {
  if (!canCalc.value) {
    message.warning('请先填写家、公司、出行方式')
    return
  }
  if (form.company_ids.length > 20) {
    message.warning(`已勾选 ${form.company_ids.length} 家，建议不超过 20 家`)
  }
  const payload: CommuteCalculateInput = {
    home_id: form.home_id!,
    company_ids: form.company_ids,
    transport_modes: form.transport_modes,
    morning: { strategy: 'depart_at', time: form.morning.format('HH:mm') },
    evening: { strategy: 'depart_at', time: form.evening.format('HH:mm') },
    buffer_minutes: form.buffer_minutes,
    force_refresh: form.force_refresh,
    save_query: true,
  }
  emit('calculate', payload)
}

function selectAllCompanies() {
  form.company_ids = companies.value.map((c) => c.id)
}
function clearCompanies() {
  form.company_ids = []
}

/**
 * 用历史查询回填表单（不立即触发计算，由父组件决定）
 */
function applyQuery(q: {
  home_id: number
  transport_modes: TransportMode[]
  morning_time: string
  evening_time: string
  buffer_minutes: number
  company_ids: number[]
}) {
  form.home_id = q.home_id
  form.transport_modes = q.transport_modes
  form.morning = dayjs(q.morning_time, 'HH:mm')
  form.evening = dayjs(q.evening_time, 'HH:mm')
  form.buffer_minutes = q.buffer_minutes
  form.company_ids = q.company_ids
}

defineExpose({
  setLoading: (v: boolean) => (loading.value = v),
  applyQuery,
  triggerCalculate: handleCalculate,
})

// 有公司变动时重置
watch(
  () => companies.value.length,
  (n) => {
    if (n === 0) form.company_ids = []
  },
)
</script>

<template>
  <a-card title="通勤查询" :bordered="false" size="small">
    <template #extra>
      <a-button
        type="primary"
        :loading="loading"
        :disabled="!canCalc"
        @click="handleCalculate"
      >
        <template #icon><RocketOutlined /></template>
        计算通勤
      </a-button>
    </template>

    <a-form layout="vertical" :colon="false" class="!mt-0">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-4">
        <a-form-item label="家庭住址" required>
          <a-select
            v-model:value="form.home_id"
            :options="homeOptions"
            placeholder="选择家庭住址"
            show-search
            option-filter-prop="label"
          />
          <div v-if="addresses.length === 0" class="text-xs text-orange-500 mt-1">
            还没有住址，请先到「设置」页添加。
          </div>
        </a-form-item>

        <a-form-item label="出行方式" required>
          <a-checkbox-group
            v-model:value="form.transport_modes"
            :options="transportOptions.map((o) => ({ value: o.value, label: `${o.icon ?? ''} ${o.label}` }))"
          />
        </a-form-item>
      </div>

      <a-form-item>
        <template #label>
          <span>目标公司（已选 {{ form.company_ids.length }} 家）</span>
          <a-button size="small" type="link" @click="selectAllCompanies">全选</a-button>
          <a-button size="small" type="link" @click="clearCompanies">清空</a-button>
        </template>
        <a-select
          v-model:value="form.company_ids"
          mode="multiple"
          :options="companyOptions"
          placeholder="选择要对比的公司（可多选）"
          :max-tag-count="5"
          show-search
          option-filter-prop="label"
          class="!w-full"
        />
        <div v-if="companies.length === 0" class="text-xs text-orange-500 mt-1">
          还没有公司，请先到「公司」页添加。
        </div>
      </a-form-item>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-x-4">
        <a-form-item label="早通勤 · 出门时间">
          <a-time-picker
            v-model:value="form.morning"
            format="HH:mm"
            class="!w-full"
            :minute-step="5"
          />
        </a-form-item>
        <a-form-item label="晚通勤 · 下班时间">
          <a-time-picker
            v-model:value="form.evening"
            format="HH:mm"
            class="!w-full"
            :minute-step="5"
          />
        </a-form-item>
        <a-form-item label="容错 buffer（分钟）">
          <a-input-number
            v-model:value="form.buffer_minutes"
            :min="0"
            :max="30"
            class="!w-full"
          />
        </a-form-item>
      </div>

      <a-form-item class="!mb-0">
        <a-checkbox v-model:checked="form.force_refresh">
          强制刷新（不使用 7 天缓存）
        </a-checkbox>
        <span class="text-xs text-slate-400 ml-4">
          默认：相同参数 7 天内复用缓存，避免高德 API 配额浪费
        </span>
      </a-form-item>
    </a-form>
  </a-card>
</template>
