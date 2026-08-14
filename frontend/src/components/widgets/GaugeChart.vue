<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">

    <div class="bg-white px-4 py-3 border-b border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide truncate pr-2">
        {{ widgetData?.widgetLabel || 'New GaugeChart' }}
      </h3>

      <span v-if="hasData && liveDeviceName" class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>

    <div class="flex-1 w-full min-h-0 relative mt-4 flex flex-col justify-center">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm text-base-content/50 italic text-center p-4">
        No device selected.
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />
      
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

// ⚡ Import the shared store
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, GaugeChart, TooltipComponent]);

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);

// ⚡ Safely grab the first device ID since Gauge only tracks one device
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);

// ⚡ Read directly from the store
const hasData = computed(() => deviceId.value && liveStreamStore.liveData[deviceId.value] !== undefined);
const liveValue = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].value : 0);
const liveDeviceName = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].name : '');

onMounted(async () => {
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  const min = customData.min !== undefined ? customData.min : 0;
  const max = customData.max !== undefined ? customData.max : 100;
  const unit = customData.unit || '';
  const showGradeText = customData.showGradeText !== undefined ? customData.showGradeText : true;
  const gradeTextSize = customData.gradeTextSize !== undefined ? customData.gradeTextSize : 14;

  const rawGrades = customData.grades || [
    { name: 'Grade D', limit: 25, color: '#FF6E76' },
    { name: 'Grade C', limit: 50, color: '#FDDD60' },
    { name: 'Grade B', limit: 75, color: '#58D9F9' },
    { name: 'Grade A', limit: 100, color: '#7CFFB2' }
  ];

  const sortedGrades = [...rawGrades].sort((a, b) => a.limit - b.limit);
  const range = max - min;

  const colorStops = sortedGrades.map(grade => {
    const clampedLimit = Math.min(grade.limit, max);
    let pct = (clampedLimit - min) / range;
    if (pct > 1) pct = 1;
    if (pct < 0) pct = 0;
    return [pct, grade.color];
  });

  if (colorStops.length === 0) colorStops.push([1, '#e2e8f0']);

  return {
    series: [
      {
        type: 'gauge',
        startAngle: 180, endAngle: 0, center: ['50%', '75%'], radius: '90%',
        min: min, max: max, splitNumber: 8,
        axisLine: { lineStyle: { width: 6, color: colorStops } },
        pointer: {
          icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
          length: '12%', width: 20, offsetCenter: [0, '-60%'],
          itemStyle: { color: 'auto' }
        },
        axisTick: { length: 12, lineStyle: { color: 'auto', width: 2 } },
        splitLine: { length: 20, lineStyle: { color: 'auto', width: 5 } },
        axisLabel: {
          show: showGradeText, color: '#464646', fontSize: gradeTextSize,
          distance: -60, rotate: 'tangential',
          formatter: function (value) {
            if (sortedGrades.length === 0) return '';
            let minDistance = Infinity; let closestGradeName = ''; let previousLimit = min;
            for (let i = 0; i < sortedGrades.length; i++) {
              const currentLimit = Math.min(sortedGrades[i].limit, max);
              const midpoint = previousLimit + (currentLimit - previousLimit) / 2;
              const dist = Math.abs(value - midpoint);
              if (dist < minDistance) { minDistance = dist; closestGradeName = sortedGrades[i].name; }
              previousLimit = currentLimit;
            }
            const tickSize = range / 8;
            if (minDistance > (tickSize / 2) + 0.1) return '';
            return closestGradeName;
          }
        },
        title: { offsetCenter: [0, '-10%'], fontSize: 20 },
        detail: {
          fontSize: 30, offsetCenter: [0, '-25%'], valueAnimation: true,
          formatter: function (value) { return (Math.round(value * 10) / 10) + unit; },
          color: 'inherit'
        },
        data: [{ value: liveValue.value, name: '' }] // ⚡ Inject the reactive value
      }
    ]
  };
});
</script>