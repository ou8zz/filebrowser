<template>
  <div class="directory-tree">
    <div class="tree-header">
      <span>{{ t("sidebar.directoryTree") }}</span>
      <button @click="refreshTree" class="refresh-btn" :title="t('sidebar.refresh')">
        <i class="material-icons">refresh</i>
      </button>
    </div>
    <div class="tree-content">
      <TreeNode
        :node="rootNode"
        :current-path="route.path"
        @select="handleSelect"
        @toggle="handleToggle"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter, useRoute } from "vue-router";
import { files as api } from "@/api";
import TreeNode from "./TreeNode.vue";

const { t } = useI18n();
const router = useRouter();
const route = useRoute();

interface TreeNodeData {
  name: string;
  path: string;
  url: string;
  isDir: boolean;
  expanded: boolean;
  loading: boolean;
  children: TreeNodeData[];
}

const rootNode = ref<TreeNodeData>({
  name: "",
  path: "/",
  url: "/files/",
  isDir: true,
  expanded: true,
  loading: false,
  children: [],
});

const loadChildren = async (node: TreeNodeData) => {
  if (node.loading || !node.isDir) return;
  
  node.loading = true;
  try {
    const res = await api.fetch(node.url);
    if (res.isDir && res.items) {
      node.children = res.items
        .filter((item: any) => item.isDir)
        .map((item: any) => {
          const childPath = node.path === "/" 
            ? `/${item.name}` 
            : `${node.path}/${item.name}`;
          const childUrl = node.url === "/files/" 
            ? `/files/${encodeURIComponent(item.name)}/` 
            : `${node.url}${encodeURIComponent(item.name)}/`;
          
          return {
            name: item.name,
            path: childPath,
            url: childUrl,
            isDir: item.isDir,
            expanded: false,
            loading: false,
            children: [],
          };
        });
    }
  } catch (err) {
    console.error("Failed to load directory:", err);
  } finally {
    node.loading = false;
  }
};

const expandToPath = async (path: string, node: TreeNodeData = rootNode.value) => {
  // 确保路径以 / 结尾，统一比较
  const normalizedPath = path.endsWith("/") ? path : path + "/";
  const normalizedNodeUrl = node.url.endsWith("/") ? node.url : node.url + "/";
  
  if (normalizedPath === normalizedNodeUrl) {
    node.expanded = true;
    return true;
  }

  if (normalizedPath.startsWith(normalizedNodeUrl)) {
    node.expanded = true;
    
    if (node.children.length === 0) {
      await loadChildren(node);
    }

    for (const child of node.children) {
      const normalizedChildUrl = child.url.endsWith("/") ? child.url : child.url + "/";
      if (normalizedPath.startsWith(normalizedChildUrl)) {
        await expandToPath(path, child);
        return true;
      }
    }
  }
  return false;
};

const handleSelect = (node: TreeNodeData) => {
  router.push(node.url);
};

const handleToggle = async (node: TreeNodeData) => {
  if (!node.expanded && node.children.length === 0) {
    await loadChildren(node);
  }
  node.expanded = !node.expanded;
};

const refreshTree = async () => {
  rootNode.value.children = [];
  await loadChildren(rootNode.value);
  await expandToPath(route.path);
};

watch(() => route.path, async (newPath) => {
  await expandToPath(newPath);
});

onMounted(async () => {
  await loadChildren(rootNode.value);
  await expandToPath(route.path);
});
</script>

<style scoped>
.directory-tree {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  font-weight: 500;
  font-size: 0.9em;
  color: var(--textPrimary);
  flex-shrink: 0;
}

.refresh-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--textSecondary);
  border-radius: 4px;
  transition: background 0.2s;
}

.refresh-btn:hover {
  background: var(--divider);
  color: var(--textPrimary);
}

.refresh-btn i {
  font-size: 16px;
}

.tree-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 4px 0;
  min-height: 0;
}
</style>
