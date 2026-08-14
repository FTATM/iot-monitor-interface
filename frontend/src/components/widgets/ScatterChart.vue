<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">

    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New ScatterChart' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
      </div>

      <div v-else-if="isLoadingHistory" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Fetching historical data...
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />

    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick, watch } from 'vue';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { ScatterChart, LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';

use([CanvasRenderer, ScatterChart, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const liveStreamStore = useLiveStreamStore();
const { data: historyData, error : historyError, execute: fetchHistoryApi } = useFetch();

const deviceSeries = ref({}); 

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => Object.keys(deviceSeries.value).length > 0 );

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    historyRange: customData.historyRange || '1h',
    customFrom: customData.customFrom || '', 
    customTo: customData.customTo || '',     
    maxPoints: customData.maxPoints || 100,
    showRegression: customData.showRegression !== undefined ? customData.showRegression : true,
    use24HourFormat: customData.use24HourFormat !== undefined ? customData.use24HourFormat : true,
    xAxisName: customData.xAxisName || 'Time',
    yAxisName: customData.yAxisName || '',
    deviceColorsMap: customData.deviceColors || {}
  };
});

const calculateRegressionLine = (data) => {
  const n = data.length;
  if (n < 2) return [];

  const minX = data[0][0];
  const maxX = data[n - 1][0];

  let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0;

  data.forEach(point => {
    // Normalizing X protects against JavaScript precision loss on large Unix timestamps
    const normX = point[0] - minX;

    sumX += normX;
    sumY += point[1];
    sumXY += normX * point[1];
    sumXX += normX * normX;
  });

  const denominator = (n * sumXX - sumX * sumX);
  if (denominator === 0) return [];

  const slope = (n * sumXY - sumX * sumY) / denominator;
  const intercept = (sumY - slope * sumX) / n;

  return [
    [minX, intercept],
    [maxX, slope * (maxX - minX) + intercept]
  ];
};

const initializeHistory = async () => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0 || config.value.historyRange === '0') return;

  if (config.value.historyRange === 'custom' && (!config.value.customFrom || !config.value.customTo)) {
    return;
  }

  isLoadingHistory.value = true;
  const idQuery = rawDeviceIds.join(',');

  let utcFrom, utcTo;

  if (config.value.historyRange === 'custom') {
    utcFrom = new Date(config.value.customFrom).toISOString();
    utcTo = new Date(config.value.customTo).toISOString();
  } else {
    const now = new Date();
    const past = new Date();

    switch (config.value.historyRange) {
      case '15m': past.setMinutes(now.getMinutes() - 15); break;
      case '30m': past.setMinutes(now.getMinutes() - 30); break;
      case '1h': past.setHours(now.getHours() - 1); break;
      case '3h': past.setHours(now.getHours() - 3); break;
      case '6h': past.setHours(now.getHours() - 6); break;
      case '24h': past.setHours(now.getHours() - 24); break;
      case '7d': past.setDate(now.getDate() - 7); break;
    }

    utcFrom = past.toISOString();
    utcTo = now.toISOString();
  }

  const fromQuery = encodeURIComponent(utcFrom);
  const toQuery = encodeURIComponent(utcTo);
  
  // Re-uses your exact same smart backend logic!
  const apiUrl = `/device/charthistory?deviceIds=${idQuery}&from=${fromQuery}&to=${toQuery}&maxPoints=${config.value.maxPoints}`;

  await fetchHistoryApi(apiUrl);

  if (!historyError.value && historyData.value) {
    const newSeries = {};

    Object.entries(historyData.value.data).forEach(([id, pointsArr]) => {
      newSeries[id] = {
        name: `Device ${id}`,
        data: pointsArr
      };
    });

    deviceSeries.value = newSeries;
  } else {
    console.error(historyError.value?.message || "Failed to fetch history chart");
  }

  isLoadingHistory.value = false;
};

onMounted(async () => {
  await initializeHistory();
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

// Listen for live updates and push them to the end of the history array
watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  const hasIncomingData = rawDeviceIds.some(id => newData[String(id)] !== undefined);
  if (!hasIncomingData) return;

  const newSeries = { ...deviceSeries.value };
  const timestamp = Date.now();

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    
    if (!newSeries[id]) {
      newSeries[id] = { name: newData[id]?.name || `Loading...`, data: [] };
    }
    
    if (newData[id]?.name) newSeries[id].name = newData[id].name;
    
    newSeries[id].data.push([timestamp, newData[id] ? newData[id].value : null]);

    // Trim history based on the user's dynamic config instead of hardcoded 20
    if (newSeries[id].data.length > config.value.maxPoints) {
      const overflow = newSeries[id].data.length - config.value.maxPoints;
      newSeries[id].data.splice(0, overflow);
    }
  });

  deviceSeries.value = newSeries;
}, { deep: true });

// Clear arrays if devices are changed
watch(() => props.widgetData?.deviceIds, (newIds, oldIds) => {
  if (JSON.stringify(newIds) === JSON.stringify(oldIds)) return; 
  deviceSeries.value = {}; 
  initializeHistory();
}, { deep: false });

// Watch for changes in history configuration
watch(
  () => [
    config.value.historyRange, 
    config.value.maxPoints, 
    config.value.customFrom, 
    config.value.customTo
  ], 
  (newVals, oldVals) => {
    if (JSON.stringify(newVals) === JSON.stringify(oldVals)) return;
    if (newVals[0] === 'custom' && (!newVals[2] || !newVals[3])) return;

    deviceSeries.value = {}; 
    initializeHistory();
  }, 
  { deep: true }
);

const chartOption = computed(() => {
  const fallbackColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];
  const rawDeviceIds = props.widgetData?.deviceIds || [];

  const dynamicSeries = [];
  const legendNames = [];

  rawDeviceIds.forEach((id, index) => {
    const dataObj = deviceSeries.value[id];
    if (!dataObj) return;

    const color = config.value.deviceColorsMap[id] || fallbackColors[index % fallbackColors.length];
    const actualName = dataObj.name;
    const rawData = dataObj.data;

    legendNames.push(actualName);

    dynamicSeries.push({
      name: actualName,
      type: 'scatter',
      itemStyle: { color: color, opacity: 0.8 },
      symbolSize: 12,
      data: rawData
    });

    if (config.value.showRegression) {
      const trendName = `${actualName} (Trend)`;
      legendNames.push(trendName);

      dynamicSeries.push({
        name: trendName,
        type: 'line',
        showSymbol: false,
        smooth: true,
        itemStyle: { color: color },
        lineStyle: { width: 2, type: 'dashed' },
        data: calculateRegressionLine(rawData)
      });
    }
  });

  return {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        const timeStr = new Date(params.value[0]).toLocaleTimeString(undefined, {
          hour12: !config.value.use24HourFormat,
          hour: '2-digit', minute: '2-digit', second: '2-digit'
        });
        const yValue = params.value[1] !== undefined && params.value[1] !== null ? params.value[1].toFixed(2) : '-';
        return `<strong>${params.seriesName}</strong><br/>Time: ${timeStr}<br/>Value: ${yValue}`;
      }
    },
    legend: { show: true, bottom: 0, data: legendNames },
    grid: { left: '3%', right: '8%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'time',
      name: config.value.xAxisName,
      nameTextStyle: { fontWeight: 'bold' },
      axisLabel: {
        formatter: (value) => {
          return new Date(value).toLocaleTimeString(undefined, {
            hour12: !config.value.use24HourFormat,
            hour: '2-digit', minute: '2-digit', second: '2-digit'
          });
        }
      }
    },
    yAxis: {
      type: 'value',
      name: config.value.yAxisName,
      nameTextStyle: { fontWeight: 'bold' },
      scale: true
    },
    series: dynamicSeries
  };
});
</script>