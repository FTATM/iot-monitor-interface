<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New BarChart' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
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
// ⚡ Notice how clean this is! No more network logic.
import { computed, ref, onMounted, nextTick } from 'vue';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';

// ⚡ Import the shared store
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);

// ⚡ Initialize the store
const liveStreamStore = useLiveStreamStore();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);

// ⚡ Read directly from the central store instead of a local map
const hasData = computed(() => {
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => liveStreamStore.liveData[id] !== undefined);
});

onMounted(async () => {
  await nextTick();
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  const textColor = customData.textColor || '#334155';
  const yAxisName = customData.yAxisName || '';
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  
  let deviceNames = [];
  let deviceValues = [];

  rawDeviceIds.forEach((id, index) => {
    // ⚡ Read directly from the central store!
    const dataObj = liveStreamStore.liveData[id];
    
    deviceNames.push(dataObj ? dataObj.name : `Loading...`);
    deviceValues.push({
      value: dataObj ? dataObj.value : 0,
      itemStyle: {
        color: deviceColorsMap[id] || fallbackColors[index % fallbackColors.length],
        borderRadius: [4, 4, 0, 0]
      }
    });
  });

  return {
    textStyle: { color: textColor },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '5%', top: '15%', containLabel: true },
    xAxis: {
      type: 'category',
      data: deviceNames,
      axisTick: { alignWithLabel: true },
      axisLabel: { color: textColor } 
    },
    yAxis: {
      type: 'value',
      name: yAxisName,
      axisLabel: { color: textColor } 
    },
    series: [
      {
        type: 'bar',
        barWidth: '60%', 
        data: deviceValues
      }
    ]
  };
});
</script>