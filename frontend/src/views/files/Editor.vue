<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>

      <action
        v-if="user.perm.modify"
        id="save-button"
        icon="save"
        :label="$t('buttons.save')"
        @action="save()"
      />
    </header-bar>

    <breadcrumbs base="/files" noLink />

    <button @click="switchAce" class="switch-btn">切换编辑器</button>
    <form id="editor"></form>
    <textarea v-if="isAce==false" class="ww" v-model="codeValue"></textarea>
  </div>
</template>

<script>
import { mapState } from "vuex";
import { files as api } from "@/api";
import { theme } from "@/utils/constants";
import buttons from "@/utils/buttons";
import url from "@/utils/url";

import { version as ace_version } from "ace-builds";
import ace from "ace-builds/src-min-noconflict/ace.js";
import modelist from "ace-builds/src-min-noconflict/ext-modelist.js";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";

export default {
  name: "editor",
  components: {
    HeaderBar,
    Action,
    Breadcrumbs,
  },
  data: function () {
    return {
      isAce: true,  // 默认Ace
      editor: null, // 文本编辑器
      isSave: true, // 文件改动状态，是否保存
      codeValue: null, // 保存后的文本
    };
  },
  computed: {
    ...mapState(["req", "user"]),
    breadcrumbs() {
      let parts = this.$route.path.split("/");

      if (parts[0] === "") {
        parts.shift();
      }

      if (parts[parts.length - 1] === "") {
        parts.pop();
      }

      let breadcrumbs = [];

      for (let i = 0; i < parts.length; i++) {
        breadcrumbs.push({ name: decodeURIComponent(parts[i]) });
      }

      breadcrumbs.shift();

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift();
        }

        breadcrumbs[0].name = "...";
      }

      return breadcrumbs;
    },
  },
  created() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
    this.editor.destroy();
  },
  mounted: function () {
    const fileContent = this.req.content || "";
    if (this.isAce) {
      this.initAce();
    } else {
      this.codeValue = fileContent;
    }
  },
  methods: {
    switchAce() {
      const fileContent = this.req.content || "";
      if(this.isAce) {
        this.isAce = false;
        document.getElementById("editor").style.display = 'none';
        this.codeValue = fileContent;
      } else {
        this.isAce = true;
        document.getElementById("editor").style.display = 'block';
      }
    },
    initAce() {
      const fileContent = this.req.content || "";
      ace.config.set(
        "basePath",
        `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
      );

      this.editor = ace.edit("editor", {
        value: fileContent,
        showPrintMargin: false,
        readOnly: this.req.type === "textImmutable",
        theme: "ace/theme/chrome",
        mode: modelist.getModeForPath(this.req.name).mode,
        wrap: true,
        keyboardHandler: "ace/keyboard/vscode",
      });

      if (theme == "dark") {
        this.editor.setTheme("ace/theme/monokai");
      }
    },
    back() {
      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
    keyEvent(event) {
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }

      if (String.fromCharCode(event.which).toLowerCase() !== "s") {
        return;
      }

      event.preventDefault();
      this.save();
    },
    async save() {
      const button = "save";
      buttons.loading("save");

      try {
        let val = this.editor.getValue();
    
        if(!this.isAce) {
          val = this.codeValue;
              console.log("save ace", this.isAce, val);
        }
        await api.put(this.$route.path, val);
        buttons.success(button);
      } catch (e) {
        buttons.done(button);
        this.$showError(e);
      }
    },
    close() {
      this.$store.commit("updateRequest", {});

      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  },
};
</script>

<style scoped>
.switch-btn {
  font-size: 12px;
  background-color: #272822;
  color: #F8F8F2;
  border: 0px;
}

.ww {
  width: 100%;
  height: 100vh;
  background-color: #272822;
  color: #F8F8F2;
  font: 12px / normal 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'Source Code Pro', 'source-code-pro', monospace;
}
</style>