<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, UserMultipleIcon } from '@hugeicons/core-free-icons'
import { starapp } from '../api/client'

const router = useRouter()
const error = ref('')
const creating = ref(false)

const form = reactive({
  displayName: '',
  allowLogin: false,
  username: '',
  password: '',
})

const loginOptions = [
  { label: 'Yes', value: true },
  { label: 'No', value: false },
]

watch(
  () => form.allowLogin,
  (enabled) => {
    if (!enabled) {
      form.username = ''
      form.password = ''
    }
  },
)

const canSubmit = computed(() => {
  if (!form.displayName.trim()) return false
  if (form.allowLogin) {
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
      ...(form.allowLogin
        ? { username: form.username.trim(), password: form.password }
        : {}),
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
      Add someone to your family. A star color is assigned automatically. Turn on login only if this
      person will sign in themselves.
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
      <FormField label="Allow login?" component-has-label>
        <RadioGroup
          v-model="form.allowLogin"
          name="child-allow-login"
          variant="boolean"
          :options="loginOptions"
        />
      </FormField>
      <template v-if="form.allowLogin">
        <FormField
          label="Username"
          for="child-username"
          description="Required when login is enabled."
        >
          <input
            id="child-username"
            v-model="form.username"
            type="text"
            autocomplete="off"
            required
            :disabled="creating"
          />
        </FormField>
        <FormField
          label="Password"
          for="child-password"
          description="At least 8 characters."
        >
          <input
            id="child-password"
            v-model="form.password"
            type="password"
            minlength="8"
            autocomplete="new-password"
            required
            :disabled="creating"
          />
        </FormField>
      </template>
      <template #actions>
        <button type="submit" class="good" :disabled="creating || !canSubmit">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
        <RouterLink :to="{ name: 'familyPeople' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
