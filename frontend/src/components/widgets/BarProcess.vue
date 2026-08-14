<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">

    <div class="bg-white px-4 py-3 border-b border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New BarProgress' }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-[150px] relative flex flex-col justify-center">
      
      <!-- 1. No Devices Selected State -->
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No devices selected. Open configuration to add data sources.
      </div>

      <!-- 2. Loading State (Devices selected, waiting for SSE ping) -->
      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <!-- 3. Actual Chart -->
      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />
      
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted,  nextTick  } from 'vue';
import VChart from 'vue-echarts';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart } from 'echarts/charts';
import { GridComponent, TooltipComponent } from 'echarts/components';

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent]);

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const hasData = computed(() => {
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => liveStreamStore.liveData[id] !== undefined);
});

onMounted(async () => {
  await nextTick();
  setTimeout(() => {
    isReady.value = true;
  }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};

  const trackColor = customData.trackColor || '#e2e8f0';
  const maxValue = customData.maxValue && customData.maxValue > 0 ? customData.maxValue : 100;
  const unit = customData.unit || '';
  const barThickness = customData.barThickness || 16;
  const borderRadius = customData.borderRadius !== undefined ? customData.borderRadius : 8;
  const showTextLabel = customData.showTextLabel !== undefined ? customData.showTextLabel : true;
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444'];

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  
  let deviceNames = [];
  let deviceValues = [];

  // ⚡ THE FIX: Loop strictly over the selected IDs to guarantee absolute visual ordering
  rawDeviceIds.forEach(id => {
    const data = liveStreamStore.liveData[id];
    deviceNames.push(data ? data.name : `Device Loading...`);
    deviceValues.push(data ? data.value : 0);
  });

  const backgroundValues = deviceNames.map(() => maxValue);

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: function (params) {
        const p = params.find(param => param.seriesName === 'Progress');
        if (!p) return '';
        const pct = Math.round((p.value / maxValue) * 100);
        return `<strong>${p.name}</strong><br/>Value: ${p.value}${unit} (${pct}%)`;
      }
    },
    grid: {
      left: '2%',
      right: '15%',
      top: '15%',
      bottom: '5%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      max: maxValue,
      show: false
    },
    yAxis: {
      type: 'category',
      data: deviceNames,
      show: false,
      inverse: true
    },
    series: [
      {
        name: 'Background',
        type: 'bar',
        itemStyle: { color: trackColor, borderRadius: borderRadius },
        silent: true,
        barWidth: barThickness,
        barGap: '-100%',
        barCategoryGap: '60%',
        data: backgroundValues,
        label: {
          show: true,
          position: ['0%', '-20px'],
          formatter: '{b}',
          color: '#0f172a',
          fontSize: 13,
          fontWeight: 'bold'
        }
      },
      {
        name: 'Progress',
        type: 'bar',
        itemStyle: {
          borderRadius: borderRadius,
          color: function (params) {
            // Because our Y-axis is now perfectly mapped to rawDeviceIds, this dataIndex lookup is bulletproof!
            const deviceId = rawDeviceIds[params.dataIndex];
            return deviceColorsMap[deviceId] || fallbackColors[params.dataIndex % fallbackColors.length];
          }
        },
        barWidth: barThickness,
        z: 3,
        label: {
          show: showTextLabel,
          position: 'right',
          formatter: function (params) {
            const pct = Math.round((params.value / maxValue) * 100);
            return `{pctStyle|${pct}%}`;
          },
          rich: {
            pctStyle: {
              color: '#64748b',
              fontWeight: 'normal',
              fontSize: 12
            }
          }
        },
        data: deviceValues
      }
    ]
  };
});

function sameIds(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  return a.every((id, i) => id === b[i]);
}
</script>