<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New ScatterChart' }}
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
import { ScatterChart, LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';

// Register both ScatterChart (for points) and LineChart (for the regression line)
use([CanvasRenderer, ScatterChart, LineChart, GridComponent, TooltipComponent, LegendComponent]);

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

// Helper function to calculate Linear Regression (Trendline)
const calculateRegressionLine = (data) => {
  let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0;
  const n = data.length;
  
  if (n === 0) return [];

  data.forEach(point => {
    sumX += point[0];
    sumY += point[1];
    sumXY += point[0] * point[1];
    sumXX += point[0] * point[0];
  });

  const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;

  // Find min and max X values to draw the line across the entire dataset
  const minX = Math.min(...data.map(p => p[0]));
  const maxX = Math.max(...data.map(p => p[0]));

  return [
    [minX, slope * minX + intercept],
    [maxX, slope * maxX + intercept]
  ];
};

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};

  const pointColor = customData.pointColor || '#3b82f6';
  const lineColor = customData.lineColor || '#ef4444';
  const showRegression = customData.showRegression !== undefined ? customData.showRegression : true;
  const xAxisName = customData.xAxisName || '';
  const yAxisName = customData.yAxisName || '';

  // Dummy data (e.g., Engine RPM vs Temperature)
  const rawData = [
    [10.0, 8.04], [8.0, 6.95], [13.0, 7.58], [9.0, 8.81], [11.0, 8.33], 
    [14.0, 9.96], [6.0, 7.24], [4.0, 4.26], [12.0, 10.84], [7.0, 4.82], [5.0, 5.68]
  ];

  const regressionData = showRegression ? calculateRegressionLine(rawData) : [];

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>X: {c0} <br/>Y: {c1}'
    },
    legend: {
      show: true,
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '8%',
      bottom: '12%',
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      name: xAxisName,
      nameTextStyle: { fontWeight: 'bold' }
    },
    yAxis: {
      type: 'value',
      name: yAxisName,
      nameTextStyle: { fontWeight: 'bold' }
    },
    series: [
      {
        name: 'Data Points',
        type: 'scatter',
        itemStyle: {
          color: pointColor,
          opacity: 0.8
        },
        symbolSize: 12,
        data: rawData
      },
      {
        name: 'Trendline',
        type: 'line',
        showSymbol: false,
        smooth: true,
        itemStyle: {
          color: lineColor
        },
        lineStyle: {
          width: 2,
          type: 'dashed'
        },
        data: regressionData
      }
    ]
  };
});
</script>