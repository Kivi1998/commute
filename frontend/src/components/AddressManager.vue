<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message, Modal, type TableColumnsType } from 'ant-design-vue'
import { PlusOutlined, StarFilled } from '@ant-design/icons-vue'
import {
  createAddress,
  deleteAddress,
  listAddresses,
  setDefaultAddress,
  updateAddress,
  type HomeAddress,
  type HomeAddressCreateInput,
} from '@/api/address'
import AddressFormModal from './AddressFormModal.vue'

const list = ref<HomeAddress[]>([])
const loading = ref(false)

const modalState = reactive<{
  open: boolean
  mode: 'create' | 'edit'
  record: HomeAddress | null
}>({
  open: false,
  mode: 'create',
  record: null,
})

async function refresh() {
  loading.value = true
  try {
    list.value = await listAddresses()
  } finally {
    loading.value = false
  }
}

onMounted(refresh)

function openCreate() {
  modalState.mode = 'create'
  modalState.record = null
  modalState.open = true
}

function openEdit(record: HomeAddress) {
  modalState.mode = 'edit'
  modalState.record = record
  modalState.open = true
}

async function handleSubmit(payload: HomeAddressCreateInput) {
  if (modalState.mode === 'create') {
    await createAddress(payload)
    message.success('已添加住址')
  } else if (modalState.record) {
    await updateAddress(modalState.record.id, payload)
    message.success('已更新住址')
  }
  modalState.open = false
  refresh()
}

function handleSetDefault(record: HomeAddress) {
  if (record.is_default) return
  Modal.confirm({
    title: '设为默认住址？',
    content: `将"${record.alias}"设为默认后，通勤计算将以此为起点。`,
    onOk: async () => {
      await setDefaultAddress(record.id)
      message.success('已切换默认住址')
      refresh()
    },
  })
}

function handleDelete(record: HomeAddress) {
  Modal.confirm({
    title: `删除住址"${record.alias}"？`,
    content: record.is_default
      ? '这是当前默认住址，删除后将自动提升最早创建的另一条为默认。'
      : '删除后不可恢复。',
    okType: 'danger',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      await deleteAddress(record.id)
      message.success('已删除')
      refresh()
    },
  })
}

const columns: TableColumnsType<HomeAddress> = [
  { title: '别名', dataIndex: 'alias', width: 140 },
  { title: '地址', dataIndex: 'address', ellipsis: true },
  { title: '经纬度', key: 'coord', width: 200 },
  { title: '默认', key: 'is_default', width: 80, align: 'center' },
  { title: '备注', dataIndex: 'note', ellipsis: true },
  { title: '操作', key: 'action', width: 220, fixed: 'right' },
]
</script>

<template>
  <div class="pt-2">
    <div class="flex justify-between items-center mb-3">
      <div class="text-slate-600 text-sm">
        支持多个家庭住址，用于租房选址对比。通勤计算将以<b>默认住址</b>为起点。
      </div>
      <a-button type="primary" @click="openCreate">
        <template #icon><PlusOutlined /></template>
        新增住址
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="list"
      :loading="loading"
      :pagination="false"
      row-key="id"
      size="middle"
    >
      <template #bodyCell="slotProps">
        <template v-if="slotProps.column.key === 'coord'">
          <span class="font-mono text-xs text-slate-500">
            {{ slotProps.record.longitude.toFixed(6) }}, {{ slotProps.record.latitude.toFixed(6) }}
          </span>
        </template>
        <template v-else-if="slotProps.column.key === 'is_default'">
          <StarFilled v-if="slotProps.record.is_default" class="!text-amber-500" />
          <span v-else class="text-slate-300">—</span>
        </template>
        <template v-else-if="slotProps.column.key === 'action'">
          <a-space size="small">
            <a-button size="small" type="link" @click="openEdit(slotProps.record as HomeAddress)">编辑</a-button>
            <a-button
              size="small"
              type="link"
              :disabled="slotProps.record.is_default"
              @click="handleSetDefault(slotProps.record as HomeAddress)"
            >设为默认</a-button>
            <a-button size="small" type="link" danger @click="handleDelete(slotProps.record as HomeAddress)">删除</a-button>
          </a-space>
        </template>
      </template>
      <template #emptyText>
        <a-empty description="还没有住址，点右上角添加" />
      </template>
    </a-table>

    <AddressFormModal
      v-model:open="modalState.open"
      :mode="modalState.mode"
      :record="modalState.record"
      @submit="handleSubmit"
    />
  </div>
</template>
