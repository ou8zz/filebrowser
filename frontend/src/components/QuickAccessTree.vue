<template>
  <div class="quick-access-tree">
    <div v-if="loading" class="loading small">加载中...</div>
    <template v-else-if="rootNodes.length > 0">
      <TreeNode
        v-for="node in rootNodes"
        :key="node.path"
        :node="node"
        :current-path="currentPath"
        @select="handleSelect"
        @toggle="handleToggle"
      />
    </template>
    <div v-else class="empty">暂无快捷访问，请在设置中配置</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { files as apiFiles, settings as apiSettings } from "@/api";
import TreeNode from "./TreeNode.vue";

interface TreeNodeData {
  name: string;
  path: string;
  url: string;
  isDir: boolean;
  expanded: boolean;
  loading: boolean;
  children: TreeNodeData[];
}

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const rootNodes = ref<TreeNodeData[]>([]);
const abortController = ref(new AbortController());

const currentPath = computed(() => route.path);

const loadNode = async (node: TreeNodeData) => {
  if (node.loading || !node.isDir) return;
  
  node.loading = true;
  try {
    const res = await apiFiles.fetch(node.url);
    if (res.items) {
      node.children = res.items.map(item => ({
        name: item.name,
        path: item.path,
        url: item.isDir ? `/files${item.path}/` : `/files${item.path}`,
        isDir: item.isDir,
        expanded: false,
        loading: false,
        children: [] as TreeNodeData[]
      })).sort((a, b) => {
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
        return a.name.localeCompare(b.name);
      });
    }
    node.expanded = !node.expanded;
  } finally {
    node.loading = false;
  }
};

const loadRoot = async () => {
  loading.value = true;
  try {
    abortController.value.abort();
    abortController.value = new AbortController();
    
    // 获取配置
    const settings = await apiSettings.get();
    const paths = settings.quickAccessPaths || [];
    
    const nodes: TreeNodeData[] = [];
    for (const path of paths) {
      try {
        const res = await apiFiles.fetch(`/files${path}`);
        
        const nodeName = path.split("/").filter(Boolean).pop() || path;
        
        nodes.push({
          name: nodeName,
          path: path,
          url: res.url,
          isDir: true,
          expanded: false,
          loading: false,
          children: res.items?.map(item => ({
            name: item.name,
            path: item.path,
            url: item.isDir ? `/files${item.path}/` : `/files${item.path}`,
            isDir: item.isDir,
            expanded: false,
            loading: false,
            children: [] as TreeNodeData[]
          })).sort((a, b) => {
            if (a.isDir && !b.isDir) return -1;
            if (!a.isDir && b.isDir) return 1;
            return a.name.localeCompare(b.name);
          }) || []
        });
      } catch (e) {
        console.error(`Failed to load quick access path ${path}:`, e);
      }
    }
    
    rootNodes.value = nodes;
  } catch (e) {
    if (e instanceof Error && e.name !== 'AbortError') {
      console.error("Failed to load quick access:", e);
    }
  } finally {
    loading.value = false;
  }
};

const handleSelect = (node: TreeNodeData) => {
  router.push(node.url);
};

const handleToggle = (node: TreeNodeData) => {
  loadNode(node);
};

onMounted(() => {
  loadRoot();
});

onUnmounted(() => {
  abortController.value.abort();
});
</script>

<style scoped>
.quick-access-tree {
  margin: 0;
}

.loading.small {
  padding: 0.5em 1em;
  color: var(--text-secondary);
  font-size: 0.9em;
}

.empty {
  padding: 0.5em 1em;
  color: var(--text-secondary);
  font-size: 0.9em;
}
</style>
