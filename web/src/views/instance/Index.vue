<template>
  <div class="page-container">
    <div class="page-header">
      <h2>实例管理</h2>
      <el-button type="primary" @click="showCreate">新建实例</el-button>
    </div>

    <el-card>
      <div style="margin-bottom: 12px; display: flex; gap: 12px">
        <el-input v-model="keyword" placeholder="搜索名称或主机" style="width: 250px" clearable @clear="loadData" @keyup.enter="loadData" />
        <el-select v-model="filterZone" placeholder="网区" clearable @change="loadData" style="width: 150px">
          <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
        </el-select>
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="instance_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="host" label="主机" width="140" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑实例' : '新建实例'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="主机" required>
          <el-input v-model="form.host" placeholder="IP或域名" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型" required>
          <el-select v-model="form.resource_type_id">
            <el-option v-for="t in types" :key="t.type_id" :label="t.name" :value="t.type_id" />
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
import { instanceApi, zoneApi, resourceTypeApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const types = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const keyword = ref('')
const filterZone = ref('')
const form = ref({ name: '', host: '', port: null, zone_id: '', resource_type_id: '' })
const editId = ref('')

const statusType = (s) => ({ active: 'success', pending: 'warning', suspended: 'info', decommissioned: 'danger' }[s] || 'info')

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    if (keyword.value) params.keyword = keyword.value
    if (filterZone.value) params.zone_id = filterZone.value
    const res = await instanceApi.list(params)
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const loadZones = async () => {
  const res = await zoneApi.list()
  zones.value = res.data?.items || []
}

const loadTypes = async () => {
  const res = await resourceTypeApi.list()
  types.value = res.data || []
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', host: '', port: null, zone_id: '', resource_type_id: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.instance_id
  form.value = { name: row.name, host: row.host, port: row.port, zone_id: row.zone_id, resource_type_id: row.resource_type_id }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await instanceApi.update(editId.value, form.value)
    } else {
      await instanceApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该实例？', '提示', { type: 'warning' })
  await instanceApi.delete(row.instance_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadZones(); loadTypes() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
