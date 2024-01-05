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

    <form id="editor"></form>
    <!-- <div id="monaco-editor" ref="monacoEditor" /> -->
    <!-- <textarea class="ww" :value="codeValue"></textarea> -->
    <!-- <vue-editor class="ww" v-model="codeValue" :editorToolbar="customToolbar"></vue-editor> -->
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
import { VueEditor } from "vue2-editor";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";

export default {
  name: "editor",
  components: {
    // VueEditor,
    HeaderBar,
    Action,
    Breadcrumbs,
  },
  data: function () {
    return {
      editor: null, // 文本编辑器
      isSave: true, // 文件改动状态，是否保存
      codeValue: null, // 保存后的文本
      // customToolbar: [
      //   ["bold", "italic", "underline"],
      //   [{ list: "ordered" }, { list: "bullet" }],
      //   ["code-block"]
      // ]
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
    // this.editor.destroy();
  },
  mounted: function () {
    const fileContent = this.req.content || "";
    // this.codeValue = fileContent;

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
      // keyboardHandler: "edit",
    });

    if (theme == "dark") {
      this.editor.setTheme("ace/theme/twilight");
    }
  },
  methods: {
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
        await api.put(this.$route.path, this.editor.getValue());
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
#monaco-editor {
  width: 100%;
  height: 100%;
}

.ww {
  width: 100%;
  height: 88%;
}
</style>