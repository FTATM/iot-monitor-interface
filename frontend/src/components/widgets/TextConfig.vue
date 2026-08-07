<template>
  <div class="flex flex-col gap-5">
    
    <!-- Content Editor -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Text Content</h4>
      
      <label class="form-control w-full">
        <textarea v-model="localConfig.content"
          class="textarea textarea-bordered w-full min-h-[150px] text-base" 
          placeholder="Type your notes, instructions, or descriptions here..."></textarea>
      </label>
    </div>

    <!-- Formatting & Display Options -->
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">Formatting & Display</h4>
      
      <div class="flex flex-col gap-3 mb-4">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" v-model="localConfig.showHeader" class="toggle toggle-primary toggle-sm" />
          <span class="label-text font-semibold">Show Widget Header Card</span>
        </label>
      </div>

      <div class="grid grid-cols-2 gap-4 mb-4">
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Text Alignment</span>
          </div>
          <select v-model="localConfig.textAlign" class="select select-bordered select-sm w-full">
            <option value="left">Left Align</option>
            <option value="center">Center Align</option>
            <option value="right">Right Align</option>
            <option value="justify">Justify</option>
          </select>
        </label>
        
        <label class="form-control w-full">
          <div class="label pb-1">
            <span class="label-text font-semibold">Font Size (px)</span>
          </div>
          <input type="number" v-model.number="localConfig.fontSize"
            class="input input-bordered input-sm w-full" placeholder="14" />
        </label>
      </div>

      <div class="grid grid-cols-1 gap-4">
        <label class="form-control w-full">
          <div class="label pb-1"><span class="label-text font-bold">Text Color</span></div>
          <input type="color" v-model="localConfig.textColor"
            class="h-10 w-full cursor-pointer rounded border border-base-300 p-0" />
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

// Provide sensible defaults for the Text block
const localConfig = ref({
  content: props.modelValue.content || '',
  showHeader: props.modelValue.showHeader !== undefined ? props.modelValue.showHeader : true,
  textAlign: props.modelValue.textAlign || 'left',
  fontSize: props.modelValue.fontSize !== undefined ? props.modelValue.fontSize : 14,
  textColor: props.modelValue.textColor || '#334155'
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