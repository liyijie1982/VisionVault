<template>
  <div class="login">
    <div class="login-shell">
      <section class="brand-panel">
        <div class="brand-header">
          <div class="brand-mark" />
          <div class="brand-copy">
            <div class="eyebrow">Industrial Visual Data Hub</div>
            <h1 class="brand-title">VisionVault</h1>
          </div>
        </div>

        <div class="brand-message">
          <div class="brand-message-title">Full Lifecycle Management for Industrial Visual Data</div>
          <p class="brand-description brand-description-full">
            Covering image acquisition, archival workflows, and intelligent retrieval,
            built with high reliability, strong security, and end-to-end traceability for critical factory visual assets.
          </p>
        </div>

        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-title">Unified Archiving</div>
            <div class="feature-text">Consolidate files from multiple sources into a clear and structured retention system.</div>
          </div>
          <div class="feature-item">
            <div class="feature-title">Automated Sync</div>
            <div class="feature-text">Enable continuous collection on the Agent side to reduce manual effort and transmission risks.</div>
          </div>
          <div class="feature-item">
            <div class="feature-title">Security Audit</div>
            <div class="feature-text">Combine log tracing and alert coordination to support daily operations and compliance needs.</div>
          </div>
        </div>

      </section>

      <section class="login-container">
        <div class="login-panel">
          <div class="panel-header">
            <div class="product-title">VisionVault</div>
            <div class="product-subtitle">Industrial Visual Data Hub</div>
          </div>

          <el-form ref="loginForm" :model="loginForm" :rules="loginRules" class="login-form">
            <el-form-item prop="username" class="input-container">
              <el-input
                v-model="loginForm.username"
                type="text"
                auto-complete="off"
                placeholder="Username"
                class="modern-input"
              >
                <svg-icon slot="prefix" icon-class="user-login" />
              </el-input>
            </el-form-item>

            <el-form-item prop="password" class="input-container">
              <el-input
                v-model="loginForm.password"
                type="password"
                auto-complete="off"
                placeholder="Password"
                class="modern-input"
                @keyup.enter.native="handleLogin"
              >
                <svg-icon slot="prefix" icon-class="password" />
              </el-input>
            </el-form-item>

            <el-form-item v-if="captchaEnabled" prop="code" class="input-container">
              <el-row :gutter="12">
                <el-col :span="15">
                  <el-input
                    v-model="loginForm.code"
                    auto-complete="off"
                    placeholder="Verification Code"
                    class="modern-input"
                    @keyup.enter.native="handleLogin"
                  >
                    <svg-icon slot="prefix" icon-class="validCode" />
                  </el-input>
                </el-col>
                <el-col :span="9">
                  <img :src="codeUrl" class="captcha-image" @click="getCode">
                </el-col>
              </el-row>
            </el-form-item>

            <div class="form-options">
              <el-checkbox v-model="loginForm.rememberMe" class="remember-checkbox">Remember me</el-checkbox>
            </div>

            <el-form-item class="login-button-container">
              <button
                class="login-button"
                :class="{'loading': loading}"
                @click.prevent="handleLogin"
              >
                <span v-if="!loading">Sign In</span>
                <span v-else>Signing In...</span>
              </button>
            </el-form-item>

            <div class="agent-download-inline">
              <span class="agent-download-label">Agent Download</span>
              <button
                class="agent-download-link"
                @click.prevent="handleDownload('windows', 'SyncAgent-install-std-x86-1.9.3.3.exe')"
              >
                <svg-icon icon-class="windows" class="download-icon" />
                <span>Standard Edition</span>
              </button>
            </div>
          </el-form>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import { getCodeImg } from '@/api/login'
import Cookies from 'js-cookie'
import { encrypt, decrypt } from '@/utils/jsencrypt'

export default {
  name: 'Login',
  data() {
    return {
      codeUrl: '',
      loginForm: {
        username: '', // "admin",
        password: '', // "admin123",
        rememberMe: false,
        code: '',
        uuid: ''
      },
      loginRules: {
        username: [
          { required: true, trigger: 'blur', message: 'Please enter your username' }
        ],
        password: [
          { required: true, trigger: 'blur', message: 'Please enter your password' }
        ],
        code: [{ required: true, trigger: 'change', message: 'Please enter the verification code' }]
      },
      loading: false,
      // Verification code toggle
      captchaEnabled: true,
      // Registration toggle
      register: false,
      redirect: undefined
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        this.redirect = route.query && route.query.redirect
      },
      immediate: true
    }
  },
  created() {
    this.getCode()
    this.getCookie()
  },
  methods: {
    // Handle Agent download
    handleDownload(type, name) {
      const downloadUrl = '/cube/agent/download/'
      this.download(downloadUrl, { type: type, name: name }, name)
    },

    getCode() {
      getCodeImg().then(res => {
        this.captchaEnabled = res.captchaEnabled === undefined ? true : res.captchaEnabled
        if (this.captchaEnabled) {
          this.codeUrl = 'data:image/gif;base64,' + res.img
          this.loginForm.uuid = res.uuid
        }
      })
    },
    getCookie() {
      const username = Cookies.get('username')
      const password = Cookies.get('password')
      const rememberMe = Cookies.get('rememberMe')
      this.loginForm = {
        username: username === undefined ? this.loginForm.username : username,
        password: password === undefined ? this.loginForm.password : decrypt(password),
        rememberMe: rememberMe === undefined ? false : Boolean(rememberMe)
      }
    },
    handleLogin() {
      this.$refs.loginForm.validate(valid => {
        if (valid) {
          this.loading = true
          if (this.loginForm.rememberMe) {
            Cookies.set('username', this.loginForm.username, { expires: 30 })
            Cookies.set('password', encrypt(this.loginForm.password), { expires: 30 })
            Cookies.set('rememberMe', this.loginForm.rememberMe, { expires: 30 })
          } else {
            Cookies.remove('username')
            Cookies.remove('password')
            Cookies.remove('rememberMe')
          }
          this.$store.dispatch('Login', this.loginForm).then(() => {
            this.$router.push({ path: this.redirect || '/' }).catch(() => {})
          }).catch(() => {
            this.loading = false
            if (this.captchaEnabled) {
              this.getCode()
            }
          })
        }
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss">
.login {
  width: 100%;
  height: 100vh;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background:
    radial-gradient(circle at top left, rgba(45, 94, 170, 0.12), transparent 35%),
    linear-gradient(135deg, #eef3f8 0%, #e5edf5 100%);
  font-family: 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', sans-serif;
  overflow: hidden;
}

.login-shell {
  width: 100%;
  height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(360px, 430px);
  background: rgba(255, 255, 255, 0.9);
  overflow: hidden;
}

.brand-panel {
  padding: 48px 56px;
  background:
    linear-gradient(180deg, rgba(20, 52, 96, 0.96) 0%, rgba(18, 43, 78, 0.92) 100%);
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: #f5f8fc;
}

.brand-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(120deg, rgba(255, 255, 255, 0.08) 0%, transparent 45%),
    linear-gradient(0deg, rgba(255, 255, 255, 0.02), rgba(255, 255, 255, 0.02));
  pointer-events: none;
}

.brand-header,
.brand-mark,
.brand-copy,
.brand-message,
.brand-description-full,
.feature-list {
  position: relative;
  z-index: 1;
}

.brand-header {
  display: flex;
  align-items: flex-start;
  gap: 22px;
}

.brand-mark {
  width: 88px;
  height: 88px;
  flex: 0 0 88px;
  background: url("../assets/logo/sidebar-logo-yhx.png") center/contain no-repeat;
}

.brand-copy {
  min-width: 0;
}

.eyebrow {
  font-size: 14px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(230, 238, 248, 0.72);
  margin: 0 0 10px;
}

.brand-title {
  margin: 0;
  font-size: 42px;
  line-height: 1.08;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.brand-message {
  width: 100%;
  max-width: 760px;
  margin-top: 22px;
  padding-left: 18px;
  border-left: 3px solid rgba(255, 255, 255, 0.24);
}

.brand-message-title {
  font-size: 30px;
  line-height: 1.35;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: #ffffff;
}

.brand-description {
  margin: 0;
  max-width: 100%;
  font-size: 16px;
  line-height: 1.9;
  color: rgba(233, 240, 248, 0.76);
}

.brand-description-full {
  display: block;
  width: 100%;
  margin-top: 12px;
}

.feature-list {
  display: grid;
  gap: 16px;
  margin: 24px 0 0;
}

.feature-item {
  padding: 18px 20px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(6px);
}

.feature-title {
  font-size: 17px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #ffffff;
}

.feature-text {
  font-size: 14px;
  line-height: 1.7;
  color: rgba(233, 240, 248, 0.74);
}

.agent-download-inline {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid #e3eaf2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.agent-download-label {
  font-size: 13px;
  color: #6a7e92;
}

.agent-download-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 14px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #234975, #2d639b);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.agent-download-link:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 18px rgba(35, 73, 117, 0.18);
}

.download-icon {
  width: 16px;
  height: 16px;
}

.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 42px;
  background: linear-gradient(180deg, #fdfefe 0%, #f6f9fc 100%);
}

.login-panel {
  width: 100%;
  max-width: 340px;
}

.panel-header {
  margin-bottom: 28px;
}

.product-title {
  font-size: 34px;
  font-weight: 700;
  line-height: 1.15;
  color: #18314a;
  letter-spacing: 0.01em;
}

.product-subtitle {
  margin-top: 10px;
  font-size: 14px;
  color: #6a7e92;
}

.login-form {
  margin-top: 0;
}

.input-container {
  margin-bottom: 20px;
}

.modern-input {
  height: 48px;
}

.captcha-image {
  display: block;
  width: 100%;
  height: 36px;
  border-radius: 0px;
  border: 1px solid #d9e3ee;
  cursor: pointer;
}

.form-options {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 24px;
}

.remember-checkbox {
  color: #5f7287;

  .el-checkbox__input.is-checked .el-checkbox__inner {
    background-color: #2d639b;
    border-color: #2d639b;
  }

  .el-checkbox__inner {
    border-color: #c6d4e2;
    background: #fff;
  }

  .el-checkbox__label {
    color: #5f7287;
    font-size: 13px;
  }
}

.login-button-container {
  margin-top: 4px;
}

.login-button {
  width: 100%;
  height: 48px;
  background: linear-gradient(135deg, #234975, #2f6ca5);
  border: none;
  border-radius: 12px;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
  box-shadow: 0 14px 28px rgba(32, 75, 121, 0.2);

  &:hover {
    background: linear-gradient(135deg, #1d3f66, #275d8f);
    transform: translateY(-1px);
    box-shadow: 0 18px 32px rgba(32, 75, 121, 0.24);
  }

  &:active {
    transform: translateY(1px);
    box-shadow: 0 8px 18px rgba(32, 75, 121, 0.18);
  }

  &.loading {
    background: linear-gradient(135deg, #5a7ea6, #6a90b8);
    animation: button-pulse 1.5s infinite ease-in-out;
  }
}

@keyframes button-pulse {
  0% { box-shadow: 0 0 0 0 rgba(47, 108, 165, 0.28); }
  70% { box-shadow: 0 0 0 10px rgba(47, 108, 165, 0); }
  100% { box-shadow: 0 0 0 0 rgba(47, 108, 165, 0); }
}

.modern-input {
  ::v-deep .el-input__inner {
    height: 48px;
    line-height: 48px;
    border-radius: 12px;
    border: 1px solid #d8e2ec;
    background: #fff;
    color: #22384d;
    font-size: 14px;
    padding-left: 42px;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  ::v-deep .el-input__inner:focus {
    border-color: #7da3c6;
    box-shadow: 0 0 0 4px rgba(45, 99, 155, 0.08);
  }

  ::v-deep .el-input__prefix {
    left: 14px;
    color: #7a8fa6;
  }
}

@media (max-width: 980px) {
  .login-shell {
    grid-template-columns: 1fr;
    height: 100vh;
  }

  .brand-panel {
    padding: 28px 24px 20px;
  }

  .brand-header {
    gap: 18px;
  }

  .brand-mark {
    width: 74px;
    height: 74px;
    flex-basis: 74px;
  }

  .brand-message {
    max-width: 100%;
    margin-top: 18px;
    padding-left: 14px;
  }

  .brand-message-title {
    font-size: 24px;
  }

  .login-container {
    padding: 24px;
  }
}

@media (max-width: 640px) {
  .brand-title {
    font-size: 36px;
  }

  .brand-header {
    align-items: center;
    gap: 14px;
  }

  .brand-mark {
    width: 64px;
    height: 64px;
    flex-basis: 64px;
  }

  .eyebrow {
    font-size: 12px;
    margin-bottom: 8px;
  }

  .brand-description {
    font-size: 15px;
  }

  .brand-message {
    margin-top: 16px;
    padding-left: 12px;
  }

  .brand-message-title {
    font-size: 20px;
    line-height: 1.45;
  }

  .login-container {
    padding: 18px 16px 20px;
  }

  .product-title {
    font-size: 28px;
  }

  .captcha-image {
    margin-top: 10px;
  }

  .login-panel {
    max-width: none;
  }

  .agent-download-inline {
    flex-direction: column;
    align-items: stretch;
  }

  .agent-download-link {
    justify-content: center;
  }
}

.login ::-webkit-scrollbar {
  width: 6px;
}

.login ::-webkit-scrollbar-thumb {
  background: rgba(122, 143, 166, 0.35);
  border-radius: 999px;
}

.login ::-webkit-scrollbar-track {
  width: 100%;
  background: transparent;
}
</style>
