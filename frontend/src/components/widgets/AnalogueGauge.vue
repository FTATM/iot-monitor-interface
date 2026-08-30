<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">

    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide truncate pr-2" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel || $t('analogueGauge.newGauge') }}
      </h3>

      <span v-if="hasData && liveDeviceName" class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>

    <div class="flex-1 w-full min-h-0 relative mt-4 flex flex-col justify-center">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevice') }}
      </div>

      <div v-else-if="!hasData" class="absolute inset-0 flex flex-col items-center justify-center text-sm gap-3" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        <span class="loading loading-spinner loading-md text-primary"></span>
        {{ $t('common.waitingData') }}
      </div>

      <v-chart v-else-if="isReady" class="absolute inset-0 w-full h-full" :option="chartOption" autoresize />
      
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { GaugeChart } from 'echarts/charts';
import { TooltipComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, GaugeChart, TooltipComponent]);

const { t } = useI18n();

const props = defineProps({
  widgetData: { type: Object, default: () => ({}) }
});

const isReady = ref(false);
const liveStreamStore = useLiveStreamStore();

const backgroundStyle = computed(() => {
  const colorObj = props.widgetData.widgetStyle || {};
  const c1 = colorObj.bgHex || '#ffffff';
  if (!colorObj.useGradient) return { backgroundColor: c1 };
  const c2 = colorObj.bgHex2 || c1; 
  const angle = colorObj.bgGradientDir || '135deg';
  return { background: `linear-gradient(${angle}, ${c1}, ${c2})` };
});

const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);
const hasData = computed(() => deviceId.value && liveStreamStore.liveData[deviceId.value] !== undefined);
const liveValue = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].value : 0);
const liveDeviceName = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].name : '');

onMounted(async () => {
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';
  
  const min = customData.min !== undefined ? customData.min : 0;
  const max = customData.max !== undefined ? customData.max : 100;
  const unit = customData.unit || '';
  const useGrades = customData.useGrades || false;
  const showGradeText = useGrades ? (customData.showGradeText !== undefined ? customData.showGradeText : true) : false;
  const gradeTextSize = customData.gradeTextSize !== undefined ? customData.gradeTextSize : 14;

  let colorStops = [[1, '#3b82f6']]; 
  let sortedGrades = [];

  if (useGrades) {
    const rawGrades = customData.grades || [];
    sortedGrades = [...rawGrades].sort((a, b) => a.limit - b.limit);
    const range = max - min;
    colorStops = sortedGrades.map(grade => {
      const clampedLimit = Math.min(grade.limit, max);
      let pct = (clampedLimit - min) / range;
      if (pct > 1) pct = 1;
      if (pct < 0) pct = 0;
      return [pct, grade.color];
    });
    if (colorStops.length === 0) colorStops.push([1, '#e2e8f0']);
  }

  return {
    textStyle: { color: chartTextColor },
    series: [
      {
        type: 'gauge', startAngle: 180, endAngle: 0, center: ['50%', '75%'], radius: '90%',
        min: min, max: max, splitNumber: 8,
        axisLine: { lineStyle: { width: 6, color: colorStops } },
        pointer: { icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z', length: '12%', width: 20, offsetCenter: [0, '-60%'], itemStyle: { color: 'auto' } },
        axisTick: { length: 12, lineStyle: { color: chartTextColor, width: 2 } },
        splitLine: { length: 20, lineStyle: { color: chartTextColor, width: 5 } },
        axisLabel: {
          show: useGrades ? showGradeText : true, 
          color: chartTextColor, fontSize: gradeTextSize,
          distance: -60, rotate: 'tangential',
          formatter: function (value) {
            if (!useGrades || sortedGrades.length === 0) return Math.round(value);
            
            let minDistance = Infinity; let closestGradeName = ''; let previousLimit = min;
            for (let i = 0; i < sortedGrades.length; i++) {
              const currentLimit = Math.min(sortedGrades[i].limit, max);
              const midpoint = previousLimit + (currentLimit - previousLimit) / 2;
              const dist = Math.abs(value - midpoint);
              if (dist < minDistance) { minDistance = dist; closestGradeName = sortedGrades[i].name; }
              previousLimit = currentLimit;
            }
            const range = max - min;
            const tickSize = range / 8;
            if (minDistance > (tickSize / 2) + 0.1) return '';
            return closestGradeName;
          }
        },
        title: { offsetCenter: [0, '-10%'], fontSize: 20, color: chartTextColor },
        detail: {
          fontSize: 30, offsetCenter: [0, '-25%'], valueAnimation: true,
          formatter: function (value) { return (Math.round(value * 10) / 10) + unit; },
          color: chartTextColor
        },
        data: [{ value: liveValue.value, name: '' }]
      }
    ]
  };
});
</script>