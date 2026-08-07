<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    <!-- Read directly from the label field, and apply the dynamic text color -->
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New BarChart' }}
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

// Import only the specific ECharts modules we need for a Bar Chart
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart } from 'echarts/charts';
import {
  GridComponent,
  TooltipComponent,
  LegendComponent
} from 'echarts/components';

// Register them
use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent]);

// Accept the configuration prop passed down from the DashboardView
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

// Wrap the chartOption in a computed property so it instantly reacts to config changes
const chartOption = computed(() => {
  // 1. Safely extract the custom data, defaulting to an empty object if it's a new widget
  const customData = props.widgetData?.customChartData || {};

  // 2. Parse the comma-separated X-Axis string into an array
  const rawXAxis = customData.xAxisData || 'Mon, Tue, Wed, Thu, Fri, Sat, Sun';
  const parsedXAxis = rawXAxis.split(',').map(item => item.trim());

  // 3. Parse the comma-separated Series Data string into an array of Numbers
  const rawSeriesData = customData.seriesData || '400, 200, 120, 900, 300, 450, 700';
  const parsedSeriesData = rawSeriesData.split(',').map(item => Number(item.trim()));

  // 4. Extract Colors
  const barColor = customData.barColor || '#3b82f6';
  const textColor = customData.textColor || '#334155';

  return {
    // Apply global text style for the chart
    textStyle: {
      color: textColor
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' } 
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '5%',
      top: '10%',
      containLabel: true 
    },
    xAxis: {
      type: 'category',
      data: parsedXAxis, // Inject parsed array here
      axisTick: { alignWithLabel: true },
      axisLabel: { color: textColor } // Ensure X labels use dynamic color
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: textColor } // Ensure Y labels use dynamic color
    },
    series: [
      {
        name: customData.seriesName || 'Revenue ($)', // Inject series name
        type: 'bar',
        barWidth: '60%', 
        data: parsedSeriesData, // Inject parsed numbers array here
        itemStyle: {
          color: barColor, // Inject bar color
          borderRadius: [4, 4, 0, 0] 
        }
      }
    ]
  };
});
</script>