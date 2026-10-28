<template>
  <div class="page-container">
    <div class="page-header">
      <h2>定时任务</h2>
      <el-button type="primary" @click="showCreate">新建任务</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="task_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="task_type" label="类型" width="140" />
        <el-table-column prop="cron_expr" label="Cron表达式" width="140" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-switch :model-value="row.enabled" @change="handleToggle(row)" size="small" />
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_run_at" label="上次执行" width="170">
          <template #default="{ row }">
            {{ row.last_run_at ? formatTime(row.last_run_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button size="small" @click="showExecutions(row)">历史</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑任务' : '新建任务'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.task_type">
            <el-option label="metric_archive" value="metric_archive" />
            <el-option label="data_cleanup" value="data_cleanup" />
            <el-option label="report_gen" value="report_gen" />
            <el-option label="sync" value="sync" />
          </el-select>
        </el-form-item>
        <el-form-item label="Cron表达式" required>
          <el-input v-model="form.cron_expr" placeholder="0 2 * * *" />
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="execDialogVisible" title="执行历史" width="700px">
      <el-table :data="executions" stripe>
        <el-table-column prop="execution_id" label="ID" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="started_at" label="开始时间" width="170">
          <template #default="{ row }">{{ formatTime(row.started_at) }}</template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时(ms)" width="100" />
        <el-table-column prop="error_message" label="错误信息">
          <template #default="{ row }">{{ row.error_message || '-' }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { scheduledTaskApi } from '@/api/modules'

const items = ref([])
const executions = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const execDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const form = ref({ name: '', task_type: 'metric_archive', cron_expr: '', config: '' })

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await scheduledTaskApi.list()
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', task_type: 'metric_archive', cron_expr: '', config: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.task_id
  form.value = { name: row.name, task_type: row.task_type, cron_expr: row.cron_expr, config: row.config }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await scheduledTaskApi.update(editId.value, form.value)
    } else {
      await scheduledTaskApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该定时任务？', '提示', { type: 'warning' })
  await scheduledTaskApi.delete(row.task_id)
  ElMessage.success('删除成功')
  loadData()
}

const handleToggle = async (row) => {
  if (row.enabled) {
    await scheduledTaskApi.disable(row.task_id)
  } else {
    await scheduledTaskApi.enable(row.task_id)
  }
  loadData()
}

const showExecutions = async (row) => {
  const res = await scheduledTaskApi.listExecutions(row.task_id, { limit: 20 })
  executions.value = res.data || []
  execDialogVisible.value = true
}

onMounted(() => loadData())
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
