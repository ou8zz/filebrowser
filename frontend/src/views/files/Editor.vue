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

    <!-- <form id="editor"></form> -->
    <div id="monaco-editor" ref="monacoEditor" />
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
import * as monaco from "monaco-editor/esm/vs/editor/editor.main";


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
  props: {
    // 编辑器支持的文本格式,自行在百度上搜索
    types: {
      type: String,
      default: 'json'
    },
    // 名称
    name: {
      type: String,
      default: 'test'
    },
    editorOptions: {
      type: Object,
      default: function() {
        return {
          selectOnLineNumbers: true,
          roundedSelection: false,
          readOnly: false, // 只读
          writeOnly: false,
          cursorStyle: 'line', // 光标样式
          automaticLayout: true, // 自动布局
          glyphMargin: true, // 字形边缘
          useTabStops: false,
          fontSize: 32, // 字体大小
          autoIndent: true // 自动布局
          // quickSuggestionsDelay: 500,   //代码提示延时
        }
      }
    },
    codes: {
      type: String,
      default: function() {
        return ''
      }
    }
  },
  data: function () {
    return {
      editor: null, // 文本编辑器
      isSave: true, // 文件改动状态，是否保存
      codeValue: null // 保存后的文本
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

    // ace.config.set(
    //   "basePath",
    //   `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
    // );

    // this.editor = ace.edit("editor", {
    //   value: fileContent,
    //   showPrintMargin: false,
    //   readOnly: this.req.type === "textImmutable",
    //   theme: "ace/theme/chrome",
    //   mode: modelist.getModeForPath(this.req.name).mode,
    //   wrap: true,
    // });

    // if (theme == "dark") {
    //   this.editor.setTheme("ace/theme/twilight");
    // }


    // 初始化编辑器，确保dom已经渲染
    const self = this;
    let themeVal = "vs-light";
    if (theme == "dark") {
        themeVal = "vs-dark";
    }
    this.editor = monaco.editor.create(self.$refs.monacoEditor, {
      value: fileContent, // 编辑器初始显示内容
      language: 'javascript', // 支持语言
      theme: themeVal, // 主题
      selectOnLineNumbers: true, //显示行号
      editorOptions: self.editorOptions,
    });

    self.editor.onDidChangeModelContent(function(event) {
      // 编辑器内容changge事件
      self.codesCopy = self.editor.getValue()
      self.$emit('onCodeChange', self.editor.getValue(), event)
    })
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
  height: 600px;
}
</style>