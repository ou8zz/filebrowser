<template>
  <div id="onlyoffice-editor"></div>
</template>

<script>
import sign from 'jwt-encode';
import { mapState } from 'vuex';
const secret = '123456B';
export default {
  name: 'OnlyOfficeEditor',
  props: {
    file: { // File object containing details like name, url, etc.
      type: Object,
      required: true
    },
    jwt: { // JWT token for authentication
      type: String,
      required: true
    }
  },
  computed: {
    ...mapState(['user'])
  },
  mounted() {
    this.loadOnlyOfficeApi();
  },
  methods: {
    loadOnlyOfficeApi() {
      // Check if the API script is already loaded
      if (typeof DocsAPI === 'undefined') {
        const script = document.createElement('script');
        script.src = 'http://localhost/web-apps/apps/api/documents/api.js';
        script.onload = this.initEditor;
        script.onerror = () => {
          console.error('Failed to load OnlyOffice API script');
        };
        document.head.appendChild(script);
      } else {
        this.initEditor();
      }
    },
    initEditor() {
      if (typeof DocsAPI === 'undefined') {
        console.error('OnlyOffice DocsAPI is not loaded.');
        return;
      }

      // 生成文档唯一标识
      const documentKey = this.generateDocumentKey(this.file.name, this.file.modified);
      const documentUrl = this.getDocumentUrl();
      
      // 存储文档密钥和文件路径的映射关系
      this.storeDocumentKeyMapping(documentKey, this.file.path);
      const userId = this.user ? this.user.id : 'anonymous';
      
      // 调试信息
      console.log('OnlyOffice配置信息:', {
        file: this.file,
        documentKey,
        documentUrl,
        jwt: this.jwt ? '已提供' : '未提供'
      });
      
      const data = {
        documentType: this.getDocumentType(this.file.name),
        document: {
          key: documentKey,
          title: this.file.name,
          url: documentUrl,
          fileType: this.getFileExtension(this.file.name),
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
            id: 'user-' + userId,
            name: '用户'
          },
          customization: {
            autosave: true,
            forcesave: false
          },
          callbackUrl: this.getCallbackUrl()
        }
      };

      const jwtToken = sign(data, secret);
      const config = {
        token: jwtToken,
        height: '100%',
        width: '100%',
        documentType: this.getDocumentType(this.file.name),
        document: {
          key: documentKey,
          title: this.file.name,
          url: this.getDocumentUrl(),
          fileType: this.getFileExtension(this.file.name),
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
            id: 'user-' + userId,
            name: '用户'
          },
          customization: {
            autosave: true,
            forcesave: false
          }
        }
      };

      // 添加事件处理
      config.events = {
        'onDocumentReady': () => {
          console.log('OnlyOffice document is ready');
        },
        'onError': (event) => {
          console.error('OnlyOffice error:', event.data);
          console.error('错误详情:', {
            errorCode: event.data.errorCode,
            errorDescription: event.data.errorDescription,
            documentUrl: documentUrl
          });
          // 尝试直接访问文档URL进行调试
          console.log('尝试直接访问文档URL:', documentUrl);
        },
        'onDocumentStateChange': (event) => {
          console.log('Document state changed:', event.data);
        }
      };

      // Initialize the editor
      const docEditor = new DocsAPI.DocEditor(
        'onlyoffice-editor', // The ID of the div element where the editor will be initialized
        config
      );
    },
    generateDocumentKey(filename, modified) {
      // 生成基于文件名和修改时间的唯一密钥
      // OnlyOffice对中文字符支持有限，使用Base64编码或MD5哈希
      const timestamp = modified ? new Date(modified).getTime() : Date.now();
      
      // 移除文件扩展名，只对文件名主体进行处理
      const nameWithoutExt = filename.replace(/\.[^/.]+$/, "");
      const extension = filename.split('.').pop();
      
      // 使用Base64编码处理中文文件名，确保OnlyOffice兼容
      const encodedName = btoa(encodeURIComponent(nameWithoutExt)).replace(/[+/=]/g, '_');
      
      return `doc_${encodedName}_${timestamp}.${extension}`;
    },
    getDocumentUrl() {
      // 动态确定filebrowser端口
      const currentUrl = window.location.href;
      let baseUrl = window.location.origin;
      
      // 如果当前是开发环境端口，替换为filebrowser端口
      if (baseUrl.includes(':5174') || baseUrl.includes(':5173')) {
        baseUrl = baseUrl.replace(/:517[34]/, ':8080');
      }
      
      const documentUrl = `${baseUrl}/api/raw${this.file.path}?auth=${this.jwt}`;
      console.log('Document URL:', documentUrl);
      return documentUrl;
    },
    getCallbackUrl() {
      // 生成回调URL，用于OnlyOffice保存文档时通知filebrowser
      let baseUrl = window.location.origin;
      
      // 如果当前是开发环境端口，替换为filebrowser端口
      if (baseUrl.includes(':5174') || baseUrl.includes(':5173')) {
        baseUrl = baseUrl.replace(/:517[34]/, ':8080');
      }
      
      // 添加用户ID参数到回调URL中
      const userId = this.user ? this.user.id : 'anonymous';
      const callbackUrl = `${baseUrl}/api/onlyoffice/callback?userId=${userId}`;
      console.log('Callback URL:', callbackUrl);
      return callbackUrl;
    },
    async storeDocumentKeyMapping(documentKey, filePath) {
      // 将文档密钥和文件路径的映射关系发送到后端存储
      try {
        let baseUrl = window.location.origin;
        
        // 如果当前是开发环境端口，替换为filebrowser端口
        if (baseUrl.includes(':5174') || baseUrl.includes(':5173')) {
          baseUrl = baseUrl.replace(/:517[34]/, ':8080');
        }
        
        const response = await fetch(`${baseUrl}/api/onlyoffice/mapping`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.jwt}`
          },
          body: JSON.stringify({
            key: documentKey,
            path: filePath
          })
        });
        
        if (!response.ok) {
          console.error('Failed to store document key mapping:', response.statusText);
        } else {
          console.log('Document key mapping stored successfully');
        }
      } catch (error) {
        console.error('Error storing document key mapping:', error);
      }
    },

    getFileExtension(filename) {
      const parts = filename.split('.');
      if (parts.length < 2) return '';
      return parts.pop().toLowerCase();
    },

    getDocumentType(filename) {
      const extension = this.getFileExtension(filename);
      switch (extension) {
        case 'docx':
        case 'doc':
        case 'odt':
        case 'rtf':
        case 'txt':
          return 'word';
        case 'xlsx':
        case 'xls':
        case 'ods':
        case 'csv':
          return 'cell';
        case 'pptx':
        case 'ppt':
        case 'odp':
          return 'slide';
        default:
          return 'word'; // Default or handle other types
      }
    }
  }
}
</script>

<style scoped>
#onlyoffice-editor {
  width: 100%;
  height: 100%;
}
</style>