<template>
  <div class="flex flex-col gap-5">
    
    <!-- Color Configuration -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Data Point Color</span></div>
        <input type="color" v-model="localConfig.pointColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Trendline Color</span></div>
        <input type="color" v-model="localConfig.lineColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" :disabled="!localConfig.showRegression" />
      </label>
    </div>

    <!-- Chart Features -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Regression Analysis</h4>
      
      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showRegression" class="toggle toggle-primary" />
          <span class="label-text font-semibold">Show Linear Regression Trendline</span>
        </label>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">X-Axis Label</span>
          </div>
          <input type="text" v-model="localConfig.xAxisName"
            class="input input-bordered input-sm w-full" placeholder="e.g., Temperature" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Y-Axis Label</span>
          </div>
          <input type="text" v-model="localConfig.yAxisName"
            class="input input-bordered input-sm w-full" placeholder="e.g., Pressure" />
        </label>
      </div>
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

// Provide sensible defaults for a Scatter Chart with Regression
const localConfig = ref({
  pointColor: props.modelValue.pointColor || '#3b82f6',
  lineColor: props.modelValue.lineColor || '#ef4444',
  showRegression: props.modelValue.showRegression !== undefined ? props.modelValue.showRegression : true,
  xAxisName: props.modelValue.xAxisName || '',
  yAxisName: props.modelValue.yAxisName || ''
});

// Watch the entire object and emit ANY change automatically
watch(localConfig, (newVal) => {
  emit('update:modelValue', { ...newVal });
}, { deep: true });

// Emit the default values immediately when the modal opens
onMounted(() => {
  emit('update:modelValue', { ...localConfig.value });
});
</script>