<template>
    <div class="sitespeed-tab">
        <div class="sitespeed-header">
            <h3>Sitespeed Metrics</h3>
            <div class="sitespeed-actions">
                <button
                    class="btn btn-secondary"
                    @click="showSettings = !showSettings"
                >
                    <icon-fa6-solid:gear />
                    Settings
                </button>
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
                    v-if="shop.sitespeedEnabled"
                    class="btn btn-primary"
                    :disabled="isAnalyzing"
                    @click="runAnalysis"
                >
                    <icon-fa6-solid:gauge :class="{ 'animate-spin': isAnalyzing }" />
                    {{ isAnalyzing ? 'Analyzing...' : 'Run Analysis' }}
                </button>
            </div>
        </div>

        <div v-if="showSettings" class="sitespeed-settings panel">
            <h4>Sitespeed Configuration</h4>
            <form @submit.prevent="saveSettings">
                <div class="form-group">
                    <label class="checkbox-label">
                        <input
                            v-model="settings.enabled"
                            type="checkbox"
                        />
                        Enable automatic daily Sitespeed analysis
                    </label>
                    <p class="form-help-text">When enabled, Sitespeed will run once daily at 3 AM automatically</p>
                </div>

                <div v-if="settings.enabled" class="form-group">
                    <label>URLs to Analyze (max 5)</label>
                    <div class="url-list">
                        <div
                            v-for="(url, index) in settings.urls"
                            :key="index"
                            class="url-item"
                        >
                            <input
                                v-model="url.label"
                                type="text"
                                placeholder="Label (e.g., Homepage)"
                                class="field"
                                maxlength="50"
                                required
                            />
                            <input
                                v-model="url.url"
                                type="url"
                                placeholder="https://example.com"
                                class="field"
                                required
                            />
                            <button
                                type="button"
                                class="btn btn-secondary icon-only"
                                :disabled="settings.urls.length === 1"
                                @click="removeUrl(index)"
                            >
                                <icon-fa6-solid:trash />
                            </button>
                        </div>
                        <button
                            v-if="settings.urls.length < 5"
                            type="button"
                            class="btn btn-secondary"
                            @click="addUrl"
                        >
                            <icon-fa6-solid:plus />
                            Add URL
                        </button>
                    </div>
                    <p class="form-help-text">If no URLs are specified, only the shop's main URL will be analyzed</p>
                </div>

                <div class="form-actions">
                    <button
                        type="submit"
                        class="btn btn-primary"
                        :disabled="isSaving"
                    >
                        {{ isSaving ? 'Saving...' : 'Save Settings' }}
                    </button>
                    <button
                        type="button"
                        class="btn btn-secondary"
                        @click="cancelSettings"
                    >
                        Cancel
                    </button>
                </div>
            </form>
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
                                <th>Page</th>
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
                                <td>{{ metric.label || 'Homepage' }}</td>
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
            :description="shop.sitespeedEnabled ? 'Run your first analysis to see performance metrics' : 'Enable Sitespeed to start monitoring performance'"
        >
            <template #actions>
                <button
                    v-if="!shop.sitespeedEnabled"
                    class="btn btn-primary"
                    @click="showSettings = true"
                >
                    <icon-fa6-solid:gear />
                    Configure Sitespeed
                </button>
                <button
                    v-else
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
    sitespeedEnabled?: boolean;
    sitespeedUrls?: Array<{ url: string; label: string }>;
    sitespeed?: Array<{
        id: number;
        url: string;
        label: string;
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
const isSaving = ref(false);
const showSettings = ref(false);
const alert = useAlert();

const settings = ref({
    enabled: props.shop.sitespeedEnabled ?? false,
    urls: props.shop.sitespeedUrls?.length
        ? [...props.shop.sitespeedUrls]
        : [{ url: '', label: '' }],
});

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
        alert.success('Sitespeed analysis completed successfully');

        // Refresh the page to show new data
        window.location.reload();
    } catch (error) {
        alert.error(
            `Failed to run sitespeed analysis: ${error.message ?? 'Unknown error'}`,
        );
    } finally {
        isAnalyzing.value = false;
    }
};

const addUrl = () => {
    settings.value.urls.push({ url: '', label: '' });
};

const removeUrl = (index: number) => {
    settings.value.urls.splice(index, 1);
};

const saveSettings = async () => {
    isSaving.value = true;
    try {
        // Filter out empty URLs
        const validUrls = settings.value.urls.filter((u) => u.url && u.label);

        await trpcClient.organization.shop.updateSitespeedSettings.mutate({
            shopId: props.shop.id,
            enabled: settings.value.enabled,
            urls: validUrls.length > 0 ? validUrls : undefined,
        });

        alert.success('Sitespeed settings saved successfully');
        showSettings.value = false;

        // Refresh the page to show updated settings
        window.location.reload();
    } catch (error) {
        alert.error(
            `Failed to save sitespeed settings: ${error.message ?? 'Unknown error'}`,
        );
    } finally {
        isSaving.value = false;
    }
};

const cancelSettings = () => {
    // Reset to original values
    settings.value = {
        enabled: props.shop.sitespeedEnabled ?? false,
        urls: props.shop.sitespeedUrls?.length
            ? [...props.shop.sitespeedUrls]
            : [{ url: '', label: '' }],
    };
    showSettings.value = false;
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

.sitespeed-settings {
    margin-top: 1rem;
    padding: 1.5rem;
}

.sitespeed-settings h4 {
    margin-top: 0;
    margin-bottom: 1.5rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-weight: 500;
}

.checkbox-label input[type="checkbox"] {
    margin: 0;
}

.form-help-text {
    margin-top: 0.25rem;
    font-size: 0.875rem;
    color: var(--text-color-muted);
}

.url-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.url-item {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.url-item .field {
    flex: 1;
}

.form-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 2rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--color-border);
}
</style>
