<template>
  <div class="flex flex-col gap-5">
    
    <!-- ECharts Colors -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Progress Color</span></div>
        <input type="color" v-model="localConfig.progressColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Pointer / Text Color</span></div>
        <input type="color" v-model="localConfig.pointerColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Gauge Scale Configuration -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Scale & Units</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Minimum Value</span></div>
          <!-- Use v-model.number to ensure ECharts gets an integer, not a string -->
          <input type="number" v-model.number="localConfig.min"
            class="input input-bordered input-sm w-full" placeholder="0" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Maximum Value</span></div>
          <input type="number" v-model.number="localConfig.max"
            class="input input-bordered input-sm w-full" placeholder="100" />
        </label>
      </div>

      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">Unit Label</span>
          <span class="label-text-alt text-base-content/60">e.g., %, °C, RPM</span>
        </div>
        <input type="text" v-model="localConfig.unit"
          class="input input-bordered input-sm w-full" placeholder="%" />
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

// Provide sensible defaults for a Gauge Chart
const localConfig = ref({
  progressColor: props.modelValue.progressColor || '#3b82f6',
  pointerColor: props.modelValue.pointerColor || '#0f172a',
  min: props.modelValue.min !== undefined ? props.modelValue.min : 0,
  max: props.modelValue.max !== undefined ? props.modelValue.max : 100,
  unit: props.modelValue.unit || '%'
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