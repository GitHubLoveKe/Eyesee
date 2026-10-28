<template>
  <div class="page-container">
    <div class="page-header">
      <h2>告警策略</h2>
      <el-button type="primary" @click="showCreate">新建策略</el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-select v-model="filter.severity" placeholder="严重级别" clearable style="width:140px" @change="loadData">
          <el-option label="critical" value="critical" />
          <el-option label="warning" value="warning" />
          <el-option label="info" value="info" />
        </el-select>
        <el-input v-model="filter.keyword" placeholder="搜索名称" clearable style="width:220px" @clear="loadData" @keyup.enter="loadData" />
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="strategy_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : row.severity === 'warning' ? 'warning' : 'info'">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scope_type" label="作用域" width="100" />
        <el-table-column prop="duration" label="持续时间" width="100" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑策略' : '新建策略'" width="650px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="指标ID" required>
          <el-input v-model="form.metric_id" />
        </el-form-item>
        <el-form-item label="表达式" required>
          <el-input v-model="form.expr_template" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input v-model="form.duration" placeholder="5m" />
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="form.severity">
            <el-option label="critical" value="critical" />
            <el-option label="warning" value="warning" />
            <el-option label="info" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="作用域类型">
          <el-select v-model="form.scope_type">
            <el-option label="instance" value="instance" />
            <el-option label="type" value="type" />
            <el-option label="zone" value="zone" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警组">
          <el-select v-model="form.alert_group_id" clearable>
            <el-option v-for="g in alertGroups" :key="g.group_id" :label="g.name" :value="g.group_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="通知渠道">
          <el-select v-model="form.channel_id" clearable>
            <el-option v-for="c in channels" :key="c.channel_id" :label="c.name" :value="c.channel_id" />
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
import { alertStrategyApi, alertGroupApi, notificationChannelApi } from '@/api/modules'

const items = ref([])
const alertGroups = ref([])
const channels = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filter = ref({ severity: '', keyword: '' })
const form = ref({
  name: '', description: '', metric_id: '', expr_template: '', duration: '5m',
  severity: 'warning', scope_type: 'instance', scope_value: '',
  alert_group_id: null, channel_id: null, labels: '', annotations: ''
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await alertStrategyApi.list({ severity: filter.value.severity, keyword: filter.value.keyword })
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  const [gRes, cRes] = await Promise.all([alertGroupApi.list(), notificationChannelApi.list()])
  alertGroups.value = gRes.data || []
  channels.value = cRes.data || []
}

const showCreate = () => {
  isEdit.value = false
  form.value = {
    name: '', description: '', metric_id: '', expr_template: '', duration: '5m',
    severity: 'warning', scope_type: 'instance', scope_value: '',
    alert_group_id: null, channel_id: null, labels: '', annotations: ''
  }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.strategy_id
  form.value = { ...row }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await alertStrategyApi.update(editId.value, form.value)
    } else {
      await alertStrategyApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该告警策略？', '提示', { type: 'warning' })
  await alertStrategyApi.delete(row.strategy_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadOptions() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
