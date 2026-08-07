<template>
  <div class="flex flex-col gap-5">
    
    <!-- ECharts Colors for Actual & Target -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Performance Bar</span></div>
        <input type="color" v-model="localConfig.barColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Target Marker</span></div>
        <input type="color" v-model="localConfig.targetColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Target Goal -->
    <label class="form-control w-full">
      <div class="label pb-1"><span class="label-text font-bold text-primary">Target Goal Value</span></div>
      <input type="number" v-model.number="localConfig.targetValue"
        class="input input-bordered input-sm w-full border-primary" placeholder="85" />
    </label>

    <!-- Dynamic Threshold Zones -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">Background Threshold Zones</h4>
        <button @click="addThreshold" class="btn btn-xs btn-primary btn-outline">
          + Add Zone
        </button>
      </div>
      
      <div class="flex flex-col gap-3">
        <div v-for="(zone, index) in localConfig.thresholds" :key="zone.id" 
             class="flex items-end gap-2 p-3 bg-base-100 border border-base-300 rounded-lg">
          
          <label class="form-control flex-1">
            <span class="label-text-alt mb-1 font-semibold">Name</span>
            <input type="text" v-model="zone.name" class="input input-bordered input-xs w-full" />
          </label>
          
          <label class="form-control w-20">
            <span class="label-text-alt mb-1 font-semibold">Max Value</span>
            <input type="number" v-model.number="zone.value" class="input input-bordered input-xs w-full" />
          </label>
          
          <label class="form-control w-12">
            <span class="label-text-alt mb-1 font-semibold">Color</span>
            <input type="color" v-model="zone.color" class="h-6 w-full cursor-pointer rounded border border-base-300 p-0" />
          </label>

          <button @click="removeThreshold(index)" class="btn btn-xs btn-ghost text-error hover:bg-error/10 px-2" title="Remove zone">
            <Icon icon="lucide:x" class="w-5 h-5" />
          </button>
        </div>

        <div v-if="localConfig.thresholds.length === 0" class="text-center text-sm text-base-content/50 py-2">
          No zones added. The background will be blank.
        </div>
      </div>
      <p class="text-xs text-base-content/60 mt-3">
        * ECharts automatically overlays these. Make sure larger values act as the background for smaller values.
      </p>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { Icon } from '@iconify/vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);

// Setup defaults with an array of thresholds
const localConfig = ref({
  barColor: props.modelValue.barColor || '#3b82f6',
  targetColor: props.modelValue.targetColor || '#0f172a',
  targetValue: props.modelValue.targetValue !== undefined ? props.modelValue.targetValue : 85,
  thresholds: props.modelValue.thresholds || [
    { id: Date.now() + 1, name: 'Excellent', value: 120, color: '#e2e8f0' },
    { id: Date.now() + 2, name: 'Satisfactory', value: 90, color: '#cbd5e1' },
    { id: Date.now() + 3, name: 'Poor', value: 60, color: '#94a3b8' }
  ]
});

const addThreshold = () => {
  localConfig.value.thresholds.push({
    id: Date.now(),
    name: 'New Zone',
    value: 0,
    color: '#e5e7eb'
  });
};

const removeThreshold = (index) => {
  localConfig.value.thresholds.splice(index, 1);
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>