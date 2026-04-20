<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/api'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'

const router = useRouter()
const route = useRoute()
const { setToken } = useAuth()
const { showToast } = useToast()

const form = reactive({
  email: '',
  password: '',
})

const submitting = ref(false)

const submit = async () => {
  if (!form.email || !form.password) {
    showToast('请输入邮箱和密码')
    return
  }
  submitting.value = true
  try {
    const data = await api.post<{ token: string; expires_in: number }>(
      '/api/v1/auth/login',
      {
        email: form.email,
        password: form.password,
      },
    )
    setToken(data.token)
    showToast('登录成功')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/home'
    router.replace(redirect)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-inner">
      <div class="page-title">账号登录</div>
      <div class="form-group">
        <input v-model.trim="form.email" class="input" type="email" placeholder="邮箱" />
        <input v-model.trim="form.password" class="input" type="password" placeholder="密码" />
      </div>
      <div class="helper-row">
        <button class="btn-text" type="button" @click="router.push('/reset')">忘记密码？</button>
        <button class="btn-text" type="button" @click="router.push('/register')">新用户注册</button>
      </div>
      <div style="margin-top: 24px">
        <button class="btn" type="button" :disabled="submitting" @click="submit">
          <span v-if="submitting" class="spinner"></span>
          登录
        </button>
      </div>
    </div>
  </div>
</template>

