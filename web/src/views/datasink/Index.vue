<template>
  <div class="page-container">
    <div class="page-header">
      <h2>数据存储管理</h2>
      <el-button type="primary" @click="showCreate">新建存储</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="sink_id" label="ID" width="130" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="sink_type" label="类型" width="140" />
        <el-table-column prop="protocol" label="协议" width="120" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">{{ row.status }}</el-tag>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑存储' : '新建存储'" width="550px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.sink_type">
            <el-option label="VictoriaMetrics" value="victoriametrics" />
            <el-option label="Elasticsearch" value="elasticsearch" />
            <el-option label="Loki" value="loki" />
            <el-option label="OpenObserve" value="openobserve" />
            <el-option label="S3" value="s3" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议" required>
          <el-select v-model="form.protocol">
            <el-option label="Prometheus" value="prometheus" />
            <el-option label="Elasticsearch" value="elasticsearch" />
            <el-option label="Loki" value="loki" />
            <el-option label="S3" value="s3" />
            <el-option label="HTTP" value="http" />
          </el-select>
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="4" placeholder='{"url":"http://vm:8428"}' />
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
import { dataSinkApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = ref({ name: '', sink_type: '', protocol: '', zone_id: '', config: '' })
const editId = ref('')

const loadData = async () => {
  loading.value = true
  try {
    const res = await dataSinkApi.list()
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
  form.value = { name: '', sink_type: '', protocol: '', zone_id: '', config: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.sink_id
  form.value = { name: row.name, sink_type: row.sink_type, protocol: row.protocol, zone_id: row.zone_id, config: row.config }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await dataSinkApi.update(editId.value, form.value)
    } else {
      await dataSinkApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该数据存储？', '提示', { type: 'warning' })
  await dataSinkApi.delete(row.sink_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadZones() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
