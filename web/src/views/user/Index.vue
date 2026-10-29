<template>
  <div class="page-container">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="showCreate">新建用户</el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-select v-model="filter.role" placeholder="角色" clearable style="width:140px" @change="loadData">
          <el-option label="admin" value="admin" />
          <el-option label="operator" value="operator" />
          <el-option label="viewer" value="viewer" />
        </el-select>
        <el-input v-model="filter.keyword" placeholder="搜索用户名/姓名" clearable style="width:220px" @clear="loadData" @keyup.enter="loadData" />
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="user_id" label="ID" width="120" />
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="display_name" label="姓名" width="140" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="phone" label="手机" width="140" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'operator' ? 'warning' : 'info'">{{ row.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新建用户'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.display_name" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="手机">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="admin" value="admin" />
            <el-option label="operator" value="operator" />
            <el-option label="viewer" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态">
          <el-select v-model="form.status">
            <el-option label="active" value="active" />
            <el-option label="disabled" value="disabled" />
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
import { userApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const filter = ref({ role: '', keyword: '' })
const form = ref({ username: '', display_name: '', email: '', phone: '', role: 'viewer', status: 'active' })

const loadData = async () => {
  loading.value = true
  try {
    const res = await userApi.list({ role: filter.value.role, keyword: filter.value.keyword })
    items.value = res.data?.items || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { username: '', display_name: '', email: '', phone: '', role: 'viewer', status: 'active' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.user_id
  form.value = { username: row.username, display_name: row.display_name, email: row.email, phone: row.phone, role: row.role, status: row.status }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await userApi.update(editId.value, form.value)
    } else {
      await userApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该用户？', '提示', { type: 'warning' })
  await userApi.delete(row.user_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => loadData())
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
