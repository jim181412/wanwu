import { useQiankun } from './qiankunUtil';
import Vue from 'vue';
import App from './App.vue';
import router from './router';
import { store } from './store';
import { i18n } from './lang';
import './router/permission';
import './assets/icons';

import ElementUi from 'element-ui';
import moment from 'moment';
import 'element-ui/lib/theme-chalk/index.css';
import '@/style/index.scss';
import { config, basePath } from './utils/config';
import { guid, copy } from '@/utils/util';

const isLocalDebugHost =
  typeof window !== 'undefined' &&
  ['localhost', '127.0.0.1'].includes(window.location.hostname);
const enableVueDevtools =
  process.env.NODE_ENV !== 'production' ||
  process.env.VUE_APP_ENABLE_VUE_DEVTOOLS === 'true' ||
  isLocalDebugHost;

function attachVueDevtools(app) {
  if (!enableVueDevtools || typeof window === 'undefined') {
    return;
  }

  window.__VUE__ = Vue;
  window.__VUE_ROOT__ = app;

  const maxAttempts = 30;
  let attempts = 0;

  const tryAttach = () => {
    attempts += 1;
    const hook = window.__VUE_DEVTOOLS_GLOBAL_HOOK__;

    if (hook && typeof hook.emit === 'function') {
      hook.enabled = true;
      hook.Vue = Vue;
      hook.emit('init', Vue);
      return;
    }

    if (attempts < maxAttempts) {
      window.setTimeout(tryAttach, 1000);
    }
  };

  tryAttach();
}

Vue.use(ElementUi, {
  i18n: (key, value) => i18n.t(key, value), // 根据选的语言切换 Element-ui 的语言
});

Vue.prototype.$config = config || {};
Vue.prototype.$basePath = basePath;
Vue.prototype.$guid = guid;
Vue.prototype.$copy = copy;

Vue.config.productionTip = false;
Vue.config.devtools = enableVueDevtools;

if (
  enableVueDevtools &&
  typeof window !== 'undefined' &&
  window.__VUE_DEVTOOLS_GLOBAL_HOOK__
) {
  window.__VUE_DEVTOOLS_GLOBAL_HOOK__.enabled = true;
}

// 定义时间格式全局过滤器
Vue.filter('dateFormat', function (daraStr, pattern = 'YYYY-MM-DD HH:mm:ss') {
  return moment(daraStr).format(pattern);
});

const vueApp = new Vue({
  router,
  store,
  i18n,
  render: function (h) {
    return h(App);
  },
}).$mount('#app');

attachVueDevtools(vueApp);

/*vueApp.$nextTick(() => {
    useQiankun()
})*/
