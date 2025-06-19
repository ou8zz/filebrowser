# OnlyOffice Vue 3 组件迁移指南

## 概述

本项目已升级到 Vue 3，并创建了新的 OnlyOffice 编辑器组件 `OnlyOfficeEditorV3.vue`，使用官方的 `@onlyoffice/document-editor-vue` 插件。

## 新组件特性

### OnlyOfficeEditorV3.vue

- 基于官方 Vue 3 插件 `@onlyoffice/document-editor-vue` <mcreference link="https://www.npmjs.com/package/@onlyoffice/document-editor-vue" index="3">3</mcreference>
- 完全兼容 Vue 3 Composition API
- 更好的类型支持（TypeScript）
- 简化的配置和事件处理
- 保持与原有组件相同的功能

## 安装依赖

```bash
# 使用 pnpm（推荐）
pnpm install

# 或使用 npm
npm install

# 或使用 yarn
yarn install
```

## 组件对比

### 原有组件 (OnlyOfficeEditor.vue)
- 手动加载 OnlyOffice API 脚本
- 直接使用 DocsAPI 初始化编辑器
- 需要手动处理脚本加载和错误

### 新组件 (OnlyOfficeEditorV3.vue)
- 使用官方 Vue 3 插件
- 声明式配置
- 内置错误处理和事件管理
- 更好的响应式支持

## 使用方法

### 在 Preview.vue 中的使用

```vue
<template>
  <OnlyOfficeEditorV3
    v-if="isOfficeFile"
    :file="fileStore.req"
    :jwt="jwt"
  />
</template>

<script setup>
import OnlyOfficeEditorV3 from "@/components/files/OnlyOfficeEditorV3.vue";
</script>
```

### 组件属性

- `file`: 文件对象，包含文件名、路径、修改时间等信息
- `jwt`: JWT 认证令牌，用于文件访问权限验证

## 配置说明

### 文档服务器配置

```javascript
const documentServerUrl = 'http://localhost'; // OnlyOffice 文档服务器地址
```

### 支持的文件类型

- **Word 文档**: .docx, .doc, .odt, .rtf, .txt
- **Excel 表格**: .xlsx, .xls, .ods, .csv  
- **PowerPoint 演示**: .pptx, .ppt, .odp

### 编辑器配置

```javascript
const editorConfig = {
  documentType: 'word', // word, cell, slide
  document: {
    key: 'unique-document-key',
    title: 'document.docx',
    url: 'http://example.com/document.docx',
    fileType: 'docx',
    permissions: {
      edit: true,
      download: true,
      print: true,
      review: true,
      comment: true
    }
  },
  editorConfig: {
    mode: 'edit',
    lang: 'zh-CN',
    user: {
      id: 'user-id',
      name: 'username'
    },
    customization: {
      autosave: true,
      forcesave: false
    },
    callbackUrl: 'http://example.com/callback'
  }
}
```

## 事件处理

### 支持的事件

- `onDocumentReady`: 文档加载完成
- `onError`: 编辑器错误
- `onDocumentStateChange`: 文档状态变化
- `onLoadComponentError`: 组件加载错误

### 错误处理

组件内置了完善的错误处理机制：

```javascript
const onLoadComponentError = (errorCode, errorDescription) => {
  switch (errorCode) {
    case -1: // 未知组件加载错误
    case -2: // 无法从服务器加载 DocsAPI
    case -3: // DocsAPI 未定义
  }
};
```

## 迁移步骤

1. ✅ 安装 `@onlyoffice/document-editor-vue` 依赖
2. ✅ 创建新的 `OnlyOfficeEditorV3.vue` 组件
3. ✅ 在 `Preview.vue` 中导入并使用新组件
4. ✅ 修复 JWT token 获取逻辑
5. ✅ 修复类型错误和引用问题
6. 🔄 测试新组件功能
7. 🔄 确认与后端 API 的兼容性

## 注意事项

1. **OnlyOffice 文档服务器**: 确保 OnlyOffice 文档服务器正在运行并可访问
2. **CORS 配置**: 确保文档服务器允许来自前端应用的跨域请求
3. **JWT 认证**: 确保 JWT token 有效且具有文件访问权限
4. **回调 URL**: 确保后端实现了 OnlyOffice 回调接口用于文档保存

## 故障排除

### 常见问题

1. **组件加载失败**: 检查 OnlyOffice 文档服务器是否运行
2. **文档无法打开**: 检查文件 URL 和 JWT token 是否正确
3. **保存失败**: 检查回调 URL 和后端接口实现

### 调试信息

组件会在控制台输出详细的调试信息，包括：
- 文档配置信息
- 文档 URL 和回调 URL
- 错误详情和建议

## 参考资源

- [OnlyOffice Vue 3 官方文档](https://api.onlyoffice.com/docs/docs-api/get-started/frontend-frameworks/vue/) <mcreference link="https://api.onlyoffice.com/docs/docs-api/get-started/frontend-frameworks/vue/" index="2">2</mcreference>
- [OnlyOffice npm 包](https://www.npmjs.com/package/@onlyoffice/document-editor-vue) <mcreference link="https://www.npmjs.com/package/@onlyoffice/document-editor-vue" index="3">3</mcreference>
- [OnlyOffice 博客文章](https://www.onlyoffice.com/blog/2022/10/deploy-onlyoffice-docs-on-react-angular-and-vue) <mcreference link="https://www.onlyoffice.com/blog/2022/10/deploy-onlyoffice-docs-on-react-angular-and-vue" index="1">1</mcreference>