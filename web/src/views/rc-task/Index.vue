<template>
  <div class="page-container">
    <div class="page-header">
      <h2>RC任务管理</h2>
      <el-button type="primary" @click="showCreate">新建任务</el-button>
    </div>

    <el-card>
      <div class="filter-bar" style="margin-bottom: 12px; display: flex; gap: 12px;">
        <el-select v-model="filterZone" placeholder="按网区筛选" clearable style="width: 180px;" @change="loadData">
          <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
        </el-select>
        <el-input v-model="filterKeyword" placeholder="搜索名称" clearable style="width: 200px;" @clear="loadData" @keyup.enter="loadData" />
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="task_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : row.status === 'error' ? 'danger' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="rule_group" label="规则组" width="120" />
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <el-button size="small" @click="showRules(row)">规则</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑RC任务' : '新建RC任务'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="查询地址">
          <el-input v-model="form.query_address" placeholder="http://vmselect:8481" />
        </el-form-item>
        <el-form-item label="规则组">
          <el-input v-model="form.rule_group" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="ruleDialogVisible" title="规则管理" width="800px">
      <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <span>任务: {{ currentTaskName }}</span>
        <el-button type="primary" size="small" @click="showCreateRule">添加规则</el-button>
      </div>
      <el-table :data="rules" stripe>
        <el-table-column prop="rule_id" label="ID" width="110" />
        <el-table-column prop="alert_name" label="告警名" />
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="severity" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : row.severity === 'warning' ? 'warning' : 'info'" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="持续时间" width="90" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="showEditRule(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="ruleFormVisible" :title="isEditRule ? '编辑规则' : '添加规则'" width="550px">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="告警名" required>
          <el-input v-model="ruleForm.alert_name" />
        </el-form-item>
        <el-form-item label="模块">
          <el-input v-model="ruleForm.module" placeholder="metrics" />
        </el-form-item>
        <el-form-item label="表达式" required>
          <el-input v-model="ruleForm.expr" type="textarea" :rows="2" placeholder="PromQL expression" />
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input v-model="ruleForm.duration" placeholder="5m" />
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="ruleForm.severity">
            <el-option label="info" value="info" />
            <el-option label="warning" value="warning" />
            <el-option label="critical" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="ruleForm.labels" placeholder='{"team":"ops"}' />
        </el-form-item>
        <el-form-item label="注解">
          <el-input v-model="ruleForm.annotations" placeholder='{"summary":"..."}' type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitRule" :loading="submittingRule">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { rcTaskApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filterZone = ref('')
const filterKeyword = ref('')
const form = ref({ name: '', zone_id: '', description: '', query_address: '', rule_group: '' })

const ruleDialogVisible = ref(false)
const ruleFormVisible = ref(false)
const currentTaskId = ref('')
const currentTaskName = ref('')
const rules = ref([])
const isEditRule = ref(false)
const editRuleId = ref('')
const submittingRule = ref(false)
const ruleForm = ref({ alert_name: '', module: 'metrics', expr: '', duration: '5m', severity: 'warning', labels: '', annotations: '' })

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    if (filterZone.value) params.zone_id = filterZone.value
    if (filterKeyword.value) params.keyword = filterKeyword.value
    const res = await rcTaskApi.list(params)
    items.value = res.data?.items || []
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
  form.value = { name: '', zone_id: '', description: '', query_address: '', rule_group: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.task_id
  form.value = { name: row.name, zone_id: row.zone_id, description: row.description, query_address: row.query_address || '', rule_group: row.rule_group || '' }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await rcTaskApi.update(editId.value, form.value)
    } else {
      await rcTaskApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该RC任务？', '提示', { type: 'warning' })
  await rcTaskApi.delete(row.task_id)
  ElMessage.success('删除成功')
  loadData()
}

const showRules = async (row) => {
  currentTaskId.value = row.task_id
  currentTaskName.value = row.name
  const res = await rcTaskApi.listRules(row.task_id)
  rules.value = res.data || []
  ruleDialogVisible.value = true
}

const showCreateRule = () => {
  isEditRule.value = false
  ruleForm.value = { alert_name: '', module: 'metrics', expr: '', duration: '5m', severity: 'warning', labels: '', annotations: '' }
  ruleFormVisible.value = true
}

const showEditRule = (row) => {
  isEditRule.value = true
  editRuleId.value = row.rule_id
  ruleForm.value = { alert_name: row.alert_name, module: row.module, expr: row.expr, duration: row.duration, severity: row.severity, labels: row.labels || '', annotations: row.annotations || '' }
  ruleFormVisible.value = true
}

const handleSubmitRule = async () => {
  submittingRule.value = true
  try {
    if (isEditRule.value) {
      await rcTaskApi.updateRule(editRuleId.value, ruleForm.value)
    } else {
      await rcTaskApi.createRule(currentTaskId.value, ruleForm.value)
    }
    ruleFormVisible.value = false
    const res = await rcTaskApi.listRules(currentTaskId.value)
    rules.value = res.data || []
  } finally {
    submittingRule.value = false
  }
}

const handleDeleteRule = async (row) => {
  await ElMessageBox.confirm('确认删除该规则？', '提示', { type: 'warning' })
  await rcTaskApi.deleteRule(row.rule_id)
  ElMessage.success('删除成功')
  const res = await rcTaskApi.listRules(currentTaskId.value)
  rules.value = res.data || []
}

onMounted(() => { loadData(); loadZones() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
