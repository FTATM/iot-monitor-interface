<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New LineChart' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-[150px] relative">
      <v-chart v-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick } from 'vue';
import VChart from 'vue-echarts';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';

// Added LegendComponent so you can see the labels for multiple lines
use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const isReady = ref(false);

onMounted(async () => {
  await nextTick();
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};

  const lineColor = customData.lineColor || '#3b82f6'; // Blue
  const isSmooth = customData.isSmooth !== undefined ? customData.isSmooth : true;
  const showArea = customData.showArea !== undefined ? customData.showArea : true;
  const isStacked = customData.isStacked !== undefined ? customData.isStacked : false;
  const yAxisName = customData.yAxisName || '';

  // The magic string that groups the lines together
  const stackName = isStacked ? 'Total' : null;

  return {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['Ingress', 'Egress'],
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '12%', // Increased slightly for the legend
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    },
    yAxis: {
      type: 'value',
      name: yAxisName,
      nameTextStyle: {
        color: '#64748b',
        fontWeight: 'bold'
      }
    },
    series: [
      {
        name: 'Ingress',
        type: 'line',
        stack: stackName, // Applies stacking if true
        smooth: isSmooth,
        itemStyle: { color: lineColor },
        areaStyle: showArea ? { color: lineColor, opacity: 0.2 } : null,
        data: [120, 132, 101, 134, 90, 230, 210]
      },
      {
        name: 'Egress',
        type: 'line',
        stack: stackName, // Must match the name above to stack properly
        smooth: isSmooth,
        itemStyle: { color: '#10b981' }, // Hardcoded Emerald for dummy data
        areaStyle: showArea ? { color: '#10b981', opacity: 0.2 } : null,
        data: [220, 182, 191, 234, 290, 330, 310]
      }
    ]
  };
});
</script>