<template>
  <!-- Main wrapper -->
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New GaugeChart' }}
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
import { GaugeChart } from 'echarts/charts';
import { TooltipComponent } from 'echarts/components';

use([CanvasRenderer, GaugeChart, TooltipComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const isReady = ref(false);

onMounted(async () => {
  // Wait for Vue's DOM updates
  await nextTick();
  // Wait 50ms for the grid-layout to finish applying its CSS dimensions
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

// Wrap the ECharts config in a computed property
const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};

  // Extract custom values or use fallbacks
  const min = customData.min !== undefined ? customData.min : 0;
  const max = customData.max !== undefined ? customData.max : 100;
  const progressColor = customData.progressColor || '#3b82f6';
  const pointerColor = customData.pointerColor || '#0f172a';
  const unit = customData.unit || '%';

  return {
    series: [
      {
        type: 'gauge',
        startAngle: 180,
        endAngle: 0,
        center: ['50%', '75%'],
        radius: '80%',
        min: min, 
        max: max, 
        splitNumber: 5,

        progress: {
          show: true,
          width: 18,
          itemStyle: {
            color: progressColor 
          }
        },

        axisLine: {
          lineStyle: {
            width: 18,
            color: [[1, '#e2e8f0']]
          }
        },

        axisTick: { show: false },

        splitLine: {
          show: true,
          length: 5,
          lineStyle: { width: 2, color: '#94a3b8' }
        },

        axisLabel: {
          show: false,
          distance: 10,
          color: '#64748b',
          fontSize: 12
        },

        pointer: {
          icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
          length: '12%',
          width: 10,
          offsetCenter: [0, '-60%'],
          itemStyle: { color: pointerColor } 
        },

        detail: {
          valueAnimation: true,
          formatter: `{value}${unit}`, 
          color: pointerColor, 
          fontSize: 16,
          fontWeight: 'bold',
          offsetCenter: [0, '0%']
        },

        data: [{ value: 20 }]
      }
    ]
  };
});
</script>