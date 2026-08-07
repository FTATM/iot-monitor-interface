<template>
  <div class="flex flex-col gap-5">
    
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Chart Display</h4>
      
      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showLegend" class="toggle toggle-primary" />
          <span class="label-text font-semibold">Show Legend</span>
        </label>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Inner Radius</span>
            <span class="label-text-alt">e.g., 40%, 0%</span>
          </div>
          <input type="text" v-model="localConfig.innerRadius"
            class="input input-bordered input-sm w-full" placeholder="40%" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Outer Radius</span>
            <span class="label-text-alt">e.g., 70%, 100%</span>
          </div>
          <input type="text" v-model="localConfig.outerRadius"
            class="input input-bordered input-sm w-full" placeholder="70%" />
        </label>
      </div>

      <label class="form-control w-full mt-3">
        <div class="label pb-1">
          <span class="label-text font-semibold">Slice Corner Radius</span>
          <span class="label-text-alt">Pixels (e.g., 5)</span>
        </div>
        <input type="number" v-model.number="localConfig.borderRadius"
          class="input input-bordered input-sm w-full" placeholder="5" />
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

// Provide sensible defaults for a Doughnut Chart
const localConfig = ref({
  showLegend: props.modelValue.showLegend !== undefined ? props.modelValue.showLegend : true,
  innerRadius: props.modelValue.innerRadius || '40%',
  outerRadius: props.modelValue.outerRadius || '70%',
  borderRadius: props.modelValue.borderRadius !== undefined ? props.modelValue.borderRadius : 5
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