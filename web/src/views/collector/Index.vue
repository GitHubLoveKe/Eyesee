<template>
  <div class="page-container">
    <div class="page-header">
      <h2>采集器管理</h2>
      <el-button type="primary" @click="showCreate">注册采集器</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="node_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="node_type" label="类型" width="120" />
        <el-table-column prop="address" label="地址" width="180" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' || row.status === 'healthy' ? 'success' : 'danger'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="task_count" label="任务数" width="80" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑采集器' : '注册采集器'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.node_type">
            <el-option label="vmagent" value="vmagent" />
            <el-option label="vmalert" value="vmalert" />
            <el-option label="otel-collector" value="otel-collector" />
            <el-option label="prometheus" value="prometheus" />
            <el-option label="alloy" value="alloy" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" required>
          <el-input v-model="form.address" placeholder="http://ip:port" />
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { collectorApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = ref({ name: '', node_type: 'vmagent', address: '', zone_id: '' })
const editId = ref('')

const loadData = async () => {
  loading.value = true
  try {
    const res = await collectorApi.list()
    items.value = res.data || []
  } finally {
    loading.value = false
  }
}

const loadZones = async () => {
  const res = await zoneApi.list()
  zones.value = res.data?.items || []
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', node_type: 'vmagent', address: '', zone_id: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.node_id
  form.value = { name: row.name, node_type: row.node_type, address: row.address, zone_id: row.zone_id }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await collectorApi.update(editId.value, form.value)
    } else {
      await collectorApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该采集器？', '提示', { type: 'warning' })
  await collectorApi.delete(row.node_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadZones() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
