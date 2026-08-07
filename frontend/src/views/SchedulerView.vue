<template>
  <div v-if="hasPermission('Scheduler', 'Display')" class="h-full w-full p-6 bg-base-200 overflow-y-auto">

    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:calendar-clock" class="w-7 h-7" />
        </div>

        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">Task Scheduler</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">Schedule and manage automated tasks for your
            devices.</p>
        </div>
      </div>

      <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
        <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
        Add Schedule
      </button>
    </div>

    <!-- Data Table Card -->
    <div class="bg-base-100 shadow-sm rounded-box border border-base-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table w-full">
          <!-- head -->
          <thead class="bg-base-200/50 text-base-content">
            <tr>
              <th>Target Device</th>
              <th>Task / Action</th>
              <th>Type</th>
              <th>Start Time</th>
              <th>Status</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="isSchedulesLoading">
              <td colspan="6" class="text-center py-8 text-base-content/50">
                <span class="loading loading-spinner loading-md"></span>
              </td>
            </tr>
            <tr v-else-if="schedules.length === 0">
              <td colspan="6" class="text-center py-8 text-base-content/50">No schedules found. Create one to get
                started.</td>
            </tr>
            <tr v-for="schedule in schedules" :key="schedule.scheduleId" class="hover:bg-base-200/30 transition-colors">
              <td class="font-semibold">{{ getDeviceName(schedule.deviceId) }}</td>
              <td>{{ schedule.action }}</td>
              <td>
                <span class="badge badge-outline badge-sm uppercase text-[10px] font-bold">
                  {{ schedule.scheduleType === 'one_time' ? 'One Time' : 'Recurring' }}
                </span>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <Icon icon="lucide:clock" class="w-4 h-4 text-base-content/50" />
                  {{ formatDateTimeDisplay(schedule.startTime) }}
                </div>
              </td>
              <td>
                <div class="badge badge-sm font-bold uppercase tracking-wider text-[10px]" :class="{
                  'badge-success text-white': schedule.status === 'completed',
                  'badge-info text-white': schedule.status === 'active',
                  'badge-error text-white': schedule.status === 'cancelled'
                }">
                  {{ schedule.status }}
                </div>
              </td>
              <td class="text-right">
                <button class="btn btn-sm btn-ghost btn-square text-primary hover:bg-primary/10"
                  @click="openEditModal(schedule)">
                  <Icon icon="lucide:pencil" class="w-4 h-4" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <dialog class="modal modal-bottom sm:modal-middle" :class="{ 'modal-open': isModalOpen }">
      <div class="modal-box sm:max-w-lg">
        <h3 class="font-bold text-lg mb-6 flex items-center gap-2">
          <Icon :icon="modalMode === 'create' ? 'lucide:calendar-plus' : 'lucide:calendar-check'"
            class="w-5 h-5 text-primary" />
          {{ modalMode === 'create' ? 'Create New Schedule' : 'Edit Schedule' }}
        </h3>

        <form @submit.prevent="saveSchedule" class="flex flex-col gap-4">

          <!-- Device Dropdown from Master Data -->
          <label class="form-control w-full">
            <div class="label pb-1">
              <span class="label-text font-semibold">Target Device</span>
            </div>
            <select v-model.number="form.deviceId" class="select select-bordered w-full" required>
              <option value="" disabled>Select a device to command...</option>
              <option v-for="device in devices" :key="device.deviceId" :value="device.deviceId">
                {{ device.deviceName }}
              </option>
            </select>
          </label>

          <!-- Action / Task Name -->
          <label class="form-control w-full">
            <div class="label pb-1">
              <span class="label-text font-semibold">Task / Action Command</span>
            </div>
            <input type="text" v-model="form.action" class="input input-bordered w-full"
              placeholder="e.g., reboot, sync_data" required />
          </label>

          <!-- Schedule Type Toggle -->
          <div class="form-control w-full mt-2">
            <div class="label pb-2">
              <span class="label-text font-semibold">Schedule Type</span>
            </div>
            <div class="flex gap-4">
              <label
                class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1">
                <input type="radio" name="scheduleType" value="one_time" class="radio radio-primary radio-sm"
                  v-model="form.scheduleType" />
                <span class="label-text font-medium">One-Time Event</span>
              </label>
              <label
                class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1">
                <input type="radio" name="scheduleType" value="recurring" class="radio radio-primary radio-sm"
                  v-model="form.scheduleType" />
                <span class="label-text font-medium">Recurring (Cron)</span>
              </label>
            </div>
          </div>

          <!-- Recurring Specific Fields (Cron & End Time) -->
          <div v-if="form.scheduleType === 'recurring'"
            class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-2 p-4 bg-base-200/50 border border-base-200 rounded-box">
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">Cron Expression</span>
                <span class="label-text-alt text-base-content/60">e.g., */5 * * * *</span>
              </div>
              <input type="text" v-model="form.cronExpression" class="input input-bordered w-full input-sm"
                placeholder="* * * * *" required />
            </label>

            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">End Time (Optional)</span>
              </div>
              <input type="datetime-local" v-model="form.endTime" class="input input-bordered w-full input-sm" />
            </label>
          </div>

          <!-- Date / Time Picker (Start Time) -->
          <label class="form-control w-full">
            <div class="label pb-1">
              <span class="label-text font-semibold">Start Date & Time</span>
            </div>
            <input type="datetime-local" v-model="form.startTime" class="input input-bordered w-full" required />
          </label>

          <!-- Status Dropdown (Only show when editing) -->
          <label v-if="modalMode === 'edit'" class="form-control w-full">
            <div class="label pb-1">
              <span class="label-text font-semibold">Status</span>
            </div>
            <select v-model="form.status" class="select select-bordered w-full">
              <option value="active">Active</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </label>

          <!-- Modal Actions -->
          <div class="modal-action mt-6">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">Cancel</button>
            <button type="submit" class="btn btn-primary text-white px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isSaving ? 'Saving...' : 'Save Schedule' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Click backdrop to close -->
      <form method="dialog" class="modal-backdrop" @click="closeModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
  <NoAccess v-else/>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';

import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';

// --- API COMPOSABLES ---
const { data: devicesData, execute: fetchDevices } = useFetch();

const { data: schedulesData, isLoading: isSchedulesLoading, execute: fetchSchedules } = useFetch();
const { error: createError, isLoading: isCreating, execute: createScheduleApi } = useMutation();
const { error: updateError, isLoading: isUpdating, execute: updateScheduleApi } = useMutation();

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

// --- STATE ---
const devices = ref([]);
const schedules = ref([]);
const isModalOpen = ref(false);
const modalMode = ref('create');
const isSaving = ref(false);

const form = ref({
  scheduleId: null,
  deviceId: '',
  action: '',
  scheduleType: 'one_time',
  status: 'active',
  startTime: '',
  endTime: '',
  cronExpression: ''
});

// --- LIFECYCLE ---
onMounted(async () => {
  await loadData();
});

// --- METHODS ---
const loadData = async () => {
  // Fetch devices master list
  await fetchDevices('/device/getalldetail');
  if (devicesData.value?.data) {
    devices.value = devicesData.value.data;
  }

  // Fetch all schedules
  await fetchSchedules('/schedule/getalldetail');
  if (schedulesData.value?.data) {
    schedules.value = schedulesData.value.data;
  }
};

const getDeviceName = (id) => {
  const device = devices.value.find(d => d.deviceId === id);
  return device ? device.deviceName : 'Unknown Device';
};

// Formats DB ISO string for the UI Table
const formatDateTimeDisplay = (isoString) => {
  if (!isoString) return '';
  const date = new Date(isoString);
  return date.toLocaleString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  });
};

// Formats DB ISO string for the HTML <input type="datetime-local">
const formatForInput = (isoString) => {
  if (!isoString) return '';
  const date = new Date(isoString);
  // Adjust for local timezone offset so the input displays correctly
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset());
  return date.toISOString().slice(0, 16);
};

// --- MODAL ACTIONS ---
const openCreateModal = () => {
  modalMode.value = 'create';
  form.value = {
    scheduleId: null,
    deviceId: '',
    action: '',
    scheduleType: 'one_time',
    status: 'active',
    startTime: '',
    endTime: '',
    cronExpression: ''
  };
  isModalOpen.value = true;
};

const openEditModal = (schedule) => {
  modalMode.value = 'edit';

  form.value = {
    scheduleId: schedule.scheduleId,
    deviceId: schedule.deviceId,
    action: schedule.action,
    scheduleType: schedule.scheduleType,
    status: schedule.status,
    startTime: formatForInput(schedule.startTime),
    endTime: formatForInput(schedule.endTime),
    cronExpression: schedule.cronExpression || ''
  };
  isModalOpen.value = true;
};

const closeModal = () => {
  if (isSaving.value) return;
  isModalOpen.value = false;
};

const saveSchedule = async () => {
  isSaving.value = true;

  // Build the payload matching your Go struct
  const payload = {
    deviceId: form.value.deviceId,
    action: form.value.action,
    scheduleType: form.value.scheduleType,
    status: form.value.status,
    // Convert local datetime back to UTC ISO for the database
    startTime: new Date(form.value.startTime).toISOString(),
  };

  if (form.value.scheduleType === 'recurring') {
    payload.cronExpression = form.value.cronExpression;
    if (form.value.endTime) {
      payload.endTime = new Date(form.value.endTime).toISOString();
    }
  }

  if (modalMode.value === 'create') {
    await createScheduleApi('/schedule/create', payload, 'POST');

    if (!createError.value) {
      toast.success("Schedule created successfully!");
      await loadData(); // Refresh the table
      closeModal();
    } else {
      toast.error(createError.value.message || "Failed to create schedule.");
    }

  } else {
    // Add the UUID for updates
    payload.scheduleId = form.value.scheduleId;

    await updateScheduleApi('/schedule/update', payload, 'PUT');

    if (!updateError.value) {
      toast.success("Schedule updated successfully!");
      await loadData(); // Refresh the table
      closeModal();
    } else {
      toast.error(updateError.value.message || "Failed to update schedule.");
    }
  }

  isSaving.value = false;
};
</script>