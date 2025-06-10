<template>
  <div id="app">
    <div class="container">
      <h1>Log Viewer</h1>
      <div class="layout">
        <div class="sidebar">
          <server-selector />
          <command-input />
          <command-history />
        </div>
        <div class="main-content">
          <log-viewer />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ServerSelector from '@/components/ServerSelector';
import CommandInput from '@/components/CommandInput';
import CommandHistory from '@/components/CommandHistory';
import LogViewer from '@/components/LogViewer';

export default {
  name: 'App',
  components: {
    ServerSelector,
    CommandInput,
    CommandHistory,
    LogViewer,
  },
  created() {
    this.$store.dispatch('fetchServers');
  },
};
</script>

<style>
/* 添加样式 */
</style>