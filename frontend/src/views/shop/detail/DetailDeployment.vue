<template>
    <div v-if="deployment" class="deployment-detail">
        <div class="deployment-header">
            <h2 class="deployment-name">{{ deployment.name }}</h2>
        </div>

        <div class="panel deployment-details-panel">
            <h3 class="panel-title">Deployment Details</h3>

            <div class="details-grid">
                <div class="detail-row">
                    <span class="detail-label">Duration:</span>
                    <span class="detail-value">{{ formatDuration(deployment.executionTime) }}</span>
                </div>

                <div class="detail-row">
                    <span class="detail-label">Command:</span>
                    <code class="detail-value detail-code">{{ deployment.command }}</code>
                </div>

                <div class="detail-row">
                    <span class="detail-label">Exit Code:</span>
                    <span class="detail-value" :class="deployment.returnCode === 0 ? 'detail-success' : 'detail-error'">
                        <icon-fa6-solid:check v-if="deployment.returnCode === 0" class="icon" />
                        <icon-fa6-solid:xmark v-else class="icon" />
                        {{ deployment.returnCode }}
                    </span>
                </div>

                <div class="detail-row">
                    <span class="detail-label">Started:</span>
                    <span class="detail-value">{{ formatDateTime(deployment.startDate) }}</span>
                </div>

                <div class="detail-row">
                    <span class="detail-label">Completed:</span>
                    <span class="detail-value">{{ formatDateTime(deployment.endDate) }}</span>
                </div>

                <div v-if="deployment.reference" class="detail-row">
                    <span class="detail-label">Reference:</span>
                    <a :href="deployment.reference" target="_blank" rel="noopener noreferrer" class="detail-value detail-link">
                        <icon-fa6-solid:arrow-up-right-from-square class="icon" />
                        View Commit
                    </a>
                </div>
            </div>
        </div>

        <div v-if="deployment.composer && Object.keys(deployment.composer).length > 0" class="panel">
            <h3 class="panel-title">
                <icon-fa6-solid:box class="icon" />
                Composer Packages
            </h3>
            <div class="composer-table-container">
                <table class="composer-table">
                    <thead>
                        <tr>
                            <th>Package</th>
                            <th>Version</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="(version, packageName) in deployment.composer"
                            :key="packageName"
                        >
                            <td class="package-name">{{ packageName }}</td>
                            <td class="package-version">{{ version }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="panel">
            <h3 class="panel-title">
                <icon-fa6-solid:terminal class="icon" />
                Output
            </h3>
            <div class="output-container">
                <pre class="output-content" v-html="formattedOutput"></pre>
            </div>
        </div>
    </div>
    <div v-else class="loading">
        Loading deployment details...
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useShopDetail } from '@/composables/useShopDetail';
import { formatDateTime } from '@/helpers/formatter';
import { trpcClient } from '@/helpers/trpc';

const route = useRoute();
const { shop } = useShopDetail();

const deployment = ref<any>(null);

const loadDeployment = async () => {
    if (!shop.value) return;

    try {
        const deploymentId = parseInt(route.params.deploymentId as string, 10);
        deployment.value = await trpcClient.organization.deployment.get.query({
            shopId: shop.value.id,
            deploymentId,
        });
    } catch (error) {
        console.error('Failed to load deployment:', error);
    }
};

const formatDuration = (seconds: string) => {
    const num = parseFloat(seconds);
    if (num < 1) {
        return `${(num * 1000).toFixed(0)}ms`;
    }
    if (num < 60) {
        return `${num.toFixed(2)}s`;
    }
    const minutes = Math.floor(num / 60);
    const secs = (num % 60).toFixed(0);
    return `${minutes}m ${secs}s`;
};

const ansiToHtml = (text: string) => {
    if (!text) return '';

    const ansiColors: Record<number, string> = {
        // Reset
        0: '</span>',
        // Regular colors
        30: '#24292e', // Black
        31: '#f85149', // Red
        32: '#3fb950', // Green
        33: '#d29922', // Yellow
        34: '#58a6ff', // Blue
        35: '#bc8cff', // Magenta
        36: '#39c5cf', // Cyan
        37: '#c9d1d9', // White
        // Bright colors
        90: '#6e7681', // Bright Black (Gray)
        91: '#ff7b72', // Bright Red
        92: '#56d364', // Bright Green
        93: '#e3b341', // Bright Yellow
        94: '#79c0ff', // Bright Blue
        95: '#d2a8ff', // Bright Magenta
        96: '#56d4dd', // Bright Cyan
        97: '#f0f6fc', // Bright White
        // Default/Reset
        39: '#c9d1d9', // Default foreground
    };

    let html = '';
    let currentColor = '';

    // Replace ANSI escape sequences with HTML
    const parts = text.split(/\x1b\[/);

    for (let i = 0; i < parts.length; i++) {
        if (i === 0) {
            // First part has no escape sequence
            html += parts[i];
            continue;
        }

        const match = parts[i].match(/^(\d+(?:;\d+)*)m(.*)$/s);
        if (match) {
            const codes = match[1].split(';').map(Number);
            const content = match[2];

            for (const code of codes) {
                if (code === 0) {
                    // Reset
                    if (currentColor) {
                        html += '</span>';
                        currentColor = '';
                    }
                } else if (ansiColors[code]) {
                    // Close previous color if exists
                    if (currentColor) {
                        html += '</span>';
                    }
                    // Open new color
                    currentColor = ansiColors[code];
                    html += `<span style="color: ${currentColor}">`;
                }
            }

            html += content;
        } else {
            html += parts[i];
        }
    }

    // Close any remaining open span
    if (currentColor) {
        html += '</span>';
    }

    return html;
};

const formattedOutput = computed(() => {
    if (!deployment.value?.output) return '';
    return ansiToHtml(deployment.value.output);
});

// Watch for shop to be available
watch(shop, (newShop) => {
    if (newShop) {
        loadDeployment();
    }
}, { immediate: true });

onMounted(() => {
    if (shop.value) {
        loadDeployment();
    }
});
</script>

<style scoped>
.deployment-detail {
    max-width: 1200px;
}

.deployment-header {
    margin-bottom: 1.5rem;
}

.deployment-name {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
    color: #fff;
}

.deployment-details-panel {
    padding: 1.5rem;
}

.deployment-details-panel .panel-title {
    margin: 0 0 1.5rem 0;
    font-size: 1.25rem;
    font-weight: 600;
}

.details-grid {
    display: flex;
    flex-direction: column;
    gap: 0;
}

.detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color);
}

.detail-row:last-child {
    border-bottom: none;
}

.detail-label {
    font-size: 1rem;
    color: var(--text-muted);
    font-weight: 400;
}

.detail-value {
    font-size: 1rem;
    color: var(--text-primary);
    font-weight: 500;
    text-align: right;
}

.detail-code {
    font-family: monospace;
    background: var(--surface-color);
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.875rem;
    max-width: 60%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.detail-success {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    color: var(--success-color);
}

.detail-success .icon {
    font-size: 1rem;
}

.detail-error {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    color: var(--error-color);
}

.detail-error .icon {
    font-size: 1rem;
}

.detail-link {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    color: var(--primary-color);
    text-decoration: none;
}

.detail-link:hover {
    text-decoration: underline;
}

.detail-link .icon {
    font-size: 1rem;
}

.composer-table-container {
    overflow-x: auto;
}

.composer-table {
    width: 100%;
    border-collapse: collapse;
}

.composer-table thead {
    background: var(--surface-color);
    border-bottom: 2px solid var(--border-color);
}

.composer-table th {
    padding: 0.5rem 1rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.composer-table td {
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--border-color);
}

.composer-table tbody tr:last-child td {
    border-bottom: none;
}

.composer-table tbody tr:hover {
    background: var(--surface-color);
}

.composer-table .package-name {
    font-family: monospace;
    color: var(--text-primary);
    font-size: 0.9rem;
}

.composer-table .package-version {
    font-family: monospace;
    color: var(--text-muted);
    font-weight: 600;
    font-size: 0.9rem;
}

.output-container {
    background: #0d1117;
    overflow: hidden;
    padding: 1rem;
    border-radius: 6px;
}

.output-content {
    margin: 0;
    padding: 0;
    overflow-x: auto;
    font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', 'Fira Mono', 'Roboto Mono', 'Courier New', monospace;
    font-size: 0.8125rem;
    line-height: 1.6;
    color: #c9d1d9;
    white-space: pre;
    word-wrap: normal;
    tab-size: 4;
    -moz-tab-size: 4;
}

/* Style ANSI-like patterns for better terminal look */
.output-content::before {
    content: '';
    display: block;
    height: 0;
}

/* Add scrollbar styling for dark theme */
.output-container::-webkit-scrollbar {
    width: 10px;
    height: 10px;
}

.output-container::-webkit-scrollbar-track {
    background: #161b22;
    border-radius: 6px;
}

.output-container::-webkit-scrollbar-thumb {
    background: #30363d;
    border-radius: 6px;
}

.output-container::-webkit-scrollbar-thumb:hover {
    background: #484f58;
}

.loading {
    padding: 3rem;
    text-align: center;
    color: var(--text-muted);
}
</style>
