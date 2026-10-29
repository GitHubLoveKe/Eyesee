<template>
  <div class="page-container">
    <div class="page-header">
      <h2>任务类型管理</h2>
      <el-button type="primary" @click="showCreate">新建任务类型</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="task_type_id" label="ID" width="130" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="source_type" label="数据源类型" width="140" />
        <el-table-column prop="collector" label="采集器" width="140" />
        <el-table-column prop="is_builtin" label="内置" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_builtin ? 'success' : 'info'" size="small">
              {{ row.is_builtin ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="80" />
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.is_builtin">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑任务类型' : '新建任务类型'" width="550px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="数据源类型" required>
          <el-input v-model="form.source_type" placeholder="如: prometheus, snmp, ping" />
        </el-form-item>
        <el-form-item label="采集器" required>
          <el-input v-model="form.collector" placeholder="如: telegraf, custom_exporter" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="默认配置">
          <el-input v-model="form.default_config" type="textarea" :rows="4" placeholder='{"interval":"30s"}' />
        </el-form-item>
        <el-form-item label="内置">
          <el-switch v-model="form.is_builtin" />
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
import { taskTypeApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const form = ref({
  name: '',
  source_type: '',
  collector: '',
  description: '',
  default_config: '',
  is_builtin: false,
})

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await taskTypeApi.list()
    items.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = {
    name: '',
    source_type: '',
    collector: '',
    description: '',
    default_config: '',
    is_builtin: false,
  }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.task_type_id
  form.value = {
    name: row.name,
    source_type: row.source_type || '',
    collector: row.collector || '',
    description: row.description || '',
    default_config: typeof row.default_config === 'object'
      ? JSON.stringify(row.default_config, null, 2)
      : (row.default_config || ''),
    is_builtin: !!row.is_builtin,
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    const payload = { ...form.value }
    if (payload.default_config) {
      try {
        payload.default_config = JSON.parse(payload.default_config)
      } catch {
        // keep as string if not valid JSON
      }
    } else {
      payload.default_config = null
    }
    if (isEdit.value) {
      await taskTypeApi.update(editId.value, payload)
    } else {
      await taskTypeApi.create(payload)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该任务类型？', '提示', { type: 'warning' })
  await taskTypeApi.delete(row.task_type_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => loadData())
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
