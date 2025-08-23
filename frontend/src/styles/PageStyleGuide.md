# 页面样式统一指南

本文档定义了SGHPC Panel前端页面的统一样式规范，确保所有页面保持一致的视觉风格。

## 页面结构模板

### 基本页面结构

```vue
<template>
  <v-container fluid class="pa-6">
    <v-row>
      <v-col cols="12">
        <v-card elevation="2">
          <!-- 主标题 -->
          <v-card-title class="text-h4 font-weight-bold d-flex align-center">
            <v-icon left class="mr-3" size="32" style="background: transparent;">[图标名称]</v-icon>
            [页面标题]
          </v-card-title>
          
          <v-card-text>
            <v-row>
              <v-col cols="12">
                <!-- 内容卡片 -->
                <v-card elevation="1">
                  <v-card-subtitle class="text-h6 font-weight-medium d-flex align-center">
                    <v-icon left class="mr-2" style="background: transparent;">[子图标名称]</v-icon>
                    [子标题]
                  </v-card-subtitle>
                  <v-card-text>
                    <!-- 页面具体内容 -->
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>
```

## 样式规范

### 1. 容器样式
- **外层容器**: `<v-container fluid class="pa-6">`
- **主卡片**: `<v-card elevation="2">`
- **内容卡片**: `<v-card elevation="1">`

### 2. 标题样式
- **主标题**: `class="text-h4 font-weight-bold d-flex align-center"`
- **子标题**: `class="text-h6 font-weight-medium d-flex align-center"`

### 3. 图标样式
- **主标题图标**: `size="32" style="background: transparent;"`
- **子标题图标**: `style="background: transparent;"`
- **所有图标**: 必须设置 `style="background: transparent;"` 确保背景透明

### 4. 间距规范
- **主标题图标间距**: `class="mr-3"`
- **子标题图标间距**: `class="mr-2"`
- **容器内边距**: `class="pa-6"`

## 页面示例

### 已实现的页面

1. **状态概览** (`Overview.vue`)
   - 主标题: "状态概览" + `mdi-view-dashboard`
   - 子模块: 管理节点信息、计算节点、SLURM作业

2. **终端管理** (`Terminal.vue`)
   - 主标题: "终端管理" + `mdi-console`
   - 子模块: Web终端

3. **文件管理** (`FileManagement.vue`)
   - 主标题: "文件管理" + `mdi-file-document-multiple`
   - 子模块: 文件浏览器

4. **Spack管理** (`Spack.vue`)
   - 主标题: "Spack管理" + `mdi-package-variant`
   - 子模块: 软件包管理器

5. **用户管理** (`User.vue`)
   - 主标题: "用户管理" + `mdi-account-multiple`
   - 子模块: 用户列表

6. **软件源管理** (`RepositoryManagement.vue`)
   - 主标题: "软件源管理" + `mdi-package-variant`
   - 子模块: 软件源配置

## 新页面开发指南

### 步骤1: 复制基本结构
从上述模板复制基本的页面结构。

### 步骤2: 设置标题和图标
- 选择合适的Material Design图标
- 设置中文标题
- 确保图标背景透明

### 步骤3: 添加内容卡片
根据页面功能需求，在内容卡片中添加具体的组件和功能。

### 步骤4: 保持一致性
- 使用统一的颜色方案
- 遵循间距规范
- 保持图标风格一致

## 注意事项

1. **图标背景**: 所有图标必须设置 `style="background: transparent;"`
2. **响应式设计**: 使用Vuetify的栅格系统确保移动端适配
3. **无障碍访问**: 为图标添加适当的aria-label属性
4. **性能优化**: 避免过深的嵌套结构

## 常用图标参考

- 仪表板: `mdi-view-dashboard`
- 终端: `mdi-console`
- 文件: `mdi-file-document-multiple`
- 软件包: `mdi-package-variant`
- 用户: `mdi-account-multiple`
- 设置: `mdi-cog`
- 网络: `mdi-network`
- 数据库: `mdi-database`
- 监控: `mdi-monitor`
- 日志: `mdi-text-box`

---

遵循此指南可确保所有页面保持一致的视觉风格和用户体验。