<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.fileInfo") }}</h2>
    </div>

    <div class="card-content">
      <p v-if="selected.length > 1">
        {{ $t("prompts.filesSelected", { count: selected.length }) }}
      </p>

      <p class="break-word" v-if="selected.length < 2">
        <strong>{{ $t("prompts.displayName") }}</strong> {{ name }}
      </p>

      <p>
        <strong>{{ $t("prompts.size") }}:</strong>
        <span v-if="humanSizeLoading">{{ $t("files.loading") }}</span>
        <span v-else>{{ humanSize }}</span>
      </p>

      <div v-if="resolution">
        <strong>{{ $t("prompts.resolution") }}:</strong>
        {{ resolution.width }} x {{ resolution.height }}
      </div>

      <p v-if="selected.length < 2" :title="modTime">
        <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
      </p>

      <template v-if="dir && selected.length === 0">
        <p>
          <strong>{{ $t("prompts.numberFiles") }}:</strong> {{ req.numFiles }}
        </p>
        <p>
          <strong>{{ $t("prompts.numberDirs") }}:</strong> {{ req.numDirs }}
        </p>
      </template>

      <template v-if="!dir">
        <p>
          <strong>MD5: </strong
          ><code
            ><a
              @click="checksum($event, 'md5')"
              @keypress.enter="checksum($event, 'md5')"
              tabindex="2"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA1: </strong
          ><code
            ><a
              @click="checksum($event, 'sha1')"
              @keypress.enter="checksum($event, 'sha1')"
              tabindex="3"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA256: </strong
          ><code
            ><a
              @click="checksum($event, 'sha256')"
              @keypress.enter="checksum($event, 'sha256')"
              tabindex="4"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA512: </strong
          ><code
            ><a
              @click="checksum($event, 'sha512')"
              @keypress.enter="checksum($event, 'sha512')"
              tabindex="5"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
      </template>
    </div>

    <div class="card-action">
      <button
        id="focus-prompt"
        type="submit"
        @click="closeHovers"
        class="button button--flat"
        :aria-label="$t('buttons.ok')"
        :title="$t('buttons.ok')"
      >
        {{ $t("buttons.ok") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { files as api } from "@/api";
import { fetchURL } from "@/api/utils";
import { ref } from 'vue';

const humanSize = ref("");
const humanSizeLoading = ref(false);

export default {
  name: "info",
  inject: ["$showError"],
  setup() {
    return {
      humanSize,
      humanSizeLoading,
    }
  },
  created() {
      this.getHumanSize();
  },
  computed: {
    ...mapState(useFileStore, [
      "req",
      "selected",
      "selectedCount",
      "isListing",
    ]),
    // humanSize: function () {
    //   if (this.selectedCount === 0 || !this.isListing) {
    //     return filesize(this.req.size);
    //   }

    //   let sum = 0;

    //   for (const selected of this.selected) {
    //     sum += this.req.items[selected].size;
    //   }

    //   return filesize(sum);
    // },
    humanTime: function () {
      if (this.selectedCount === 0) {
        return dayjs(this.req.modified).fromNow();
      }

      return dayjs(this.req.items[this.selected[0]].modified).fromNow();
    },
    modTime: function () {
      if (this.selectedCount === 0) {
        return new Date(Date.parse(this.req.modified)).toLocaleString();
      }

      return new Date(
        Date.parse(this.req.items[this.selected[0]].modified)
      ).toLocaleString();
    },
    name: function () {
      return this.selectedCount === 0
        ? this.req.name
        : this.req.items[this.selected[0]].name;
    },
    dir: function () {
      return (
        this.selectedCount > 1 ||
        (this.selectedCount === 0
          ? this.req.isDir
          : this.req.items[this.selected[0]].isDir)
      );
    },
    resolution: function () {
      if (this.selectedCount === 1) {
        const selectedItem = this.req.items[this.selected[0]];
        if (selectedItem && selectedItem.type === "image") {
          return selectedItem.resolution;
        }
      } else if (this.req && this.req.type === "image") {
        return this.req.resolution;
      }
      return null;
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    checksum: async function (event, algo) {
      event.preventDefault();

      let link;

      if (this.selectedCount) {
        link = this.req.items[this.selected[0]].url;
      } else {
        link = this.$route.path;
      }

      try {
        const hash = await api.checksum(link, algo);
        event.target.textContent = hash;
      } catch (e) {
        this.$showError(e);
      }
    },
    getHumanSize: async function () {
      let sumFiles = 0;
      const dirPaths = [];

      if (this.selectedCount === 0 || !this.isListing) {
        if (this.req?.isDir) {
          dirPaths.push(this.req.path);
        } else if (this.req) {
          humanSize.value = filesize(this.req.size);
          return;
        }
      } else {
        for (const selected of this.selected) {
          const item = this.req.items[selected];
          if (item.isDir) {
            dirPaths.push(item.path);
          } else {
            sumFiles += item.size;
          }
        }
      }

      if (dirPaths.length === 0) {
        humanSize.value = filesize(sumFiles);
        return;
      }

      humanSizeLoading.value = true;
      try {
        const urls = dirPaths.join(",");
        const res = await fetchURL(`/api/size?paths=${encodeURIComponent(urls)}`, {});
        const data = await res.json();
        const total = (data?.size ?? 0) + sumFiles;
        humanSize.value = filesize(total);
      } catch (e) {
        humanSize.value = filesize(sumFiles);
        this.$showError(e);
      } finally {
        humanSizeLoading.value = false;
      }
    },
  },
};
</script>
