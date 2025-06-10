<template>
  <div class="command-input">
    <select v-model="selectedServer">
      <option value="">Select Server</option>
      <option v-for="server in servers" :key="server.name" :value="server.name">
        {{ server.name }}
      </option>
    </select>

    <input
        type="text"
        v-model="command"
        placeholder="Enter command..."
        @keyup.enter="executeCommand"
        :disabled="!selectedServer"
    >

    <button @click="executeCommand" :disabled="!selectedServer || !command">
      Execute
    </button>

    <div v-if="output" class="command-output">
      <pre>{{ output }}</pre>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex';

export default {
  name: 'CommandInput',
  data() {
    return {
      selectedServer: '',
      command: '',
      output: '',
    };
  },
  computed: {
    ...mapState(['servers']),
  },
  methods: {
    async executeCommand() {
      if (!this.selectedServer || !this.command) return;

      try {
        this.output = await this.$store.dispatch('executeCommand', {
          server: this.selectedServer,
          command: this.command,
        });

        this.command = '';
      } catch (error) {
        this.output = `Error: ${error.message}`;
      }
    },
  },
};
</script>

<style scoped>
/* 添加样式 */
</style>