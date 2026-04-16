<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import type { Enums } from '@/api/meta'
import {
  fetchProfile,
  upsertProfile,
  type CompanyType,
  type ProfileUpsertInput,
} from '@/api/profile'

defineProps<{ enums: Enums }>()

const form = reactive<ProfileUpsertInput>({
  full_name: '',
  phone: '',
  email: '',
  current_city: '',
  current_city_code: '',
  target_position: '',
  experience_years: undefined,
  preferred_company_types: [],
})

const submitting = ref(false)
const loading = ref(true)

onMounted(async () => {
  try {
    const p = await fetchProfile()
    if (p) {
      form.full_name = p.full_name ?? ''
      form.phone = p.phone ?? ''
      form.email = p.email ?? ''
      form.current_city = p.current_city
      form.current_city_code = p.current_city_code ?? ''
      form.target_position = p.target_position
      form.experience_years = p.experience_years
      form.preferred_company_types = p.preferred_company_types ?? []
    }
  } catch {
    // 客户端拦截器已提示
  } finally {
    loading.value = false
  }
})

const cityOptions = [
  { code: '110000', name: '北京' },
  { code: '310000', name: '上海' },
  { code: '440100', name: '广州' },
  { code: '440300', name: '深圳' },
  { code: '330100', name: '杭州' },
  { code: '510100', name: '成都' },
  { code: '320100', name: '南京' },
  { code: '420100', name: '武汉' },
]

function handleCityChange(name: string) {
  const item = cityOptions.find((c) => c.name === name)
  form.current_city_code = item?.code ?? ''
}

async function handleSubmit() {
  if (!form.current_city || !form.target_position) {
    message.warning('请填写城市与岗位')
    return
  }
  submitting.value = true
  try {
    const payload: ProfileUpsertInput = {
      full_name: form.full_name?.trim() || undefined,
      phone: form.phone?.trim() || undefined,
      email: form.email?.trim() || undefined,
      current_city: form.current_city,
      current_city_code: form.current_city_code || undefined,
      target_position: form.target_position,
      experience_years: form.experience_years,
      preferred_company_types: form.preferred_company_types as CompanyType[],
    }
    await upsertProfile(payload)
    message.success('画像已保存')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-spin :spinning="loading">
    <a-form layout="vertical" class="max-w-xl pt-2">
      <div class="text-xs text-slate-500 mb-2">联系人信息（用于地址复制卡片展示）</div>
      <div class="grid grid-cols-3 gap-3">
        <a-form-item label="姓名">
          <a-input v-model:value="form.full_name" :maxlength="32" placeholder="如：张三" />
        </a-form-item>
        <a-form-item label="电话">
          <a-input v-model:value="form.phone" :maxlength="20" placeholder="手机号" />
        </a-form-item>
        <a-form-item label="邮箱">
          <a-input v-model:value="form.email" :maxlength="128" placeholder="name@example.com" />
        </a-form-item>
      </div>

      <div class="text-xs text-slate-500 mt-2 mb-2">求职信息</div>
      <a-form-item label="当前所在城市" required>
        <a-select
          v-model:value="form.current_city"
          show-search
          placeholder="选择或搜索城市"
          :options="cityOptions.map((c) => ({ value: c.name, label: c.name }))"
          @change="(v: any) => handleCityChange(v as string)"
        />
      </a-form-item>

      <a-form-item label="求职岗位" required>
        <a-input
          v-model:value="form.target_position"
          placeholder="如：后台开发、产品经理"
          :maxlength="128"
          show-count
        />
      </a-form-item>

      <a-form-item label="工作经验（年）">
        <a-input-number
          v-model:value="form.experience_years"
          :min="0"
          :max="30"
          placeholder="0-30"
          class="!w-40"
        />
      </a-form-item>

      <a-form-item label="偏好公司类型">
        <a-checkbox-group
          v-model:value="form.preferred_company_types"
          :options="enums.company_type.map((e) => ({ value: e.value, label: e.label }))"
        />
      </a-form-item>

      <a-form-item>
        <a-button type="primary" :loading="submitting" @click="handleSubmit">
          保存画像
        </a-button>
      </a-form-item>
    </a-form>
  </a-spin>
</template>
