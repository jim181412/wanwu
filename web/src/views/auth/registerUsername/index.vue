<template>
  <overview @getCommonInfo="handleCommonInfo">
    <template #default="{ commonInfo }">
      <div class="auth-box">
        <p class="auth-header">
          <span style="font-weight: bold">{{ $t('register.title') }}</span>
        </p>
        <div class="auth-form">
          <el-form ref="form" :model="form" :rules="rules" label-position="top">
            <el-form-item class="auth-form-item" prop="username">
              <img class="auth-icon" src="@/assets/imgs/user.png" alt="" />
              <el-input
                v-model.trim="form.username"
                :placeholder="
                  $t('common.input.placeholder') + $t('register.form.username')
                "
                clearable
              />
            </el-form-item>
            <el-form-item class="auth-form-item" prop="password">
              <img class="auth-icon" src="@/assets/imgs/pwd.png" alt="" />
              <el-input
                v-model.trim="form.password"
                :type="isShowPwd1 ? '' : 'password'"
                class="auth-pwd-input"
                :placeholder="
                  $t('common.input.placeholder') + $t('register.form.password')
                "
                @keyup.enter.native="addByEnterKey"
                clearable
              />
              <img
                v-if="!isShowPwd1"
                class="pwd-icon"
                src="@/assets/imgs/hidePwd.png"
                alt=""
                @click="isShowPwd1 = true"
              />
              <img
                v-else
                class="pwd-icon"
                src="@/assets/imgs/showPwd.png"
                alt=""
                @click="isShowPwd1 = false"
              />
            </el-form-item>
            <el-form-item class="auth-form-item" prop="passwordAgain">
              <img class="auth-icon" src="@/assets/imgs/pwd.png" alt="" />
              <el-input
                v-model.trim="form.passwordAgain"
                :type="isShowPwd2 ? '' : 'password'"
                class="auth-pwd-input"
                :placeholder="
                  $t('common.input.placeholder') +
                  $t('register.form.confirmPassword')
                "
                @keyup.enter.native="addByEnterKey"
                clearable
              />
              <img
                v-if="!isShowPwd2"
                class="pwd-icon"
                src="@/assets/imgs/hidePwd.png"
                alt=""
                @click="isShowPwd2 = true"
              />
              <img
                v-else
                class="pwd-icon"
                src="@/assets/imgs/showPwd.png"
                alt=""
                @click="isShowPwd2 = false"
              />
            </el-form-item>
          </el-form>
          <div class="nav-bt">
            {{ $t('register.askAccount') }}
            <span
              :style="{ color: 'var(--color)', cursor: 'pointer' }"
              @click="$router.push({ path: `/login` })"
            >
              {{ $t('register.login') }}
            </span>
          </div>
          <div class="auth-bt">
            <p
              class="primary-bt"
              :style="`background: ${commonInfo.login.loginButtonColor} !important`"
              @click="doRegister"
            >
              {{ $t('register.button') }}
            </p>
          </div>
          <div class="bottom-text">{{ commonInfo.login.platformDesc }}</div>
        </div>
      </div>
    </template>
  </overview>
</template>

<script>
import overview from '@/views/auth/layout';
import { registerByUsername } from '@/api/user';
import { urlEncrypt } from '@/utils/crypto';

export default {
  components: { overview },
  data() {
    const checkPassword = (rule, value, callback) => {
      const reg =
        /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/;
      if (!reg.test(value)) {
        callback(new Error(this.$t('resetPwd.pwdError')));
        return;
      }
      callback();
    };
    const checkPasswordAgain = (rule, value, callback) => {
      if (value !== this.form.password) {
        callback(new Error(this.$t('resetPwd.differError')));
        return;
      }
      callback();
    };
    return {
      form: {
        username: '',
        password: '',
        passwordAgain: '',
      },
      rules: {
        username: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 20,
            message: this.$t('common.hint.userNameLimit'),
            trigger: 'blur',
          },
          {
            pattern: /^(?!_)[a-zA-Z0-9_.\u4e00-\u9fa5]+$/,
            message: this.$t('common.hint.userName'),
            trigger: 'blur',
          },
        ],
        password: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            validator: checkPassword,
            trigger: 'blur',
          },
        ],
        passwordAgain: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            validator: checkPassword,
            trigger: 'blur',
          },
          {
            validator: checkPasswordAgain,
            trigger: 'blur',
          },
        ],
      },
      isShowPwd1: false,
      isShowPwd2: false,
    };
  },
  methods: {
    handleCommonInfo(commonInfo) {
      if (!commonInfo.register.email.status) {
        this.$router.push({ path: `/login` });
      }
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.doRegister();
      }
    },
    doRegister() {
      this.$refs.form.validate(valid => {
        if (!valid) return;
        registerByUsername({
          username: this.form.username,
          password: urlEncrypt(this.form.password),
        }).then(res => {
          if (res.code === 0) {
            this.$router.push({ path: `/login` });
          }
        });
      });
    },
  },
};
</script>
