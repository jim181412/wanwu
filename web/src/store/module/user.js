import {
  login,
  getPermission,
  getCommonInfo,
  login2FA2new,
  login2FA2exist,
  login2FA1,
  loginSSOExchange,
} from '@/api/user';
import { fetchOrgs } from '@/api/permission/org';
import { jumpOAuth, redirectUserInfoPage } from '@/utils/util';
import { formatPerms } from '@/router/permission';
import { replaceRouter } from '@/router';

const processLogin = (res, commit, params, authType = 'local') => {
  const orgs = res.data.orgs || [];
  const orgPermission = res.data.orgPermission || {};
  const orgId = orgPermission.org ? orgPermission.org.id : '';
  const { isAdmin, isSystem } = orgPermission || {};

  let permission = {};
  permission.orgPermission = formatPerms(orgPermission.permissions);
  permission.roles = orgPermission.roles || [];

  if (res.code === 0) {
    commit('setUserInfo', {
      uid: res.data.uid,
      userName: res.data.username,
      orgId,
      userCategory: res.data.userCategory,
    });
    commit('setOrgInfo', { orgs });
    commit('setToken', res.data.token);
    commit('setExpiresAt', res.data.expiresAt);
    commit('setAuthType', authType);
    commit('setPermission', {
      ...permission,
      isAdmin,
      isSystem,
      isUpdatePassword: res.data.isUpdatePassword,
    });
    //配置导航用户logo和名称以及欢迎文字
    commit('setCommonInfo', { data: res.data.custom || {} });

    commit('setIs2FA', false);

    if (params.client_id) {
      // 重定向到OAuth页面
      jumpOAuth(params);
      return;
    }

    // 更新权限路由
    replaceRouter(permission.orgPermission);
    // 重定向到修改密码或者有权限的页面
    redirectUserInfoPage(res.data.isUpdatePassword, null, true);
  }
};

export const user = {
  namespaced: true,
  state: {
    userInfo: { uid: '', userName: '', orgId: '' },
    orgInfo: { orgs: [] },
    token: '',
    expiresAt: 0,
    authType: '',
    is2FA: false,
    permission: {},
    commonInfo: {},
    lang: '',
    defaultIcons: {
      agentIcon: '',
      ragIcon: '',
    },
    userAvatar: '',
  },

  mutations: {
    setUserAvatar(state, userAvatar) {
      state.userAvatar = userAvatar;
    },
    setDefaultIcons(state, defaultIcons) {
      state.defaultIcons = { ...state.defaultIcons, ...defaultIcons };
    },
    setUserInfo(state, userInfo) {
      state.userInfo = { ...state.userInfo, ...userInfo };
    },
    setOrgInfo(state, orgInfo) {
      state.orgInfo = { ...state.orgInfo, ...orgInfo };
    },
    setToken(state, token) {
      state.token = token;
    },
    setExpiresAt(state, expiresAt) {
      state.expiresAt = expiresAt;
    },
    setAuthType(state, authType) {
      state.authType = authType;
    },
    setIs2FA(state, is2FA) {
      state.is2FA = is2FA;
    },
    setLang(state, lang) {
      if (lang.code) {
        state.lang = lang.code;
        window.localStorage.setItem('locale', lang.code);
      }
    },
    setPermission(state, permission) {
      state.permission = permission;
    },
    LoginOut(state) {
      state.userInfo = {};
      state.token = '';
      state.authType = '';
      state.permission = {};
      localStorage.setItem('access_cert', JSON.stringify(state));
      window.location.reload();
    },
    setCommonInfo(state, commonInfo) {
      state.commonInfo = { ...state.commonInfo, ...commonInfo };
    },
  },
  actions: {
    async LoginIn({ commit }, { loginInfo, params }) {
      const res = await login(loginInfo);
      processLogin(res, commit, params, 'local');
    },

    async LoginInSSO({ commit }, { ticket, callbackUrl, params }) {
      const res = await loginSSOExchange({ ticket, callbackUrl });
      processLogin(res, commit, params, 'sso');
      return res;
    },

    async LoginIn2FA1({ commit }, loginInfo) {
      const res = await login2FA1(loginInfo);
      if (res.code === 0) {
        commit('setToken', res.data.token);
        commit('setExpiresAt', 0);
        return res.data;
      }
    },

    async LoginIn2FA2({ commit }, { loginInfo, params }) {
      const res = await ('newPassword' in loginInfo &&
      'oldPassword' in loginInfo
        ? login2FA2new(loginInfo)
        : login2FA2exist(loginInfo));
      processLogin(res, commit, params, 'local');
    },

    // 获取权限
    async getPermissionInfo({ commit }) {
      return new Promise(async (resolve, reject) => {
        let res = await getPermission();
        const orgPermission = res.data.orgPermission || {};
        const { isAdmin, isSystem } = orgPermission || {};
        const permissions = {};
        permissions.orgPermission = formatPerms(orgPermission.permissions);
        permissions.roles = orgPermission.roles || [];

        const permission = {
          ...permissions,
          isAdmin,
          isSystem,
          isUpdatePassword: res.data.isUpdatePassword,
        };
        if (res.code === 0) {
          commit('setUserAvatar', res.data.avatar.path);
          commit('setPermission', permission);
          if (res.data.language) commit('setLang', res.data.language);
          replaceRouter(permission.orgPermission || []);
          resolve(permission);
        } else {
          commit('setPermission', {});
          replaceRouter([]);
          reject();
        }

        const orgRes = (await fetchOrgs()) || {};
        if (orgRes.code === 0)
          commit('setOrgInfo', { orgs: orgRes.data.select || [] });
      });
    },

    async LoginOut({ commit }) {
      commit('LoginOut');
    },

    async getOrgInfo({ commit }) {
      const res = (await fetchOrgs()) || {};
      if (res.code === 0) {
        commit('setOrgInfo', { orgs: res.data.select || [] });
      }
    },
    async getCommonInfo({ commit }) {
      const res = (await getCommonInfo()) || {};
      if (res.code === 0) {
        commit('setCommonInfo', { data: res.data || {} });
        // 存储默认图标信息
        const defaultIcons = {
          agentIcon: res.data.defaultIcon.agentIcon || '',
          ragIcon: res.data.defaultIcon.ragIcon || '',
        };
        commit('setDefaultIcons', defaultIcons);
      }
    },
  },
  getters: {
    commonInfo(state) {
      return state.commonInfo;
    },
    lang(state) {
      return state.lang;
    },
    userInfo(state) {
      return state.userInfo;
    },
    orgInfo(state) {
      return state.orgInfo;
    },
    token(state) {
      return state.token;
    },
    expiresAt(state) {
      return state.expiresAt;
    },
    authType(state) {
      return state.authType;
    },
    permission(state) {
      return state.permission;
    },
    defaultIcons(state) {
      return state.defaultIcons;
    },
    userAvatar(state) {
      return state.userAvatar;
    },
  },
};
