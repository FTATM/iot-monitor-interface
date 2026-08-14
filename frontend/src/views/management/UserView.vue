<template>
  <div v-if="hasPermission('User', 'Display')" class="w-full mx-auto p-4">
    <!-- Page Header Card -->
    <div
      class="bg-base-100 shadow-sm rounded-box border border-base-200 p-6 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">

      <div class="flex items-center gap-4">
        <!-- Iconify User Icon -->
        <div class="p-3 bg-primary/10 text-primary rounded-xl flex items-center justify-center">
          <Icon icon="lucide:users" class="w-7 h-7" />
        </div>

        <!-- Header Text -->
        <div>
          <h2 class="m-0 text-2xl font-extrabold text-base-content tracking-tight">User Management</h2>
          <p class="mt-1 mb-0 text-base-content/60 text-sm font-medium">Manage system access and provision new accounts.
          </p>
        </div>
      </div>

    </div>

    <!-- Users Table -->
    <TableData :data="userTable" :columns="tableColumns" :initial-sorting="[{ id: 'userId', desc: false }]"
      :is-loading="isLoading">
      <!-- Toolbar Action Slot -->
      <template #toolbar-actions>
        <button class="btn btn-primary shadow-sm hover:shadow-md transition-all" @click="openCreateModal">
          <Icon icon="lucide:plus" class="w-5 h-5" />
          Add User
        </button>
      </template>

      <!-- Custom Cell Slot: 'deviceId' -->
      <template #cell-userId="{ value }">
        <span class="font-medium text-base-content/50">{{ value }}</span>
      </template>

      <template #cell-firstName="{ row }">
        <span class="font-medium">{{ row.firstName }} {{ row.lastName }}</span>
      </template>

      <template #cell-roleId="{ value }">
        <span class="badge badge-outline badge-primary badge-sm font-semibold">
          {{ getRoleName(value) }}
        </span>
      </template>

      <template #cell-active="{ value }">
        <span :class="['badge', value ? 'badge-success text-success-content' : 'badge-ghost']">
          {{ value ? 'Active' : 'Disabled' }}
        </span>
      </template>


      <!-- Custom Cell Slot: 'actions' -->
      <template #cell-actions="{ row }">
        <div class="flex justify-end gap-2">
          <button @click="openEditModal(row)" class="btn btn-sm btn-primary">
            <Icon icon="lucide:pencil" class="w-5 h-5" />
          </button>
          <button @click="openDeleteModal(row)" class="btn btn-sm btn-error text-white">
            <Icon icon="lucide:trash" class="w-5 h-5" />
          </button>
        </div>
      </template>
    </TableData>
    <!-- DaisyUI Native Dialog Modal -->
    <dialog ref="userModal" class="modal">
      <!-- Increased max-width and removed default padding to build our own header/body/footer -->
      <div class="modal-box sm:w-11/12 sm:max-w-xl p-0 overflow-hidden shadow-2xl">

        <!-- Modal Header -->
        <div class="px-6 py-5 border-b border-base-200 bg-base-100 flex justify-between items-center">
          <h3 class="m-0 text-xl font-extrabold text-base-content">
            {{ isEditing ? 'Edit User Profile' : 'Create New User' }}
          </h3>
          <button class="btn btn-sm btn-circle btn-ghost" @click="closeModal">
            <Icon icon="lucide:x" class="w-4 h-4" />
          </button>
        </div>

        <!-- Modal Body (Form) -->
        <form @submit.prevent="submitForm" autocomplete="off" class="p-6 bg-base-100">

          <!-- Use CSS Grid for perfect alignment -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-5 gap-y-1">

            <!-- First Name -->
            <label class="form-control w-full">
              <div class="label pb-1"><span class="label-text font-semibold">First Name</span></div>
              <input type="text" v-model="form.firstName" placeholder="Jane" @blur="v$.firstName.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.firstName.$error }]" />
              <!-- Fixed height (h-6) prevents the form from jumping when errors appear -->
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.firstName.$error" class="label-text-alt text-error font-medium">{{
                  v$.firstName.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Last Name -->
            <label class="form-control w-full">
              <div class="label pb-1"><span class="label-text font-semibold">Last Name</span></div>
              <input type="text" v-model="form.lastName" placeholder="Doe" @blur="v$.lastName.$touch()"
                :class="['input input-bordered w-full', { 'input-error': v$.lastName.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.lastName.$error" class="label-text-alt text-error font-medium">{{
                  v$.lastName.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Username (Spans both columns) -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">Username</span>
                <span v-if="isEditing" class="badge badge-neutral badge-sm">Read-only</span>
              </div>
              <input type="text" v-model="form.username" placeholder="jdoe" @blur="v$.username.$touch()"
                :disabled="isEditing" autocomplete="none"
                :class="['input input-bordered w-full disabled:bg-base-200/50 disabled:text-base-content/60', { 'input-error': v$.username.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.username.$error" class="label-text-alt text-error font-medium">{{
                  v$.username.$errors[0].$message }}</span>
              </div>
            </label>

            <!-- Password (Spans both columns) -->
            <label class="form-control w-full sm:col-span-2">
              <div class="label pb-1">
                <span class="label-text font-semibold">Password</span>
                <span v-if="isEditing" class="label-text-alt text-info font-medium">Leave blank to keep current</span>
              </div>
              <input type="password" v-model="form.password" @blur="v$.password.$touch()"
                placeholder="Enter secure password" autocomplete="new-password" spellcheck="false"
                :class="['input input-bordered w-full', { 'input-error': v$.password.$error }]" />
              <div class="label px-1 py-1 h-6">
                <span v-if="v$.password.$error" class="label-text-alt text-error font-medium">{{
                  v$.password.$errors[0].$message }}</span>
              </div>
            </label>
          </div>

          <!-- Role Assignment (Searchable Dropdown) -->
          <label class="form-control w-full sm:col-span-2 relative">
            <div class="label pb-1">
              <span class="label-text font-semibold">Assigned Role</span>
              <span class="label-text-alt text-error">*</span>
            </div>

            <SearchableDropdown v-model="form.roleId" :options="Array.from(rolesMaster.values())" label-key="roleName"
              value-key="roleId" placeholder="Search for a role..." :error="v$.roleId.$error"
              @blur="v$.roleId.$touch()" />

            <div class="label px-1 py-1 h-6">
              <span v-if="v$.roleId.$error" class="label-text-alt text-error font-medium">
                {{ v$.roleId.$errors[0].$message }}
              </span>
            </div>
          </label>

          <!-- Redesigned Active Toggle Card -->
          <div class="mt-2 p-5 bg-base-200/40 rounded-xl border border-base-200 flex items-center justify-between">
            <div>
              <p class="font-bold text-base-content m-0 text-sm">Account Status</p>
              <p class="text-xs text-base-content/60 m-0 mt-1">Allow this user to log in and access the system.</p>
            </div>
            <input type="checkbox" v-model="form.active" class="toggle toggle-success toggle-lg" />
          </div>


          <!-- Modal Footer -->
          <div class="border-t border-base-200 mt-6 pt-5 flex justify-end gap-3">
            <button type="button" class="btn btn-ghost" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary px-8">
              {{ isEditing ? 'Save Changes' : 'Create User' }}
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
          Are you sure you want to permanently delete the user <span class="font-bold text-base-content">"{{
            userToDelete?.username }}"</span>? This action cannot be undone.
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
import { required, requiredIf, minLength, helpers } from '@vuelidate/validators';
import { toast } from 'vue3-toastify';
import { Icon } from '@iconify/vue';
import { usePermissionStore } from '@/stores/usePermissionStore';
import NoAccess from '@/components/NoAccess.vue';
import SearchableDropdown from '@/components/SearchableDropdown.vue';
import TableData from '@/components/TableData.vue';


// --- COMPOSABLES ---
const { error: userAddedError, execute: userAddedApi } = useMutation();
const { error: userUpdatedError, execute: userUpdatedApi } = useMutation();
const { data: userAllFetch, isLoading, error: userAllFetchError, execute: userAllFetchApi } = useFetch();
const { error: userDeletedError, isLoading: isDeleting, execute: userDeletedApi } = useMutation();
const { data: roleData, error: roleAllError, execute: roleFetchApi } = useFetch();

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission } = permissionStore;

// --- STATE ---
const userModal = ref(null);
const isEditing = ref(false);
const editingUserId = ref(null);
const userTable = ref([]);
const deleteModal = ref(null);
const userToDelete = ref(null);
const rolesMaster = ref(new Map());

const tableColumns = [
  {
    header: 'ID',
    accessorKey: 'userId',
    meta: { headerClass: 'w-16', cellClass: 'font-bold' }
  },
  {
    header: 'Name',
    accessorKey: 'firstName',
  },
  {
    header: 'Username',
    accessorKey: 'username',
  },
  {
    header: 'Role',
    accessorKey: 'roleId',
  },
  {
    header: 'Active',
    accessorKey: 'active',
  },
  {
    header: 'Actions',
    id: 'actions',
    enableSorting: false,
    meta: { headerClass: 'text-right', cellClass: 'text-right' }
  }
];

const form = ref({
  firstName: '',
  lastName: '',
  username: '',
  password: '',
  active: true,
  roleId: null,
});

// --- VALIDATION RULES ---
const rules = computed(() => ({
  firstName: {
    required: helpers.withMessage('First name is required', required)
  },
  lastName: {
    required: helpers.withMessage('Last name is required', required)
  },
  username: {
    required: helpers.withMessage('Username is required', required),
    minLength: helpers.withMessage('Username must be at least 4 characters', minLength(4))
  },
  password: {
    required: helpers.withMessage(
      'Password is required',
      requiredIf(() => !isEditing.value)
    ),
    validLength: helpers.withMessage('Password must be at least 6 characters', (value) => {
      if (isEditing.value && (!value || value.length === 0)) {
        return true;
      }
      return value && value.length >= 6;
    })
  },
  roleId: {
    required: helpers.withMessage('Selecting a role is required', required)
  }
}));

const v$ = useVuelidate(rules, form);

// --- HELPER METHODS ---
const getRoleName = (id) => {
  const found = rolesMaster.value.has(id) ? rolesMaster.value.get(id) : null;
  return found ? found.roleName : 'No Role';
};

// --- METHODS ---
const openCreateModal = () => {
  isEditing.value = false;
  editingUserId.value = null;
  form.value = { firstName: '', lastName: '', username: '', password: '', active: true, roleId: null };
  v$.value.$reset();
  userModal.value.showModal();
};

const openEditModal = (user) => {
  isEditing.value = true;
  editingUserId.value = user.userId;
  form.value = { firstName: user.firstName, lastName: user.lastName, username: user.username, password: '', active: user.active, roleId: user.roleId };
  v$.value.$reset();
  userModal.value.showModal();
};

const closeModal = () => {
  userModal.value.close();
};

const openDeleteModal = (user) => {
  userToDelete.value = user;
  deleteModal.value.showModal();
};

const closeDeleteModal = () => {
  deleteModal.value.close();
  userToDelete.value = null;
};

const confirmDelete = async () => {
  if (!userToDelete.value) return;

  // Assuming your Go backend uses a DELETE request with the ID in the URL. 
  // Adjust the endpoint if your backend expects the ID in a payload instead.
  await userDeletedApi(`/user/delete/${userToDelete.value.userId}`, null, 'DELETE');

  if (!userDeletedError.value) {
    toast.success(`User ${userToDelete.value.username} deleted successfully!`);
    await loadTable(); // Refresh the table
    closeDeleteModal();
  } else {
    toast.error(userDeletedError.value?.message || "Failed to delete user");
  }
};

const loadTable = async () => {
  await userAllFetchApi('/user/getalldetail')

  if (!userAllFetchError.value && userAllFetch.value) {
    userTable.value = []
    for (let i of userAllFetch.value.data) {
      userTable.value.push({
        userId: i.userId,
        firstName: i.firstName,
        lastName: i.lastName,
        username: i.username,
        active: i.active,
        roleId: i.roleId
      })
    }
  }

}

const loadRoles = async () => {
  await roleFetchApi('/role/getall');
  if (!roleAllError.value && roleData.value) {
    rolesMaster.value.clear();
    for (let i of roleData.value.data) {
      rolesMaster.value.set(i.roleId, i)
    }
  }
  else {
    toast.error(roleAllError.value?.message || "Failed to load Role");
  }
};

const setupData = async () => {
  await loadTable();
  await loadRoles();
}

// --- SUBMIT ACTION ---
const submitForm = async () => {
  const isFormValid = await v$.value.$validate();
  if (!isFormValid) return;

  if (isEditing.value) {
    const payload = { ...form.value, userId: editingUserId.value };
    if (!payload.password) delete payload.password;
    delete payload.username;

    await userUpdatedApi('/user/update', payload, 'PUT');

    if (!userUpdatedError.value) {
      await loadTable();
      closeModal();
      toast.success("User updated successfully!");
    } else {
      toast.error(userUpdatedError.value?.message || "Failed to update user");
    }
  } else {
    await userAddedApi('/user/create', form.value, 'POST');

    if (!userAddedError.value) {
      await loadTable();
      closeModal();
      toast.success("New user created successfully!");
    } else {
      toast.error(userAddedError.value.message || "Failed to create user");
    }
  }
};

onMounted(async () => {
  await setupData();
});
</script>