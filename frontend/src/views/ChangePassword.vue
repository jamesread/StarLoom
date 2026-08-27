<script setup lang="ts">
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'

const router = useRouter()
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)
const error = ref('')
const success = ref('')

async function save() {
  error.value = ''
  success.value = ''
  if (newPassword.value.length < 8) {
    error.value = 'New password must be at least 8 characters'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  saving.value = true
  try {
    const res = await starapp.changePassword({
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    })
    if (!res.standardResponse?.success) {
      error.value = res.standardResponse?.message || 'Password change failed'
      return
    }
    success.value = 'Password changed'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch {
    error.value = 'Password change failed'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Section title="Change password" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" @click="router.push({ name: 'userControlPanel' })">
        Back
      </button>
    </template>

    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <p v-if="success" class="inline-notification note">{{ success }}</p>

    <FormLayout @submit.prevent="save">
      <FormField label="Current password" for="current-password">
        <input id="current-password" v-model="currentPassword" type="password" autocomplete="current-password" required />
      </FormField>
      <FormField label="New password" for="new-password" description="At least 8 characters.">
        <input id="new-password" v-model="newPassword" type="password" autocomplete="new-password" required />
      </FormField>
      <FormField label="Confirm new password" for="confirm-password">
        <input id="confirm-password" v-model="confirmPassword" type="password" autocomplete="new-password" required />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Change password' }}</button>
      </template>
    </FormLayout>
  </Section>
</template>
