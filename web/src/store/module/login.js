import { getCommonInfo } from '@/api/user';

export const login = {
  namespaced: true,
  state: {
    init: true,
    commonInfo: {
      login: {
        logo: {},
        loginButtonColor: '#5983FF',
        unifiedAuth: {
          enabled: false,
        },
      },
      home: {},
      tab: {},
      loginEmail: {
        email: {
          status: false,
        },
      },
      register: {
        email: {
          status: false,
        },
      },
      resetPassword: {
        email: {
          status: false,
        },
      },
    },
  },

  mutations: {
    setCommonInfo(state, commonInfo) {
      state.commonInfo = { ...state.commonInfo, ...commonInfo };
    },
  },
  actions: {
    async getCommonInfo({ state, commit }) {
      if (!state.init) return;
      const res = (await getCommonInfo()) || {};
      if (res.code === 0) {
        commit('setCommonInfo', res.data || {});
        state.init = false;
      }
    },
  },
  getters: {
    commonInfo(state) {
      return state.commonInfo;
    },
  },
};
