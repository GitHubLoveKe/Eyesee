<template>
  <div class="page-container">
    <div class="page-header">
      <h2>网区管理</h2>
      <el-button type="primary" @click="showCreate">新建网区</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="zone_id" label="ID" width="140" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="collector_type" label="采集器类型" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button size="small" @click="showSegments(row)">网段</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑网区' : '新建网区'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" />
        </el-form-item>
        <el-form-item label="采集器类型">
          <el-select v-model="form.collector_type">
            <el-option label="vmagent" value="vmagent" />
            <el-option label="prometheus" value="prometheus" />
            <el-option label="alloy" value="alloy" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="segmentDialogVisible" title="网段管理" width="600px">
      <el-button size="small" type="primary" @click="showCreateSegment" style="margin-bottom: 12px">添加网段</el-button>
      <el-table :data="segments" stripe size="small">
        <el-table-column prop="cidr" label="CIDR" />
        <el-table-column prop="name" label="名称" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button size="small" type="danger" link @click="handleDeleteSegment(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="segmentFormVisible" title="添加网段" width="400px">
      <el-form :model="segmentForm" label-width="60px">
        <el-form-item label="CIDR" required>
          <el-input v-model="segmentForm.cidr" placeholder="10.0.1.0/24" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="segmentForm.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="segmentFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitSegment">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { zoneApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = ref({ name: '', description: '', collector_type: 'vmagent' })
const editId = ref('')

const segmentDialogVisible = ref(false)
const segmentFormVisible = ref(false)
const segments = ref([])
const currentZoneId = ref('')
const segmentForm = ref({ cidr: '', name: '' })

const loadData = async () => {
  loading.value = true
  try {
    const res = await zoneApi.list()
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', description: '', collector_type: 'vmagent' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.zone_id
  form.value = { name: row.name, description: row.description, collector_type: row.collector_type }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await zoneApi.update(editId.value, form.value)
    } else {
      await zoneApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该网区？', '提示', { type: 'warning' })
  await zoneApi.delete(row.zone_id)
  ElMessage.success('删除成功')
  loadData()
}

const showSegments = async (row) => {
  currentZoneId.value = row.zone_id
  const res = await zoneApi.listSegments(row.zone_id)
  segments.value = res.data || []
  segmentDialogVisible.value = true
}

const showCreateSegment = () => {
  segmentForm.value = { cidr: '', name: '' }
  segmentFormVisible.value = true
}

const handleSubmitSegment = async () => {
  await zoneApi.createSegment(currentZoneId.value, segmentForm.value)
  segmentFormVisible.value = false
  const res = await zoneApi.listSegments(currentZoneId.value)
  segments.value = res.data || []
}

const handleDeleteSegment = async (row) => {
  await zoneApi.deleteSegment(currentZoneId.value, row.segment_id)
  const res = await zoneApi.listSegments(currentZoneId.value)
  segments.value = res.data || []
}

onMounted(loadData)
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
