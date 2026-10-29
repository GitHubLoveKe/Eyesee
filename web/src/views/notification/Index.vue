<template>
  <div class="page-container">
    <div class="page-header">
      <h2>通知渠道</h2>
      <el-button type="primary" @click="showCreate">新建渠道</el-button>
    </div>

    <el-card>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="channel_id" label="ID" width="120" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="channel_type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.channel_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_test_result" label="测试结果" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.last_test_result" :type="row.last_test_result === 'success' ? 'success' : 'danger'">{{ row.last_test_result }}</el-tag>
            <span v-else style="color:#999">未测试</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240">
          <template #default="{ row }">
            <el-button size="small" @click="handleTest(row)">测试</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑渠道' : '新建渠道'" width="550px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.channel_type">
            <el-option label="email" value="email" />
            <el-option label="dingtalk" value="dingtalk" />
            <el-option label="wechat" value="wechat" />
            <el-option label="feishu" value="feishu" />
            <el-option label="slack" value="slack" />
            <el-option label="webhook" value="webhook" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="4" placeholder='{"webhook_url":"..."}' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
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
import { notificationChannelApi } from '@/api/modules'

const items = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const form = ref({ name: '', channel_type: 'dingtalk', config: '', enabled: true })

const loadData = async () => {
  loading.value = true
  try {
    const res = await notificationChannelApi.list()
    items.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  isEdit.value = false
  form.value = { name: '', channel_type: 'dingtalk', config: '', enabled: true }
  dialogVisible.value = true
}

const showEdit = (row) => {
  isEdit.value = true
  editId.value = row.channel_id
  form.value = { name: row.name, channel_type: row.channel_type, config: row.config, enabled: row.enabled }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await notificationChannelApi.update(editId.value, form.value)
    } else {
      await notificationChannelApi.create(form.value)
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

const handleTest = async (row) => {
  await notificationChannelApi.test(row.channel_id)
  ElMessage.success('测试完成')
  loadData()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除该通知渠道？', '提示', { type: 'warning' })
  await notificationChannelApi.delete(row.channel_id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => loadData())
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
