<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Message } from "@arco-design/web-vue";
import { fetchCaptcha } from "../api/auth";
import { downloadAgentPackageById, fetchCurrentVersion } from "../api/skybase";
import { login } from "../utils/auth";

const router = useRouter();
const route = useRoute();
const loading = ref(false);
const form = reactive({
  username: "admin",
  password: "",
  code: "",
  uuid: "",
  remember: true
});
const captchaImage = ref("");
const captchaLoading = ref(false);

async function loadCaptcha() {
  captchaLoading.value = true;
  try {
    const result = await fetchCaptcha();
    form.code = "";
    form.uuid = result.uuid;
    captchaImage.value = result.img ? `data:image/svg+xml;base64,${result.img}` : "";
  } catch (error) {
    form.uuid = "";
    captchaImage.value = "";
    Message.error(error instanceof Error ? error.message : "Failed to load verification code");
  } finally {
    captchaLoading.value = false;
  }
}

async function handleLogin() {
  if (!form.username.trim() || !form.password.trim() || !form.code.trim()) {
    Message.error("Enter username, password, and verification code");
    return;
  }
  if (!form.uuid) {
    Message.error("Verification code is not ready yet");
    return;
  }

  loading.value = true;
  try {
    await login(form.username.trim(), form.password, form.code.trim(), form.uuid);
    Message.success("Welcome back");
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
    router.push(redirect);
  } catch (error) {
    await loadCaptcha();
    Message.error(error instanceof Error ? error.message : "Sign in failed");
  } finally {
    loading.value = false;
  }
}

async function downloadAgent() {
  try {
    const current = await fetchCurrentVersion();
    if (!current?.id) {
      Message.error("No active package available");
      return;
    }
    window.open(downloadAgentPackageById(current.id), "_blank");
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to locate package");
  }
}

onMounted(() => {
  loadCaptcha();
});
</script>

<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-brand">
        <div class="login-brand__top">
          <img class="login-brand__mark" src="/logo.png" alt="SkyView logo" />
          <div class="login-brand__copy">
            <div class="login-brand__eyebrow">Industrial Visual Data Hub</div>
            <h1>VisionVault</h1>
          </div>
        </div>

        <div class="login-brand__message">
          <div class="login-brand__title">Full Lifecycle Management for Industrial Visual Data</div>
          <p>
            Covering image acquisition, archival workflows, and intelligent retrieval,
            built with high reliability, strong security, and end-to-end traceability for critical factory visual assets.
          </p>
        </div>

        <div class="login-brand__features">
          <div class="login-feature">
            <div class="login-feature__title">Unified Archiving</div>
            <div class="login-feature__text">Consolidate files from multiple sources into a clear and structured retention system.</div>
          </div>
          <div class="login-feature">
            <div class="login-feature__title">Automated Sync</div>
            <div class="login-feature__text">Enable continuous collection on the Agent side to reduce manual effort and transmission risks.</div>
          </div>
          <div class="login-feature">
            <div class="login-feature__title">Security Audit</div>
            <div class="login-feature__text">Combine log tracing and alert coordination to support daily operations and compliance needs.</div>
          </div>
        </div>
      </section>

      <section class="login-panel">
        <div class="login-panel__card">
          <div class="login-panel__header">
            <div class="login-panel__brand">VisionVault</div>
            <div class="login-panel__subtitle">Industrial Visual Data Hub</div>
          </div>

          <a-form layout="vertical" :model="form" class="login-form" @submit.prevent="handleLogin">
            <a-form-item field="username" label="Username">
              <a-input v-model="form.username" size="large" placeholder="Enter username" />
            </a-form-item>
            <a-form-item field="password" label="Password">
              <a-input-password v-model="form.password" size="large" placeholder="Enter password" />
            </a-form-item>
            <a-form-item field="code" label="Verification Code">
              <div class="login-captcha">
                <a-input
                  v-model="form.code"
                  size="large"
                  placeholder="Enter verification code"
                  @press-enter="handleLogin"
                />
                <button
                  class="login-captcha__image"
                  type="button"
                  :disabled="captchaLoading"
                  @click="loadCaptcha"
                >
                  <img v-if="captchaImage" :src="captchaImage" alt="Verification code" />
                  <span v-else>{{ captchaLoading ? "Loading..." : "Retry" }}</span>
                </button>
              </div>
            </a-form-item>
            <div class="login-form__options">
              <a-checkbox class="login-form__checkbox" v-model="form.remember">Remember me</a-checkbox>
            </div>
            <a-button class="login-form__submit" type="primary" long size="large" :loading="loading" @click="handleLogin">
              Sign In
            </a-button>
          </a-form>

          <div class="login-download">
            <span>Agent package</span>
            <a-button class="login-download__button" type="text" @click="downloadAgent">Download latest package</a-button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
:deep(.login-form .arco-input-wrapper),
:deep(.login-form .arco-input-password) {
  border: 1px solid #cfd6df;
  border-radius: 10px;
  background: #ffffff;
}

:deep(.login-form .arco-input-wrapper:hover),
:deep(.login-form .arco-input-password:hover) {
  border-color: #9aa8b6;
}

:deep(.login-form .arco-input-wrapper.arco-input-focus),
:deep(.login-form .arco-input-password.arco-input-focus) {
  border-color: #2f6f99;
  box-shadow: 0 0 0 2px rgba(47, 111, 153, 0.12);
}

.login-captcha {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 132px;
  gap: 12px;
  align-items: center;
}

.login-captcha__image {
  height: 44px;
  padding: 0;
  border: 1px solid var(--color-border-2, #d9dde4);
  border-radius: 10px;
  background: #f7f8fa;
  color: #4e5969;
  cursor: pointer;
  overflow: hidden;
}

.login-captcha__image:disabled {
  cursor: wait;
  opacity: 0.72;
}

.login-captcha__image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.login-captcha__image span {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  font-size: 12px;
  font-weight: 600;
}

:deep(.login-download__button) {
  color: #32485a;
}

:deep(.login-download__button:hover) {
  color: #32485a;
}
</style>
