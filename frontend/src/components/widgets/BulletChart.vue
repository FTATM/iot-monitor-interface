<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Solid White Header Box -->
    <div class="bg-white px-4 py-3 border-b border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New BulletChart' }}
      </h3>
    </div>
    
    <!-- Padded Chart Area -->
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

onMounted(async () => {
  // Wait for Vue's DOM updates
  await nextTick();
  // Wait 50ms for the grid-layout to finish applying its CSS dimensions
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  
  const barColor = customData.barColor || '#3b82f6';
  const targetColor = customData.targetColor || '#0f172a';
  const target = customData.targetValue !== undefined ? customData.targetValue : 85;

  // Grab the dynamic array (or fall back to defaults if not set yet)
  const thresholds = customData.thresholds || [
    { name: 'Excellent', value: 120, color: '#e2e8f0' },
    { name: 'Satisfactory', value: 90, color: '#cbd5e1' },
    { name: 'Poor', value: 60, color: '#94a3b8' }
  ];

  // IMPORTANT: Sort thresholds descending! 
  // We want the largest bar (e.g. 120) to render first in the background, 
  // and the smaller bars (e.g. 60) to render on top of it.
  const sortedThresholds = [...thresholds].sort((a, b) => b.value - a.value);

  // Map the sorted zones into ECharts bar series
  const dynamicSeries = sortedThresholds.map((zone, index) => {
    return {
      name: zone.name,
      type: 'bar',
      barWidth: '50%',
      // The first background layer doesn't need a gap shift, everything else overlaps it
      barGap: index === 0 ? '0%' : '-100%',
      data: [zone.value],
      itemStyle: { color: zone.color }, 
      animation: false
    };
  });

  // Push the Actual Performance Bar on top
  dynamicSeries.push({
    name: 'Actual Value',
    type: 'bar',
    barWidth: '20%',
    barGap: '-100%',
    data: [75], // Dummy data - replace with API data later!
    itemStyle: { color: barColor }, 
    z: 3 // Ensures it stays above all threshold backgrounds
  });

  // Push the Target Scatter Marker absolute top
  dynamicSeries.push({
    name: 'Target',
    type: 'scatter',
    symbol: 'rect',
    symbolSize: [4, 40],
    data: [target],
    itemStyle: { color: targetColor }, 
    z: 4
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
      splitLine: { show: false }
    },
    yAxis: {
      type: 'category',
      data: ['Metric'],
      axisLine: { show: false },
      axisTick: { show: false }
    },
    series: dynamicSeries
  };
});
</script>