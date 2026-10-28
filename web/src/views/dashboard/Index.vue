<template>
  <div class="dashboard">
    <h2>系统概览</h2>
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ summary.total_zones || 0 }}</div>
            <div class="stat-label">网区数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ summary.total_instances || 0 }}</div>
            <div class="stat-label">实例总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ summary.total_tasks || 0 }}</div>
            <div class="stat-label">采集任务</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ summary.active_alerts || 0 }}</div>
            <div class="stat-label">活跃告警</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>节点状态</span>
          </template>
          <div class="node-stats">
            <div class="node-stat">
              <span class="label">在线节点</span>
              <el-tag type="success">{{ summary.healthy_nodes || 0 }}</el-tag>
            </div>
            <div class="node-stat">
              <span class="label">离线节点</span>
              <el-tag type="danger">{{ summary.offline_nodes || 0 }}</el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="label">平台版本</span>
              <span class="value">Phase 1 v0.1.0</span>
            </div>
            <div class="info-item">
              <span class="label">后端框架</span>
              <span class="value">Go + Chi</span>
            </div>
            <div class="info-item">
              <span class="label">前端框架</span>
              <span class="value">Vue 3 + Element Plus</span>
            </div>
            <div class="info-item">
              <span class="label">数据库</span>
              <span class="value">PostgreSQL 17</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { statusApi } from '@/api/modules'

const summary = ref({})

onMounted(async () => {
  try {
    const res = await statusApi.global()
    summary.value = res.data || {}
  } catch (e) {
    // ignore
  }
})
</script>

<style scoped>
.dashboard h2 {
  margin-bottom: 20px;
  font-size: 20px;
}

.stat-cards .el-card {
  text-align: center;
}

.stat-item {
  padding: 10px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
}

.stat-label {
  margin-top: 8px;
  font-size: 14px;
  color: #999;
}

.node-stats {
  display: flex;
  gap: 30px;
}

.node-stat {
  display: flex;
  align-items: center;
  gap: 10px;
}

.node-stat .label {
  color: #666;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  justify-content: space-between;
}

.info-item .label {
  color: #666;
}

.info-item .value {
  color: #333;
  font-weight: 500;
}
</style>
