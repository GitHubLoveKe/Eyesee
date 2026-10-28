<template>
  <div class="page-container">
    <div class="page-header">
      <h2>告警组</h2>
      <el-button type="primary" @click="showCreate">新建告警组</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="group_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="parent_id" label="父组" width="120">
          <template #default="{ row }">
            {{ row.parent_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="escalation_minutes" label="升级超时(分)" width="120" />
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <el-button size="small" @click="showMembers(row)">成员</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑告警组' : '新建告警组'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="父组">
          <el-select v-model="form.parent_id" clearable placeholder="无">
            <el-option v-for="g in items" :key="g.group_id" :label="g.name" :value="g.group_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="升级超时(分)">
          <el-input-number v-model="form.escalation_minutes" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="memberDialogVisible" title="组成员管理" width="550px">
      <div style="margin-bottom:12px;display:flex;gap:8px">
        <el-select v-model="memberForm.user_id" placeholder="选择用户" style="flex:1">
          <el-option v-for="u in users" :key="u.user_id" :label="u.display_name || u.username" :value="u.user_id" />
        </el-select>
        <el-select v-model="memberForm.role" style="width:120px">
          <el-option label="member" value="member" />
          <el-option label="manager" value="manager" />
        </el-select>
        <el-button type="primary" @click="handleAddMember">添加</el-button>
      </div>
      <el-table :data="members" stripe>
        <el-table-column prop="user_id" label="用户ID" />
        <el-table-column prop="role" label="角色" width="100" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="handleRemoveMember(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertGroupApi, userApi } from '@/api/modules'

const items = ref([])
const users = ref([])
const members = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const memberDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const currentGroupId = ref('')
const form = ref({ name: '', description: '', parent_id: null, escalation_minutes: null })
const memberForm = ref({ user_id: '', role: 'member' })

const loadData = async () => {
  loading.value = true
  try {
    const res = await alertGroupApi.list()
    items.value = res.data || []
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  const res = await userApi.list()
  users.value = res.data?.items || []
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', description: '', parent_id: null, escalation_minutes: null }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.group_id
  form.value = { name: row.name, description: row.description, parent_id: row.parent_id, escalation_minutes: row.escalation_minutes }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await alertGroupApi.update(editId.value, form.value)
    } else {
      await alertGroupApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该告警组？', '提示', { type: 'warning' })
  await alertGroupApi.delete(row.group_id)
  ElMessage.success('删除成功')
  loadData()
}

const showMembers = async (row) => {
  currentGroupId.value = row.group_id
  const res = await alertGroupApi.listMembers(row.group_id)
  members.value = res.data || []
  memberDialogVisible.value = true
}

const handleAddMember = async () => {
  await alertGroupApi.addMember(currentGroupId.value, memberForm.value)
  memberForm.value = { user_id: '', role: 'member' }
  const res = await alertGroupApi.listMembers(currentGroupId.value)
  members.value = res.data || []
}

const handleRemoveMember = async (row) => {
  await alertGroupApi.removeMember(currentGroupId.value, row.user_id)
  const res = await alertGroupApi.listMembers(currentGroupId.value)
  members.value = res.data || []
}

onMounted(() => { loadData(); loadUsers() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
