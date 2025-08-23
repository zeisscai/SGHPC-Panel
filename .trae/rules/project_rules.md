# 项目名称：SGHPC-Panel
项目开发环境：Node.js v22.14.0;Go 1.25.0;
## Node.js依赖包版本信息
运行Vue 3.5.18 + vue-router 4.5.1 + Vuetify 3.9.3
构建：Vite 6 + @vitejs/plugin-vue + vite-plugin-vuetify

## 设计
Material Design
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



## 开发环境
代码开发环境时 macOS
项目运行环境是：Rocky Linux 9.6或 openEuler 24
因此项目的功能不需要适配macOS，只需要适配Rocky Linux 9.6和 openEuler 24即可。