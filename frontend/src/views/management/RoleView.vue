<template>
  <div v-if="hasPermission('Role', 'Display')" class="h-full w-full p-6 bg-base-200 overflow-y-auto">

    <!-- Page Header Card -->
    <div class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:shield" class="w-7 h-7" />
        </div>
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">{{ $t('role.title') }}</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">{{ $t('role.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <TableData 
      :data="roleTable" 
      :columns="tableColumns" 
      :initial-sorting="[{ id: 'roleId', desc: false }]"
      :is-loading="isLoading">
      
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5 stroke-[3]" />
          {{ $t('role.addRole') }}
        </button>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button class="btn btn-sm btn-primary" @click="openEditModal(row)">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>

    <!-- Create / Edit Modal -->
    <dialog ref="roleModal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box p-0 sm:max-w-md overflow-hidden shadow-2xl">
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content flex items-center gap-2">
            <Icon :icon="isEditing ? 'lucide:shield-check' : 'lucide:shield-plus'" class="w-5 h-5 text-primary" />
            {{ isEditing ? $t('role.editRole') : $t('role.createRole') }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal" :disabled="isSaving">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100 flex flex-col gap-4">
          <!-- Role Name Input -->
          <label class="form-control w-full">
            <div class="label pb-1">
              <span class="label-text font-semibold">{{ $t('common.roleName') }}</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <input 
              type="text" 
              v-model="form.roleName" 
              class="input input-bordered w-full"
              :placeholder="$t('role.roleNamePlaceholder')" 
              @blur="v$.roleName.$touch()"
              :class="{ 'input-error': v$.roleName.$error }" 
            />
            <div class="label px-1 py-1 h-6">
              <span v-if="v$.roleName.$error" class="label-text-alt text-error font-medium">
                {{ v$.roleName.$errors[0].$message }}
              </span>
            </div>
          </label>

          <!-- Permissions Checklist -->
          <div class="divider text-sm font-bold text-base-content/50 mt-0 mb-0">{{ $t('role.permissions') }}</div>

          <div v-if="isLoadingMenuAvailable || isLoadingRole" class="flex justify-center py-8">
            <span class="loading loading-spinner loading-lg text-primary"></span>
          </div>

          <div v-else class="flex flex-col gap-4 max-h-64 overflow-y-auto pr-2">
            <div v-for="menu in menuAvailables" :key="menu.menuId" class="bg-base-200/30 p-3 rounded-lg border border-base-200">
              <div class="flex justify-between items-center mb-3 border-b border-base-300 pb-2">
                <h4 class="font-extrabold text-base-content text-lg m-0">{{ menu.menuName }}</h4>
                <button type="button" class="btn btn-xs"
                  :class="isAllSelected(menu) ? 'btn-ghost text-error' : 'btn-outline btn-primary'"
                  @click="toggleSelectAll(menu)">
                  {{ isAllSelected(menu) ? $t('role.deselectAll') : $t('role.selectAll') }}
                </button>
              </div>

              <!-- Flat Actions -->
              <div v-if="menu.availableActions && menu.availableActions.length > 0" class="flex flex-wrap gap-4 pl-2 mb-2">
                <label v-for="action in menu.availableActions" :key="action.actionId" class="cursor-pointer label p-0 flex gap-2">
                  <input type="checkbox" class="checkbox checkbox-sm checkbox-primary"
                    :value="`${menu.menuId}-${action.actionId}`" v-model="form.selectedPermissions" />
                  <span class="label-text font-medium">{{ action.actionName }}</span>
                </label>
              </div>

              <!-- Nested Submenus -->
              <div v-if="menu.submenus && menu.submenus.length > 0" class="flex flex-col gap-3 pl-2 mt-2">
                <div v-for="sub in menu.submenus" :key="sub.menuId">
                  <h5 class="font-semibold text-xs text-base-content/70 mb-1 border-l-2 border-primary pl-2">
                    {{ sub.menuName }}
                  </h5>
                  <div class="flex flex-wrap gap-4 pl-3">
                    <label v-for="action in sub.availableActions" :key="action.actionId" class="cursor-pointer label p-0 flex gap-2">
                      <input type="checkbox" class="checkbox checkbox-sm checkbox-primary"
                        :value="`${sub.menuId}-${action.actionId}`" v-model="form.selectedPermissions" />
                      <span class="label-text font-medium">{{ action.actionName }}</span>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Modal Actions -->
          <div class="modal-action mt-2 border-t border-base-200 pt-5">
            <button type="button" class="btn btn-ghost" @click="closeModal" :disabled="isSaving">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary text-white px-8" :disabled="isSaving">
              <span v-if="isSaving" class="loading loading-spinner loading-sm"></span>
              {{ isEditing ? $t('common.save') : $t('role.createRole') }}
            </button>
          </div>
        </form>
      </div>

      <form method="dialog" class="modal-backdrop" @click="closeModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
  <NoAccess v-else />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { toast } from 'vue3-toastify';
import { useFetch } from '@/composables/useFetch';
import { useMutation } from '@/composables/useMutation';
import { useVuelidate } from '@vuelidate/core';
import { required, helpers } from '@vuelidate/validators';

import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import TableData from '@/components/TableData.vue';

const { t } = useI18n();

const { data: roleAllFetch, isLoading, error: roleAllFetchError, execute: roleAllFetchApi } = useFetch();
const { data: menuAvailableData, isLoading: isLoadingMenuAvailable, execute: menuAvailableFetchApi } = useFetch();
const { data: roleDetailData, isLoading: isLoadingRole, execute: roleDetailFetchApi } = useFetch();
const { error: roleUpsertError, execute: roleUpsertApi } = useMutation();

const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

const roleTable = ref([]);
const menuAvailables = ref([]);
const roleModal = ref(null);
const isEditing = ref(false);
const editingRoleId = ref(null);
const isSaving = ref(false);

const tableColumns = computed(() => [
  { header: t('common.id'), accessorKey: 'roleId', meta: { headerClass: 'w-16', cellClass: 'font-bold' } },
  { header: t('common.roleName'), accessorKey: 'roleName', meta: { headerClass: 'w-20', cellClass: 'font-bold' } },
  { header: t('common.actions'), id: 'actions', enableSorting: false, meta: { headerClass: 'text-right', cellClass: 'text-right' } }
]);

const form = ref({
  roleName: '',
  selectedPermissions: []
});

const rules = computed(() => ({
  roleName: {
    required: helpers.withMessage(t('role.validation.roleNameRequired'), required)
  }
}));

const v$ = useVuelidate(rules, form);

onMounted(async () => {
  await loadData();
});

const loadData = async () => {
  await roleAllFetchApi('/role/getall');
  if (!roleAllFetchError.value && roleAllFetch.value) {
    roleTable.value = roleAllFetch.value.data.map(r => ({
      roleId: r.roleId || r.role_id,
      roleName: r.roleName || r.role_name
    }));
  }

  await menuAvailableFetchApi('/role/getmenuavailable');
  if (menuAvailableData.value) {
    menuAvailables.value = menuAvailableData.value.data;
  }
};

const getAllPermsForMenu = (menu) => {
  const allPerms = [];
  if (menu.availableActions) {
    menu.availableActions.forEach(action => allPerms.push(`${menu.menuId}-${action.actionId}`));
  }
  if (menu.submenus) {
    menu.submenus.forEach(sub => {
      if (sub.availableActions) {
        sub.availableActions.forEach(action => allPerms.push(`${sub.menuId}-${action.actionId}`));
      }
    });
  }
  return allPerms;
};

const isAllSelected = (menu) => {
  const allPerms = getAllPermsForMenu(menu);
  if (allPerms.length === 0) return false;
  return allPerms.every(perm => form.value.selectedPermissions.includes(perm));
};

const toggleSelectAll = (menu) => {
  const allPerms = getAllPermsForMenu(menu);
  if (isAllSelected(menu)) {
    form.value.selectedPermissions = form.value.selectedPermissions.filter(p => !allPerms.includes(p));
  } else {
    allPerms.forEach(p => {
      if (!form.value.selectedPermissions.includes(p)) form.value.selectedPermissions.push(p);
    });
  }
};

const openCreateModal = () => {
  isEditing.value = false;
  editingRoleId.value = null;
  form.value = { roleName: '', selectedPermissions: [] };
  v$.value.$reset();
  roleModal.value.showModal();
};

const openEditModal = async (role) => {
  isEditing.value = true;
  editingRoleId.value = role.roleId;
  form.value = { roleName: role.roleName, selectedPermissions: [] };
  v$.value.$reset();
  roleModal.value.showModal();

  await roleDetailFetchApi(`/role/getdetailbyid/${role.roleId}`);
  if (roleDetailData.value) {
    const detail = roleDetailData.value.data;
    if (detail.rolePermissions) {
      form.value.selectedPermissions = detail.rolePermissions.map(p => `${p.menuId}-${p.actionId}`);
    }
  }
};

const closeModal = () => {
  if (isSaving.value) return;
  roleModal.value.close();
};

const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  isSaving.value = true;
  const formattedPermissions = form.value.selectedPermissions.map(str => {
    const [menuId, actionId] = str.split('-');
    return { menuId: parseInt(menuId), actionId: parseInt(actionId) };
  });

  const payload = {
    roleId: isEditing.value ? editingRoleId.value : 0,
    roleName: form.value.roleName,
    rolePermissions: formattedPermissions
  };

  await roleUpsertApi('/role/upsert', payload, 'POST');

  if (!roleUpsertError.value) {
    toast.success(isEditing.value ? t('common.messages.updated') : t('common.messages.created'));
    await loadData();
    closeModal();
  } else {
    toast.error(roleUpsertError.value.message || t('common.messages.saveError'));
  }

  isSaving.value = false;
};
</script>