<template>
    <Alert v-if="shop && !shop.sitespeedEnabled" type="info">
        Sitespeed monitoring is not enabled for this shop. Please enable it in the shop settings to view performance metrics.
    </Alert>
    <div class="mb-1">
        <a v-if="shop && shop.sitespeed && shop.sitespeed.length > 0" class="btn btn-primary-outline" :href="'/sitespeed/' + shop.id + '/index.html'" target="_blank">
            <i class="fa fa-chart-line"></i> View Sitespeed Report
        </a>
    </div>
    <div v-if="shop && shop.sitespeed && shop.sitespeed.length > 0" class="panel panel-table chart-panel">
      <canvas ref="timeChartCanvas" width="800" height="400"></canvas>
    </div>
    <div v-if="shop && shop.sitespeed && shop.sitespeed.length > 0" class="panel panel-table chart-panel">
      <canvas ref="transferSizeChartCanvas" width="800" height="400"></canvas>
    </div>
    <div v-if="shop && shop.sitespeed && shop.sitespeed.length > 0" class="panel panel-table chart-panel">
      <canvas ref="clsChartCanvas" width="800" height="400"></canvas>
    </div>
    <div class="panel panel-table">
        <data-table
            v-if="shop"
            :columns="[
                { key: 'createdAt', name: 'Checked At' },
                { key: 'ttfb', name: 'TTFB' },
                { key: 'fullyLoaded', name: 'Fully Loaded' },
                { key: 'largestContentfulPaint', name: 'Largest Contentful Paint' },
                { key: 'firstContentfulPaint', name: 'First Contentful Paint' },
                { key: 'cumulativeLayoutShift', name: 'Cumulative Layout Shift' },
                { key: 'transferSize', name: 'Transfer Size' },
            ]"
            :data="shop.sitespeed || []"
            >
            <template #cell-ttfb="{ row }">
                {{ row.ttfb ? `${row.ttfb} ms` : '-' }}
            </template>
            <template #cell-fullyLoaded="{ row }">
                {{ row.fullyLoaded ? `${row.fullyLoaded} ms` : '-' }}
            </template>
            <template #cell-largestContentfulPaint="{ row }">
                {{ row.largestContentfulPaint ? `${row.largestContentfulPaint} ms` : '-' }}
            </template>
            <template #cell-firstContentfulPaint="{ row }">
                {{ row.firstContentfulPaint ? `${row.firstContentfulPaint} ms` : '-' }}
            </template>
            <template #cell-cumulativeLayoutShift="{ row }">
                {{ row.cumulativeLayoutShift ? row.cumulativeLayoutShift : '-' }}
            </template>
            <template #cell-transferSize="{ row }">
                {{ row.transferSize ? formatBytes(row.transferSize) : '-' }}
            </template>
            <template #cell-createdAt="{ row }">
                {{ formatDateTime(row.createdAt) }}
            </template>
        </data-table>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { Chart, registerables } from 'chart.js';
import 'chartjs-adapter-date-fns';
import { formatDateTime } from '@/helpers/formatter';
import {useShopDetail} from "@/composables/useShopDetail";

const {
    shop,
} = useShopDetail();

function formatBytes(bytes: number, decimals = 2): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

Chart.register(...registerables);

const timeChartCanvas = ref<HTMLCanvasElement | null>(null);
const timeChartInstance = ref<Chart | null>(null);
const transferSizeChartCanvas = ref<HTMLCanvasElement | null>(null);
const transferSizeChartInstance = ref<Chart | null>(null);
const clsChartCanvas = ref<HTMLCanvasElement | null>(null);
const clsChartInstance = ref<Chart | null>(null);

const timeMetrics = [
  { key: 'ttfb', label: 'TTFB' },
  { key: 'fullyLoaded', label: 'Fully Loaded' },
  { key: 'largestContentfulPaint', label: 'LCP' },
  { key: 'firstContentfulPaint', label: 'FCP' },
];

const createTimeChart = () => {
  if (!timeChartCanvas.value || !shop.value?.sitespeed) return;
  
  if (timeChartInstance.value) {
    timeChartInstance.value.destroy();
  }

  const ctx = timeChartCanvas.value.getContext('2d');
  if (!ctx) return;

  const sortedData = [...shop.value.sitespeed].sort((a, b) => 
    new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );

  const datasets = timeMetrics.map(metric => ({
    label: metric.label,
    data: sortedData.map(item => ({
      x: new Date(item.createdAt).getTime(),
      y: item[metric.key as keyof typeof item] as number
    })),
  }));

  timeChartInstance.value = new Chart(ctx, {
    type: 'line',
    data: { datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        title: {
          display: true,
          text: 'Performance Metrics Over Time'
        },
        legend: {
          display: true,
          position: 'top'
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              let label = context.dataset.label ?? '';
              if (label) {
                label += ': ';
              }
              if (context.parsed.y !== null) {
                label += context.parsed.y + 'ms';
              }
              return label;
            },
            title: function(tooltipItems) {
              if (tooltipItems.length > 0) {
                const timestamp = tooltipItems[0].parsed.x;
                return new Date(timestamp).toLocaleString();
              }
              return '';
            }
          }
        }
      },
      scales: {
        x: {
          type: 'time',
          time: {
            unit: 'day',
            displayFormats: {
              day: 'MMM dd'
            }
          },
          title: {
            display: true,
            text: 'Date'
          }
        },
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: 'Time (ms)'
          }
        }
      }
    }
  });
};

const createTransferSizeChart = () => {
  if (!transferSizeChartCanvas.value || !shop.value?.sitespeed) return;
  
  if (transferSizeChartInstance.value) {
    transferSizeChartInstance.value.destroy();
  }

  const ctx = transferSizeChartCanvas.value.getContext('2d');
  if (!ctx) return;

  const sortedData = [...shop.value.sitespeed].sort((a, b) => 
    new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );

  const datasets = [{
    label: 'Transfer Size',
    data: sortedData.map(item => ({
      x: new Date(item.createdAt).getTime(),
      y: item.transferSize ? Math.round(item.transferSize / 1024) : 0
    })),
  }];

  transferSizeChartInstance.value = new Chart(ctx, {
    type: 'line',
    data: { datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        title: {
          display: true,
          text: 'Transfer Size Over Time'
        },
        legend: {
          display: true,
          position: 'top'
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              return `Transfer Size: ${context.parsed.y} KB`;
            },
            title: function(tooltipItems) {
              if (tooltipItems.length > 0) {
                const timestamp = tooltipItems[0].parsed.x;
                return new Date(timestamp).toLocaleString();
              }
              return '';
            }
          }
        }
      },
      scales: {
        x: {
          type: 'time',
          time: {
            unit: 'day',
            displayFormats: {
              day: 'MMM dd'
            }
          },
          title: {
            display: true,
            text: 'Date'
          }
        },
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: 'Size (KB)'
          }
        }
      }
    }
  });
};

const createClsChart = () => {
  if (!clsChartCanvas.value || !shop.value?.sitespeed) return;
  
  if (clsChartInstance.value) {
    clsChartInstance.value.destroy();
  }

  const ctx = clsChartCanvas.value.getContext('2d');
  if (!ctx) return;

  const sortedData = [...shop.value.sitespeed].sort((a, b) => 
    new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );

  const datasets = [{
    label: 'Cumulative Layout Shift',
    data: sortedData.map(item => ({
      x: new Date(item.createdAt).getTime(),
      y: item.cumulativeLayoutShift ?? 0
    })),
  }];

  clsChartInstance.value = new Chart(ctx, {
    type: 'line',
    data: { datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        title: {
          display: true,
          text: 'Cumulative Layout Shift Over Time'
        },
        legend: {
          display: true,
          position: 'top'
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              return `CLS: ${context.parsed.y}`;
            },
            title: function(tooltipItems) {
              if (tooltipItems.length > 0) {
                const timestamp = tooltipItems[0].parsed.x;
                return new Date(timestamp).toLocaleString();
              }
              return '';
            }
          }
        }
      },
      scales: {
        x: {
          type: 'time',
          time: {
            unit: 'day',
            displayFormats: {
              day: 'MMM dd'
            }
          },
          title: {
            display: true,
            text: 'Date'
          }
        },
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: 'CLS Score'
          }
        }
      }
    }
  });
};

onMounted(async () => {
  await nextTick();
  if (shop.value?.sitespeed) {
    createTimeChart();
    createTransferSizeChart();
    createClsChart();
  }
});

onUnmounted(() => {
  if (timeChartInstance.value) {
    timeChartInstance.value.destroy();
  }
  if (transferSizeChartInstance.value) {
    transferSizeChartInstance.value.destroy();
  }
  if (clsChartInstance.value) {
    clsChartInstance.value.destroy();
  }
});

watch(() => shop.value?.sitespeed, async () => {
  await nextTick();
  createTimeChart();
  createTransferSizeChart();
  createClsChart();
}, { deep: true });

</script>
