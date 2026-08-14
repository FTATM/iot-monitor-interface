<template>
  <div class="flex flex-col gap-5">
    
    <!-- General Pie Settings -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Chart Appearance</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Inner Radius</span>
          </div>
          <input type="text" v-model="localConfig.innerRadius"
            class="input input-bordered input-sm w-full" placeholder="40%" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Outer Radius</span>
          </div>
          <input type="text" v-model="localConfig.outerRadius"
            class="input input-bordered input-sm w-full" placeholder="70%" />
        </label>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Corner Radius (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.borderRadius"
            class="input input-bordered input-sm w-full" placeholder="5" />
        </label>
        
        <label class="cursor-pointer label justify-start gap-4 mt-6">
          <input type="checkbox" v-model="localConfig.showLegend" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Show Legend</span>
        </label>
      </div>
    </div>

    <!-- Device Specific Colors -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">Device Colors</h4>
      </div>
      
      <div class="flex flex-col gap-2">
        <div v-for="device in activeDevices" :key="device.deviceId" 
             class="flex items-center gap-3 p-2 bg-base-100 border border-base-300 rounded-lg">
          
          <input type="color" v-model="localConfig.deviceColors[device.deviceId]" 
                 class="h-8 w-12 cursor-pointer rounded border border-base-300 p-0 shrink-0" />
          
          <span class="text-sm font-semibold flex-1">{{ device.deviceName }}</span>
        </div>

        <div v-if="activeDevices.length === 0" class="text-sm text-base-content/50 py-2 text-center">
          No devices selected. Please select devices in the General Settings tab first.
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  },
  selectedDeviceIds: {
    type: Array,
    default: () => []
  },
  allDevices: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['update:modelValue']);

// ⚡ THE FIX: Use .map() to strictly preserve the dropdown selection order
const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean);
});

const defaultColorPalette = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

const localConfig = ref({
  innerRadius: props.modelValue.innerRadius || '40%',
  outerRadius: props.modelValue.outerRadius || '70%',
  borderRadius: props.modelValue.borderRadius !== undefined ? props.modelValue.borderRadius : 5,
  showLegend: props.modelValue.showLegend !== undefined ? props.modelValue.showLegend : true,
  deviceColors: props.modelValue.deviceColors || {}
});

// Auto-assign colors to new devices
watch(() => props.selectedDeviceIds, (newIds) => {
  newIds.forEach((id, index) => {
    if (!localConfig.value.deviceColors[id]) {
      localConfig.value.deviceColors[id] = defaultColorPalette[index % defaultColorPalette.length];
    }
  });
}, { immediate: true, deep: true });

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>