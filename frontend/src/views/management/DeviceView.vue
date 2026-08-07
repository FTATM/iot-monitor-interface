<template>
  <div v-if="hasPermission('Device', 'Display')"  class="w-full mx-auto p-4">
    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">

      <div class="flex items-center gap-4">
        <!-- Iconify Device Icon -->
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:cpu" class="w-7 h-7" />
        </div>

        <!-- Header Text -->
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">Device Management</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">Manage system devices and their configuration
            values.</p>
        </div>
      </div>

      <!-- Action Button with Iconify Plus Icon -->
      <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
        <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
        Add Device
      </button>

    </div>

    <!-- Devices Table -->
    <div class="overflow-x-auto shadow rounded-box bg-base-100 border border-base-200">
      <table class="table w-full text-left">
        <thead class="bg-base-200 text-base-content">
          <tr>
            <th>Device Name</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="device in deviceTable" :key="device.deviceId" class="hover">
            <td class="font-medium">{{ device.deviceName }}</td>
            <td>
              <div class="flex items-center gap-3">

                <!-- 1. The Real-Time Connection Dot -->
                <div class="tooltip" :data-tip="device.isConnected ? 'Online' : 'Offline'">
                  <div class="w-3 h-3 rounded-full"
                    :class="device.isConnected ? 'bg-success shadow-[0_0_8px_rgba(0,255,0,0.5)]' : 'bg-error'">
                  </div>
                </div>

                <!-- 2. The Administrative Status Badge -->
                <div class="badge badge-sm font-semibold uppercase tracking-wider"
                  :class="device.isActive ? 'badge-success' : 'badge-ghost text-base-content/50'">
                  {{ device.isActive ? 'Active' : 'Inactive' }}
                </div>

              </div>
            </td>
            <td>
              <div class="flex gap-2">
                <button @click="openEditModal(device)" class="btn btn-sm btn-primary">
                  <Icon icon="lucide:pencil" class="w-5 h-5" />
                </button>

                <button @click="openDeleteModal(device)" class="btn btn-sm btn-error text-white">
                  <Icon icon="lucide:trash" class="w-5 h-5" />
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="deviceTable.length === 0">
            <td colspan="5" class="text-center text-base-content/50 p-8">No devices found.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- DaisyUI Native Dialog Modal -->
    <dialog ref="deviceModal" class="modal">
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl">

        <!-- Modal Header -->
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? 'Edit Device Configuration' : 'Register New Device' }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Modal Body (Form) -->
        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100">

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-5 gap-y-3">

            <!-- Device Name -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">Device Name</span>
                <span v-if="isEditing" class="badge badge-neutral badge-sm">Read-only</span>
              </div>
              <input type="text" v-model="form.deviceName" placeholder="e.g. Main Sensor" @blur="v$.deviceName.$touch()"
                :disabled="isEditing"
                :class="['input input-bordered w-full disabled:bg-base-200/50 disabled:text-base-content/60', { 'input-error': v$.deviceName.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.deviceName.$error" class="label-text-alt text-error font-medium">
                  {{ v$.deviceName.$errors[0].$message }}
                </span>
              </div>
            </label>

            <!-- Value Data -->
            <!-- <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1"><span class="label-text font-semibold">Initial Value Data</span></div>
              <input type="number" v-model.number="form.valueData" placeholder="0" @blur="v$.valueData.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.valueData.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.valueData.$error" class="label-text-alt text-error font-medium">
                  {{ v$.valueData.$errors[0].$message }}
                </span>
              </div>
            </label> -->

            <!-- ADMINISTRATIVE STATUS (User Controlled) -->
            <div class="sm:col-span-2 p-4 bg-base-200/50 rounded-box border border-base-200">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold block text-sm">Active Status</span>
                  <span class="text-xs text-base-content/60">Toggle to enable or disable monitoring.</span>
                </div>
                <input type="checkbox" v-model="form.isActive" class="toggle toggle-primary" />
              </div>
            </div>

            <!-- NETWORK STATUS (System Managed - Only shows when editing) -->
            <div v-if="isEditing"
              class="sm:col-span-2 p-4 bg-base-200/50 rounded-box border border-base-200 flex flex-col gap-3 mt-1">
              <h4 class="font-bold text-sm text-base-content uppercase tracking-wider m-0">Network Status</h4>

              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold">Connection</span>
                <div class="badge gap-1 border-none shadow-sm"
                  :class="form.isConnected ? 'bg-success text-white' : 'bg-error text-white'">
                  {{ form.isConnected ? 'Online' : 'Offline' }}
                </div>
              </div>

              <!-- Only show Last Seen if the device is currently Offline (isConnected is false) -->
              <div v-if="!form.isConnected" class="flex items-center justify-between">
                <span class="text-sm font-semibold">Last Seen At</span>
                <span class="text-sm font-mono text-base-content/70">
                  {{ formatLastSeen(form.lastSeenAt) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="border-t border-base-200 mt-6 pt-5 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary px-8 text-white">
              {{ isEditing ? 'Save Changes' : 'Register Device' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Clickable backdrop to close -->
      <form method="dialog" class="modal-backdrop">
        <button @click="closeModal">close</button>
      </form>
    </dialog>
    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" />
          Confirm Deletion
        </h3>
        <p class="py-4 text-base-content/80">
          Delete the <span class="font-bold text-base-content">"{{ deviceToDelete?.deviceName }}"</span>?
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">No,
            Cancel</button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span>
            Yes, Delete
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeDeleteModal">close</button>
      </form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useMutation } from '@/composables/useMutation';
import { useFetch } from '@/composables/useFetch';
import { useVuelidate } from '@vuelidate/core';
import { required, numeric, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';

// --- COMPOSABLES ---
const { data: deviceAdded, error: deviceAddedError, execute: deviceAddedApi } = useMutation();
const { data: deviceUpdated, error: deviceUpdatedError, execute: deviceUpdatedApi } = useMutation();
const { data: deviceAllFetch, isLoading, error: deviceAllFetchError, execute: deviceAllFetchApi } = useFetch();
const { error: deviceDeletedError, isLoading: isDeleting, execute: deviceDeletedApi } = useMutation();

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

// --- STATE ---
const deviceModal = ref(null);
const isEditing = ref(false);
const editingDeviceId = ref(null);
const deviceTable = ref([]);
const deleteModal = ref(null);
const deviceToDelete = ref(null);

// Updated form state to map to your boolean database fields
const form = ref({
  deviceName: '',
  // valueData: 0,
  isActive: true,
  isConnected: false,
  lastSeenAt: null
});

// --- HELPER FUNCTIONS ---
const formatLastSeen = (timestamp) => {
  if (!timestamp) return 'Never connected';

  const date = new Date(timestamp);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  });
};

// --- VALIDATION RULES ---
const rules = computed(() => ({
  deviceName: {
    required: helpers.withMessage('Device name is required', required)
  }
}));

const v$ = useVuelidate(rules, form);

// --- METHODS ---
const openCreateModal = () => {
  isEditing.value = false;
  editingDeviceId.value = null;
  form.value = {
    deviceName: '',
    // valueData: 0,
    isActive: true,
    isConnected: false,
    lastSeenAt: null
  };
  v$.value.$reset();
  deviceModal.value.showModal();
};

const openEditModal = (device) => {
  isEditing.value = true;
  editingDeviceId.value = device.deviceId;
  form.value = {
    deviceName: device.deviceName,
    // valueData: device.valueData,
    isActive: device.isActive,
    isConnected: device.isConnected,
    lastSeenAt: device.lastSeenAt
  };
  v$.value.$reset();
  deviceModal.value.showModal();
};

const closeModal = () => {
  deviceModal.value.close();
};


const openDeleteModal = (device) => {
  deviceToDelete.value = device;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  deleteModal.value.close();
  deviceToDelete.value = null;
};

const confirmDelete = async () => {
  if (!deviceToDelete.value) return;

  await deviceDeletedApi(`/device/delete/${deviceToDelete.value.deviceId}`, null, 'DELETE');

  if (!deviceDeletedError.value) {
    toast.success(`Device ${deviceToDelete.value.deviceName} deleted successfully!`);
    await loadTable();
    closeDeleteModal();
  } else {
    toast.error(userDeletedError.value?.message || "Failed to delete device");
  }
};

const loadTable = async () => {
  await deviceAllFetchApi('/device/getalldetail');

  if (!deviceAllFetchError.value && deviceAllFetch.value) {
    deviceTable.value = []
    for (let i of deviceAllFetch.value.data) {
      deviceTable.value.push({
        deviceId: i.deviceId,
        deviceName: i.deviceName,
        valueData: i.valueData,
        isActive: i.isActive,
        isConnected: i.isConnected,
        lastSeenAt: i.lastSeenAt
      });
    }
  }

}

const setupData = async () => {
  await loadTable
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();

  if (!isFormValid) {
    return;
  }

  const payload = {
    deviceName: form.value.deviceName,
    // valueData: Number(form.value.valueData),
    isActive: form.value.isActive
  };

  if (isEditing.value) {
    payload.deviceId = editingDeviceId.value;
    delete payload.deviceName;

    await deviceUpdatedApi('/device/update', payload, 'PUT');

    if (!deviceUpdatedError.value) {
      closeModal();
      toast.success("Device updated successfully!");

      await loadTable();
    } else {
      toast.error(deviceUpdatedError.value?.message || "Failed to update device");
    }

  } else {
    await deviceAddedApi('/device/create', [payload], 'POST');

    if (!deviceAddedError.value) {
      closeModal();
      toast.success("Device create successfully!");

      await loadTable();
    } else {
      toast.error(deviceUpdatedError.value?.message || "Failed to update device");
    }
  }
};


onMounted(async () => {
  await loadTable();
  // await setupData();
});
</script>