<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden"
    :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">

    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New LineChart' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices"
        class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
      </div>

      <div v-else-if="isLoadingHistory"
        class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Fetching historical data...
      </div>

      <div v-else-if="!hasData"
        class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
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
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const liveStreamStore = useLiveStreamStore();
const { data: historyData, error : historyError, execute: fetchHistoryApi } = useFetch();

// Store data points as coordinate arrays: { "1": { name: "Device 1", data: [[timestamp, value], [timestamp, value]] } }
const deviceSeries = ref({});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => Object.keys(deviceSeries.value).length > 0);

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    historyRange: customData.historyRange || '1h',
    customFrom: customData.customFrom || '', // ⚡ Capture From
    customTo: customData.customTo || '',     // ⚡ Capture To
    maxPoints: customData.maxPoints || 100,
    isSmooth: customData.isSmooth !== undefined ? customData.isSmooth : true,
    showArea: customData.showArea !== undefined ? customData.showArea : true,
    isStacked: customData.isStacked !== undefined ? customData.isStacked : false,
    use24HourFormat: customData.use24HourFormat !== undefined ? customData.use24HourFormat : true,
    yAxisName: customData.yAxisName || '',
    deviceColorsMap: customData.deviceColors || {}
  };
});

const initializeHistory = async () => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0 || config.value.historyRange === '0') return;

  // Prevent fetching if custom is selected but dates are missing
  if (config.value.historyRange === 'custom' && (!config.value.customFrom || !config.value.customTo)) {
    return;
  }

  isLoadingHistory.value = true;
  const idQuery = rawDeviceIds.join(',');

  let utcFrom, utcTo;

  // 1. Determine the timestamps based on the dropdown selection
  if (config.value.historyRange === 'custom') {
    // Use the explicit dates from the date pickers
    utcFrom = new Date(config.value.customFrom).toISOString();
    utcTo = new Date(config.value.customTo).toISOString();
  } else {
    // Dynamically calculate the relative time from right "now"
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

  // 2. Encode them safely for the URL
  const fromQuery = encodeURIComponent(utcFrom);
  const toQuery = encodeURIComponent(utcTo);

  // ⚡ 3. Construct the API URL to perfectly match your Go handler
  // (Notice I updated this to use 'deviceId=' instead of 'deviceIds=' to match your Go code!)
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

    // Push the native time coordinate
    newSeries[id].data.push([timestamp, newData[id] ? newData[id].value : null]);

    // Trim history based on max configuration
    if (newSeries[id].data.length > config.value.maxPoints) {
      // Calculate how many to remove, safely handling large bulk history fetches
      const overflow = newSeries[id].data.length - config.value.maxPoints;
      newSeries[id].data.splice(0, overflow);
    }
  });

  deviceSeries.value = newSeries;
}, { deep: true });


// Clear arrays if devices are changed in config
watch(() => props.widgetData?.deviceIds, (newIds, oldIds) => {
  if (JSON.stringify(newIds) === JSON.stringify(oldIds)) return;
  deviceSeries.value = {};
  initializeHistory();
}, { deep: false });

// ⚡ Watch for changes in the history configuration and re-fetch automatically!
watch(
  () => [
    config.value.historyRange, 
    config.value.maxPoints, 
    config.value.customFrom, 
    config.value.customTo
  ], 
  (newVals, oldVals) => {
    // Prevent re-fetching if the values didn't actually change
    if (JSON.stringify(newVals) === JSON.stringify(oldVals)) return;
    
    // If they switched to 'custom' but haven't picked dates yet, do nothing
    if (newVals[0] === 'custom' && (!newVals[2] || !newVals[3])) return;

    // Clear the old lines off the chart and fetch the new time range
    deviceSeries.value = {}; 
    initializeHistory();
  }, 
  { deep: true }
);


const chartOption = computed(() => {
  const fallbackColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];
  const rawDeviceIds = props.widgetData?.deviceIds || [];

  const legendNames = [];
  const dynamicSeries = [];

  rawDeviceIds.forEach((id, index) => {
    const dataObj = deviceSeries.value[id];
    if (!dataObj) return;

    const color = config.value.deviceColorsMap[id] || fallbackColors[index % fallbackColors.length];
    legendNames.push(dataObj.name);

    dynamicSeries.push({
      name: dataObj.name,
      type: 'line',
      smooth: config.value.isSmooth,
      stack: config.value.isStacked ? 'Total' : null,
      itemStyle: { color: color },
      areaStyle: config.value.showArea ? { color: color, opacity: 0.2 } : null,
      data: dataObj.data
    });
  });

  return {
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        if (!params.length) return '';

        // Extract time from the first point
        const timeStr = new Date(params[0].value[0]).toLocaleTimeString(undefined, {
          hour12: !config.value.use24HourFormat,
          hour: '2-digit', minute: '2-digit', second: '2-digit'
        });

        let tipHtml = `<strong>${timeStr}</strong><br/>`;
        params.forEach(p => {
          const val = p.value[1] !== undefined && p.value[1] !== null ? p.value[1].toFixed(2) : '-';
          tipHtml += `${p.marker} ${p.seriesName}: <b>${val}</b><br/>`;
        });
        return tipHtml;
      }
    },
    legend: { data: legendNames, bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '12%', top: '15%', containLabel: true },

    // ⚡ Switched to 'time' axis to support native timestamp arrays
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLabel: {
        formatter: (value) => {
          return new Date(value).toLocaleTimeString(undefined, {
            hour12: !config.value.use24HourFormat,
            hour: '2-digit', minute: '2-digit', second: '2-digit'
          });
        }
      }
    },
    yAxis: { type: 'value', name: config.value.yAxisName },
    series: dynamicSeries
  };
});
</script>