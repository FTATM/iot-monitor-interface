<template>
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="backgroundStyle">
    
    <div class="backdrop-blur-md px-4 py-3 shadow-sm z-10 flex justify-between items-center rounded-lg">
      <h3 class="m-0 text-base font-extrabold tracking-wide" :style="{ color: widgetData.widgetStyle?.textHex || '#334155' }">
        {{ widgetData?.widgetLabel || $t('barChart.newChart') }}
      </h3>
    </div>

    <div class="flex-1 w-full min-h-0 relative flex flex-col justify-center mt-2">
      
      <div v-if="!hasDevices" class="absolute inset-0 flex items-center justify-center text-sm italic text-center p-4" :style="{ color: widgetData.widgetStyle?.textHex || '#64748b' }">
        {{ $t('common.noDevicesConfig') }}
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
import { BarChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent]);

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

const hasData = computed(() => {
  const ids = props.widgetData?.deviceIds || [];
  return ids.some(id => liveStreamStore.liveData[id] !== undefined);
});

onMounted(async () => {
  await nextTick();
  setTimeout(() => { isReady.value = true; }, 50);
});

const chartOption = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  const chartTextColor = props.widgetData.widgetStyle?.textHex || '#334155';
  const yAxisName = customData.yAxisName || '';
  const deviceColorsMap = customData.deviceColors || {};
  const fallbackColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

  const rawDeviceIds = props.widgetData?.deviceIds || [];
  let deviceNames = [];
  let deviceValues = [];

  rawDeviceIds.forEach((id, index) => {
    const dataObj = liveStreamStore.liveData[id];
    deviceNames.push(dataObj ? dataObj.name : t('common.loading'));
    deviceValues.push({
      value: dataObj ? dataObj.value : 0,
      itemStyle: { color: deviceColorsMap[id] || fallbackColors[index % fallbackColors.length], borderRadius: [4, 4, 0, 0] }
    });
  });

  return {
    textStyle: { color: chartTextColor },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '5%', top: '15%', containLabel: true },
    xAxis: {
      type: 'category', data: deviceNames,
      axisTick: { alignWithLabel: true },
      axisLabel: { color: chartTextColor } 
    },
    yAxis: {
      type: 'value', name: yAxisName,
      nameTextStyle: { color: chartTextColor },
      axisLabel: { color: chartTextColor } 
    },
    series: [{ type: 'bar', barWidth: '60%', data: deviceValues }]
  };
});
</script>