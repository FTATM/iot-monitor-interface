<template>
  <div class="flex flex-col gap-5">
    
    <!-- ECharts Colors -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Text Color</span></div>
        <input type="color" v-model="localConfig.textColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Bar Color</span></div>
        <input type="color" v-model="localConfig.barColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- X-Axis Configuration -->
    <label class="form-control w-full">
      <div class="label pb-1">
        <span class="label-text font-bold">X-Axis Labels</span>
        <span class="label-text-alt text-base-content/60">Comma-separated</span>
      </div>
      <input type="text" v-model="localConfig.xAxisData"
        class="input input-bordered input-sm w-full font-mono" 
        placeholder="Mon, Tue, Wed, Thu, Fri, Sat, Sun" />
    </label>

    <!-- Series Configuration -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Series Data</h4>
      
      <label class="form-control w-full mb-3">
        <div class="label pb-1"><span class="label-text font-semibold">Series Name</span></div>
        <input type="text" v-model="localConfig.seriesName"
          class="input input-bordered input-sm w-full" placeholder="Revenue ($)" />
      </label>
      
      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">Data Values</span>
          <span class="label-text-alt text-base-content/60">Comma-separated numbers</span>
        </div>
        <input type="text" v-model="localConfig.seriesData"
          class="input input-bordered input-sm w-full font-mono" 
          placeholder="400, 200, 120, 900, 300, 450, 700" />
      </label>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

// Provide sensible defaults if this is a brand new widget
const localConfig = ref({
  textColor: props.modelValue.textColor || '#334155',
  barColor: props.modelValue.barColor || '#3b82f6',
  xAxisData: props.modelValue.xAxisData || 'Mon, Tue, Wed, Thu, Fri, Sat, Sun',
  seriesName: props.modelValue.seriesName || 'Revenue ($)',
  seriesData: props.modelValue.seriesData || '400, 200, 120, 900, 300, 450, 700'
});

// FIX 1: Watch the entire object and emit ANY change automatically. 
// No more manual @input bindings needed!
watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal });
}, { deep: true });

// FIX 2: Emit the default values immediately when the modal opens!
// This prevents DashboardView from saving an empty {} if you don't edit every field.
onMounted(() => {
  emit('update:modelValue', { ...localConfig.value });
});
</script>