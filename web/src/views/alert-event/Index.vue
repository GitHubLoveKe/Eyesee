<template>
  <div class="page-container">
    <div class="page-header">
      <h2>告警事件</h2>
      <div style="display: flex; gap: 8px;">
        <el-button @click="showSilences">静默管理</el-button>
        <el-button type="primary" @click="showCreate">上报告警</el-button>
      </div>
    </div>

    <el-card style="margin-bottom: 16px;">
      <div style="display: flex; gap: 24px; text-align: center;">
        <div>
          <div style="font-size: 28px; font-weight: bold; color: #f56c6c;">{{ stats.total_firing || 0 }}</div>
          <div style="color: #999; font-size: 13px;">触发中</div>
        </div>
        <div>
          <div style="font-size: 28px; font-weight: bold; color: #67c23a;">{{ stats.total_resolved || 0 }}</div>
          <div style="color: #999; font-size: 13px;">已恢复</div>
        </div>
        <div v-for="(count, sev) in stats.by_severity" :key="sev">
          <div style="font-size: 28px; font-weight: bold;" :style="{ color: sevColor(sev) }">{{ count }}</div>
          <div style="color: #999; font-size: 13px;">{{ sev }}</div>
        </div>
      </div>
    </el-card>

    <el-card>
      <div class="filter-bar" style="margin-bottom: 12px; display: flex; gap: 12px; flex-wrap: wrap;">
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 120px;" @change="loadData">
          <el-option label="firing" value="firing" />
          <el-option label="resolved" value="resolved" />
        </el-select>
        <el-select v-model="filterSeverity" placeholder="级别" clearable style="width: 120px;" @change="loadData">
          <el-option label="critical" value="critical" />
          <el-option label="warning" value="warning" />
          <el-option label="info" value="info" />
        </el-select>
        <el-select v-model="filterZone" placeholder="网区" clearable style="width: 150px;" @change="loadData">
          <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
        </el-select>
        <el-input v-model="filterAlertName" placeholder="告警名称" clearable style="width: 200px;" @clear="loadData" @keyup.enter="loadData" />
        <el-select v-model="filterPeriod" style="width: 100px;" @change="loadStats">
          <el-option label="1小时" value="1h" />
          <el-option label="6小时" value="6h" />
          <el-option label="24小时" value="24h" />
          <el-option label="7天" value="7d" />
        </el-select>
      </div>
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="event_id" label="ID" width="70" />
        <el-table-column prop="alert_name" label="告警名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="severity" label="级别" width="90">
          <template #default="{ row }">
            <el-tag :type="sevType(row.severity)" size="small">{{ row.severity || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="zone_id" label="网区" width="120" />
        <el-table-column prop="instance_id" label="实例" width="130" show-overflow-tooltip />
        <el-table-column prop="starts_at" label="触发时间" width="170">
          <template #default="{ row }">{{ formatTime(row.starts_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button v-if="row.status === 'firing'" size="small" type="warning" @click="handleAck(row)">认领</el-button>
            <span v-else style="color: #999; font-size: 12px;">已处理</span>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-if="total > 0"
        style="margin-top: 12px; justify-content: flex-end;"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </el-card>

    <el-dialog v-model="createDialogVisible" title="上报告警事件" width="550px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="告警名称" required>
          <el-input v-model="createForm.alert_name" />
        </el-form-item>
        <el-form-item label="网区">
          <el-select v-model="createForm.zone_id" clearable>
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="实例ID">
          <el-input v-model="createForm.instance_id" />
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="createForm.severity">
            <el-option label="info" value="info" />
            <el-option label="warning" value="warning" />
            <el-option label="critical" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="createForm.labels" placeholder='{"key":"value"}' />
        </el-form-item>
        <el-form-item label="注解">
          <el-input v-model="createForm.annotations" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="submitting">上报</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="ackDialogVisible" title="认领告警" width="400px">
      <el-form :model="ackForm" label-width="80px">
        <el-form-item label="用户ID" required>
          <el-input v-model="ackForm.user_id" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="ackForm.note" type="textarea" :rows="2" placeholder="已收到，正在排查..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ackDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAckSubmit" :loading="submitting">确认认领</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="silenceDialogVisible" title="静默管理" width="800px">
      <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <span>告警静默列表</span>
        <el-button type="primary" size="small" @click="showCreateSilence">新建静默</el-button>
      </div>
      <el-table :data="silences" stripe>
        <el-table-column prop="silence_id" label="ID" width="120" />
        <el-table-column prop="zone_id" label="网区" width="120">
          <template #default="{ row }">{{ row.zone_id || '全局' }}</template>
        </el-table-column>
        <el-table-column prop="matchers" label="匹配条件" show-overflow-tooltip />
        <el-table-column prop="duration" label="时长" width="80" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'warning' : row.status === 'revoked' ? 'info' : 'success'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" show-overflow-tooltip />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button v-if="row.status === 'active'" size="small" type="danger" @click="handleRevokeSilence(row)">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="silenceFormVisible" title="新建静默" width="500px">
      <el-form :model="silenceForm" label-width="100px">
        <el-form-item label="匹配条件" required>
          <el-input v-model="silenceForm.matchers" type="textarea" :rows="2" placeholder='[{"name":"alertname","value":"MySQL.*","is_regex":true}]' />
        </el-form-item>
        <el-form-item label="网区">
          <el-select v-model="silenceForm.zone_id" clearable>
            <el-option v-for="z in zones" :key="z.zone_id" :label="z.name" :value="z.zone_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长">
          <el-select v-model="silenceForm.duration">
            <el-option label="30分钟" value="30m" />
            <el-option label="1小时" value="1h" />
            <el-option label="2小时" value="2h" />
            <el-option label="6小时" value="6h" />
            <el-option label="1天" value="1d" />
          </el-select>
        </el-form-item>
        <el-form-item label="原因">
          <el-input v-model="silenceForm.reason" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="silenceFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateSilence" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertEventApi, alertSilenceApi, zoneApi } from '@/api/modules'

const items = ref([])
const zones = ref([])
const stats = ref({})
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filterStatus = ref('')
const filterSeverity = ref('')
const filterZone = ref('')
const filterAlertName = ref('')
const filterPeriod = ref('24h')

const createDialogVisible = ref(false)
const submitting = ref(false)
const createForm = ref({ alert_name: '', zone_id: '', instance_id: '', severity: 'warning', labels: '', annotations: '' })

const ackDialogVisible = ref(false)
const ackEventId = ref(null)
const ackForm = ref({ user_id: '', note: '' })

const silenceDialogVisible = ref(false)
const silences = ref([])
const silenceFormVisible = ref(false)
const silenceForm = ref({ matchers: '', zone_id: '', duration: '1h', reason: '' })

const sevType = (s) => s === 'critical' ? 'danger' : s === 'warning' ? 'warning' : 'info'
const sevColor = (s) => s === 'critical' ? '#f56c6c' : s === 'warning' ? '#e6a23c' : '#909399'
const formatTime = (t) => t ? new Date(t).toLocaleString('zh-CN') : '-'

const loadData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterSeverity.value) params.severity = filterSeverity.value
    if (filterZone.value) params.zone_id = filterZone.value
    if (filterAlertName.value) params.alert_name = filterAlertName.value
    const res = await alertEventApi.list(params)
    items.value = res.data?.items || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  const params = { period: filterPeriod.value }
  if (filterZone.value) params.zone_id = filterZone.value
  const res = await alertEventApi.statistics(params)
  stats.value = res.data || {}
}

const loadZones = async () => {
  const res = await zoneApi.list()
  zones.value = res.data?.items || []
}

const handlePageChange = (p) => { page.value = p; loadData() }

const showCreate = () => {
  createForm.value = { alert_name: '', zone_id: '', instance_id: '', severity: 'warning', labels: '', annotations: '' }
  createDialogVisible.value = true
}

const handleCreate = async () => {
  submitting.value = true
  try {
    await alertEventApi.create(createForm.value)
    createDialogVisible.value = false
    loadData()
    loadStats()
  } finally {
    submitting.value = false
  }
}

const handleAck = (row) => {
  ackEventId.value = row.event_id
  ackForm.value = { user_id: '', note: '' }
  ackDialogVisible.value = true
}

const handleAckSubmit = async () => {
  submitting.value = true
  try {
    await alertEventApi.acknowledge(ackEventId.value, ackForm.value)
    ackDialogVisible.value = false
    ElMessage.success('认领成功')
    loadData()
  } finally {
    submitting.value = false
  }
}

const showSilences = async () => {
  const res = await alertSilenceApi.list()
  silences.value = res.data || []
  silenceDialogVisible.value = true
}

const showCreateSilence = () => {
  silenceForm.value = { matchers: '', zone_id: '', duration: '1h', reason: '' }
  silenceFormVisible.value = true
}

const handleCreateSilence = async () => {
  submitting.value = true
  try {
    await alertSilenceApi.create(silenceForm.value)
    silenceFormVisible.value = false
    const res = await alertSilenceApi.list()
    silences.value = res.data || []
  } finally {
    submitting.value = false
  }
}

const handleRevokeSilence = async (row) => {
  await ElMessageBox.confirm('确认撤销该静默规则？', '提示', { type: 'warning' })
  await alertSilenceApi.revoke(row.silence_id)
  ElMessage.success('撤销成功')
  const res = await alertSilenceApi.list()
  silences.value = res.data || []
}

onMounted(() => { loadZones(); loadData(); loadStats() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 18px; }
</style>
