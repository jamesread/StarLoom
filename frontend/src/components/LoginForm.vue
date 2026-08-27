<script setup lang="ts">
import { ref } from 'vue'
import Login from 'picocrank/vue/components/Login.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { starapp } from '../api/client'

const emit = defineEmits<{ 'login-success': [] }>()
const loginRef = ref<InstanceType<typeof Login> | null>(null)

async function onLocalLogin(payload: { username: string; password: string }) {
  try {
    const res = await starapp.loginWithUsernameAndPassword(payload)
    if (res.standardResponse?.success) {
      loginRef.value?.resetLocalForm()
      emit('login-success')
      return
    }
    loginRef.value?.setLocalLoginError(res.standardResponse?.message || 'Login failed')
  } catch {
    loginRef.value?.setLocalLoginError('Login failed')
  }
}
</script>

<template>
  <Section title="Login" subtitle="Sign in with your username and password" :padding="true">
    <Login
      ref="loginRef"
      :show-default-tabs="false"
      :custom-tabs="[{ id: 'local', label: 'Username & Password' }]"
      @local-login="onLocalLogin"
    />
  </Section>
</template>
