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
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('scheduler.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('scheduler.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <TableData :data="schedules" :columns="tableColumns" :is-loading="isSchedulesLoading"
      :initial-sorting="[{ id: 'status', desc: false }]">
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
          {{ $t('scheduler.addSchedule') }}
        </button>
      </template>

      <template #cell-target="{ row }">
        <div v-if="row.deviceGroupId" class="flex items-center gap-2">
          <div class="p-1.5 bg-secondary/10 rounded-md">
            <Icon icon="lucide:layers" class="w-4 h-4 text-secondary" />
          </div>
          <span class="font-semibold">{{ getGroupName(row.deviceGroupId) }}</span>
        </div>
        <div v-else-if="row.deviceId" class="flex items-center gap-2">
          <div class="p-1.5 bg-primary/10 rounded-md">
            <Icon icon="lucide:cpu" class="w-4 h-4 text-primary" />
          </div>
          <span class="font-semibold">{{ getDeviceName(row.deviceId) }}</span>
        </div>
        <div v-else class="badge badge-error badge-sm text-white font-bold">
          {{ $t('scheduler.noTarget') }}
        </div>
      </template>

      <template #cell-taskAction="{ value, row }">
        <div v-if="row.deviceGroupId" class="flex flex-col gap-1">
          <span class="text-sm font-medium whitespace-nowrap">
            {{ $t('common.command') }}: <span
              class="font-mono bg-base-200 px-1.5 py-0.5 rounded border border-base-300">{{ value?.command || '-'
              }}</span>
          </span>
          <!-- ⚡ Uses the new helper function to filter out ghost data in the UI -->
          <span v-if="getActiveOverridesCount(row) > 0"
            class="text-[10px] text-base-content/60 font-bold uppercase tracking-wide">
            {{ $t('common.overridesCount', { count: getActiveOverridesCount(row) }) }}
          </span>
        </div>
        <div v-else>
          <span class="font-mono bg-base-200 px-2 py-1 rounded text-sm border border-base-300">{{ value?.command ||
            value || '-' }}</span>
        </div>
      </template>

      <template #cell-scheduleType="{ value }">
        <span class="badge badge-outline badge-sm uppercase text-[10px] font-bold">
          {{ value === 'one_time' ? $t('scheduler.typeOneTime') : $t('scheduler.typeRecurring') }}
        </span>
      </template>

      <template #cell-startTime="{ value }">
        <div class="flex items-center gap-2">
          <Icon icon="lucide:clock" class="w-4 h-4 text-base-content/50" />
          {{ formatDateTimeDisplay(value) }}
        </div>
      </template>

      <template #cell-status="{ value }">
        <div class="badge badge-sm font-bold uppercase tracking-wider text-[10px]" :class="{
          'badge-success text-white': value === 'completed',
          'badge-info text-white': value === 'active',
          'badge-error text-white': value === 'cancelled'
        }">
          {{ value }}
        </div>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button class="btn btn-sm btn-primary" @click="openEditModal(row)">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
          <button class="btn btn-sm btn-error text-white shadow-sm hover:shadow-md transition-all"
            @click="openDeleteModal(row)">
            <Icon icon="lucide:trash-2" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>

    <!-- Create / Edit Modal -->
    <dialog class="modal modal-bottom sm:modal-middle" :class="{ 'modal-open': isModalOpen }">
      <div class="modal-box sm:max-w-lg p-0 flex flex-col max-h-[90vh]">

        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center shrink-0">
          <h3 class="m-0 font-bold text-lg flex items-center gap-2">
            <Icon :icon="modalMode === 'create' ? 'lucide:calendar-plus' : 'lucide:calendar-check'"
              class="w-5 h-5 text-primary" />
            {{ modalMode === 'create' ? $t('scheduler.createSchedule') : $t('scheduler.editSchedule') }}
          </h3>
          <button type="button" class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="saveSchedule" class="flex flex-col flex-1 min-h-0">
          <div class="p-6 flex flex-col gap-2 flex-1 overflow-y-auto min-h-0">

            <div class="form-control w-full">
              <div class="label pb-2">
                <span class="label-text font-semibold">{{ $t('scheduler.targetType') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="device" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="v$.$reset()" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:cpu" class="w-4 h-4" /> {{ $t('common.device') }}
                  </span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="group" class="radio radio-primary radio-sm" v-model="form.targetType"
                    @change="v$.$reset()" />
                  <span class="label-text font-medium flex items-center gap-2">
                    <Icon icon="lucide:layers" class="w-4 h-4" /> {{ $t('common.group') }}
                  </span>
                </label>
              </div>
            </div>

            <div class="form-control w-full mt-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ form.targetType === 'device' ?
                  $t('scheduler.selectTargetDevice')
                  : $t('scheduler.selectTargetGroup') }}</span>
              </div>

              <template v-if="form.targetType === 'device'">
                <SearchableDropdown v-model="form.deviceId" :options="devices" label-key="deviceName"
                  value-key="deviceId" :placeholder="$t('common.searchDevice')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.deviceId.$error" class="label-text-alt text-error font-medium">{{
                    v$.deviceId.$errors[0].$message }}</span>
                </div>
              </template>
              <template v-else>
                <SearchableDropdown v-model="form.deviceGroupId" :options="groups" label-key="groupName"
                  value-key="groupId" :placeholder="$t('common.searchGroup')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.deviceGroupId.$error" class="label-text-alt text-error font-medium">{{
                    v$.deviceGroupId.$errors[0].$message }}</span>
                </div>
              </template>
            </div>

            <!-- Single Device Command -->
            <template v-if="form.targetType === 'device'">
              <label class="form-control w-full">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.taskCommand') }}</span>
                </div>
                <input type="text" v-model="form.taskActionPayload.command"
                  @blur="v$.taskActionPayload.command.$touch()"
                  :class="['input input-bordered w-full', { 'input-error': v$.taskActionPayload.command.$error }]"
                  :placeholder="$t('scheduler.taskCommandPlaceholder')" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.taskActionPayload.command.$error" class="label-text-alt text-error font-medium">{{
                    v$.taskActionPayload.command.$errors[0].$message }}</span>
                </div>
              </label>
            </template>

            <!-- Group Device Command -->
            <template v-if="form.targetType === 'group' && form.deviceGroupId">
              <div class="p-4 bg-base-200/50 rounded-box border border-base-200 mt-2">
                <h4 class="font-bold text-sm mb-3">{{ $t('common.groupCommandSettings') }}</h4>

                <label class="form-control w-full">
                  <div class="label pb-1"><span class="label-text font-semibold text-primary">{{
                    $t('common.baseCommandGroup') }}</span></div>
                  <input type="text" v-model="form.taskActionPayload.command"
                    @blur="v$.taskActionPayload.command.$touch()"
                    :class="['input input-bordered w-full border-primary', { 'input-error': v$.taskActionPayload.command.$error }]"
                    placeholder="e.g., 0" />
                  <div class="label px-1 py-0 h-5">
                    <span v-if="v$.taskActionPayload.command.$error" class="label-text-alt text-error font-medium">{{
                      v$.taskActionPayload.command.$errors[0].$message }}</span>
                  </div>
                </label>

                <div class="flex justify-between items-center mt-2 border-t border-base-200/50 pt-4 mb-2">
                  <span class="label-text font-semibold text-primary">{{ $t('common.enableOverrides') }}</span>
                  <input type="checkbox" v-model="form.enableOverrides" @change="handleOverrideToggle"
                    class="toggle toggle-primary toggle-sm" />
                </div>

                <div v-if="form.enableOverrides" class="flex flex-col gap-2 max-h-48 overflow-y-auto pr-2">
                  <div v-for="device in activeGroupDevices" :key="device.deviceId"
                    class="flex items-center gap-3 p-2 bg-base-100 rounded-lg border border-base-200">
                    <span class="flex-1 text-sm font-semibold truncate" :title="device.deviceName">{{ device.deviceName
                    }}</span>

                    <input type="text" v-model="form.taskActionPayload.deviceOverrides[device.deviceId]"
                      class="input input-bordered input-sm w-32" :placeholder="$t('common.default')" />
                  </div>
                  <div v-if="activeGroupDevices.length === 0"
                    class="text-sm italic text-base-content/50 text-center py-4">
                    {{ $t('common.noDevicesAssigned') }}
                  </div>
                </div>
              </div>
            </template>

            <div class="form-control w-full mt-2">
              <div class="label pb-2">
                <span class="label-text font-semibold">{{ $t('scheduler.scheduleType') }}</span>
              </div>
              <div class="flex gap-4">
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="one_time" class="radio radio-primary radio-sm"
                    v-model="form.scheduleType" />
                  <span class="label-text font-medium">{{ $t('scheduler.oneTimeEvent') }}</span>
                </label>
                <label
                  class="label cursor-pointer justify-start gap-2 bg-base-200/50 p-2 rounded-lg border border-base-200 flex-1 hover:bg-base-200 transition-colors">
                  <input type="radio" value="recurring" class="radio radio-primary radio-sm"
                    v-model="form.scheduleType" />
                  <span class="label-text font-medium">{{ $t('scheduler.recurringCron') }}</span>
                </label>
              </div>
            </div>

            <div v-if="form.scheduleType === 'recurring'"
              class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1 mt-2 p-4 bg-base-200/50 border border-base-200 rounded-box">
              <label class="form-control w-full sm:col-span-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.runFrequency') }}</span>
                </div>
                <select v-model="cronPreset" @change="handleCronPresetChange" class="select select-bordered w-full">
                  <option value="*/5 * * * *">{{ $t('scheduler.freq.min5') }}</option>
                  <option value="*/15 * * * *">{{ $t('scheduler.freq.min15') }}</option>
                  <option value="0 * * * *">{{ $t('scheduler.freq.hour1') }}</option>
                  <option value="0 0 * * *">{{ $t('scheduler.freq.midnight') }}</option>
                  <option value="0 8 * * *">{{ $t('scheduler.freq.am8') }}</option>
                  <option value="custom">{{ $t('scheduler.freq.custom') }}</option>
                </select>
              </label>

              <label v-if="cronPreset === 'custom'" class="form-control w-full sm:col-span-2 mt-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.customCron') }}</span>
                  <span class="label-text-alt text-base-content/60">{{ $t('scheduler.customCronHint') }}</span>
                </div>
                <input type="text" v-model="form.cronExpression" @blur="v$.cronExpression.$touch()"
                  :class="['input input-bordered w-full input-sm', { 'input-error': v$.cronExpression.$error }]"
                  placeholder="* * * * *" />
                <div class="label px-1 py-0 h-5">
                  <span v-if="v$.cronExpression.$error" class="label-text-alt text-error font-medium">{{
                    v$.cronExpression.$errors[0].$message }}</span>
                </div>
              </label>

              <label class="form-control w-full sm:col-span-2 mt-2">
                <div class="label pb-1">
                  <span class="label-text font-semibold">{{ $t('scheduler.endTime') }}</span>
                </div>
                <VueDatePicker v-model="form.endTime" :is-24="true" auto-apply :preset-dates="presetDates"
                  format="yyyy-MM-dd HH:mm" :placeholder="$t('scheduler.selectEndTime')" teleport-center>
                  <template #input-icon>
                    <Icon icon="lucide:calendar-clock" class="w-5 h-5 ml-3 text-base-content/50" />
                  </template>
                </VueDatePicker>
              </label>
            </div>

            <label class="form-control w-full mt-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('scheduler.startTime') }}</span>
              </div>
              <VueDatePicker v-model="form.startTime" :is-24="true" auto-apply :preset-dates="presetDates"
                format="yyyy-MM-dd HH:mm" :placeholder="$t('scheduler.selectStartTime')" teleport-center
                @closed="v$.startTime.$touch()" />
              <div class="label px-1 py-0 h-5">
                <span v-if="v$.startTime.$error" class="label-text-alt text-error font-medium">{{
                  v$.startTime.$errors[0].$message }}</span>
              </div>
            </label>

            <label v-if="modalMode === 'edit'" class="form-control w-full">
              <div class="label pb-1">
                <span class="label-text font-semibold">{{ $t('common.status') }}</span>
              </div>
              <select v-model="form.status" class="select select-bordered w-full">
                <option value="active">Active</option>
                <option value="completed">Completed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </label>
          </div>

          <div class="px-6 py-4 border-t border-base-200 bg-base-100 flex justify-end gap-3 shrink-0">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">{{ $t('common.cancel')
            }}</button>
            <button type="submit" class="btn btn-primary text-white px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isSaving ? $t('scheduler.saving') : $t('scheduler.saveSchedule') }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeModal"><button>close</button></form>
    </dialog>

    <dialog ref="deleteModal" class="modal z-[200]">
      <div class="modal-box">
        <h3 class="font-bold text-lg text-error flex items-center gap-2">
          <Icon icon="lucide:alert-triangle" class="w-6 h-6" /> {{ $t('common.confirmDelete') || 'Confirm Deletion' }}
        </h3>
        <p class="py-4 text-base-content/80">
          {{ $t('scheduler.deleteWarning') || 'Are you sure you want to delete this schedule? This action cannot be undone.' }}
        </p>
        <div class="modal-action">
          <button type="button" @click="closeDeleteModal" class="btn btn-ghost" :disabled="isDeleting">
            {{ $t('common.noCancel') || 'Cancel' }}
          </button>
          <button type="button" @click="confirmDelete" class="btn btn-error text-white" :disabled="isDeleting">
            <span v-if="isDeleting" class="loading loading-spinner loading-sm"></span> 
            {{ $t('common.yesDelete') || 'Delete' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="closeDeleteModal">close</button></form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';

// ⚡ IMPORT VUELIDATE
import { useVuelidate } from '@vuelidate/core';
import { required, requiredIf, helpers } from '@vuelidate/validators';

import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';

import { VueDatePicker } from '@vuepic/vue-datepicker';

const { t } = useI18n();

const { data: devicesData, error: deviceDataError, execute: fetchDevices } = useFetch();
const { data: groupsData, error: groupDataError, execute: fetchGroups } = useFetch();
const { error: deleteError, isLoading: isDeleting, execute: deleteScheduleApi } = useMutation();

const { data: schedulesData, error: schedulesDataError, isLoading: isSchedulesLoading, execute: fetchSchedules } = useFetch();
const { error: createError, execute: createScheduleApi } = useMutation();
const { error: updateError, execute: updateScheduleApi } = useMutation();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const devices = ref([]);
const groups = ref([]);
const schedules = ref([]);
const isModalOpen = ref(false);
const modalMode = ref('create');
const isSaving = ref(false);
const presetDates = ref([{ label: 'Today', value: new Date() }]);
const cronPreset = ref('*/5 * * * *');
const deleteModal = ref(null);
const scheduleToDelete = ref(null);

const tableColumns = computed(() => [
  { header: t('scheduler.table.target'), id: 'target', enableSorting: false },
  { header: t('scheduler.table.taskAction'), accessorKey: 'taskAction' },
  { header: t('scheduler.table.type'), accessorKey: 'scheduleType' },
  { header: t('scheduler.table.startTime'), accessorKey: 'startTime' },
  {
    header: t('common.status'),
    accessorKey: 'status',
    sortingFn: (rowA, rowB) => {
      const priority = { 'active': 1, 'completed': 2, 'cancelled': 3 };

      // Get the weight for each row based on the status string
      const weightA = priority[rowA.original.status] || 99;
      const weightB = priority[rowB.original.status] || 99;

      return weightA - weightB;
    }
  },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  scheduleId: null,
  targetType: 'device',
  deviceId: null,
  deviceGroupId: null,
  enableOverrides: false,
  taskActionPayload: {
    command: '',
    deviceOverrides: {}
  },
  scheduleType: 'one_time',
  status: 'active',
  startTime: null,
  endTime: null,
  cronExpression: '*/5 * * * *'
});

// ⚡ VUELIDATE RULES
const rules = computed(() => ({
  deviceId: {
    requiredIfDevice: helpers.withMessage(t('scheduler.validation.selectDevice'), requiredIf(() => form.value.targetType === 'device'))
  },
  deviceGroupId: {
    requiredIfGroup: helpers.withMessage(t('scheduler.validation.selectGroup'), requiredIf(() => form.value.targetType === 'group'))
  },
  taskActionPayload: {
    command: {
      required: helpers.withMessage(t('scheduler.validation.taskCommandRequired'), required)
    }
  },
  startTime: {
    required: helpers.withMessage(t('scheduler.validation.startTimeRequired'), required)
  },
  cronExpression: {
    requiredIfRecurring: helpers.withMessage(t('scheduler.validation.cronRequired'), requiredIf(() => form.value.scheduleType === 'recurring' && cronPreset.value === 'custom'))
  }
}));

const v$ = useVuelidate(rules, form);

const activeGroupDevices = computed(() => {
  if (form.value.targetType !== 'group' || !form.value.deviceGroupId) return [];
  const group = groups.value.find(g => g.groupId === form.value.deviceGroupId);

  if (group && group.devices) return group.devices;
  if (group && group.deviceIds) return group.deviceIds.map(id => devices.value.find(d => d.deviceId === id)).filter(Boolean);

  return devices.value.filter(d => d.deviceGroupId === form.value.deviceGroupId);
});

const handleOverrideToggle = () => {
  if (!form.value.enableOverrides) {
    form.value.taskActionPayload.deviceOverrides = {};
  }
};

onMounted(async () => {
  await loadData();
});

const loadData = async () => {
  await fetchDevices('/device/getalldetail');
  if (deviceDataError.value) toast.error(deviceDataError.value.message || t('common.messages.loadError'));
  if (devicesData.value?.data) devices.value = devicesData.value.data;

  await fetchGroups('/device/group/getalldetail');
  if (groupDataError.value) toast.error(groupDataError.value.message || t('common.messages.loadError'));
  if (groupsData.value?.data) groups.value = groupsData.value.data;

  await fetchSchedules('/schedule/getalldetail');
  if (schedulesDataError.value) toast.error(schedulesDataError.value.message || t('common.messages.loadError'));
  if (schedulesData.value?.data) schedules.value = schedulesData.value.data;
};

const getDeviceName = (id) => {
  const device = devices.value.find(d => d.deviceId === id);
  return device ? device.deviceName : t('scheduler.unknownDevice');
};

const getGroupName = (id) => {
  const group = groups.value.find(g => g.groupId === id);
  return group ? group.groupName : t('scheduler.unknownGroup');
};

const getActiveOverridesCount = (row) => {
  if (!row.deviceGroupId || !row.taskAction?.deviceOverrides) return 0;

  const group = groups.value.find(g => g.groupId === row.deviceGroupId);
  let activeIds = [];

  // Extract the current valid device IDs for this group
  if (group && group.devices) {
    activeIds = group.devices.map(d => Number(d.deviceId));
  } else if (group && group.deviceIds) {
    activeIds = group.deviceIds.map(id => Number(id));
  } else {
    activeIds = devices.value.filter(d => d.deviceGroupId === row.deviceGroupId).map(d => Number(d.deviceId));
  }

  // Count how many keys in the JSON actually exist in the activeIds array
  let count = 0;
  for (const idStr of Object.keys(row.taskAction.deviceOverrides)) {
    if (activeIds.includes(Number(idStr))) {
      count++;
    }
  }

  return count;
};

const formatDateTimeDisplay = (isoString) => {
  if (!isoString) return '';
  const date = new Date(isoString);
  return date.toLocaleString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  });
};

const openCreateModal = () => {
  modalMode.value = 'create';
  form.value = {
    scheduleId: null,
    targetType: 'device',
    deviceId: null,
    deviceGroupId: null,
    enableOverrides: false,
    taskActionPayload: { command: '', deviceOverrides: {} },
    scheduleType: 'one_time',
    status: 'active',
    startTime: null,
    endTime: null,
    cronExpression: '*/5 * * * *'
  };
  cronPreset.value = '*/5 * * * *';

  v$.value.$reset(); // ⚡ Reset validation state
  isModalOpen.value = true;
};

const openEditModal = (schedule) => {
  modalMode.value = 'edit';
  const isGroup = schedule.deviceGroupId !== null && schedule.deviceGroupId !== undefined;

  let parsedAction = { command: '', deviceOverrides: {} };

  if (typeof schedule.taskAction === 'object' && schedule.taskAction !== null) {
    parsedAction = { ...parsedAction, ...schedule.taskAction };
  } else if (typeof schedule.taskAction === 'string') {
    parsedAction.command = schedule.taskAction;
  }

  if (!parsedAction.deviceOverrides) parsedAction.deviceOverrides = {};

  form.value = {
    scheduleId: schedule.scheduleId,
    targetType: isGroup ? 'group' : 'device',
    deviceId: schedule.deviceId || null,
    deviceGroupId: schedule.deviceGroupId || null,
    enableOverrides: Object.keys(parsedAction.deviceOverrides).length > 0,
    taskActionPayload: parsedAction,
    scheduleType: schedule.scheduleType,
    status: schedule.status,
    startTime: schedule.startTime ? new Date(schedule.startTime) : null,
    endTime: schedule.endTime ? new Date(schedule.endTime) : null,
    cronExpression: schedule.cronExpression || ''
  };

  if (isGroup && form.value.deviceGroupId) {
    const cleanOverrides = {};
    activeGroupDevices.value.forEach(d => {
      if (parsedAction.deviceOverrides[d.deviceId]) {
        cleanOverrides[d.deviceId] = parsedAction.deviceOverrides[d.deviceId];
      }
    });
    form.value.taskActionPayload.deviceOverrides = cleanOverrides;
    form.value.enableOverrides = Object.keys(cleanOverrides).length > 0;
  }

  if (form.value.scheduleType === 'recurring') {
    cronPreset.value = ['*/5 * * * *', '*/15 * * * *', '0 * * * *', '0 0 * * *', '0 8 * * *'].includes(form.value.cronExpression)
      ? form.value.cronExpression
      : 'custom';
  }

  v$.value.$reset(); // ⚡ Reset validation state
  isModalOpen.value = true;
};

const closeModal = () => {
  isModalOpen.value = false;
};

const saveSchedule = async () => {
  // ⚡ Execute Vuelidate check instead of manual if/else blocks
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  if (!form.value.enableOverrides) {
    form.value.taskActionPayload.deviceOverrides = {};
  }

  isSaving.value = true;

  const payload = {
    deviceId: form.value.targetType === 'device' ? Number(form.value.deviceId) : null,
    deviceGroupId: form.value.targetType === 'group' ? Number(form.value.deviceGroupId) : null,
    taskAction: form.value.taskActionPayload,
    scheduleType: form.value.scheduleType,
    status: form.value.status,
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
      toast.success(t('common.messages.created'));
      await loadData();
      closeModal();
    } else {
      toast.error(createError.value.message || t('common.messages.createError'));
    }
  } else {
    payload.scheduleId = form.value.scheduleId;
    await updateScheduleApi('/schedule/update', payload, 'PUT');

    if (!updateError.value) {
      toast.success(t('common.messages.updated'));
      await loadData();
      closeModal();
    } else {
      toast.error(updateError.value.message || t('common.messages.updateError'));
    }
  }

  isSaving.value = false;
};

const handleCronPresetChange = () => {
  if (cronPreset.value !== 'custom') {
    form.value.cronExpression = cronPreset.value;
  } else {
    // This will only run if a USER manually clicks "Custom Expression..."
    form.value.cronExpression = '';
  }
};

// Add these functions near your other modal controls
const openDeleteModal = (schedule) => {
  scheduleToDelete.value = schedule;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  deleteModal.value.close();
  scheduleToDelete.value = null;
};

// The new function that actually fires when they click "Yes, Delete" inside the modal
const confirmDelete = async () => {
  if (!scheduleToDelete.value) return;
  
  await deleteScheduleApi(`/schedule/delete/${scheduleToDelete.value.scheduleId}`, null, 'DELETE');
  
  if (!deleteError.value) {
    toast.success(t('common.messages.deleted') || 'Schedule deleted successfully');
    await loadData();
    closeDeleteModal();
  } else {
    toast.error(deleteError.value?.message || t('common.messages.deleteError') || 'Failed to delete schedule');
  }
};
</script>