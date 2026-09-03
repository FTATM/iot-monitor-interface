<template>
  <div class="flex flex-col gap-5">

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('lineChart.config.dataOptions') }}</h4>

      <div class="grid grid-cols-2 gap-4 mb-3">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.historyRange') }}</span></div>
          <select v-model="localConfig.historyRange" class="select select-bordered select-sm w-full">
            <option value="0">{{ $t('common.timeRanges.live') }}</option>
            <option value="15m">{{ $t('common.timeRanges.m15') }}</option>
            <option value="30m">{{ $t('common.timeRanges.m30') }}</option>
            <option value="1h">{{ $t('common.timeRanges.h1') }}</option>
            <option value="3h">{{ $t('common.timeRanges.h3') }}</option>
            <option value="6h">{{ $t('common.timeRanges.h6') }}</option>
            <option value="24h">{{ $t('common.timeRanges.h24') }}</option>
            <option value="7d">{{ $t('common.timeRanges.d7') }}</option>
            <option value="custom">{{ $t('common.timeRanges.custom') }}</option>
          </select>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.maxDataPoints') }}</span></div>
          <input type="number" v-model.number="localConfig.maxPoints" class="input input-bordered input-sm w-full"
            placeholder="100" />
        </label>
      </div>

      <div v-if="localConfig.historyRange === 'custom'"
        class="grid grid-cols-2 gap-4 p-3 mt-2 bg-base-100 rounded-lg border border-base-300">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">{{ $t('common.from') }}</span>
          </div>
          <VueDatePicker v-model="localConfig.customFrom" :is-24="true" auto-apply :preset-dates="presetDates"
            :dark="themeStore.isDarkTheme" format="yyyy-MM-dd HH:mm" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>

        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-semibold text-primary">{{ $t('common.to') }}</span></div>
          <VueDatePicker v-model="localConfig.customTo" :is-24="true" auto-apply :preset-dates="presetDates"
            :dark="themeStore.isDarkTheme" format="yyyy-MM-dd HH:mm" teleport-center>
            <template #input-icon>
              <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
            </template>
          </VueDatePicker>
        </label>
      </div>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('lineChart.config.lineStyleLayout') }}</h4>

      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isSmooth" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('common.smoothCurve') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showArea" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('lineChart.config.fillArea') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isStacked" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('lineChart.config.stackLines') }}</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.use24HourFormat" class="toggle toggle-secondary toggle-sm" />
          <span class="label-text font-semibold">{{ $t('common.use24HourFormat') }}</span>
        </label>
      </div>

      <label class="form-control w-full mt-2">
        <div class="label pb-1">
          <span class="label-text font-semibold">{{ $t('lineChart.config.yAxisLabel') }}</span>
        </div>
        <input type="text" v-model="localConfig.yAxisName" class="input input-bordered input-sm w-full"
          :placeholder="$t('lineChart.config.yAxisPlaceholder')" />
      </label>
    </div>

    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <div class="flex justify-between items-center mb-4">
        <h4 class="font-bold text-sm text-base-content m-0">{{ $t('common.deviceColors') }}</h4>
      </div>

      <div class="flex flex-col gap-2">
        <div v-for="device in activeDevices" :key="device.deviceId"
          class="flex items-center gap-3 p-2 bg-base-100 border border-base-300 rounded-lg">

          <input type="color" v-model="localConfig.deviceColors[device.deviceId]"
            class="h-8 w-12 cursor-pointer rounded border border-base-300 p-0 shrink-0" />

          <span class="text-sm font-semibold flex-1">{{ device.deviceName }}</span>
        </div>

        <div v-if="activeDevices.length === 0" class="text-sm text-base-content/50 py-2 text-center">
          {{ $t('common.noDevicesSelected') }}
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { Icon } from '@iconify/vue';
import { VueDatePicker } from '@vuepic/vue-datepicker';
import { useThemeStore } from '@/stores/useThemeStore';
const themeStore = useThemeStore();

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  selectedDeviceIds: { type: Array, default: () => [] },
  allDevices: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const presetDates = ref([{ label: 'Today', value: new Date() }]);

const activeDevices = computed(() => {
  if (!props.selectedDeviceIds) return [];
  return props.selectedDeviceIds
    .map(id => props.allDevices.find(device => device.deviceId === id))
    .filter(Boolean);
});

const defaultColorPalette = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

const localConfig = ref({
  historyRange: props.modelValue.historyRange || '1h',
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