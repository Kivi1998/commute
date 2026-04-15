<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal, type TableColumnsType } from 'ant-design-vue'
import {
  CarOutlined,
  EnvironmentOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import {
  deleteCommuteQuery,
  listCommuteQueries,
  type CommuteQueryListItem,
  type TransportMode,
} from '@/api/commute'

dayjs.extend(relativeTime)

const router = useRouter()

const list = ref<CommuteQueryListItem[]>([])
const loading = ref(false)

const modeIconLabel: Record<TransportMode, string> = {
  transit: '🚇 公交',
  driving: '🚗 驾车',
  cycling: '🚴 骑行',
  walking: '🚶 步行',
}

async function refresh() {
  loading.value = true
  try {
    list.value = await listCommuteQueries()
  } finally {
    loading.value = false
  }
}

onMounted(refresh)

function handleRestore(record: CommuteQueryListItem) {
  router.push({ path: '/commute', query: { from_query: record.id } })
}

function handleDelete(record: CommuteQueryListItem) {
  Modal.confirm({
    title: '删除这条查询历史？',
    content: '关联的计算结果将解除关联（不删除缓存）。',
    okType: 'danger',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      await deleteCommuteQuery(record.id)
      message.success('已删除')
      refresh()
    },
  })
}

const columns: TableColumnsType<CommuteQueryListItem> = [
  { title: '时间', key: 'time', width: 200 },
  { title: '家', key: 'home' },
  { title: '公司', key: 'companies' },
  { title: '出行方式', key: 'modes', width: 180 },
  { title: '早/晚', key: 'times', width: 120 },
  { title: '操作', key: 'action', width: 180, fixed: 'right' },
]
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <a-card :bordered="false">
      <template #title>
        <span>历史查询</span>
        <span class="text-xs text-slate-400 ml-2 font-normal">
          每次通勤计算都会自动保存
        </span>
      </template>
      <template #extra>
        <a-button size="small" @click="refresh">
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="list"
        :loading="loading"
        row-key="id"
        :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 条记录` }"
      >
        <template #bodyCell="slotProps">
          <template v-if="slotProps.column.key === 'time'">
            <div class="text-slate-800">
              {{ dayjs(slotProps.record.created_at).format('MM-DD HH:mm') }}
            </div>
            <div class="text-xs text-slate-400">
              {{ dayjs(slotProps.record.created_at).fromNow() }}
            </div>
          </template>

          <template v-else-if="slotProps.column.key === 'home'">
            <div class="flex items-center gap-1 text-slate-700">
              <EnvironmentOutlined class="!text-blue-500" />
              <span class="font-medium">{{ slotProps.record.home_alias }}</span>
            </div>
            <div class="text-xs text-slate-400 truncate max-w-xs">
              {{ slotProps.record.home_address }}
            </div>
          </template>

          <template v-else-if="slotProps.column.key === 'companies'">
            <div class="text-sm">
              <a-tag color="blue">{{ slotProps.record.company_count }} 家</a-tag>
              <span class="text-slate-600">
                {{ slotProps.record.company_names.slice(0, 3).join(' / ') }}
                <span v-if="slotProps.record.company_names.length > 3" class="text-slate-400">
                  等 {{ slotProps.record.company_names.length }} 家
                </span>
              </span>
            </div>
          </template>

          <template v-else-if="slotProps.column.key === 'modes'">
            <a-space :size="4" wrap>
              <a-tag
                v-for="m in slotProps.record.transport_modes"
                :key="m"
                class="!m-0"
              >
                {{ modeIconLabel[m as TransportMode] ?? m }}
              </a-tag>
            </a-space>
          </template>

          <template v-else-if="slotProps.column.key === 'times'">
            <div class="text-sm text-slate-600">
              {{ slotProps.record.morning_time }} 出
            </div>
            <div class="text-sm text-slate-600">
              {{ slotProps.record.evening_time }} 回
            </div>
          </template>

          <template v-else-if="slotProps.column.key === 'action'">
            <a-space size="small">
              <a-button size="small" type="link" @click="handleRestore(slotProps.record as CommuteQueryListItem)">
                <template #icon><ReloadOutlined /></template>
                恢复并重算
              </a-button>
              <a-button size="small" type="link" danger @click="handleDelete(slotProps.record as CommuteQueryListItem)">
                删除
              </a-button>
            </a-space>
          </template>
        </template>

        <template #emptyText>
          <div class="py-8 text-slate-400">
            <CarOutlined class="text-3xl mb-2" />
            <div>还没有查询历史，去「通勤」页开始第一次计算</div>
            <a-button type="link" class="mt-2" @click="router.push('/commute')">
              去查询 →
            </a-button>
          </div>
        </template>
      </a-table>
    </a-card>
  </div>
</template>
