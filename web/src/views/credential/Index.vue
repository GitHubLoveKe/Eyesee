<template>
  <div class="page-container">
    <div class="page-header">
      <h2>凭据管理</h2>
      <el-button type="primary" @click="showCreate">新建凭据</el-button>
    </div>

    <el-card>
      <div class="filter-bar" style="margin-bottom: 12px; display: flex; gap: 12px;">
        <el-select v-model="filterZone" placeholder="按网区筛选" clearable style="width: 180px;" @change="loadData">
          <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
        </el-select>
        <el-select v-model="filterType" placeholder="按类型筛选" clearable style="width: 160px;" @change="loadData">
          <el-option label="basic_auth" value="basic_auth" />
          <el-option label="token" value="token" />
          <el-option label="tls_cert" value="tls_cert" />
          <el-option label="ssh_key" value="ssh_key" />
          <el-option label="snmp_community" value="snmp_community" />
        </el-select>
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="credential_id" label="ID" width="130" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="zone_id" label="网区" width="130" />
        <el-table-column prop="cred_type" label="类型" width="140">
          <template #default="{ row }">
            <el-tag size="small">{{ row.cred_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="expires_at" label="过期时间" width="170">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isExpired(row.expires_at) }">
              {{ row.expires_at ? formatTime(row.expires_at) : '永不过期' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑凭据' : '新建凭据'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="网区" required>
          <el-select v-model="form.zone_id" :disabled="isEdit">
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.cred_type">
            <el-option label="basic_auth" value="basic_auth" />
            <el-option label="token" value="token" />
            <el-option label="tls_cert" value="tls_cert" />
            <el-option label="ssh_key" value="ssh_key" />
            <el-option label="snmp_community" value="snmp_community" />
          </el-select>
        </el-form-item>
        <el-form-item label="凭据数据" required>
          <el-input v-model="form.data" type="textarea" :rows="4" :placeholder="dataPlaceholder" />
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker v-model="form.expires_at" type="datetime" placeholder="留空表示永不过期" style="width: 100%;" />
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
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { credentialApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filterZone = ref('')
const filterType = ref('')
const form = ref({ name: '', zone_id: '', cred_type: 'basic_auth', data: '', expires_at: null })

const dataPlaceholder = computed(() => {
  const placeholders = {
    basic_auth: '{"username":"admin","password":"secret"}',
    token: '{"token":"eyJhbG..."}',
    tls_cert: '{"ca_cert":"...","client_cert":"...","client_key":"..."}',
    ssh_key: '{"private_key":"...","passphrase":"..."}',
    snmp_community: '{"community":"public","version":"v2c"}'
  }
  return placeholders[form.value.cred_type] || ''
})

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const isExpired = (t) => {
  if (!t) return false
  return new Date(t) < new Date()
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    if (filterZone.value) params.zone_id = filterZone.value
    if (filterType.value) params.cred_type = filterType.value
    const res = await credentialApi.list(params)
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
  form.value = { name: '', zone_id: '', cred_type: 'basic_auth', data: '', expires_at: null }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.credential_id
  form.value = {
    name: row.name, zone_id: row.zone_id, cred_type: row.cred_type,
    data: '', expires_at: row.expires_at ? new Date(row.expires_at) : null
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    const payload = { ...form.value }
    if (payload.expires_at) {
      payload.expires_at = payload.expires_at.toISOString()
    } else {
      payload.expires_at = null
    }
    if (isEdit.value) {
      await credentialApi.update(editId.value, payload)
    } else {
      await credentialApi.create(payload)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该凭据？删除前请确认没有实例引用。', '提示', { type: 'warning' })
  await credentialApi.delete(row.credential_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadZones() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
.text-danger { color: #f56c6c; }
</style>
