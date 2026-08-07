<template>
  <div class="flex flex-col gap-5">
    
    <!-- Color Configuration -->
    <div class="grid grid-cols-1 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Main Text Color</span></div>
        <input type="color" v-model="localConfig.valueColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
    </div>

    <!-- Text Formatting -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Text & Value Formatting</h4>
      
      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Prefix</span>
          </div>
          <input type="text" v-model="localConfig.prefix"
            class="input input-bordered input-sm w-full" placeholder="e.g., $" />
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Unit (Suffix)</span>
          </div>
          <input type="text" v-model="localConfig.unit"
            class="input input-bordered input-sm w-full" placeholder="e.g., °C, MB/s" />
        </label>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Font Size (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.fontSize"
            class="input input-bordered input-sm w-full" placeholder="56" />
        </label>

        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Subtext / Description</span>
          </div>
          <input type="text" v-model="localConfig.subtext"
            class="input input-bordered input-sm w-full" placeholder="e.g., Last updated 2m ago" />
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

// Provide sensible defaults for a Status Widget
const localConfig = ref({
  valueColor: props.modelValue.valueColor || '#10b981',
  fontSize: props.modelValue.fontSize !== undefined ? props.modelValue.fontSize : 56,
  prefix: props.modelValue.prefix || '',
  unit: props.modelValue.unit || '',
  subtext: props.modelValue.subtext || ''
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