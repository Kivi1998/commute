<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import type { Company, CompanyCreateInput, CompanyStatus } from '@/api/company'
import type { EnumItem } from '@/api/meta'
import type { CompanyType } from '@/api/profile'
import AmapPicker, { type AmapPickerValue } from './AmapPicker.vue'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  record: Company | null
  statusOptions: EnumItem[]
  categoryOptions: EnumItem[]
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: CompanyCreateInput): void
}>()

const form = reactive<CompanyCreateInput>({
  name: '',
  address: '',
  province: '',
  city: '',
  district: '',
  longitude: 0,
  latitude: 0,
  category: undefined,
  industry: '',
  status: 'watching',
  note: '',
})

const pickerValue = ref<AmapPickerValue | null>(null)
const submitting = ref(false)

function resetForm() {
  form.name = ''
  form.address = ''
  form.province = ''
  form.city = ''
  form.district = ''
  form.longitude = 0
  form.latitude = 0
  form.category = undefined
  form.industry = ''
  form.status = 'watching'
  form.note = ''
  pickerValue.value = null
}

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.mode === 'edit' && props.record) {
      const r = props.record
      form.name = r.name
      form.address = r.address
      form.province = r.province ?? ''
      form.city = r.city ?? ''
      form.district = r.district ?? ''
      form.longitude = r.longitude
      form.latitude = r.latitude
      form.category = r.category
      form.industry = r.industry ?? ''
      form.status = r.status as CompanyStatus
      form.note = r.note ?? ''
      if (Math.abs(r.longitude) <= 180 && Math.abs(r.latitude) <= 90 && r.longitude !== 0) {
        pickerValue.value = {
          address: r.address,
          longitude: r.longitude,
          latitude: r.latitude,
          province: r.province,
          city: r.city,
          district: r.district,
        }
      } else {
        pickerValue.value = null
      }
    } else {
      resetForm()
    }
  },
)

function handlePickerChange(v: AmapPickerValue) {
  pickerValue.value = v
  form.address = v.address
  form.longitude = v.longitude
  form.latitude = v.latitude
  form.province = v.province ?? ''
  form.city = v.city ?? ''
  form.district = v.district ?? ''
}

async function handleOk() {
  if (!form.name.trim() || !form.address.trim()) return
  if (!form.longitude || !form.latitude) return

  submitting.value = true
  try {
    const payload: CompanyCreateInput = {
      name: form.name.trim(),
      address: form.address.trim(),
      province: form.province?.trim() || undefined,
      city: form.city?.trim() || undefined,
      district: form.district?.trim() || undefined,
      longitude: form.longitude,
      latitude: form.latitude,
      category: form.category as CompanyType | undefined,
      industry: form.industry?.trim() || undefined,
      status: form.status,
      note: form.note?.trim() || undefined,
    }
    emit('submit', payload)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    :title="mode === 'create' ? '新增公司' : '编辑公司'"
    :confirm-loading="submitting"
    ok-text="保存"
    cancel-text="取消"
    width="760px"
    destroy-on-close
    @ok="handleOk"
    @cancel="emit('update:open', false)"
    @update:open="(v: boolean) => emit('update:open', v)"
  >
    <a-form layout="vertical" class="pt-2">
      <a-form-item label="公司名称" required>
        <a-input v-model:value="form.name" :maxlength="128" placeholder="如：字节跳动" />
      </a-form-item>

      <a-form-item label="办公地址（搜索或点选地图）" required>
        <AmapPicker
          :model-value="pickerValue"
          @update:model-value="handlePickerChange"
        />
      </a-form-item>

      <div class="grid grid-cols-2 gap-3">
        <a-form-item label="公司类型">
          <a-select
            v-model:value="form.category"
            placeholder="可选"
            allow-clear
            :options="categoryOptions.map((o) => ({ value: o.value, label: o.label }))"
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model:value="form.status"
            :options="statusOptions.map((o) => ({ value: o.value, label: o.label }))"
          />
        </a-form-item>
      </div>

      <a-form-item label="行业">
        <a-input v-model:value="form.industry" placeholder="如：互联网、金融、医疗" />
      </a-form-item>

      <a-form-item label="备注">
        <a-textarea v-model:value="form.note" :rows="2" placeholder="HR 联系方式、薪资范围、面试进度等" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>
