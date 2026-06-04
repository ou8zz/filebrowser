<template>
  <div>
    <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <div class="sidebar-content">
    <button @click="toggleUserPad" class="action">
      <i class="material-icons">person</i>
      <span>用户面板</span>
    </button>
    <div id="user-pad" v-show="showUserPad">
      <template v-if="isLoggedIn">
      
        <button @click="toAccountSettings" class="action">
          <i class="material-icons">person</i>
          <span>{{ user.username }}</span>
        </button>
        <button
          class="action"
          @click="toRoot"
          :aria-label="$t('sidebar.myFiles')"
          :title="$t('sidebar.myFiles')"
        >
          <i class="material-icons">folder</i>
          <span>{{ $t("sidebar.myFiles") }}</span>
        </button>

        <div v-if="user.perm.create">
          <button
            @click="showHover('newDir')"
            class="action"
            :aria-label="$t('sidebar.newFolder')"
            :title="$t('sidebar.newFolder')"
          >
            <i class="material-icons">create_new_folder</i>
            <span>{{ $t("sidebar.newFolder") }}</span>
          </button>

          <button
            @click="showHover('newFile')"
            class="action"
            :aria-label="$t('sidebar.newFile')"
            :title="$t('sidebar.newFile')"
          >
            <i class="material-icons">note_add</i>
            <span>{{ $t("sidebar.newFile") }}</span>
          </button>
        </div>

        <div v-if="user.perm.admin">
          <button
            class="action"
            @click="toGlobalSettings"
            :aria-label="$t('sidebar.settings')"
            :title="$t('sidebar.settings')"
          >
            <i class="material-icons">settings_applications</i>
            <span>{{ $t("sidebar.settings") }}</span>
          </button>
        </div>
        <button
          v-if="canLogout"
          @click="logout"
          class="action"
          id="logout"
          :aria-label="$t('sidebar.logout')"
          :title="$t('sidebar.logout')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t("sidebar.logout") }}</span>
        </button>
      </template>
      <template v-else>
        <router-link
          v-if="!hideLoginButton"
          class="action"
          to="/login"
          :aria-label="$t('sidebar.login')"
          :title="$t('sidebar.login')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t("sidebar.login") }}</span>
        </router-link>

        <router-link
          v-if="signup"
          class="action"
          to="/login"
          :aria-label="$t('sidebar.signup')"
          :title="$t('sidebar.signup')"
        >
          <i class="material-icons">person_add</i>
          <span>{{ $t("sidebar.signup") }}</span>
        </router-link>
      </template>
      </div>

      <!-- Quick Access -->
      <div v-if="isLoggedIn && isFiles" class="sidebar-quick-access">
        <div class="section-title">快捷访问</div>
        <QuickAccessTree />
      </div>

      <!-- Directory Tree -->
      <DirectoryTree v-if="isLoggedIn && isFiles" class="sidebar-directory-tree" />

      <div
        class="credits"
        v-if="isFiles && !disableUsedPercentage"
      >
        <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
        <br />
        {{ usage.used }} of {{ usage.total }} used
      </div>

      <p class="credits">
        <span>
          <span v-if="disableExternal">File Browser</span>
          <a
            v-else
            rel="noopener noreferrer"
            target="_blank"
            href="https://github.com/filebrowser/filebrowser"
            >File Browser</a
          >
          <span> {{ " " }} {{ version }}</span>
        </span>
        <span>
          <a @click="help">{{ $t("sidebar.help") }}</a>
        </span>
      </p>
    </div>
  </nav>
  </div>
</template>

<script>
import { reactive, ref } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import DirectoryTree from "@/components/DirectoryTree.vue";
import QuickAccessTree from "@/components/QuickAccessTree.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    const showUserPad = ref(false);
    return { 
      usage, usageAbortController: new AbortController(), 
      showUserPad };
  },
  components: {
    ProgressBar,
    DirectoryTree,
    QuickAccessTree,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    toggleUserPad() {
      this.showUserPad = !this.showUserPad;
    },
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      this.$router.push({ path: "/files" });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>

<style scoped>
.sidebar-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.sidebar-quick-access {
  padding: 0.5em 0;
  border-bottom: 1px solid var(--divider);
}

.section-title {
  padding: 0.5em 1em;
  font-size: 0.9em;
  font-weight: 500;
  color: var(--text-secondary);
}

.sidebar-directory-tree {
  flex: 1;
  min-height: 0;
  margin: 0.5em 0;
  overflow: hidden;
}
</style>
