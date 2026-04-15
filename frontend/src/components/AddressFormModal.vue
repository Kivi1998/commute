<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import type { HomeAddress, HomeAddressCreateInput } from '@/api/address'
import AmapPicker, { type AmapPickerValue } from './AmapPicker.vue'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  record: HomeAddress | null
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: HomeAddressCreateInput): void
}>()

const form = reactive<HomeAddressCreateInput>({
  alias: '',
  address: '',
  province: '',
  city: '',
  district: '',
  longitude: 0,
  latitude: 0,
  is_default: false,
  note: '',
})

const pickerValue = ref<AmapPickerValue | null>(null)
const submitting = ref(false)

function resetForm() {
  form.alias = ''
  form.address = ''
  form.province = ''
  form.city = ''
  form.district = ''
  form.longitude = 0
  form.latitude = 0
  form.is_default = false
  form.note = ''
  pickerValue.value = null
}

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.mode === 'edit' && props.record) {
      const r = props.record
      form.alias = r.alias
      form.address = r.address
      form.province = r.province ?? ''
      form.city = r.city ?? ''
      form.district = r.district ?? ''
      form.longitude = r.longitude
      form.latitude = r.latitude
      form.is_default = r.is_default
      form.note = r.note ?? ''
      // 仅在有合法坐标时初始化地图
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
  if (!form.alias.trim()) return
  if (!form.address.trim()) return
  if (!form.longitude || !form.latitude) return

  submitting.value = true
  try {
    const payload: HomeAddressCreateInput = {
      alias: form.alias.trim(),
      address: form.address.trim(),
      province: form.province?.trim() || undefined,
      city: form.city?.trim() || undefined,
      district: form.district?.trim() || undefined,
      longitude: form.longitude,
      latitude: form.latitude,
      is_default: form.is_default,
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
    :title="mode === 'create' ? '新增家庭住址' : '编辑家庭住址'"
    :confirm-loading="submitting"
    ok-text="保存"
    cancel-text="取消"
    width="720px"
    destroy-on-close
    @ok="handleOk"
    @cancel="emit('update:open', false)"
    @update:open="(v: boolean) => emit('update:open', v)"
  >
    <a-form layout="vertical" class="pt-2">
      <a-form-item label="别名" required>
        <a-input v-model:value="form.alias" placeholder="如：我家、候选 A" :maxlength="64" />
      </a-form-item>

      <a-form-item label="地址（搜索或点选地图）" required>
        <AmapPicker
          :model-value="pickerValue"
          @update:model-value="handlePickerChange"
        />
      </a-form-item>

      <a-form-item label="备注">
        <a-textarea v-model:value="form.note" :rows="2" placeholder="如：月租 5000、地铁 5 分钟" />
      </a-form-item>

      <a-form-item>
        <a-checkbox v-model:checked="form.is_default">设为默认住址</a-checkbox>
      </a-form-item>
    </a-form>
  </a-modal>
</template>
