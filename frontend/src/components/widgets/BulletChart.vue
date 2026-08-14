<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Solid White Header Box with Device Name Badge -->
    <div class="bg-white px-4 py-3 border-b border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide truncate pr-2">
        {{ widgetData?.widgetLabel || 'New BulletChart' }}
      </h3>
      <span v-if="hasData && liveDeviceName" class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>
    
    <!-- Padded Chart Area -->
    <div class="flex-1 w-full min-h-[150px] relative flex flex-col justify-center">
      
      <!-- 1. No Devices Selected State -->
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No device selected. Open configuration to add a data source.
      </div>

      <!-- 2. Loading State (Devices selected, waiting for SSE ping) -->
      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <!-- 3. Actual Chart -->
      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />

    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick  } from 'vue';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart, ScatterChart } from 'echarts/charts';
import { 
  GridComponent, 
  TooltipComponent, 
  LegendComponent 
} from 'echarts/components';

use([CanvasRenderer, BarChart, ScatterChart, GridComponent, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();

// ⚡ UI State helper
const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);
const hasData = computed(() => deviceId.value && liveStreamStore.liveData[deviceId.value] !== undefined);
const liveValue = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].value : 0);
const liveDeviceName = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].name : '');

onMounted(async () => {
  await nextTick();
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  
  const barColor = customData.barColor || '#1A5FB4';
  const targetColor = customData.targetColor || '#26A269';
  const target = customData.targetValue !== undefined ? customData.targetValue : 85;
  const unit = customData.unit || '';
  const xAxisMax = customData.xAxisMax || null; 

  const thresholds = customData.thresholds || [
    { name: 'Excellent', value: 120, color: '#e2e8f0' },
    { name: 'Poor', value: 60, color: '#94a3b8' }
  ];

  const sortedThresholds = [...thresholds].sort((a, b) => b.value - a.value);

  const dynamicSeries = sortedThresholds.map((zone, index) => {
    return {
      name: zone.name,
      type: 'bar',
      barWidth: '50%',
      barGap: index === 0 ? '0%' : '-100%',
      data: [zone.value],
      itemStyle: { color: zone.color }, 
      animation: false,
      tooltip: { valueFormatter: (value) => `${value}${unit}` }
    };
  });

  // Inject the live SSE data here
  dynamicSeries.push({
    name: 'Actual Value',
    type: 'bar',
    barWidth: '20%',
    barGap: '-100%',
    data: [liveValue.value], 
    itemStyle: { color: barColor }, 
    z: 3,
    tooltip: { valueFormatter: (value) => `${value}${unit}` }
  });

  dynamicSeries.push({
    name: 'Target',
    type: 'scatter',
    symbol: 'rect',
    symbolSize: [4, 40],
    data: [target],
    itemStyle: { color: targetColor }, 
    z: 4,
    tooltip: { valueFormatter: (value) => `${value}${unit}` }
  });

  return {
    tooltip: {
      trigger: 'item',
      axisPointer: { type: 'shadow' }
    },
    legend: {
      bottom: 0,
      icon: 'circle',
      itemWidth: 10,
      textStyle: { color: '#64748b' }
    },
    grid: {
      left: '2%',
      right: '5%',
      bottom: '15%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      max: xAxisMax, 
      splitLine: { show: false },
      axisLabel: { formatter: `{value}${unit}` } 
    },
    yAxis: {
      type: 'category',
      data: ['Metric'],
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { show: false } 
    },
    series: dynamicSeries
  };
});
</script>