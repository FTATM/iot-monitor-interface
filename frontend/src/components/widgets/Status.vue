<template>
  <!-- Added overflow-hidden to the main wrapper to respect grid boundaries -->
  <div class="flex flex-col h-full w-full p-4 overflow-hidden" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Added shrink-0 so the header never squishes -->
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center shrink-0">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide truncate pr-2">
        {{ widgetData?.widgetLabel || 'New Status' }}
      </h3>
      <span v-if="hasData && liveDeviceName" class="badge badge-neutral badge-lg py-4 px-4 text-sm font-bold shrink-0 shadow-sm">
        {{ liveDeviceName }}
      </span>
    </div>

    <!-- Main Status Display Area: Added min-h-0 and overflow-y-auto for scrolling -->
    <div class="flex-1 w-full min-h-0 relative p-4 overflow-y-auto flex flex-col">
      
      <!-- Replaced justify-center with m-auto to fix the overflow centering bug -->
      <div v-if="!hasDevices" class="m-auto text-sm text-base-content/50 italic text-center">
        No device selected.
      </div>

      <div v-else-if="!hasData" class="m-auto flex flex-col items-center text-sm text-base-content/50 gap-3">
        <span class="loading loading-spinner loading-md text-primary"></span>
        Waiting for live data...
      </div>
      
      <div v-else class="m-auto flex flex-col items-center justify-center text-center w-full">
        <div class="flex items-baseline justify-center text-center">
          <span v-if="config.prefix" class="font-semibold text-base-content/50 mr-2 text-2xl">
            {{ config.prefix }}
          </span>
          
          <span class="font-extrabold leading-none tracking-tight drop-shadow-sm transition-colors duration-300" 
                :style="{ color: evaluatedStatus.color, fontSize: config.fontSize + 'px' }">
            {{ evaluatedStatus.text }}
          </span>
          
          <span v-if="config.unit" class="font-bold text-base-content/50 ml-2 text-2xl">
            {{ config.unit }}
          </span>
        </div>

        <div v-if="config.subtext" class="mt-2 text-base-content/60 font-medium text-sm text-center whitespace-pre-line w-full">
          {{ config.subtext }}
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { computed  } from 'vue';
import { useLiveStreamStore } from '@/stores/useLiveStreamStore';

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

const liveStreamStore = useLiveStreamStore();
const hasDevices = computed(() => props.widgetData?.deviceIds && props.widgetData.deviceIds.length > 0);
const deviceId = computed(() => hasDevices.value ? String(props.widgetData.deviceIds[0]) : null);
const hasData = computed(() => deviceId.value && liveStreamStore.liveData[deviceId.value] !== undefined);
const liveValue = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].value : 0);
const liveDeviceName = computed(() => hasData.value ? liveStreamStore.liveData[deviceId.value].name : '');

const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    valueColor: customData.valueColor || '#10b981', 
    fontSize: customData.fontSize || 56, 
    prefix: customData.prefix || '',
    unit: customData.unit || '',
    subtext: customData.subtext || '',
    enableConditions: customData.enableConditions || false,
    conditions: customData.conditions || []
  };
});

const evaluatedStatus = computed(() => {
  const currentValue = liveValue.value !== null ? liveValue.value : '0';

  if (!config.value.enableConditions || config.value.conditions.length === 0) {
    return { text: currentValue, color: config.value.valueColor };
  }

  for (const cond of config.value.conditions) {
    let match = false;
    
    const numCurrent = Number(currentValue);
    const numCompare = Number(cond.compareValue);
    const isNumeric = !isNaN(numCurrent) && !isNaN(numCompare) && cond.compareValue !== '';

    switch (cond.operator) {
      case '==': 
        match = currentValue == cond.compareValue; 
        break;
      case '!=': 
        match = currentValue != cond.compareValue; 
        break;
      case '>': 
        match = isNumeric ? numCurrent > numCompare : currentValue > cond.compareValue; 
        break;
      case '<': 
        match = isNumeric ? numCurrent < numCompare : currentValue < cond.compareValue; 
        break;
      case '>=': 
        match = isNumeric ? numCurrent >= numCompare : currentValue >= cond.compareValue; 
        break;
      case '<=': 
        match = isNumeric ? numCurrent <= numCompare : currentValue <= cond.compareValue; 
        break;
    }

    if (match) {
      return { 
        text: cond.displayText, 
        color: cond.color || config.value.valueColor 
      };
    }
  }

  return { text: currentValue, color: config.value.valueColor };
});
</script>