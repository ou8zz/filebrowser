<template>
  <div class="tree-node">
    <div
      class="tree-node-content"
      :class="{ active: isActive, 'has-children': node.isDir }"
      @click="handleClick"
    >
      <span class="toggle-icon" v-if="node.isDir" @click.stop="handleToggle">
        <i v-if="node.loading" class="material-icons rotating">refresh</i>
        <i v-else-if="node.expanded" class="material-icons">expand_more</i>
        <i v-else class="material-icons">chevron_right</i>
      </span>
      <span class="toggle-placeholder" v-else></span>
      <i class="material-icons file-icon" :class="{ 'folder-icon': node.isDir }">
        {{ getFileIcon() }}
      </i>
      <span class="node-name">{{ node.name || "/" }}</span>
    </div>
    <div v-if="node.isDir && node.expanded && node.children.length > 0" class="tree-children">
      <TreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :current-path="currentPath"
        @select="$emit('select', $event)"
        @toggle="$emit('toggle', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface TreeNodeData {
  name: string;
  path: string;
  url: string;
  isDir: boolean;
  expanded: boolean;
  loading: boolean;
  children: TreeNodeData[];
}

const props = defineProps<{
  node: TreeNodeData;
  currentPath: string;
}>();

const emit = defineEmits<{
  select: [node: TreeNodeData];
  toggle: [node: TreeNodeData];
}>();

const isActive = computed(() => {
  return props.node.url === props.currentPath;
});

const handleClick = () => {
  emit("select", props.node);
};

const handleToggle = () => {
  emit("toggle", props.node);
};

const getFileIcon = () => {
  if (props.node.isDir) {
    return props.node.expanded ? "folder_open" : "folder";
  }
  return "insert_drive_file";
};
</script>

<style scoped>
.tree-node {
  user-select: none;
}

.tree-node-content {
  display: flex;
  align-items: center;
  padding: 6px 12px 6px 8px;
  cursor: pointer;
  transition: background 0.15s;
  gap: 6px;
}

.tree-node-content:hover {
  background: var(--background-hover);
}

.tree-node-content.active {
  background: var(--primary-light);
  color: var(--primary-dark);
}

.toggle-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.toggle-placeholder {
  width: 20px;
  flex-shrink: 0;
}

.toggle-icon i {
  font-size: 18px;
  color: var(--text-secondary);
}

.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.file-icon {
  font-size: 20px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.folder-icon {
  color: var(--color-warning);
}

.tree-node-content.active .folder-icon {
  color: var(--primary-dark);
}

.node-name {
  font-size: 16px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tree-children {
  padding-left: 20px;
}
</style>
