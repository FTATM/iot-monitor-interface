<template>
  <div class="flex flex-col gap-5">
    
    <!-- Data Structure -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Table Structure</h4>
      
      <label class="form-control w-full">
        <div class="label pb-1">
          <span class="label-text font-semibold">Column Headers</span>
          <span class="label-text-alt">Comma-separated</span>
        </div>
        <input type="text" v-model="localConfig.columns"
          class="input input-bordered input-sm w-full" placeholder="ID, Name, Role, Status" />
      </label>
    </div>

    <!-- UI Toggles -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Display Options</h4>
      
      <div class="flex flex-col gap-3">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isDense" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Dense Spacing (Compact)</span>
        </label>
        
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.isStriped" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Striped Rows (Zebra)</span>
        </label>

        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showRowCount" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Show Row Count in Header</span>
        </label>
      </div>
    </div>

    <!-- Header Colors -->
    <div class="grid grid-cols-2 gap-4">
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Header Background</span></div>
        <input type="color" v-model="localConfig.headerColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
      </label>
      <label class="form-control w-full">
        <div class="label pb-1"><span class="label-text font-bold">Header Text Color</span></div>
        <input type="color" v-model="localConfig.headerTextColor"
          class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
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

// Provide sensible defaults for a Table
const localConfig = ref({
  columns: props.modelValue.columns || 'ID, Date, Description, Status, Amount',
  isStriped: props.modelValue.isStriped !== undefined ? props.modelValue.isStriped : true,
  isDense: props.modelValue.isDense !== undefined ? props.modelValue.isDense : false,
  showRowCount: props.modelValue.showRowCount !== undefined ? props.modelValue.showRowCount : true,
  headerColor: props.modelValue.headerColor || '#f8fafc',
  headerTextColor: props.modelValue.headerTextColor || '#334155'
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