<script setup lang="ts">
import { onBeforeUnmount, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const { showToast } = useToast()

const form = reactive({
  email: '',
  password: '',
  code: '',
})

const sending = ref(false)
const countdown = ref(0)
const submitting = ref(false)
let timer: number | undefined

const startCountdown = () => {
  countdown.value = 60
  timer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      if (timer) {
        window.clearInterval(timer)
        timer = undefined
      }
    }
  }, 1000)
}

const sendCode = async () => {
  if (!form.email) {
    showToast('请输入邮箱')
    return
  }
  if (sending.value || countdown.value > 0) {
    return
  }
  sending.value = true
  try {
    await api.post('/api/v1/auth/send-code', {
      email: form.email,
      business_type: 1,
    })
    showToast('验证码已发送')
    startCountdown()
  } finally {
    sending.value = false
  }
}

const submit = async () => {
  if (form.password.length < 8) {
    showToast('密码不得少于8位')
    return
  }
  if (!form.email || !form.code) {
    showToast('请完整填写信息')
    return
  }
  submitting.value = true
  try {
    await api.post('/api/v1/auth/register', {
      email: form.email,
      password: form.password,
      code: form.code,
    })
    showToast('注册成功')
    window.setTimeout(() => router.replace('/login'), 3000)
  } finally {
    submitting.value = false
  }
}

onBeforeUnmount(() => {
  if (timer) {
    window.clearInterval(timer)
  }
})
</script>

<template>
  <div class="page">
    <div class="page-inner">
      <div class="header-bar">
        <button class="back-btn" type="button" @click="router.back()">←</button>
        <span class="section-title">返回</span>
      </div>
      <div class="page-title">欢迎注册</div>
      <div class="form-group">
        <input v-model.trim="form.email" class="input" type="email" placeholder="邮箱" />
        <input v-model.trim="form.password" class="input" type="password" placeholder="密码（至少8位）" />
        <div class="input-with-action">
          <input v-model.trim="form.code" class="input" type="text" placeholder="验证码" />
          <button class="text-btn" type="button" :disabled="countdown > 0 || sending" @click="sendCode">
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </button>
        </div>
      </div>
      <div style="margin-top: 24px">
        <button class="btn" type="button" :disabled="submitting" @click="submit">
          <span v-if="submitting" class="spinner"></span>
          注册
        </button>
      </div>
    </div>
  </div>
</template>

