<template>
  <div class="widget-container">
    <h3 class="widget-title">Revenue (ECharts)</h3>
    
    <div class="chart-wrapper">
      <v-chart class="chart" :option="chartOption" autoresize />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
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

// The ECharts Configuration Object
const chartOption = ref({
  // Built-in hover tooltips that snap to the axis
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' } // Creates a nice gray highlight behind the hovered bar
  },
  
  // 'grid' controls the spacing of the chart inside the canvas
  grid: {
    left: '3%',
    right: '4%',
    bottom: '5%',
    top: '10%',
    containLabel: true // Ensures the Y-axis numbers don't get cut off when resizing
  },
  
  xAxis: {
    type: 'category',
    data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    axisTick: { alignWithLabel: true }
  },
  
  yAxis: {
    type: 'value'
  },
  
  series: [
    {
      name: 'Revenue ($)',
      type: 'bar',
      barWidth: '60%', // Controls how thick the bars are
      data: [400, 200, 120, 900, 300, 450, 700],
      itemStyle: {
        color: '#3b82f6', // The same blue theme
        borderRadius: [4, 4, 0, 0] // Rounds the top-left and top-right corners
      }
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

/* No crazy Flexbox overrides needed here! */
.chart-wrapper {
  flex: 1;
  width: 100%;
}

.chart {
  height: 100%;
  width: 100%;
}
</style>