<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <div class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10 flex justify-between items-center">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New Status' }}
      </h3>
    </div>

    <!-- Main Status Display Area -->
    <div class="flex-1 w-full min-h-[120px] flex flex-col items-center justify-center relative p-4">
      
      <div class="flex items-baseline justify-center text-center">
        <!-- Optional Prefix (e.g., $ or ~) -->
        <span v-if="config.prefix" class="font-semibold text-base-content/50 mr-2 text-2xl">
          {{ config.prefix }}
        </span>
        
        <!-- Main Value -->
        <!-- In a real app, 'dummyValue' would come from your websocket or API -->
        <span class="font-extrabold leading-none tracking-tight drop-shadow-sm" 
              :style="{ color: config.valueColor, fontSize: config.fontSize + 'px' }">
          {{ dummyValue }}
        </span>
        
        <!-- Optional Unit (e.g., MB/s or %) -->
        <span v-if="config.unit" class="font-bold text-base-content/50 ml-2 text-2xl">
          {{ config.unit }}
        </span>
      </div>

      <!-- Optional Subtext -->
      <div v-if="config.subtext" class="mt-2 text-base-content/60 font-medium text-sm text-center">
        {{ config.subtext }}
      </div>

    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  widgetData: {
    type: Object,
    default: () => ({})
  }
});

// A dummy value so you can see it while designing the canvas
const dummyValue = 'Online';

// Computed property to safely extract all config values with fallbacks
const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  
  return {
    valueColor: customData.valueColor || '#10b981', // Default to Emerald Green
    fontSize: customData.fontSize || 56, // Default to a very large font size
    prefix: customData.prefix || '',
    unit: customData.unit || '',
    subtext: customData.subtext || ''
  };
});
</script>