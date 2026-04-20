<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, ApiError } from '../api/api'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'

const { subscribed, setSubscribed } = useAuth()
const { showToast } = useToast()

const profile = reactive({
  birthday: '',
  constellation: '',
  gender: '',
  city: '',
  occupation: '',
})

const profileLoading = ref(false)
const profileSaving = ref(false)
const profileHint = ref(false)

const fortune = ref('')
const fortuneDate = ref('')
const fortuneLoading = ref(false)

const subscribeLoading = ref(false)

const isProfileComplete = () => {
  return (
    Boolean(profile.birthday) &&
    Boolean(profile.constellation) &&
    Boolean(profile.gender) &&
    Boolean(profile.city) &&
    Boolean(profile.occupation)
  )
}

const loadProfile = async () => {
  profileLoading.value = true
  try {
    const data = await api.get<typeof profile>('/api/v1/user/profile')
    Object.assign(profile, data)
    profileHint.value = !isProfileComplete()
  } catch {
    profileHint.value = true
  } finally {
    profileLoading.value = false
  }
}

const saveProfile = async () => {
  if (!isProfileComplete()) {
    showToast('请完整填写资料')
    return
  }
  profileSaving.value = true
  try {
    await api.put('/api/v1/user/profile', {
      birthday: profile.birthday,
      constellation: profile.constellation,
      gender: profile.gender,
      city: profile.city,
      occupation: profile.occupation,
    })
    showToast('用户资料已保存')
    profileHint.value = false
  } finally {
    profileSaving.value = false
  }
}

const fetchFortune = async () => {
  fortuneLoading.value = true
  try {
    const data = await api.get<{ date: string; content: string }>('/api/v1/fortune/today')
    fortune.value = data.content
    fortuneDate.value = data.date
  } catch (error) {
    if (error instanceof ApiError && error.code === 4003) {
      profileHint.value = true
    }
  } finally {
    fortuneLoading.value = false
  }
}

const toggleSubscribe = async () => {
  if (subscribeLoading.value) {
    return
  }
  subscribeLoading.value = true
  try {
    if (subscribed.value) {
      await api.post('/api/v1/user/unsubscribe')
      setSubscribed(false)
    } else {
      await api.post('/api/v1/user/subscribe')
      setSubscribed(true)
    }
  } finally {
    subscribeLoading.value = false
  }
}

onMounted(() => {
  loadProfile()
})
</script>

<template>
  <div class="page is-gray">
    <div class="page-inner">
      <div class="page-title">每日运势</div>

      <div class="card">
        <div class="section-title">个人资料</div>
        <p v-if="profileHint" style="color: #999999; margin-bottom: 12px">
          请先完善资料，生成运势需要这些信息
        </p>
        <div class="form-group">
          <input v-model="profile.birthday" class="input" type="date" placeholder="出生日期" />
          <div class="grid-two">
            <input v-model.trim="profile.constellation" class="input" type="text" placeholder="星座" />
            <input v-model.trim="profile.gender" class="input" type="text" placeholder="性别" />
          </div>
          <div class="grid-two">
            <input v-model.trim="profile.city" class="input" type="text" placeholder="城市" />
            <input v-model.trim="profile.occupation" class="input" type="text" placeholder="职业" />
          </div>
        </div>
        <div class="card-footer">
          <button class="btn small" type="button" :disabled="profileSaving" @click="saveProfile">
            <span v-if="profileSaving" class="spinner dark"></span>
            保存
          </button>
        </div>
        <p v-if="profileLoading" style="margin-top: 8px; color: #999999">资料加载中...</p>
      </div>

      <div class="card">
        <div class="section-title">今日运势</div>
        <div v-if="fortune" style="line-height: 1.6">
          <div class="badge" style="margin-bottom: 12px">{{ fortuneDate }}</div>
          <div>{{ fortune }}</div>
        </div>
        <div v-else style="color: #999999">
          <p style="margin-bottom: 12px">今天还没有生成运势，点击下方按钮获取</p>
          <button class="btn" type="button" :disabled="fortuneLoading" @click="fetchFortune">
            <span v-if="fortuneLoading" class="spinner"></span>
            生成今日运势
          </button>
          <p v-if="fortuneLoading" style="margin-top: 12px; color: #999999">获取今天运势中...</p>
        </div>
      </div>

      <div class="card">
        <div class="section-title">订阅服务</div>
        <p style="color: #999999; margin-bottom: 16px">
          订阅后每天早上发送今日运势到你的邮箱
        </p>
        <div style="display: flex; align-items: center; justify-content: space-between">
          <span>{{ subscribed ? '已订阅' : '未订阅' }}</span>
          <button
            class="switch"
            :class="{ 'is-on': subscribed }"
            type="button"
            :disabled="subscribeLoading"
            @click="toggleSubscribe"
          ></button>
        </div>
        <p v-if="subscribeLoading" style="margin-top: 8px; color: #999999">正在更新订阅状态...</p>
      </div>
    </div>
  </div>
</template>

