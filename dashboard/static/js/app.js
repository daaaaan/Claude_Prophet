document.addEventListener('alpine:init', () => {
    Alpine.store('dashboard', {
        // Data from WebSocket snapshots
        account: { cash: 0, equity: 0, buying_power: 0 },
        positions: [],
        activities: [],
        managedPositions: [],
        botHealth: { alive: false, last_activity: null, uptime: '' },

        // Connection state
        connected: false,
        lastUpdate: null,
        _ws: null,
        _reconnectDelay: 1000,

        init() {
            this.connectWebSocket();
            // Update connectionStatus reactively every second
            setInterval(() => { this.lastUpdate = this.lastUpdate; }, 1000);
        },

        connectWebSocket() {
            const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
            const ws = new WebSocket(`${protocol}//${location.host}/ws`);

            ws.onopen = () => {
                this.connected = true;
                this._reconnectDelay = 1000; // reset backoff
            };

            ws.onclose = () => {
                this.connected = false;
                this._ws = null;
                // Reconnect with exponential backoff, max 30s
                setTimeout(() => this.connectWebSocket(), this._reconnectDelay);
                this._reconnectDelay = Math.min(this._reconnectDelay * 2, 30000);
            };

            ws.onerror = () => {
                ws.close();
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.lastUpdate = new Date();
                if (data.type === 'snapshot') {
                    if (data.account) this.account = data.account;
                    if (data.positions) this.positions = data.positions;
                    if (data.activity) this.activities = data.activity;
                    if (data.bot_health) this.botHealth = data.bot_health;
                    if (data.managed_positions) this.managedPositions = data.managed_positions;
                }
            };

            this._ws = ws;
        },

        get connectionStatus() {
            if (!this.connected) return 'disconnected';
            if (!this.lastUpdate) return 'connecting';
            const age = (Date.now() - this.lastUpdate.getTime()) / 1000;
            if (age < 5) return 'live';
            if (age < 30) return 'stale';
            return 'disconnected';
        },

        get totalUnrealizedPL() {
            return this.positions.reduce((sum, p) => sum + (p.unrealized_pl || 0), 0);
        },

        formatMoney(value) {
            if (value == null) return '$0.00';
            return (value < 0 ? '-' : '') + '$' + Math.abs(value).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
        },

        formatPL(value) {
            if (value == null) return '$0.00';
            return (value >= 0 ? '+' : '-') + '$' + Math.abs(value).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
        },

        formatPct(value) {
            if (value == null) return '0.00%';
            return (value >= 0 ? '+' : '') + value.toFixed(2) + '%';
        },

        formatTimeAgo(timestamp) {
            if (!timestamp) return '';
            const diff = (Date.now() - Date.parse(timestamp)) / 1000;
            if (diff < 60) return Math.floor(diff) + 's ago';
            if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
            if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
            return Math.floor(diff / 86400) + 'd ago';
        },

        activityIcon(type) {
            const icons = {
                'POSITION_OPENED': '\u2197',
                'POSITION_CLOSED': '\u2198',
                'ANALYSIS': '\u2315',
                'INTELLIGENCE': '\u2605',
                'DECISION': '\u2302'
            };
            return icons[type] || '\u2022';
        },

        activityColor(type) {
            const colors = {
                'POSITION_OPENED': 'text-green-400',
                'POSITION_CLOSED': 'text-red-400',
                'ANALYSIS': 'text-blue-400',
                'INTELLIGENCE': 'text-purple-400',
                'DECISION': 'text-yellow-400'
            };
            return colors[type] || 'text-gray-400';
        }
    });
});
