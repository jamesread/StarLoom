<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { UserMultipleIcon } from '@hugeicons/core-free-icons'
import { starapp, type FamilyMember } from '../api/client'

const members = ref<FamilyMember[]>([])
const error = ref('')
const showCreate = ref(false)
const creating = ref(false)
const form = reactive({
  displayName: '',
  username: '',
  password: '',
})

async function load() {
  try {
    const res = await starapp.listMembers()
    members.value = (res.members || []).filter((m) => m.role === 'child')
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function createChild() {
  creating.value = true
  try {
    await starapp.createChildMember({
      displayName: form.displayName.trim(),
      username: form.username.trim(),
      password: form.password,
    })
    form.displayName = ''
    form.username = ''
    form.password = ''
    showCreate.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section title="Children" :icon="UserMultipleIcon">
    <p class="subtle">Add children with their own login accounts.</p>
    <p v-if="error" class="error">{{ error }}</p>
    <button type="button" @click="showCreate = true">Add child</button>

    <table v-if="members.length">
      <thead>
        <tr>
          <th>Name</th>
          <th>Username</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in members" :key="m.id">
          <td>{{ m.displayName }}</td>
          <td>{{ m.username }}</td>
          <td>
            <RouterLink :to="{ name: 'familyChildDetail', params: { id: m.id } }">Manage</RouterLink>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="subtle">No children yet.</p>

    <div v-if="showCreate" class="create-form">
      <FormField label="Display name">
        <input v-model="form.displayName" type="text" required />
      </FormField>
      <FormField label="Username">
        <input v-model="form.username" type="text" required autocomplete="off" />
      </FormField>
      <FormField label="Password">
        <input v-model="form.password" type="password" required minlength="8" autocomplete="new-password" />
      </FormField>
      <button type="button" class="secondary" @click="showCreate = false">Cancel</button>
      <button type="button" :disabled="creating" @click="createChild">Create</button>
    </div>
  </Section>
</template>

<style scoped>
.create-form {
  margin: 1rem 0;
  padding: 1rem;
  border: 1px solid var(--pico-muted-border-color);
  border-radius: var(--pico-border-radius);
}
.error {
  color: var(--pico-del-color);
}
</style>
