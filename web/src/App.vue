<template>
  <el-container v-if="authStore.isLoggedIn" class="app-container">
    <el-aside width="220px" class="app-aside">
      <div class="logo">
        <h2>DCS</h2>
        <span>监控采集管理平台</span>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        class="aside-menu"
        background-color="#1d1e1f"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-sub-menu index="resource">
          <template #title>
            <el-icon><Grid /></el-icon>
            <span>资源管理</span>
          </template>
          <el-menu-item index="/business">业务系统</el-menu-item>
          <el-menu-item index="/zones">网区管理</el-menu-item>
          <el-menu-item index="/resource-types">资源类型</el-menu-item>
          <el-menu-item index="/instances">实例管理</el-menu-item>
          <el-menu-item index="/metric-definitions">指标管理</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="task">
          <template #title>
            <el-icon><List /></el-icon>
            <span>任务管理</span>
          </template>
          <el-menu-item index="/collection-tasks">采集任务</el-menu-item>
          <el-menu-item index="/task-types">任务类型</el-menu-item>
          <el-menu-item index="/rc-tasks">RC任务</el-menu-item>
          <el-menu-item index="/scheduled-tasks">定时任务</el-menu-item>
          <el-menu-item index="/data-sinks">数据存储</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="alert">
          <template #title>
            <el-icon><Bell /></el-icon>
            <span>告警管理</span>
          </template>
          <el-menu-item index="/alert-strategies">告警策略</el-menu-item>
          <el-menu-item index="/alert-events">告警事件</el-menu-item>
          <el-menu-item index="/alert-groups">告警组</el-menu-item>
          <el-menu-item index="/notification-channels">通知渠道</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="infra">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>基础设施</span>
          </template>
          <el-menu-item index="/collectors">采集器管理</el-menu-item>
          <el-menu-item index="/zone-coordinators">协调器</el-menu-item>
          <el-menu-item index="/credentials">凭据管理</el-menu-item>
          <el-menu-item index="/users">用户管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="app-header">
        <div class="header-right">
          <el-tag type="success">Phase 5</el-tag>
          <span class="user-info">{{ authStore.username || 'admin' }}</span>
          <el-button size="small" @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
  <router-view v-else />
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Monitor, Grid, List, Setting, Bell } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

onMounted(() => {
  authStore.loadFromStorage()
})

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.app-container {
  height: 100vh;
}

.app-aside {
  background-color: #1d1e1f;
  overflow-y: auto;
}

.logo {
  padding: 20px;
  text-align: center;
  color: #fff;
  border-bottom: 1px solid #333;
}

.logo h2 {
  font-size: 24px;
  margin-bottom: 4px;
}

.logo span {
  font-size: 12px;
  color: #999;
}

.aside-menu {
  border-right: none;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  border-bottom: 1px solid #e6e6e6;
  background: #fff;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  cursor: pointer;
  font-size: 14px;
  color: #333;
}

.app-main {
  background-color: #f5f7fa;
  overflow-y: auto;
}
</style>
