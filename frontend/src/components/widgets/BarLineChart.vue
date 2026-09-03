<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden rounded-box" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
        :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel || $t('barLineChart.newChart') }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="isLoadingHistory" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.fetchingHistory') }}
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />

    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart, LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useFetch } from '@/composables/useFetch';

use([CanvasRenderer, BarChart, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const { t, locale } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const aggregatedCategories = ref([]);
const aggregatedValues = ref([]);

const { data: historyData, error: historyError, execute: fetchHistoryApi } = useFetch();

const chartConfig = computed(() => props.widgetData?.customChartData || {});
const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1;
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const initializeAndBucketHistory = async () => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;
  const deviceId = rawDeviceIds[0];

  isLoadingHistory.value = true;

  const range = chartConfig.value.historyRange || '7d';
  const now = new Date();
  const past = new Date();

  switch (range) {
    case '24h': past.setHours(now.getHours() - 24); break;
    case '7d': past.setDate(now.getDate() - 7); break;
    case '30d': past.setDate(now.getDate() - 30); break;
  }

  const utcFrom = encodeURIComponent(past.toISOString());
  const utcTo = encodeURIComponent(now.toISOString());

  await fetchHistoryApi(`/device/charthistory?deviceIds=${deviceId}&from=${utcFrom}&to=${utcTo}&maxPoints=5000`);

  if (!historyError.value && historyData.value && historyData.value.data[deviceId]) {
    const rawPoints = historyData.value.data[deviceId];

    const interval = chartConfig.value.bucketInterval || 'day';
    const aggregation = chartConfig.value.aggregationMode || 'sum';
    const buckets = {};

    const currentLocale = locale.value === 'th' ? 'th-TH' : 'en-GB';

    rawPoints.forEach(p => {
      if (p[1] === null || p[1] === undefined) return;
      const d = new Date(p[0]);
      let key = '';

      if (interval === 'hour') {
        key = d.toLocaleString(currentLocale, { calendar: 'gregory', month: 'short', day: 'numeric', hour: '2-digit', hour12: false }) + ':00';
      } else if (interval === 'day') {
        key = d.toLocaleString(currentLocale, { calendar: 'gregory', month: 'short', day: 'numeric' });
      } else if (interval === 'month') {
        key = d.toLocaleString(currentLocale, { calendar: 'gregory', month: 'short', year: 'numeric' });
      }

      if (!buckets[key]) buckets[key] = [];
      buckets[key].push(Number(p[1]));
    });

    const categories = Object.keys(buckets);
    const values = categories.map(key => {
      const arr = buckets[key];
      if (aggregation === 'sum') return arr.reduce((a, b) => a + b, 0);
      if (aggregation === 'max') return Math.max(...arr);
      if (aggregation === 'min') return Math.min(...arr);
      return arr.reduce((a, b) => a + b, 0) / arr.length;
    });

    aggregatedCategories.value = categories;
    aggregatedValues.value = values;
  }

  isLoadingHistory.value = false;
};

onMounted(async () => {
  await initializeAndBucketHistory();
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

watch(
  () => [
    props.widgetData?.deviceIds,
    chartConfig.value.historyRange,
    chartConfig.value.bucketInterval,
    chartConfig.value.aggregationMode
  ],
  (newVals, oldVals) => {
    if (JSON.stringify(newVals) === JSON.stringify(oldVals)) return;
    initializeAndBucketHistory();
  },
  { deep: true }
);

const chartOption = computed(() => {
  const textColor = props.widgetData.widgetStyle?.textHex || '#334155';
  const barColor = chartConfig.value.barColor || '#3b82f6';
  const lineColor = chartConfig.value.lineColor || '#f97316';

  const barName = chartConfig.value.barName || 'Bar Value';
  const lineName = chartConfig.value.lineName || 'Trend Line';
  const yAxisName = chartConfig.value.yAxisName || '';

  return {
    textStyle: { color: textColor },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      valueFormatter: (value) => {
        if (value !== undefined && value !== null) {
          return Number(value).toFixed(2);
        }
        return '-';
      }
    },
    legend: {
      data: [barName, lineName],
      bottom: 0,
      textStyle: { color: textColor }
    },
    grid: { left: '3%', right: '4%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'category',
      data: aggregatedCategories.value,
      axisPointer: { type: 'shadow' },
      axisLabel: {
        color: textColor,
        hideOverlap: true,
        formatter: (value) => {
          // ⚡ NEW: Split the category string onto two lines
          // For example, "Sep 3" becomes "Sep" over "3", or "Sep 3, 14:00" splits nicely.
          return value.replace(' ', '\n');
        }
      }
    },
    yAxis: {
      type: 'value',
      name: yAxisName,
      nameTextStyle: { color: textColor },
      axisLabel: { color: textColor },
      splitLine: { lineStyle: { color: textColor, opacity: 0.1 } }
    },
    series: [
      {
        name: barName,
        type: 'bar',
        barWidth: '40%',
        itemStyle: { color: barColor, borderRadius: [4, 4, 0, 0] },
        data: aggregatedValues.value
      },
      {
        name: lineName,
        type: 'line',
        smooth: chartConfig.value.isSmooth ?? true,
        itemStyle: { color: lineColor },
        lineStyle: { width: 3, color: lineColor },
        symbolSize: 8,
        data: aggregatedValues.value
      }
    ]
  };
});
</script>