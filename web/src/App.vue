<template>
  <el-container class="app-container">
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
        </el-sub-menu>
        <el-sub-menu index="task">
          <template #title>
            <el-icon><List /></el-icon>
            <span>任务管理</span>
          </template>
          <el-menu-item index="/collection-tasks">采集任务</el-menu-item>
          <el-menu-item index="/data-sinks">数据存储</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="infra">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>基础设施</span>
          </template>
          <el-menu-item index="/collectors">采集器管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="app-header">
        <div class="header-right">
          <el-tag type="success">Phase 1</el-tag>
          <el-dropdown>
            <span class="user-info">admin</span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人设置</el-dropdown-item>
                <el-dropdown-item divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { Monitor, Grid, List, Setting } from '@element-plus/icons-vue'
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
