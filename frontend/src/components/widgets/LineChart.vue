<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden"
    :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide"
          :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel || $t('lineChart.newChart') }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">

      <div v-if="!hasDevices"
        class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
      </div>

      <div v-else-if="isLoadingHistory"
        class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.fetchingHistory') }}
      </div>

      <div v-else-if="!hasData"
        class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3"
        :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
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
import { LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';
import { useFetch } from '@/composables/useFetch';
import { useFormatter } from '@/composables/useFormatter';
const { formatTime } = useFormatter();

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  
  if (!colorObj.useGradient) {
    return { backgroundColor: c1 };
  }

  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  
  return {
    background: `linear-gradient(${angle}, ${c1}, ${c2})`
  };
});

const isReady = ref(false);
const isLoadingHistory = ref(false);
const liveStreamStore = useLiveStreamStore();
const { data: historyData, error : historyError, execute: fetchHistoryApi } = useFetch();

const deviceSeries = ref({});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => Object.keys(deviceSeries.value).length > 0);

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    historyRange: customData.historyRange || '1h',
    customFrom: customData.customFrom || '', 
    customTo: customData.customTo || '',     
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
  const apiUrl = `/device/charthistory?deviceIds=${idQuery}&from=${fromQuery}&to=${toQuery}&maxPoints=${config.value.maxPoints}`;

  await fetchHistoryApi(apiUrl);

  if (!historyError.value && historyData.value) {
    const newSeries = {};
    Object.entries(historyData.value.data).forEach(([id, pointsArr]) => {
      newSeries[id] = { name: `${t('common.device')} ${id}`, data: pointsArr };
    });
    deviceSeries.value = newSeries;
  } else {
    console.error(historyError.value?.message || t('lineChart.messages.fetchHistoryFailed'));
  }

  isLoadingHistory.value = false;
};

onMounted(async () => {
  await initializeHistory();
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

watch(() => liveStreamStore.liveData, (newData) => {
  const rawDeviceIds = props.widgetData?.deviceIds || [];
  if (rawDeviceIds.length === 0) return;

  const hasIncomingData = rawDeviceIds.some(id => newData[String(id)] !== undefined);
  if (!hasIncomingData) return;

  const newSeries = { ...deviceSeries.value };
  const timestamp = Date.now();

  rawDeviceIds.forEach(rawId => {
    const id = String(rawId);
    if (!newSeries[id]) { newSeries[id] = { name: newData[id]?.name || t('common.loading'), data: [] }; }
    if (newData[id]?.name) newSeries[id].name = newData[id].name;
    newSeries[id].data.push([timestamp, newData[id] ? newData[id].value : null]);

    if (newSeries[id].data.length > config.value.maxPoints) {
      const overflow = newSeries[id].data.length - config.value.maxPoints;
      newSeries[id].data.splice(0, overflow);
    }
  });

  deviceSeries.value = newSeries;
}, { deep: true });

watch(() => props.widgetData?.deviceIds, (newIds, oldIds) => {
  if (JSON.stringify(newIds) === JSON.stringify(oldIds)) return;
  deviceSeries.value = {};
  initializeHistory();
}, { deep: false });

watch(
  () => [config.value.historyRange, config.value.maxPoints, config.value.customFrom, config.value.customTo], 
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
  
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';

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
      stack: config.value.isStacked ? t('lineChart.total') : null,
      itemStyle: { color: color },
      areaStyle: config.value.showArea ? { color: color, opacity: 0.2 } : null,
      data: dataObj.data
    });
  });

  return {
    textStyle: {
      color: chartTextColor
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        if (!params.length) return '';
        const timeStr = formatTime(params[0].value[0]);

        let tipHtml = `<strong>${timeStr}</strong><br/>`;
        params.forEach(p => {
          const val = p.value[1] !== undefined && p.value[1] !== null ? p.value[1].toFixed(2) : '-';
          tipHtml += `${p.marker} <span style="color:${chartTextColor}">${p.seriesName}: <b>${val}</b></span><br/>`;
        });
        return tipHtml;
      }
    },
    legend: { 
      data: legendNames, 
      bottom: 0,
      textStyle: { color: chartTextColor }
    },
    grid: { left: '3%', right: '4%', bottom: '12%', top: '15%', containLabel: true },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLabel: {
        color: chartTextColor,
        formatter: (value) => {
          const fullDateTime = formatTime(value);
          return fullDateTime.replace(' ', '\n');
        }
      }
    },
    yAxis: { 
      type: 'value', 
      name: config.value.yAxisName,
      nameTextStyle: { color: chartTextColor },
      axisLabel: { color: chartTextColor }
    },
    series: dynamicSeries
  };
});
</script>