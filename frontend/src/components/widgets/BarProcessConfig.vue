<template>
  <div class="flex flex-col gap-5">

    <!-- Target & Label -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Data Configuration</h4>

      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Maximum Goal Value</span>
          </div>
          <input type="number" v-model.number="localConfig.maxValue" class="input input-bordered input-sm w-full"
            placeholder="1000" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Unit Label</span>
          </div>
          <input type="text" v-model="localConfig.unit" class="input input-bordered input-sm w-full"
            placeholder="e.g. GB, MB/s" />
        </label>
      </div>

      <label class="cursor-pointer label justify-start gap-4">
        <input type="checkbox" v-model="localConfig.showTextLabel" class="toggle toggle-primary" />
        <span class="label-text font-semibold">Show Progress Values on Right</span>
      </label>
    </div>

    <!-- Sizing & Style -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Bar Styling</h4>

      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Thickness (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.barThickness" class="input input-bordered input-sm w-full"
            placeholder="16" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Corner Radius (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.borderRadius" class="input input-bordered input-sm w-full"
            placeholder="8" />
        </label>
      </div>

      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Background Track Color</span></div>
        <input type="color" v-model="localConfig.trackColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Dynamic Colors -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">Device Colors</h4>
      </div>

      <div class="flex flex-col gap-2">
        <!-- Loop through the currently selected devices from the General tab -->
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
  // ⚡ NEW: Accept the props from the parent modal
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

// Map the selected IDs to their actual names using the master list
const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean); // Removes any undefined results if a device was deleted
});

// Setup default colors for new devices
const defaultColorPalette = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6'];

const localConfig = ref({
  // ⚡ NEW: Use an object mapping { deviceId: '#color' } instead of an array
  deviceColors: props.modelValue.deviceColors || {},
  trackColor: props.modelValue.trackColor || '#e2e8f0',
  maxValue: props.modelValue.maxValue !== undefined ? props.modelValue.maxValue : 100,
  unit: props.modelValue.unit || '',
  barThickness: props.modelValue.barThickness !== undefined ? props.modelValue.barThickness : 16,
  borderRadius: props.modelValue.borderRadius !== undefined ? props.modelValue.borderRadius : 8,
  showTextLabel: props.modelValue.showTextLabel !== undefined ? props.modelValue.showTextLabel : true
});

// Automatically assign a color to a device if it doesn't have one yet
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