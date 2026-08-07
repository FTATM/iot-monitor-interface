<template>
  <div class="flex flex-col h-full w-full p-4" :style="{ backgroundColor: widgetData.widgetColor?.bgHex || '#ffffff' }">
    
    <!-- Optional Header -->
    <div v-if="config.showHeader" class="bg-white px-4 py-3 border border-slate-200 shadow-sm z-10">
      <h3 class="m-0 text-base font-extrabold text-black tracking-wide">
        {{ widgetData?.widgetLabel || 'New Text' }}
      </h3>
    </div>

    <!-- Text Content Area -->
    <div class="flex-1 w-full relative overflow-y-auto"
         :class="{ 'bg-white border-x border-b border-slate-200 p-4': config.showHeader, 'p-2': !config.showHeader }">
      
      <!-- The whiteSpace: 'pre-wrap' rule is the magic that preserves newlines! -->
      <div class="w-full break-words"
           :style="{ 
             whiteSpace: 'pre-wrap', 
             textAlign: config.textAlign, 
             fontSize: config.fontSize + 'px', 
             color: config.textColor 
           }">
        {{ config.content || 'Double-click or right-click to configure this text block...' }}
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

// Extract configurations with safe fallbacks
const config = computed(() => {
  const customData = props.widgetData?.customChartData || {};
  return {
    content: customData.content || '',
    showHeader: customData.showHeader !== undefined ? customData.showHeader : true,
    textAlign: customData.textAlign || 'left',
    fontSize: customData.fontSize !== undefined ? customData.fontSize : 14,
    textColor: customData.textColor || '#334155' // Slate 700
  };
});
</script>