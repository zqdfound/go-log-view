<template>
  <div class="log-viewer">
    <div class="log-controls">
      <select v-model="selectedServer" @change="updateLogFiles">
        <option value="">Select Server</option>
        <option v-for="server in servers" :key="server.name" :value="server.name">
          {{ server.name }}
        </option>
      </select>

      <select v-model="selectedFile" :disabled="!selectedServer">
        <option value="">Select Log File</option>
        <option v-for="file in logFiles" :key="file.alias" :value="file.alias">
          {{ file.alias }}
        </option>
      </select>

      <input
          type="number"
          v-model.number="lines"
          min="10"
          max="1000"
          placeholder="Lines to show"
      >

      <button
          @click="startLog"
          :disabled="!selectedServer || !selectedFile"
      >
        Start
      </button>

      <button
          @click="stopLog"
          :disabled="!selectedServer || !selectedFile"
      >
        Stop
      </button>

      <button @click="clearLog">Clear</button>
    </div>

    <div class="log-container">
      <pre v-if="activeLog">{{ activeLog }}</pre>
      <div v-else class="empty-log">No log selected or active</div>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from 'vuex';

export default {
  name: 'LogViewer',
  data() {
    return {
      selectedServer: '',
      selectedFile: '',
      lines: 100,
    };
  },
  computed: {
    ...mapState(['servers', 'activeLogs']),
    ...mapGetters(['getLogFilesForServer']),

    logFiles() {
      if (!this.selectedServer) return [];
      const server = this.servers.find(s => s.name === this.selectedServer);
      return server ? server.log_files : [];
    },

    activeLog() {
      if (!this.selectedServer || !this.selectedFile) return null;
      return this.activeLogs[this.selectedServer] &&
          this.activeLogs[this.selectedServer][this.selectedFile];
    },
  },
  created() {
    this.$store.dispatch('fetchServers');

    // 设置WebSocket监听
    this.unsubscribe = this.$ws.addListener('log', (message) => {
      if (message.server === this.selectedServer && message.file === this.selectedFile) {
        this.$store.commit('addLogEntry', {
          server: message.server,
          file: message.file,
          content: message.content,
        });
      }
    });
  },
  beforeDestroy() {
    this.unsubscribe && this.unsubscribe();
  },
  methods: {
    updateLogFiles() {
      this.selectedFile = '';
    },

    async startLog() {
      try {
        await this.$store.dispatch('startLog', {
          server: this.selectedServer,
          file: this.selectedFile,
          lines: this.lines,
        });
      } catch (error) {
        this.$notify({
          type: 'error',
          title: 'Error',
          text: 'Failed to start log: ' + error.message,
        });
      }
    },

    async stopLog() {
      try {
        await this.$store.dispatch('stopLog', {
          server: this.selectedServer,
          file: this.selectedFile,
        });
      } catch (error) {
        this.$notify({
          type: 'error',
          title: 'Error',
          text: 'Failed to stop log: ' + error.message,
        });
      }
    },

    clearLog() {
      if (this.selectedServer && this.selectedFile) {
        this.$store.commit('clearLog', {
          server: this.selectedServer,
          file: this.selectedFile,
        });
      }
    },
  },
};
</script>

<style scoped>
/* 添加样式 */
</style>