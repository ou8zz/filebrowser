<template>
  <div id="onlyoffice-editor"></div>
</template>

<script setup>
import { onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { fetchURL } from "@/api/utils.js";

// Props
const props = defineProps({
  file: { // File object containing details like name, url, etc.
    type: Object,
    required: true,
  },
  jwt: { // JWT token for authentication
    type: String,
    required: true,
  },
});

// Store
const authStore = useAuthStore();
const user = authStore.user;

// Lifecycle
onMounted(() => {
  initializeEditor();
});

const initializeEditor = async () => {
  try {
    const onlyOfficeConfig = await getEditorConfigFromBackend();
    await loadOnlyOfficeApi(onlyOfficeConfig.host);
    initEditor(onlyOfficeConfig);
  } catch (error) {
    console.error("初始化OnlyOffice编辑器失败:", error);
  }
};

const getEditorConfigFromBackend = async () => {
  const response = await fetchURL(`/api/onlyoffice/mapping`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${props.jwt}`,
    },
    body: JSON.stringify({
      filePath: props.file.path,
      fileName: props.file.name,
      fileModified: props.file.modified,
      userId: user ? user.id : 0,
      username: user ? user.username : "anonymous",
    }),
  });

  if (!response.ok) {
    throw new Error(`获取编辑器配置失败: ${response.statusText}`);
  }

  return await response.json();
};

const loadOnlyOfficeApi = (documentServerUrl) => {
  if (typeof DocsAPI !== "undefined") {
    return Promise.resolve();
  }

  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    const base = String(documentServerUrl || "").replace(/\/+$/, "");
    script.src = `${base}/web-apps/apps/api/documents/api.js`;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Failed to load OnlyOffice API script"));
    document.head.appendChild(script);
  });
};

const initEditor = (onlyOfficeConfig) => {
  if (typeof DocsAPI === "undefined") {
    console.error("OnlyOffice DocsAPI is not loaded.");
    return;
  }

  const documentUrl = onlyOfficeConfig?.document?.url;

  const config = {
    token: onlyOfficeConfig.token,
    height: "100%",
    width: "100%",
    documentType: onlyOfficeConfig.documentType,
    document: onlyOfficeConfig.document,
    editorConfig: onlyOfficeConfig.editorConfig,
  };

  // 添加事件处理
  config.events = {
    onDocumentReady: () => {
      console.log("OnlyOffice document is ready");
    },
    onError: (event) => {
      console.error("OnlyOffice error:", event.data);
      console.error("错误详情:", {
        errorCode: event.data.errorCode,
        errorDescription: event.data.errorDescription,
        documentUrl: documentUrl,
      });
      // 尝试直接访问文档URL进行调试
      console.log("尝试直接访问文档URL:", documentUrl);
    },
    onDocumentStateChange: (event) => {
      console.log("Document state changed:", event.data);
    },
  };

  // Initialize the editor
  new DocsAPI.DocEditor(
    "onlyoffice-editor", // The ID of the div element where the editor will be initialized
    config
  );
};
</script>

<style scoped>
#onlyoffice-editor {
  width: 100%;
  height: 100%;
}
</style>
