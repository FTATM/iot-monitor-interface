<template>
  <div class="widget-container">
    <h3 class="widget-title">Server CPU (ECharts)</h3>
    
    <div class="chart-wrapper">
      <v-chart class="chart" :option="chartOption" autoresize />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

// 1. Import the Vue wrapper
import VChart from 'vue-echarts';

// 2. Import the ECharts core and the specific pieces we need
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { GaugeChart } from 'echarts/charts';
import { TooltipComponent } from 'echarts/components';

// 3. Register the components (this keeps your bundle size small!)
use([CanvasRenderer, GaugeChart, TooltipComponent]);

// 4. The ECharts Configuration Object
// ECharts has a native, highly customizable 'gauge' type
const chartOption = ref({
  series: [
    {
      type: 'gauge',
      startAngle: 180,
      endAngle: 0,
      center: ['50%', '75%'], // Push it down slightly to fit the half-circle perfectly
      radius: '80%',
      min: 0,
      max: 100,
      splitNumber: 5,
      
      // The colored progress bar inside the gauge
      progress: {
        show: true,
        width: 18,
        itemStyle: {
          color: '#3b82f6' // Your dashboard blue
        }
      },
      
      // The background track of the gauge
      axisLine: {
        lineStyle: {
          width: 18,
          color: [[1, '#e2e8f0']] // Light gray
        }
      },
      
      // Hide the tiny ticks
      axisTick: { show: false },
      
      // The larger dividing lines
      splitLine: {
        show: true,
        length: 5,
        lineStyle: { width: 2, color: '#94a3b8' }
      },
      
      // The numbers around the gauge
      axisLabel: {
        show: false,
        distance: 10,
        color: '#64748b',
        fontSize: 12
      },
      
      // The physical needle pointing to the value!
      pointer: {
        icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
        length: '12%',
        width: 10,
        offsetCenter: [0, '-60%'],
        itemStyle: { color: '#0f172a' }
      },
      
      // The text in the center
      detail: {
        valueAnimation: true,
        formatter: '{value}%',
        color: '#0f172a',
        fontSize: 16,
        fontWeight: 'bold',
        offsetCenter: [0, '0%']
      },
      
      // The actual data value
      data: [{ value: 20 }]
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
  margin: 0 0 0 0;
  font-size: 1rem;
  color: #334155;
  font-weight: bold;
}

.chart-wrapper {
  flex: 1;
  width: 100%;
  /* No crazy flex hacks needed here, just let the canvas fill the div */
}

.chart {
  height: 100%;
  width: 100%;
}
</style>