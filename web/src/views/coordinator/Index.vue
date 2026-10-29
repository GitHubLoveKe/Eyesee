<template>
  <div class="page-container">
    <div class="page-header">
      <h2>协调器管理</h2>
      <el-button type="primary" @click="showCreate">注册协调器</el-button>
    </div>

    <el-card>
      <div class="filter-bar" style="margin-bottom: 12px;">
        <el-select v-model="filterZone" placeholder="按网区筛选" clearable style="width: 200px;" @change="loadData">
          <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
        </el-select>
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="coordinator_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="address" label="HTTP地址" width="200" />
        <el-table-column prop="gossip_addr" label="Gossip地址" width="180" />
        <el-table-column prop="role" label="角色" width="80">
          <template #default="{ row }">
            <el-tag :type="row.role === 'leader' ? 'primary' : 'info'" size="small">{{ row.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="managed_nodes" label="管理节点" width="90" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'offline' ? 'danger' : 'warning'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_heartbeat" label="最后心跳" width="170">
          <template #default="{ row }">{{ formatTime(row.last_heartbeat) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button size="small" @click="handleHeartbeat(row)">心跳</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑协调器' : '注册协调器'" width="550px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="HTTP地址" required>
          <el-input v-model="form.address" placeholder="http://ip:port" />
        </el-form-item>
        <el-form-item label="Gossip地址" required>
          <el-input v-model="form.gossip_addr" placeholder="ip:7946" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="leader" value="leader" />
            <el-option label="standby" value="standby" />
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
import { zoneCoordinatorApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filterZone = ref('')
const form = ref({ name: '', zone_id: '', address: '', gossip_addr: '', role: 'leader' })

const formatTime = (t) => t ? new Date(t).toLocaleString('zh-CN') : '-'

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    if (filterZone.value) params.zone_id = filterZone.value
    const res = await zoneCoordinatorApi.list(params)
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
  form.value = { name: '', zone_id: '', address: '', gossip_addr: '', role: 'leader' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.coordinator_id
  form.value = { name: row.name, zone_id: row.zone_id, address: row.address, gossip_addr: row.gossip_addr, role: row.role }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await zoneCoordinatorApi.update(editId.value, form.value)
    } else {
      await zoneCoordinatorApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该协调器？', '提示', { type: 'warning' })
  await zoneCoordinatorApi.delete(row.coordinator_id)
  ElMessage.success('删除成功')
  loadData()
}

const handleHeartbeat = async (row) => {
  await zoneCoordinatorApi.heartbeat(row.coordinator_id)
  ElMessage.success('心跳已更新')
  loadData()
}

onMounted(() => { loadData(); loadZones() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
