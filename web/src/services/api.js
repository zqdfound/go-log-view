import axios from 'axios';

const api = axios.create({
    baseURL: process.env.VUE_APP_API_URL || '/api',
});

export default {
    getServers() {
        return api.get('/servers');
    },

    startLog(serverName, fileAlias, lines = 100) {
        return api.post('/log/start', {
            serverName,
            fileAlias,
            lines,
        });
    },

    stopLog(serverName, fileAlias) {
        return api.post('/log/stop', {
            serverName,
            fileAlias,
        });
    },

    executeCommand(serverName, command) {
        return api.post('/command', {
            serverName,
            command,
        });
    },
};