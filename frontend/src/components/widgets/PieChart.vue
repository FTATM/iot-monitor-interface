<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New PieChart' }}
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
import { PieChart } from 'echarts/charts';
import { TooltipComponent, LegendComponent } from 'echarts/components';

// Register the required ECharts components for a pie chart
use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent]);

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

  // Extract custom values or use Doughnut fallbacks
  const showLegend = customData.showLegend !== undefined ? customData.showLegend : true;
  const innerRadius = customData.innerRadius || '40%';
  const outerRadius = customData.outerRadius || '70%';
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 5;

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
        // The first value is the inner hole, the second is the outer edge
        radius: [innerRadius, outerRadius],
        itemStyle: {
          borderRadius: borderRadius,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        // Hardcoded dummy data for now
        data: [
          { value: 1048, name: 'System' },
          { value: 735, name: 'Database' },
          { value: 580, name: 'Cache' },
          { value: 484, name: 'Logs' },
          { value: 300, name: 'Other' }
        ]
      }
    ]
  };
});
</script>