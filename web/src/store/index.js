import Vue from 'vue';
import Vuex from 'vuex';
import api from '@/services/api';

Vue.use(Vuex);

export default new Vuex.Store({
    state: {
        servers: [],
        activeLogs: {},
        commandHistory: [],
    },
    mutations: {
        setServers(state, servers) {
            state.servers = servers;
        },
        addLogEntry(state, { server, file, content }) {
            if (!state.activeLogs[server]) {
                Vue.set(state.activeLogs, server, {});
            }
            if (!state.activeLogs[server][file]) {
                Vue.set(state.activeLogs[server], file, '');
            }
            state.activeLogs[server][file] += content;
        },
        clearLog(state, { server, file }) {
            if (state.activeLogs[server] && state.activeLogs[server][file]) {
                Vue.set(state.activeLogs[server], file, '');
            }
        },
        addCommandToHistory(state, command) {
            state.commandHistory.unshift(command);
            if (state.commandHistory.length > 50) {
                state.commandHistory.pop();
            }
        },
    },
    actions: {
        async fetchServers({ commit }) {
            try {
                const response = await api.getServers();
                commit('setServers', response.data);
            } catch (error) {
                console.error('Error fetching servers:', error);
            }
        },
        async startLog({ commit }, { server, file, lines }) {
            try {
                await api.startLog(server, file, lines);
            } catch (error) {
                console.error('Error starting log:', error);
                throw error;
            }
        },
        async stopLog({ commit }, { server, file }) {
            try {
                await api.stopLog(server, file);
                commit('clearLog', { server, file });
            } catch (error) {
                console.error('Error stopping log:', error);
            }
        },
        async executeCommand({ commit }, { server, command }) {
            try {
                const response = await api.executeCommand(server, command);
                commit('addCommandToHistory', { server, command, output: response.data.output });
                return response.data.output;
            } catch (error) {
                console.error('Error executing command:', error);
                throw error;
            }
        },
    },
});