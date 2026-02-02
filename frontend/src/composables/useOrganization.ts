import { ref, computed, watch } from 'vue'
import type { Organization } from '../types/organization'
import { listOrganizations } from '../api/organizations'

const organizations = ref<Organization[]>([])
const currentOrgId = ref<string | null>(localStorage.getItem('current_org_id'))
const loading = ref(false)
const error = ref<string | null>(null)

const currentOrg = computed(() => {
  if (!currentOrgId.value) return organizations.value[0] || null
  return organizations.value.find((org) => org.id === currentOrgId.value) || organizations.value[0] || null
})

watch(currentOrgId, (newId) => {
  if (newId) {
    localStorage.setItem('current_org_id', newId)
  } else {
    localStorage.removeItem('current_org_id')
  }
})

async function fetchOrganizations() {
  loading.value = true
  error.value = null
  try {
    organizations.value = await listOrganizations()
    // If current org is not in the list anymore, select the first one
    if (currentOrgId.value && !organizations.value.find((org) => org.id === currentOrgId.value)) {
      currentOrgId.value = organizations.value[0]?.id || null
    }
    // If no current org selected, select the first one
    if (!currentOrgId.value && organizations.value.length > 0) {
      currentOrgId.value = organizations.value[0].id
    }
  } catch (e) {
    error.value = 'Failed to load organizations'
    console.error('Failed to load organizations:', e)
  } finally {
    loading.value = false
  }
}

function selectOrganization(orgId: string) {
  currentOrgId.value = orgId
}

function clearOrganizations() {
  organizations.value = []
  currentOrgId.value = null
  localStorage.removeItem('current_org_id')
}

export function useOrganization() {
  return {
    organizations,
    currentOrg,
    currentOrgId,
    loading,
    error,
    fetchOrganizations,
    selectOrganization,
    clearOrganizations,
  }
}
