<template>
  <div class="agent-mobile-wrapper right-page-content-body">
    <CommonLayout
      :aside-title="asideTitle"
      :isButton="true"
      :asideWidth="asideWidth"
      @handleBtnClick="handleBtnClick"
      :isBtnDisabled="sessionStatus === 0"
      :class="[chatType === 'webChat' ? 'chatBg' : '']"
      :showAside="showAside"
    >
      <template #aside-content>
        <transition name="fade">
          <div class="explore-aside-app">
            <div
              v-for="(n, i) in historyList"
              class="appList"
              :class="[
                'appList',
                { disabled: sessionStatus === 0 },
                { active: n.active },
              ]"
              @click="historyClick(n)"
              @touchstart="historyClick(n)"
              @mouseenter="mouseEnter(n)"
              @mouseleave="mouseLeave(n)"
              :key="'history' + i"
            >
              <span class="appName">
                <span class="appTag"></span>
                {{ n.title }}
              </span>
              <span
                class="el-icon-delete appDelete"
                v-if="n.hover || n.active"
                @click.stop="deleteConversation(n)"
              ></span>
            </div>
          </div>
        </transition>
      </template>
      <template #main-content>
        <div class="app-content">
          <Chat
            :chatType="'chat'"
            :editForm="editForm"
            :appUrlInfo="appUrlInfo"
            :type="chatType"
            ref="agentChat"
            @reloadList="reloadList"
            @setHistoryStatus="setHistoryStatus"
          />
        </div>
      </template>
    </CommonLayout>
  </div>
</template>
<script>
import CommonLayout from '@/components/exploreContainer.vue';
import Chat from './components/chat.vue';
import { mapGetters, mapActions } from 'vuex';
import {
  getAgentPublishedInfo,
  getOpenurlInfo,
  OpenurlConverList,
  getConversationlist,
} from '@/api/agent';
import { getApiKeyRoot } from '@/api/appspace';
import sseMethod from '@/mixins/sseMethod';
import { MULTIPLE_AGENT, SINGLE_AGENT } from '@/views/agent/constants';
import { getCustomSkillConversationList } from '@/api/templateSquare';
export default {
  components: { CommonLayout, Chat },
  mixins: [sseMethod],
  provide() {
    return {
      getHeaderConfig: this.headerConfig,
    };
  },
  data() {
    return {
      showAside: true,
      asideWidth: '260px',
      apiURL: '',
      asideTitle: this.$t('app.createConversation'),
      assistantId: '',
      historyList: [],
      appUrlInfo: {},
      editForm: {
        assistantId: '',
        category: SINGLE_AGENT,
        avatar: {},
        name: '',
        desc: '',
        prologue: '',
        recommendQuestion: [],
        recommendConfig: {
          recommendEnable: false,
          modelConfig: {
            config: {
              temperature: 0.7,
              temperatureEnable: true,
              topP: 1,
              topPEnable: true,
              frequencyPenalty: 0,
              frequencyPenaltyEnable: true,
              presencePenalty: 0,
              presencePenaltyEnable: true,
              maxTokens: 512,
              maxTokensEnable: true,
            },
            displayName: '',
            model: '',
            modelId: '',
            modelType: '',
            provider: '',
          },
          promptEnable: false,
          prompt: '',
          maxHistory: 3,
        },
      },
      chatType: 'agentChat',
      apiStrategies: {
        agentChat_info: getAgentPublishedInfo,
        webChat_info: getOpenurlInfo,
        agentChat_converstionList: getConversationlist,
        webChat_converstionList: OpenurlConverList,
      },
      uuid: '',
      STORAGE_KEY: 'chatUUID',
    };
  },
  computed: {
    ...mapGetters('app', ['sessionStatus']),
  },
  created() {
    const id = this.$route.query.id || this.$route.params.id;
    if (id) {
      this.assistantId = id;
      this.editForm.assistantId = id;
    }
    if (this.$route.path.includes('/webChat')) {
      this.chatType = 'webChat';
      this.initUUID();
    } else {
      this.chatType = 'agentChat';
    }
    // this.getDetail();
    console.log(this.editForm);

    this.getList();
  },
  mounted() {
    if (!localStorage.getItem(this.STORAGE_KEY)) {
      localStorage.setItem(this.STORAGE_KEY, '');
    }
    window.addEventListener('storage', this.handleStorageEvent);
  },
  beforeDestroy() {
    window.removeEventListener('storage', this.handleStorageEvent);
    this.clearMaxPicNum();
  },
  methods: {
    ...mapActions('app', ['setMaxPicNum', 'clearMaxPicNum']),
    initUUID() {
      const storedUUID = localStorage.getItem('chatUUID');
      this.uuid = storedUUID || this.$guid();
      if (!storedUUID) {
        localStorage.setItem('chatUUID', this.uuid);
      }
    },
    handleStorageEvent(event) {
      if (event.key === this.STORAGE_KEY && !event.newValue) {
        this.clearUUID();
      }
    },
    clearUUID() {
      localStorage.removeItem('chatUUID');
      this.uuid = this.$guid();
      localStorage.setItem('chatUUID', this.uuid);
    },
    reloadList(val) {
      this.getList(val);
    },
    headerConfig() {
      if (!this.uuid) {
        return { headers: { 'X-Client-ID': '' } };
      }
      const config = {
        headers: { 'X-Client-ID': this.uuid },
      };
      return config;
    },
    async getDetail() {
      let res = null;
      let data = null;
      if (this.chatType === 'agentChat') {
        res = await getAgentPublishedInfo({
          assistantId: this.editForm.assistantId,
        });
      } else {
        const config = this.headerConfig();
        res = await getOpenurlInfo(this.assistantId, config);
      }
      if (res.code === 0) {
        if (this.chatType === 'agentChat') {
          data = res.data;
          this.editForm.category = data.category;
        } else {
          data = res.data.assistant;
          this.appUrlInfo = data.appUrlInfo;
        }
        this.editForm.avatar = data.avatar;
        this.editForm.name = data.name;
        this.editForm.desc = data.desc;
        this.editForm.prologue = data.prologue;
        this.setMaxPicNum(data.visionConfig.picNum);
        this.editForm.recommendQuestion = data.recommendQuestion.map(item => ({
          value: item,
        }));
        this.editForm.recommendConfig = data.recommendConfig;
      }
    },
    async getList(noInit) {
      const res = await getCustomSkillConversationList({
        pageNo: 1,
        pageSize: 1000,
      });
      if (res.code === 0) {
        if (res.data.list && res.data.list.length > 0) {
          this.historyList = res.data.list.map(n => {
            return { ...n, hover: false, active: false };
          });
          if (noInit) {
            this.historyList[0].active = true;
          } else {
            this.historyClick[this.historyList[0]];
          }
        } else {
          this.historyList = [];
        }
      } else {
        this.historyList = [];
      }
    },
    setHistoryStatus() {
      this.historyList.forEach(m => {
        m.active = false;
      });
    },
    historyClick(n) {
      //切换对话
      n.hover = true;
      this.$refs['agentChat'].conversationClick(n);
    },
    deleteConversation(n) {
      this.$refs['agentChat'].preDelConversation(n);
    },
    handleBtnClick() {
      //新建对话
      this.$refs['agentChat'].createConversion();
    },
    mouseEnter(n) {
      n.hover = true;
    },
    mouseLeave(n) {
      n.hover = false;
    },
    apiKeyRootUrl() {
      const data = { appId: this.editForm.assistantId, appType: 'agent' };
      getApiKeyRoot(data).then(res => {
        if (res.code === 0) {
          this.apiURL = res.data || '';
        }
      });
    },
    openApiDialog() {
      this.$refs.apiKeyDialog.showDialog();
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/chat.scss';
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}
.chatBg {
  background: linear-gradient(
    1deg,
    rgb(255, 255, 255) 42%,
    rgb(255, 255, 255) 42%,
    rgb(235, 237, 254) 98%,
    rgb(238, 240, 255) 98%
  );
}
.active {
  background-color: $color_opacity !important;
  .appTag {
    background-color: $color !important;
  }
}
.agent-mobile-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  .mobile-menu-btn {
    display: none;
    position: fixed;
    top: 5px;
    z-index: 1001;
    border-radius: 4px;
    padding: 5px 12px;
    cursor: pointer;
    .mobile-menu-img {
      width: 24px;
    }
  }
  .mobile-overlay {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 999;
    transition: opacity 0.3s ease;

    &.show {
      display: block;
    }
  }
}
.explore-aside-app {
  .appList:hover {
    background-color: $color_opacity !important;
  }
  .appList {
    margin: 10px 20px;
    padding: 10px;
    border-radius: 6px;
    margin-bottom: 6px;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
    position: relative;
    .appDelete {
      color: $color;
      margin-right: -5px;
      cursor: pointer;
    }
    .appName {
      display: block;
      max-width: 130px;
      overflow: hidden;
      white-space: nowrap;
      pointer-events: none;
      text-overflow: ellipsis;
      .appTag {
        display: inline-block;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #ccc;
      }
    }
  }
}
.app-content {
  width: 100%;
  height: 100%;
}

// weburl适配移动端样式
::v-deep .chatBg,
::v-deep .explore-container {
  @media (max-width: 768px) {
    .el-aside {
      position: fixed !important;
      top: 0 !important;
      left: 0 !important;
      height: 100vh !important;
      z-index: 1000 !important;
      transition: transform 0.3s ease !important;
      border-radius: 0 !important;
      box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15) !important;
      width: 60vw !important;
      .mobile-menu-open & {
        transform: translateX(0) !important;
      }
    }

    .el-main {
      width: 99% !important;
      padding-top: 16px;
      margin-left: 0 !important;
      .center-editable {
        left: 0;
        right: 0;
      }
      .center-session .history-box {
        padding: 0;
      }
      .session-answer .session-answer-wrapper {
        padding-left: 0;
      }
      .session .session-item {
        padding-right: 0;
      }
      .edtable--wrap {
        z-index: 99;
        .editable--send {
          padding: 5px 12px;
          span img {
            width: 12px;
            height: 12px;
          }
        }
      }
    }
    &.el-container {
      width: 100% !important;
    }
  }
}

@media (max-width: 768px) {
  .agent-mobile-wrapper .mobile-menu-btn {
    display: block;
  }
}
</style>
