<template>
  <div class="page-container">
    <div class="page-header">
      <h2>指标管理</h2>
      <el-button type="primary" @click="showCreate">新建指标</el-button>
    </div>

    <el-card>
      <div class="filter-bar" style="margin-bottom: 12px; display: flex; gap: 12px;">
        <el-input v-model="filterCategory" placeholder="按分类筛选" clearable style="width: 180px;" @clear="loadData" @keyup.enter="loadData" />
        <el-input v-model="filterKeyword" placeholder="搜索名称" clearable style="width: 200px;" @clear="loadData" @keyup.enter="loadData" />
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="metric_id" label="ID" width="130" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="metric_type" label="类型" width="100" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="help" label="说明" show-overflow-tooltip />
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button size="small" @click="showQueries(row)">查询</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑指标' : '新建指标'" width="600px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="metric" />
        </el-form-item>
        <el-form-item label="指标类型">
          <el-select v-model="form.metric_type" clearable>
            <el-option label="counter" value="counter" />
            <el-option label="gauge" value="gauge" />
            <el-option label="histogram" value="histogram" />
            <el-option label="summary" value="summary" />
          </el-select>
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="form.unit" placeholder="bytes, seconds, etc." />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.help" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="PromQL模板">
          <el-input v-model="form.promql_template" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="变量">
          <el-input v-model="form.variables" placeholder='{"var1":"desc"}' />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.labels" placeholder='{"job":"node"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="queryDialogVisible" title="常用查询" width="750px">
      <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <span>指标: {{ currentMetricName }}</span>
        <el-button type="primary" size="small" @click="showCreateQuery">添加查询</el-button>
      </div>
      <el-table :data="queries" stripe>
        <el-table-column prop="query_id" label="ID" width="110" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="promql" label="PromQL" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="handleDeleteQuery(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="queryFormVisible" title="添加查询" width="550px">
      <el-form :model="queryForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="queryForm.name" />
        </el-form-item>
        <el-form-item label="PromQL" required>
          <el-input v-model="queryForm.promql" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="queryForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="queryForm.category" />
        </el-form-item>
        <el-form-item label="变量">
          <el-input v-model="queryForm.variables" placeholder='{"var1":"value"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="queryFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitQuery" :loading="submittingQuery">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { metricDefinitionApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filterCategory = ref('')
const filterKeyword = ref('')
const form = ref({ name: '', category: 'metric', metric_type: '', unit: '', help: '', promql_template: '', variables: '', labels: '' })

const queryDialogVisible = ref(false)
const queryFormVisible = ref(false)
const currentMetricId = ref('')
const currentMetricName = ref('')
const queries = ref([])
const submittingQuery = ref(false)
const queryForm = ref({ name: '', promql: '', description: '', category: '', variables: '' })

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    if (filterCategory.value) params.category = filterCategory.value
    if (filterKeyword.value) params.keyword = filterKeyword.value
    const res = await metricDefinitionApi.list(params)
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', category: 'metric', metric_type: '', unit: '', help: '', promql_template: '', variables: '', labels: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.metric_id
  form.value = {
    name: row.name, category: row.category, metric_type: row.metric_type || '',
    unit: row.unit || '', help: row.help || '', promql_template: row.promql_template || '',
    variables: row.variables || '', labels: row.labels || ''
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    const payload = { ...form.value }
    Object.keys(payload).forEach(k => { if (payload[k] === '') payload[k] = null })
    if (isEdit.value) {
      await metricDefinitionApi.update(editId.value, payload)
    } else {
      await metricDefinitionApi.create(payload)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该指标定义？', '提示', { type: 'warning' })
  await metricDefinitionApi.delete(row.metric_id)
  ElMessage.success('删除成功')
  loadData()
}

const showQueries = async (row) => {
  currentMetricId.value = row.metric_id
  currentMetricName.value = row.name
  const res = await metricDefinitionApi.listQueries(row.metric_id)
  queries.value = res.data || []
  queryDialogVisible.value = true
}

const showCreateQuery = () => {
  queryForm.value = { name: '', promql: '', description: '', category: '', variables: '' }
  queryFormVisible.value = true
}

const handleSubmitQuery = async () => {
  submittingQuery.value = true
  try {
    const payload = { ...queryForm.value }
    Object.keys(payload).forEach(k => { if (payload[k] === '') payload[k] = null })
    await metricDefinitionApi.createQuery(currentMetricId.value, payload)
    queryFormVisible.value = false
    const res = await metricDefinitionApi.listQueries(currentMetricId.value)
    queries.value = res.data || []
  } finally {
    submittingQuery.value = false
  }
}

const handleDeleteQuery = async (row) => {
  await ElMessageBox.confirm('确认删除该查询？', '提示', { type: 'warning' })
  await metricDefinitionApi.deleteQuery(row.query_id)
  ElMessage.success('删除成功')
  const res = await metricDefinitionApi.listQueries(currentMetricId.value)
  queries.value = res.data || []
}

onMounted(() => loadData())
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
