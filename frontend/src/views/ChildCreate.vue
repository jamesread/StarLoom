<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, UserMultipleIcon } from '@hugeicons/core-free-icons'
import { starapp } from '../api/client'

const router = useRouter()
const error = ref('')
const creating = ref(false)

const form = reactive({
  displayName: '',
  username: '',
  password: '',
})

const canSubmit = computed(() => {
  if (!form.displayName.trim()) return false
  if (form.password.trim()) {
    if (!form.username.trim()) return false
    if (form.password.length < 8) return false
  }
  return true
})

async function createPerson() {
  creating.value = true
  error.value = ''
  try {
    await starapp.createChildMember({
      displayName: form.displayName.trim(),
      username: form.username.trim(),
      password: form.password,
    })
    router.push({ name: 'familyPeople' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Section title="Add person" :icon="UserMultipleIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyPeople' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>People</span>
      </RouterLink>
    </template>

    <p class="subtle">
      Add someone to your family. A star color is assigned automatically. Login details are optional for
      members who will not sign in themselves.
    </p>

    <FormLayout @submit.prevent="createPerson">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Display name" for="child-display-name">
        <input
          id="child-display-name"
          v-model="form.displayName"
          type="text"
          required
          placeholder="Alex"
          :disabled="creating"
        />
      </FormField>
      <FormField
        label="Username"
        for="child-username"
        description="Optional. Required if you set a password so this person can sign in."
      >
        <input
          id="child-username"
          v-model="form.username"
          type="text"
          autocomplete="off"
          :disabled="creating"
        />
      </FormField>
      <FormField
        label="Password"
        for="child-password"
        description="Optional. Leave blank if this person will not sign in. At least 8 characters when set."
      >
        <input
          id="child-password"
          v-model="form.password"
          type="password"
          minlength="8"
          autocomplete="new-password"
          :disabled="creating"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="creating || !canSubmit">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
        <RouterLink :to="{ name: 'familyPeople' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
