<template>
    <div class="sitespeed-tab">
        <div class="sitespeed-header">
            <h3>Sitespeed Metrics</h3>
            <div class="sitespeed-actions">
                <a
                    v-if="shop.sitespeed && shop.sitespeed.length > 0"
                    :href="`/sitespeed/result/${shop.id}/`"
                    target="_blank"
                    class="btn btn-secondary"
                >
                    <icon-fa6-solid:arrow-up-right-from-square />
                    View Details
                </a>
                <button
                    class="btn btn-primary"
                    :disabled="isAnalyzing"
                    @click="runAnalysis"
                >
                    <icon-fa6-solid:gauge :class="{ 'animate-spin': isAnalyzing }" />
                    {{ isAnalyzing ? 'Analyzing...' : 'Run Analysis' }}
                </button>
            </div>
        </div>

        <div v-if="shop.sitespeed && shop.sitespeed.length > 0" class="sitespeed-content">
            <div class="sitespeed-latest">
                <h4>Latest Analysis</h4>
                <div class="metrics-grid">
                    <div class="metric-card">
                        <div class="metric-label">TTFB</div>
                        <div class="metric-value">{{ latestMetrics.ttfb || 'N/A' }}<span v-if="latestMetrics.ttfb">ms</span></div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">Fully Loaded</div>
                        <div class="metric-value">{{ formatTime(latestMetrics.fullyLoaded) }}</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">LCP</div>
                        <div class="metric-value">{{ formatTime(latestMetrics.largestContentfulPaint) }}</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">FCP</div>
                        <div class="metric-value">{{ formatTime(latestMetrics.firstContentfulPaint) }}</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">Speed Index</div>
                        <div class="metric-value">{{ formatTime(latestMetrics.speedIndex) }}</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">Transfer Size</div>
                        <div class="metric-value">{{ formatBytes(latestMetrics.transferSize) }}</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-label">CLS</div>
                        <div class="metric-value">{{ formatCLS(latestMetrics.cumulativeLayoutShift) }}</div>
                    </div>
                </div>
            </div>

            <div class="sitespeed-history">
                <h4>Analysis History</h4>
                <div class="history-table">
                    <table>
                        <thead>
                            <tr>
                                <th>Date</th>
                                <th>TTFB</th>
                                <th>Fully Loaded</th>
                                <th>LCP</th>
                                <th>Speed Index</th>
                                <th>Transfer Size</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="metric in shop.sitespeed" :key="metric.id">
                                <td>{{ formatDateTime(metric.createdAt) }}</td>
                                <td>{{ metric.ttfb || 'N/A' }}<span v-if="metric.ttfb">ms</span></td>
                                <td>{{ formatTime(metric.fullyLoaded) }}</td>
                                <td>{{ formatTime(metric.largestContentfulPaint) }}</td>
                                <td>{{ formatTime(metric.speedIndex) }}</td>
                                <td>{{ formatBytes(metric.transferSize) }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <element-empty
            v-else
            title="No sitespeed data available"
            description="Run your first analysis to see performance metrics"
        >
            <template #actions>
                <button
                    class="btn btn-primary"
                    :disabled="isAnalyzing"
                    @click="runAnalysis"
                >
                    <icon-fa6-solid:gauge :class="{ 'animate-spin': isAnalyzing }" />
                    {{ isAnalyzing ? 'Analyzing...' : 'Run First Analysis' }}
                </button>
            </template>
        </element-empty>
    </div>
</template>

<script setup lang="ts">
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import { useAlert } from '@/composables/useAlert';
import { formatDateTime } from '@/helpers/formatter';
import { trpcClient } from '@/helpers/trpc';
import { computed, ref } from 'vue';

interface Shop {
    id: number;
    sitespeed?: Array<{
        id: number;
        ttfb: number | null;
        fullyLoaded: number | null;
        largestContentfulPaint: number | null;
        firstContentfulPaint: number | null;
        cumulativeLayoutShift: number | null;
        speedIndex: number | null;
        transferSize: number | null;
        createdAt: Date;
    }>;
}

interface Props {
    shop: Shop;
}

const props = defineProps<Props>();

const isAnalyzing = ref(false);
const { showAlert } = useAlert();

const latestMetrics = computed(() => {
    if (!props.shop.sitespeed || props.shop.sitespeed.length === 0) {
        return {};
    }
    return props.shop.sitespeed[0];
});

const formatTime = (timeMs: number | null) => {
    if (!timeMs) return 'N/A';
    if (timeMs < 1000) return `${timeMs}ms`;
    return `${(timeMs / 1000).toFixed(2)}s`;
};

const formatBytes = (bytes: number | null) => {
    if (!bytes) return 'N/A';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / 1024 ** i).toFixed(2)} ${sizes[i]}`;
};

const formatCLS = (cls: number | null) => {
    if (cls === null || cls === undefined) return 'N/A';
    // Convert from stored integer (multiplied by 1000) back to decimal
    const clsValue = cls / 1000;
    return clsValue.toFixed(3);
};

const runAnalysis = async () => {
    isAnalyzing.value = true;
    try {
        await trpcClient.organization.shop.runSitespeedAnalysis.mutate({
            shopId: props.shop.id,
        });
        showAlert('Sitespeed analysis completed successfully', 'success');

        // Refresh the page to show new data
        window.location.reload();
    } catch (error) {
        console.error('Failed to run sitespeed analysis:', error);
        showAlert(
            'Failed to run sitespeed analysis. Please try again.',
            'error',
        );
    } finally {
        isAnalyzing.value = false;
    }
};
</script>

<style scoped>
.sitespeed-tab {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.sitespeed-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.sitespeed-header h3 {
    margin: 0;
}

.sitespeed-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-top: 1rem;
}

.metric-card {
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 0.5rem;
    padding: 1rem;
    text-align: center;
}

.metric-label {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
    margin-bottom: 0.5rem;
}

.metric-value {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--color-text-primary);
}

.history-table {
    margin-top: 1rem;
    overflow-x: auto;
}

.history-table table {
    width: 100%;
    border-collapse: collapse;
}

.history-table th,
.history-table td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--color-border);
}

.history-table th {
    background: var(--color-surface);
    font-weight: 600;
    color: var(--color-text-secondary);
}

.history-table tr:hover {
    background: var(--color-surface);
}
</style>