<template>
  <div class="flex flex-col gap-5">
    
    <div class="grid grid-cols-1 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Primary Line Color</span></div>
        <input type="color" v-model="localConfig.lineColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Line Style & Layout</h4>
      
      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isSmooth" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Smooth Curve</span>
        </label>
        
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showArea" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Fill Area Under Line</span>
        </label>

        <!-- NEW TOGGLE FOR STACKING -->
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isStacked" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">Stack Lines (Sum values)</span>
        </label>
      </div>

      <label class="form-control w-full mt-2">
        <div class="label pb-1">
          <span class="label-text font-semibold">Y-Axis Unit Label</span>
        </div>
        <input type="text" v-model="localConfig.yAxisName"
          class="input input-bordered input-sm w-full" placeholder="e.g., MB/s" />
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

const localConfig = ref({
  lineColor: props.modelValue.lineColor || '#3b82f6',
  isSmooth: props.modelValue.isSmooth !== undefined ? props.modelValue.isSmooth : true,
  showArea: props.modelValue.showArea !== undefined ? props.modelValue.showArea : true,
  isStacked: props.modelValue.isStacked !== undefined ? props.modelValue.isStacked : false,
  yAxisName: props.modelValue.yAxisName || ''
});

watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal });
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', { ...localConfig.value });
});
</script>