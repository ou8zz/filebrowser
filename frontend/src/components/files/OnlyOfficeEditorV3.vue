<template>
  <div class="onlyoffice-editor-container">
    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>正在加载编辑器配置...</p>
    </div>
    
    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <div class="error-message">
        <h3>加载失败,请检查 OnlyOffice 服务配置</h3>
        <p>{{ error }}</p>
        <button @click="initializeEditor" class="retry-button">重试</button>
      </div>
    </div>
    
    <!-- OnlyOffice编辑器 -->
    <DocumentEditor
      v-else-if="editorConfig"
      id="onlyoffice-editor-v3"
      :documentServerUrl="documentServerUrl"
      :config="editorConfig"
      :events_onDocumentReady="onDocumentReady"
      :events_onError="onError"
      :events_onDocumentStateChange="onDocumentStateChange"
      :onLoadComponentError="onLoadComponentError"
      height="100%"
      width="100%"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { DocumentEditor } from '@onlyoffice/document-editor-vue';
import { useAuthStore } from '@/stores/auth';
import { fetchURL } from "@/api/utils.js";

// Props
interface Props {
  file: {
    name: string;
    path: string;
    modified?: string;
    [key: string]: any;
  };
  jwt: string;
}

const props = defineProps<Props>();

// Store
const authStore = useAuthStore();
const user = authStore.user;

// Reactive data
const documentServerUrl = ref("");
const editorConfig = ref(null);
const isLoading = ref(true);
const error = ref(null);

// Lifecycle
onMounted(() => {
  initializeEditor();
});

// Methods
const initializeEditor = async () => {
  try {
    isLoading.value = true;
    error.value = null;
    
    // 从后端获取完整的编辑器配置
    const config = await getEditorConfigFromBackend();
    editorConfig.value = config;
    
    console.log('OnlyOffice V3 配置信息:', {
      file: props.file,
      config: config,
      jwt: props.jwt ? '已提供' : '未提供'
    });
  } catch (err) {
    console.error('初始化编辑器失败:', err);
    error.value = err.message || '初始化编辑器失败';
  } finally {
    isLoading.value = false;
  }
};

const getEditorConfigFromBackend = async () => {
  // 从后端获取完整的OnlyOffice编辑器配置
  try {
    const response = await fetchURL(`/api/onlyoffice/mapping`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${props.jwt}`
      },
      body: JSON.stringify({
        filePath: props.file.path,
        fileName: props.file.name,
        fileModified: props.file.modified,
        userId: user ? user.id : 'anonymous',
        username: user?.username || '用户'
      })
    });
    
    if (!response.ok) {
      throw new Error(`获取编辑器配置失败: ${response.statusText}`);
    }
    
    const config = await response.json();
    documentServerUrl.value = config.host;
    return config;
  } catch (error) {
    console.error('获取编辑器配置时出错:', error);
    throw error;
  }
};

// 文件类型检查和配置生成逻辑已移到后端处理

// Event handlers
const onDocumentReady = () => {
  console.log('OnlyOffice V3 document is ready');
};

const onError = (event: any) => {
  console.error('OnlyOffice V3 error:', event);
  console.error('错误详情:', {
    errorCode: event.errorCode,
    errorDescription: event.errorDescription,
    documentUrl: documentUrl.value
  });
  // 尝试直接访问文档URL进行调试
  console.log('尝试直接访问文档URL:', documentUrl.value);
};

const onDocumentStateChange = (event: any) => {
  console.log('Document state changed:', event);
};

const onLoadComponentError = (errorCode: number, errorDescription: string) => {
  console.error('OnlyOffice V3 component load error:', { errorCode, errorDescription });
  switch (errorCode) {
    case -1: // Unknown error loading component
      console.log('Unknown error loading component:', errorDescription);
      break;
    case -2: // Error load DocsAPI from documentServerUrl
      console.log('Error load DocsAPI from server:', errorDescription);
      break;
    case -3: // DocsAPI is not defined
      console.log('DocsAPI is not defined:', errorDescription);
      break;
    default:
      console.log('Component error:', errorDescription);
  }
};
</script>

<style scoped>
.onlyoffice-editor-container {
  width: 100%;
  height: 100%;
  min-height: 600px;
}

#onlyoffice-editor-v3 {
  width: 100%;
  height: 100%;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #3498db;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.error-message {
  text-align: center;
  padding: 32px;
  border: 1px solid #e74c3c;
  border-radius: 8px;
  background-color: #fdf2f2;
  max-width: 400px;
}

.error-message h3 {
  color: #e74c3c;
  margin: 0 0 16px 0;
  font-size: 18px;
}

.error-message p {
  color: #666;
  margin: 0 0 20px 0;
  line-height: 1.5;
}

.retry-button {
  background-color: #3498db;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.retry-button:hover {
  background-color: #2980b9;
}

.loading-container p {
  color: #666;
  font-size: 16px;
  margin: 0;
}
</style>