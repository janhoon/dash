<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, UserPlus, Trash2, Shield, Edit2, Users } from 'lucide-vue-next'
import type { Organization, Member, MembershipRole } from '../types/organization'
import {
  getOrganization,
  updateOrganization,
  deleteOrganization,
  listMembers,
  createInvitation,
  updateMemberRole,
  removeMember,
} from '../api/organizations'
import { useOrganization } from '../composables/useOrganization'

const route = useRoute()
const router = useRouter()
const { fetchOrganizations } = useOrganization()

const orgId = computed(() => route.params.id as string)
const org = ref<Organization | null>(null)
const members = ref<Member[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

// Edit form
const editMode = ref(false)
const editName = ref('')
const editSlug = ref('')
const editLoading = ref(false)
const editError = ref<string | null>(null)

// Invite form
const showInviteForm = ref(false)
const inviteEmail = ref('')
const inviteRole = ref<MembershipRole>('viewer')
const inviteLoading = ref(false)
const inviteError = ref<string | null>(null)
const inviteSuccess = ref<string | null>(null)

// Delete confirmation
const showDeleteConfirm = ref(false)
const deleteLoading = ref(false)

const isAdmin = computed(() => org.value?.role === 'admin')

onMounted(async () => {
  await loadData()
})

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const [orgData, membersData] = await Promise.all([
      getOrganization(orgId.value),
      listMembers(orgId.value),
    ])
    org.value = orgData
    members.value = membersData
    editName.value = orgData.name
    editSlug.value = orgData.slug
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load organization'
  } finally {
    loading.value = false
  }
}

function startEdit() {
  editMode.value = true
  editName.value = org.value?.name || ''
  editSlug.value = org.value?.slug || ''
  editError.value = null
}

function cancelEdit() {
  editMode.value = false
  editError.value = null
}

async function saveEdit() {
  if (!editName.value.trim()) {
    editError.value = 'Name is required'
    return
  }

  editLoading.value = true
  editError.value = null
  try {
    org.value = await updateOrganization(orgId.value, {
      name: editName.value.trim(),
      slug: editSlug.value.trim(),
    })
    editMode.value = false
    await fetchOrganizations()
  } catch (e) {
    editError.value = e instanceof Error ? e.message : 'Failed to update organization'
  } finally {
    editLoading.value = false
  }
}

async function handleInvite() {
  if (!inviteEmail.value.trim()) {
    inviteError.value = 'Email is required'
    return
  }

  inviteLoading.value = true
  inviteError.value = null
  inviteSuccess.value = null
  try {
    const invitation = await createInvitation(orgId.value, {
      email: inviteEmail.value.trim(),
      role: inviteRole.value,
    })
    inviteSuccess.value = `Invitation sent! Token: ${invitation.token}`
    inviteEmail.value = ''
    inviteRole.value = 'viewer'
  } catch (e) {
    inviteError.value = e instanceof Error ? e.message : 'Failed to send invitation'
  } finally {
    inviteLoading.value = false
  }
}

async function handleRoleChange(member: Member, newRole: MembershipRole) {
  try {
    await updateMemberRole(orgId.value, member.user_id, { role: newRole })
    member.role = newRole
  } catch (e) {
    alert(e instanceof Error ? e.message : 'Failed to update role')
  }
}

async function handleRemoveMember(member: Member) {
  if (!confirm(`Remove ${member.email} from this organization?`)) {
    return
  }
  try {
    await removeMember(orgId.value, member.user_id)
    members.value = members.value.filter((m) => m.id !== member.id)
  } catch (e) {
    alert(e instanceof Error ? e.message : 'Failed to remove member')
  }
}

async function handleDelete() {
  deleteLoading.value = true
  try {
    await deleteOrganization(orgId.value)
    await fetchOrganizations()
    router.push('/dashboards')
  } catch (e) {
    alert(e instanceof Error ? e.message : 'Failed to delete organization')
  } finally {
    deleteLoading.value = false
    showDeleteConfirm.value = false
  }
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="org-settings">
    <header class="page-header">
      <button class="btn-back" @click="goBack">
        <ArrowLeft :size="20" />
      </button>
      <div class="header-content">
        <h1>Organization Settings</h1>
        <p v-if="org">{{ org.name }}</p>
      </div>
    </header>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="org" class="settings-content">
      <!-- General Settings -->
      <section class="settings-section">
        <div class="section-header">
          <h2>General</h2>
          <button v-if="isAdmin && !editMode" class="btn btn-secondary btn-sm" @click="startEdit">
            <Edit2 :size="16" />
            Edit
          </button>
        </div>

        <div v-if="editMode" class="edit-form">
          <div class="form-group">
            <label>Organization Name</label>
            <input v-model="editName" type="text" :disabled="editLoading" />
          </div>
          <div class="form-group">
            <label>URL Slug</label>
            <input v-model="editSlug" type="text" :disabled="editLoading" />
          </div>
          <div v-if="editError" class="error-message">{{ editError }}</div>
          <div class="form-actions">
            <button class="btn btn-secondary" @click="cancelEdit" :disabled="editLoading">Cancel</button>
            <button class="btn btn-primary" @click="saveEdit" :disabled="editLoading">
              {{ editLoading ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </div>
        <div v-else class="info-grid">
          <div class="info-item">
            <span class="info-label">Name</span>
            <span class="info-value">{{ org.name }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Slug</span>
            <span class="info-value">{{ org.slug }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Your Role</span>
            <span class="info-value role-badge" :class="org.role">{{ org.role }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Created</span>
            <span class="info-value">{{ new Date(org.created_at).toLocaleDateString() }}</span>
          </div>
        </div>
      </section>

      <!-- Members Section -->
      <section class="settings-section">
        <div class="section-header">
          <h2><Users :size="20" /> Members ({{ members.length }})</h2>
          <button v-if="isAdmin" class="btn btn-primary btn-sm" @click="showInviteForm = !showInviteForm">
            <UserPlus :size="16" />
            Invite
          </button>
        </div>

        <!-- Invite Form -->
        <div v-if="showInviteForm && isAdmin" class="invite-form">
          <div class="form-row">
            <input
              v-model="inviteEmail"
              type="email"
              placeholder="Email address"
              :disabled="inviteLoading"
            />
            <select v-model="inviteRole" :disabled="inviteLoading">
              <option value="viewer">Viewer</option>
              <option value="editor">Editor</option>
              <option value="admin">Admin</option>
            </select>
            <button class="btn btn-primary" @click="handleInvite" :disabled="inviteLoading">
              {{ inviteLoading ? 'Sending...' : 'Send Invite' }}
            </button>
          </div>
          <div v-if="inviteError" class="error-message">{{ inviteError }}</div>
          <div v-if="inviteSuccess" class="success-message">{{ inviteSuccess }}</div>
        </div>

        <!-- Members List -->
        <div class="members-list">
          <div v-for="member in members" :key="member.id" class="member-item">
            <div class="member-avatar">
              {{ (member.name || member.email).charAt(0).toUpperCase() }}
            </div>
            <div class="member-info">
              <span class="member-name">{{ member.name || member.email }}</span>
              <span class="member-email">{{ member.email }}</span>
            </div>
            <div class="member-actions">
              <select
                v-if="isAdmin"
                :value="member.role"
                @change="handleRoleChange(member, ($event.target as HTMLSelectElement).value as MembershipRole)"
                class="role-select"
              >
                <option value="viewer">Viewer</option>
                <option value="editor">Editor</option>
                <option value="admin">Admin</option>
              </select>
              <span v-else class="role-badge" :class="member.role">{{ member.role }}</span>
              <button
                v-if="isAdmin"
                class="btn-icon danger"
                @click="handleRemoveMember(member)"
                title="Remove member"
              >
                <Trash2 :size="16" />
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Danger Zone -->
      <section v-if="isAdmin" class="settings-section danger-zone">
        <div class="section-header">
          <h2><Shield :size="20" /> Danger Zone</h2>
        </div>
        <div class="danger-content">
          <div class="danger-item">
            <div class="danger-info">
              <strong>Delete Organization</strong>
              <p>Permanently delete this organization and all its data. This action cannot be undone.</p>
            </div>
            <button class="btn btn-danger" @click="showDeleteConfirm = true">Delete Organization</button>
          </div>
        </div>
      </section>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal modal-sm">
        <h3>Delete Organization?</h3>
        <p>
          This will permanently delete <strong>{{ org?.name }}</strong> and all its dashboards, panels, and
          settings. This action cannot be undone.
        </p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showDeleteConfirm = false" :disabled="deleteLoading">
            Cancel
          </button>
          <button class="btn btn-danger" @click="handleDelete" :disabled="deleteLoading">
            {{ deleteLoading ? 'Deleting...' : 'Delete Organization' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.org-settings {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 2rem;
}

.btn-back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-back:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.header-content h1 {
  margin: 0 0 0.25rem 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

.header-content p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.error {
  color: var(--accent-danger);
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  padding: 1.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.section-header h2 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-value {
  font-size: 0.875rem;
  color: var(--text-primary);
}

.role-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.role-badge.admin {
  background: rgba(102, 126, 234, 0.15);
  color: var(--accent-primary);
}

.role-badge.editor {
  background: rgba(78, 205, 196, 0.15);
  color: var(--accent-success);
}

.role-badge.viewer {
  background: rgba(255, 159, 67, 0.15);
  color: var(--accent-warning);
}

.edit-form,
.invite-form {
  padding: 1rem;
  background: var(--bg-tertiary);
  border-radius: 8px;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.form-group input,
select {
  width: 100%;
  padding: 0.625rem 0.875rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.form-group input:focus,
select:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.form-row {
  display: flex;
  gap: 0.75rem;
}

.form-row input {
  flex: 1;
}

.form-row select {
  width: 120px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
}

.error-message {
  padding: 0.625rem 0.875rem;
  background: rgba(255, 107, 107, 0.1);
  border: 1px solid rgba(255, 107, 107, 0.3);
  border-radius: 6px;
  color: var(--accent-danger);
  font-size: 0.875rem;
  margin-top: 0.75rem;
}

.success-message {
  padding: 0.625rem 0.875rem;
  background: rgba(78, 205, 196, 0.1);
  border: 1px solid rgba(78, 205, 196, 0.3);
  border-radius: 6px;
  color: var(--accent-success);
  font-size: 0.875rem;
  margin-top: 0.75rem;
  word-break: break-all;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.member-avatar {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-primary);
  border-radius: 50%;
  font-size: 0.875rem;
  font-weight: 600;
  color: white;
  flex-shrink: 0;
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-name {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.member-email {
  display: block;
  font-size: 0.75rem;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.member-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.role-select {
  width: auto;
  padding: 0.375rem 0.5rem;
  font-size: 0.75rem;
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: var(--bg-hover);
}

.btn-icon.danger:hover {
  background: rgba(255, 107, 107, 0.1);
  color: var(--accent-danger);
}

.danger-zone {
  border-color: var(--accent-danger);
}

.danger-zone .section-header h2 {
  color: var(--accent-danger);
}

.danger-content {
  padding: 1rem;
  background: rgba(255, 107, 107, 0.05);
  border-radius: 8px;
}

.danger-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.danger-info strong {
  display: block;
  font-size: 0.875rem;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.danger-info p {
  margin: 0;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.btn-secondary {
  background: var(--bg-tertiary);
  border-color: var(--border-primary);
  color: var(--text-primary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-hover);
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-primary-hover);
}

.btn-danger {
  background: var(--accent-danger);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #e55b5b;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  padding: 1.5rem;
  max-width: 400px;
}

.modal h3 {
  margin: 0 0 0.75rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal p {
  margin: 0 0 1.5rem 0;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
