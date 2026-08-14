<!-- src/components/SearchableDropdown.vue -->
<template>
  <div class="relative w-full">
    <!-- Input Container (Acts like the standard input box but holds tags) -->
    <div 
      :class="[
        'input input-bordered w-full flex flex-wrap gap-1.5 items-center h-auto min-h-[3rem] py-1.5 px-3 cursor-text transition-shadow', 
        { 'input-error': error, 'outline outline-2 outline-offset-2 outline-base-content/20': isOpen }
      ]"
      @click="focusInput"
    >
      <!-- Badges for Multiple Selection -->
      <template v-if="multiple && Array.isArray(modelValue)">
        <span 
          v-for="val in modelValue" 
          :key="val" 
          class="badge badge-primary badge-sm py-3 font-semibold gap-1 z-10"
        >
          {{ getDisplayLabel(val) }}
          <!-- Delete tag button -->
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            class="h-3.5 w-3.5 cursor-pointer hover:text-base-100 transition-colors" 
            fill="none" viewBox="0 0 24 24" stroke="currentColor"
            @click.stop="removeItem(val)"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </span>
      </template>

      <!-- Search Input -->
      <input 
        ref="searchInput"
        type="text" 
        v-model="searchQuery" 
        @focus="isOpen = true" 
        @blur="handleBlur"
        :placeholder="showPlaceholder ? placeholder : ''" 
        class="flex-1 min-w-[60px] bg-transparent outline-none border-none p-0 m-0 h-full text-sm" 
      />
    </div>
    
    <!-- Dropdown List -->
    <ul 
      v-show="isOpen" 
      class="absolute z-50 w-full p-2 mt-1 shadow-xl bg-base-100 rounded-box max-h-48 overflow-y-auto border border-base-200"
    >
      <li v-for="option in filteredOptions" :key="option[valueKey]">
        <a 
          :class="[
            'block px-4 py-2 cursor-pointer rounded-lg font-medium flex justify-between items-center transition-colors',
            isSelected(option) ? 'bg-primary/10 text-primary' : 'hover:bg-base-200'
          ]" 
          @mousedown.prevent="selectOption(option)"
        >
          {{ option[labelKey] }}
          
          <!-- Checkmark for selected items in multiple mode -->
          <svg v-if="multiple && isSelected(option)" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </a>
      </li>
      <li v-if="filteredOptions.length === 0" class="px-4 py-2 text-base-content/50 text-sm">
        No results found for "{{ searchQuery }}"
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

const props = defineProps({
  // modelValue now accepts an Array for multiple selection
  modelValue: {
    type: [String, Number, Array],
    default: null
  },
  // ⚡ NEW: Toggle between single and multiple selection mode
  multiple: {
    type: Boolean,
    default: false
  },
  options: {
    type: Array,
    default: () => []
  },
  labelKey: {
    type: String,
    default: 'name' 
  },
  valueKey: {
    type: String,
    default: 'id'
  },
  placeholder: {
    type: String,
    default: 'Search...'
  },
  error: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:modelValue', 'blur']);

const searchQuery = ref('');
const isOpen = ref(false);
const searchInput = ref(null);

// Focus the hidden input when clicking anywhere on the container wrapper
const focusInput = () => {
  searchInput.value?.focus();
};

// Check if an option is currently selected (handles both single and array)
const isSelected = (option) => {
  if (props.multiple && Array.isArray(props.modelValue)) {
    return props.modelValue.includes(option[props.valueKey]);
  }
  return props.modelValue === option[props.valueKey];
};

// Get the text label for a badge based on its ID
const getDisplayLabel = (val) => {
  const opt = props.options.find(o => o[props.valueKey] === val);
  return opt ? opt[props.labelKey] : val;
};

// Hide the placeholder text if badges are currently being displayed
const showPlaceholder = computed(() => {
  if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length > 0) {
    return false;
  }
  return true;
});

// Filter options based on search query
const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options;
  return props.options.filter(opt => 
    String(opt[props.labelKey]).toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

// Select or Deselect an option
const selectOption = (option) => {
  if (props.multiple) {
    // Array Logic
    let newValue = Array.isArray(props.modelValue) ? [...props.modelValue] : [];
    const val = option[props.valueKey];
    
    if (newValue.includes(val)) {
      newValue = newValue.filter(v => v !== val); // Remove if already selected
    } else {
      newValue.push(val); // Add if not selected
    }
    
    emit('update:modelValue', newValue);
    searchQuery.value = ''; // Clear the search box so they can search the next item
    
    // Notice we do NOT close `isOpen = false` here so they can keep clicking multiple items
  } else {
    // Single Logic
    searchQuery.value = option[props.labelKey];
    isOpen.value = false;
    emit('update:modelValue', option[props.valueKey]);
  }
  
  emit('blur');
};

// Remove a specific badge (Multiple mode only)
const removeItem = (val) => {
  if (props.multiple && Array.isArray(props.modelValue)) {
    const newValue = props.modelValue.filter(v => v !== val);
    emit('update:modelValue', newValue);
    emit('blur');
  }
};

const handleBlur = () => {
  isOpen.value = false;
  
  if (props.multiple) {
    searchQuery.value = ''; // Always clear search text on blur in multiple mode
  } else {
    // Reset to the selected single item text
    const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
    searchQuery.value = selected ? selected[props.labelKey] : '';
  }
  
  emit('blur');
};

watch(
  () => [props.modelValue, props.options],
  () => {
    if (props.multiple) {
      if (!isOpen.value) {
        searchQuery.value = '';
      }
    } else {
      const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
      searchQuery.value = selected ? selected[props.labelKey] : '';
    }
  },
  { immediate: true, deep: true }
);
</script>