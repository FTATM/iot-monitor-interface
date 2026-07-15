<template>
  <div class="widget-container">
    <h3 class="widget-title">Q3 Revenue Target</h3>
    
    <div class="chart-wrapper">
      <v-chart class="chart" :option="chartOption" autoresize />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import VChart from 'vue-echarts';

// We need BarChart for the ranges/actual, and ScatterChart for the target marker
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart, ScatterChart } from 'echarts/charts';
import { 
  GridComponent, 
  TooltipComponent, 
  LegendComponent 
} from 'echarts/components';

// Register all necessary components
use([CanvasRenderer, BarChart, ScatterChart, GridComponent, TooltipComponent, LegendComponent]);

const chartOption = ref({
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
  // Bullet charts use a horizontal layout, so Category is on the Y-Axis
  xAxis: {
    type: 'value',
    splitLine: { show: false } // Hides the vertical grid lines for a cleaner look
  },
  yAxis: {
    type: 'category',
    data: ['Revenue'],
    axisLine: { show: false },
    axisTick: { show: false }
  },
  
  // THE MAGIC LAYER CAKE
  series: [
    // 1. The "Good" Range (Background Base)
    {
      name: 'Excellent',
      type: 'bar',
      barWidth: '50%',
      data: [120], // Max scale of the chart
      itemStyle: { color: '#e2e8f0' }, // Light gray
      animation: false // Background doesn't need to animate in
    },
    // 2. The "Satisfactory" Range
    {
      name: 'Satisfactory',
      type: 'bar',
      barWidth: '50%',
      barGap: '-100%', // Pulls this bar directly over the previous one
      data: [90],
      itemStyle: { color: '#cbd5e1' }, // Medium gray
      animation: false
    },
    // 3. The "Poor" Range
    {
      name: 'Poor',
      type: 'bar',
      barWidth: '50%',
      barGap: '-100%',
      data: [60],
      itemStyle: { color: '#94a3b8' }, // Dark gray
      animation: false
    },
    // 4. The Actual Performance Bar
    {
      name: 'Actual Revenue',
      type: 'bar',
      barWidth: '20%', // Thinner so the background ranges show around it
      barGap: '-100%',
      data: [75],
      itemStyle: { color: '#3b82f6' }, // Dashboard blue
      z: 3 // Ensures it renders on top of the gray backgrounds
    },
    // 5. The Target Marker
    {
      name: 'Target',
      type: 'scatter',
      symbol: 'rect', // Turns the default circle into a rectangle
      symbolSize: [4, 40], // 4px wide, 40px tall (creates a vertical line)
      data: [85],
      itemStyle: { color: '#0f172a' }, // Very dark slate
      z: 4 // Renders on the absolute top layer
    }
  ]
});
</script>

<style scoped>
.widget-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 16px;
  background-color: white;
  box-sizing: border-box;
}

.widget-title {
  margin: 0 0 12px 0;
  font-size: 1rem;
  color: #334155;
  font-weight: bold;
}

.chart-wrapper {
  flex: 1;
  width: 100%;
}

.chart {
  height: 100%;
  width: 100%;
}
</style>