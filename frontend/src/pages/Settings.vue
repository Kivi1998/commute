<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import ProfileForm from '@/components/ProfileForm.vue'
import AddressManager from '@/components/AddressManager.vue'
import { fetchEnums, type Enums } from '@/api/meta'

const activeTab = ref<string>('profile')
const enums = ref<Enums | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    enums.value = await fetchEnums()
  } catch (e) {
    message.error('加载字典失败：' + (e as Error).message)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="max-w-5xl mx-auto">
    <a-spin :spinning="loading">
      <a-tabs v-model:active-key="activeTab" class="bg-white rounded-md px-4 pb-4">
        <a-tab-pane key="profile" tab="个人画像">
          <ProfileForm v-if="enums" :enums="enums" />
        </a-tab-pane>
        <a-tab-pane key="addresses" tab="家庭住址">
          <AddressManager />
        </a-tab-pane>
      </a-tabs>
    </a-spin>
  </div>
</template>
