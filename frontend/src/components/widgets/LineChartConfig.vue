<template>
  <div class="flex flex-col gap-5">
    
    <!-- Data Source & History Options -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Data Options</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">History Range</span></div>
          <select v-model="localConfig.historyRange" class="select select-bordered select-sm w-full">
            <option value="0">Live Data Only</option>
            <option value="15m">Last 15 Minutes</option>
            <option value="30m">Last 30 Minutes</option>
            <option value="1h">Last 1 Hour</option>
            <option value="3h">Last 3 Hours</option>
            <option value="6h">Last 6 Hours</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <!-- ⚡ NEW: Custom Range Option -->
            <option value="custom">Custom Absolute Range...</option>
          </select>
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">Max Data Points</span></div>
          <input type="number" v-model.number="localConfig.maxPoints" class="input input-bordered input-sm w-full" placeholder="100" />
        </label>
      </div>

      <!-- ⚡ NEW: "From" and "To" Date/Time Pickers -->
      <div v-if="localConfig.historyRange === 'custom'" class="grid grid-cols-2 gap-4 p-3 mt-2 bg-base-100 rounded-lg border border-base-300">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">From</span></div>
          <!-- step="1" allows selection down to the second -->
          <input type="datetime-local" step="1" v-model="localConfig.customFrom" class="input input-bordered input-sm w-full border-primary" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">To</span></div>
          <input type="datetime-local" step="1" v-model="localConfig.customTo" class="input input-bordered input-sm w-full border-primary" />
        </label>
      </div>
    </div>

    <!-- Line Style & Layout -->
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

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isStacked" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">Stack Lines (Sum values)</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.use24HourFormat" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">Use 24-Hour Time Format</span>
        </label>
      </div>

      <label class="form-control w-full mt-2">
        <div class="label pb-1">
          <span class="label-text font-semibold">Y-Axis Unit Label</span>
        </div>
        <input type="text" v-model="localConfig.yAxisName" class="input input-bordered input-sm w-full" placeholder="e.g., MB/s" />
      </label>
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
          No devices selected.
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  selectedDeviceIds: { type: Array, default: () => [] },
  allDevices: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean);
});

const defaultColorPalette = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

const localConfig = ref({
  historyRange: props.modelValue.historyRange || '1h',
  // ⚡ NEW: Default empty strings for the custom range
  customFrom: props.modelValue.customFrom || '', 
  customTo: props.modelValue.customTo || '',
  maxPoints: props.modelValue.maxPoints || 100,
  isSmooth: props.modelValue.isSmooth !== undefined ? props.modelValue.isSmooth : true,
  showArea: props.modelValue.showArea !== undefined ? props.modelValue.showArea : true,
  isStacked: props.modelValue.isStacked !== undefined ? props.modelValue.isStacked : false,
  use24HourFormat: props.modelValue.use24HourFormat !== undefined ? props.modelValue.use24HourFormat : true,
  yAxisName: props.modelValue.yAxisName || '',
  deviceColors: props.modelValue.deviceColors || {}
});

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