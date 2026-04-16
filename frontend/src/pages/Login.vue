<script setup lang="ts">
import { reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { LockOutlined, UserOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = reactive({
  email: '',
  password: '',
})

async function handleLogin() {
  if (!form.email || !form.password) {
    message.warning('请填写账号和密码')
    return
  }
  try {
    await auth.login({ email: form.email, password: form.password })
    message.success(`欢迎回来，${auth.displayName}`)
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch {
    // 拦截器已提示
  }
}

function quickFill(email: string, password: string) {
  form.email = email
  form.password = password
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50 flex items-center justify-center px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <div class="text-4xl mb-2">🚇</div>
        <h1 class="text-2xl font-semibold text-slate-800">通勤查询系统</h1>
        <p class="text-sm text-slate-500 mt-1">Commute</p>
      </div>

      <a-card :bordered="false" class="shadow-lg">
        <a-form layout="vertical" @submit.prevent="handleLogin">
          <a-form-item label="账号">
            <a-input
              v-model:value="form.email"
              size="large"
              placeholder="账号名"
              autocomplete="username"
            >
              <template #prefix><UserOutlined class="text-slate-400" /></template>
            </a-input>
          </a-form-item>

          <a-form-item label="密码">
            <a-input-password
              v-model:value="form.password"
              size="large"
              placeholder="请输入密码"
              autocomplete="current-password"
              @press-enter="handleLogin"
            >
              <template #prefix><LockOutlined class="text-slate-400" /></template>
            </a-input-password>
          </a-form-item>

          <a-form-item class="!mb-2">
            <a-button
              type="primary"
              size="large"
              block
              :loading="auth.loading"
              @click="handleLogin"
            >
              登录
            </a-button>
          </a-form-item>
        </a-form>

        <a-divider class="!my-4">
          <span class="text-xs text-slate-400">测试账号</span>
        </a-divider>

        <div class="space-y-2 text-xs">
          <div
            class="flex justify-between items-center p-2 rounded bg-slate-50 hover:bg-slate-100 cursor-pointer"
            @click="quickFill('kivi', '542426')"
          >
            <div>
              <div class="font-medium text-slate-700">kivi</div>
              <div class="text-slate-400">主账号（含已有数据）</div>
            </div>
            <a-button size="small" type="link">填入</a-button>
          </div>
          <div
            class="flex justify-between items-center p-2 rounded bg-slate-50 hover:bg-slate-100 cursor-pointer"
            @click="quickFill('dudu', '311416')"
          >
            <div>
              <div class="font-medium text-slate-700">dudu</div>
              <div class="text-slate-400">第二账号</div>
            </div>
            <a-button size="small" type="link">填入</a-button>
          </div>
        </div>
      </a-card>
    </div>
  </div>
</template>
