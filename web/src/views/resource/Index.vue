<template>
  <div class="page-container">
    <div class="page-header">
      <h2>资源类型树</h2>
      <el-button type="primary" @click="showCreate">新建类型</el-button>
    </div>

    <el-card>
      <el-table :data="tree" v-loading="loading" row-key="type_id" :tree-props="{ children: 'children' }" stripe>
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type_id" label="ID" width="140" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="icon" label="图标" width="80" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑类型' : '新建类型'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="父类型">
          <el-select v-model="form.parent_id" clearable placeholder="无(根节点)">
            <el-option v-for="t in flatList" :key="t.type_id" :label="t.name" :value="t.type_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" />
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
import { resourceTypeApi } from '@/api/modules'

const tree = ref([])
const flatList = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const form = ref({ name: '', parent_id: null, icon: '', description: '' })
const editId = ref('')

const buildTree = (items) => {
  const map = {}
  const roots = []
  items.forEach(i => { map[i.type_id] = { ...i, children: [] } })
  items.forEach(i => {
    if (i.parent_id && map[i.parent_id]) {
      map[i.parent_id].children.push(map[i.type_id])
    } else {
      roots.push(map[i.type_id])
    }
  })
  return roots
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await resourceTypeApi.tree()
    const items = res.data?.tree || []
    flatList.value = items
    tree.value = buildTree(items)
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', parent_id: null, icon: '', description: '' }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.type_id
  form.value = { name: row.name, parent_id: row.parent_id, icon: row.icon, description: row.description }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await resourceTypeApi.update(editId.value, form.value)
    } else {
      await resourceTypeApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该资源类型？', '提示', { type: 'warning' })
  try {
    await resourceTypeApi.delete(row.type_id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    // error handled by interceptor
  }
}

onMounted(loadData)
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
