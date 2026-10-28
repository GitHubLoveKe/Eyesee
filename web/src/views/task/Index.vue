<template>
  <div class="page-container">
    <div class="page-header">
      <h2>采集任务管理</h2>
      <el-button type="primary" @click="showCreate">新建任务</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="task_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="instance_id" label="实例ID" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-switch :model-value="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑任务' : '新建任务'" width="550px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="实例ID" required>
          <el-input v-model="form.instance_id" />
        </el-form-item>
        <el-form-item label="类型ID" required>
          <el-input v-model="form.type_id" />
        </el-form-item>
        <el-form-item label="模块">
          <el-select v-model="form.module">
            <el-option label="metrics" value="metrics" />
            <el-option label="logs" value="logs" />
            <el-option label="traces" value="traces" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="4" />
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
import { collectionTaskApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = ref({ name: '', instance_id: '', type_id: '', module: 'metrics', config: '' })
const editId = ref('')

const statusType = (s) => ({ running: 'success', pending: 'warning', error: 'danger', stopped: 'info' }[s] || 'info')

const loadData = async () => {
  loading.value = true
  try {
    const res = await collectionTaskApi.list()
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', instance_id: '', type_id: '', module: 'metrics', config: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.task_id
  form.value = { name: row.name, instance_id: row.instance_id, type_id: row.type_id, module: row.module, config: row.config }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await collectionTaskApi.update(editId.value, form.value)
    } else {
      await collectionTaskApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const toggleEnabled = async (row) => {
  if (row.enabled) {
    await collectionTaskApi.disable(row.task_id)
  } else {
    await collectionTaskApi.enable(row.task_id)
  }
  loadData()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该采集任务？', '提示', { type: 'warning' })
  await collectionTaskApi.delete(row.task_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
