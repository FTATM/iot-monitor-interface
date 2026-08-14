<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New PieChart' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-[150px] relative mt-2 flex flex-col justify-center">
      
      <!-- 1. No Devices Selected State -->
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
      </div>

      <!-- 2. Loading State -->
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
import { computed, ref, onMounted, nextTick } from 'vue';
import VChart from 'vue-echarts';

// ⚡ Import the shared store
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { PieChart } from 'echarts/charts';
import { TooltipComponent, LegendComponent } from 'echarts/components';

use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
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

  const showLegend = customData.showLegend !== undefined ? customData.showLegend : true;
  const innerRadius = customData.innerRadius || '40%';
  const outerRadius = customData.outerRadius || '70%';
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 5;
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  
  let dynamicData = [];
  rawDeviceIds.forEach((id, index) => {
    // ⚡ Read directly from the central store!
    const dataObj = liveStreamStore.liveData[id];
    
    dynamicData.push({
      name: dataObj ? dataObj.name : `Device Loading...`,
      value: dataObj ? dataObj.value : 0,
      itemStyle: {
        color: deviceColorsMap[id] || fallbackColors[index % fallbackColors.length]
      }
    });
  });

  return {
    tooltip: {
      trigger: 'item'
    },
    legend: {
      show: showLegend,
      bottom: '0%',
      left: 'center'
    },
    series: [
      {
        name: 'Data',
        type: 'pie',
        radius: [innerRadius, outerRadius],
        itemStyle: {
          borderRadius: borderRadius,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: { show: false, position: 'center' },
        emphasis: {
          label: { show: true, fontSize: 16, fontWeight: 'bold' }
        },
        labelLine: { show: false },
        data: dynamicData
      }
    ]
  };
});
</script>