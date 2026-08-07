<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New BarProgress' }}
      </h3>
      <!-- Optional top-right text label -->
      <span v-if="chartOption.showTextLabel" class="font-bold text-lg" :style="{ color: chartOption.progressColor }">
        {{ dummyProgress }} / {{ chartOption.maxValue }}
      </span>
    </div>

    <!-- We can use a smaller min-height here since progress bars are usually thin -->
    <div class="flex-1 w-full min-h-[80px] relative flex flex-col justify-center">
      <v-chart v-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick } from 'vue';
import VChart from 'vue-echarts';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart } from 'echarts/charts';
import { GridComponent } from 'echarts/components';

// We only need BarChart and GridComponent for this
use([CanvasRenderer, BarChart, GridComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const isReady = ref(false);

// Dummy data for visual layout
const dummyProgress = 75;

onMounted(async () => {
  await nextTick();
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};

  const progressColor = customData.progressColor || '#10b981'; // Emerald Green
  const trackColor = customData.trackColor || '#e2e8f0';     // Slate 200
  const maxValue = customData.maxValue !== undefined ? customData.maxValue : 100;
  const barThickness = customData.barThickness || 24;
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 12;
  const showTextLabel = customData.showTextLabel !== undefined ? customData.showTextLabel : true;

  return {
    // Expose these to the template so we can use them in the HTML header
    progressColor,
    maxValue,
    showTextLabel,

    // Remove the grid padding entirely
    grid: {
      left: '2%',
      right: '2%',
      top: 'center', // Vertically center the bar
      bottom: 'center',
      height: barThickness, // Lock the grid height to the bar thickness
      containLabel: false
    },
    xAxis: {
      type: 'value',
      max: maxValue, // Lock the max value
      show: false    // Hide the axis completely
    },
    yAxis: {
      type: 'category',
      show: false,   // Hide the axis completely
      data: ['Progress']
    },
    series: [
      // 1. The Background Track
      {
        type: 'bar',
        itemStyle: {
          color: trackColor,
          borderRadius: borderRadius
        },
        silent: true, // Disable hover animations on the background
        barWidth: barThickness,
        // The barGap trick is what makes the next series overlap this one perfectly
        barGap: '-100%', 
        data: [maxValue] 
      },
      // 2. The Foreground Progress Bar
      {
        type: 'bar',
        itemStyle: {
          color: progressColor,
          borderRadius: borderRadius
        },
        barWidth: barThickness,
        z: 3, // Ensure this draws on top of the background track
        data: [dummyProgress]
      }
    ]
  };
});
</script>