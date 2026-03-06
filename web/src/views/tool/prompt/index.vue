<template>
  <div class="mcp-content-box customize">
    <div class="mcp-content">
      <div class="card-search card-search-cust">
        <div style="width: 100%">
          <search-input
            :placeholder="$t('tool.prompt.search')"
            ref="searchInput"
            @handleSearch="fetchList"
          />
        </div>
        <el-button
          type="primary"
          size="small"
          @click="$router.push({ path: `/promptEvaluate` })"
        >
          {{ $t('promptEvaluate.title') }}
        </el-button>
      </div>
      <div class="card-box">
        <div class="card card-item-create">
          <div class="app-card-create" @click="createPrompt">
            <div class="create-img-wrap">
              <img
                class="create-img"
                src="@/assets/imgs/card_create_icon_prompt.svg"
                alt=""
              />
            </div>
            <span>{{ $t('tool.prompt.create') }}</span>
          </div>
        </div>
        <div
          v-if="list && list.length"
          class="card"
          v-for="(item, index) in list"
          :key="index"
          @click.stop="handleClick(item)"
        >
          <div class="card-title">
            <img
              class="card-logo"
              :src="avatarSrc(item.avatar.path, defaultAvatar)"
            />
            <div class="mcp_detailBox">
              <span class="mcp_name">{{ item.name }}</span>
              <span class="mcp_from">
                <label>{{ item.desc }}</label>
              </span>
            </div>
            <el-dropdown placement="bottom">
              <span class="el-dropdown-link">
                <i class="el-icon-more" @click.stop />
              </span>
              <el-dropdown-menu slot="dropdown" style="margin-top: -10px">
                <el-dropdown-item @click.native="editPrompt(item)">
                  {{ $t('common.button.edit') }}
                </el-dropdown-item>
                <el-dropdown-item @click.native="copyPrompt(item)">
                  {{ $t('common.button.copy') }}
                </el-dropdown-item>
                <el-dropdown-item @click.native="handleDelete(item)">
                  {{ $t('common.button.delete') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <div class="card-des">{{ item.prompt || '--' }}</div>
        </div>
      </div>

      <el-empty
        class="noData"
        v-if="!(list && list.length)"
        :description="$t('common.noData')"
      />
    </div>
    <CreatePrompt
      :isCustom="true"
      :type="promptType"
      ref="createPrompt"
      @reload="fetchList"
    />
  </div>
</template>
<script>
import CreatePrompt from '@/components/createApp/createPrompt.vue';
import SearchInput from '@/components/searchInput.vue';
import {
  getCustomPromptList,
  deleteCustomPrompt,
  copyCustomPrompt,
} from '@/api/templateSquare';
import { avatarSrc } from '@/utils/util';
export default {
  components: { SearchInput, CreatePrompt },
  data() {
    return {
      basePath: this.$basePath,
      defaultAvatar: require('@/assets/imgs/prompt.png'),
      list: [],
      promptType: 'create',
    };
  },
  mounted() {
    this.fetchList();
  },
  methods: {
    avatarSrc,
    fetchList() {
      const searchInput = this.$refs.searchInput;
      const params = {
        name: searchInput.value,
      };
      getCustomPromptList(params).then(res => {
        this.list = res.data.list || [];
      });
    },
    handleClick(item) {
      this.promptType = 'detail';
      this.showPromptDialog(item);
    },
    showPromptDialog(item) {
      this.$refs.createPrompt.openDialog(item);
    },
    createPrompt() {
      this.promptType = 'create';
      this.showPromptDialog();
    },
    editPrompt(item) {
      this.promptType = 'edit';
      this.showPromptDialog(item);
    },
    copyPrompt(item) {
      copyCustomPrompt({ customPromptId: item.customPromptId }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('tempSquare.copySuccess'));
          this.fetchList();
        }
      });
    },
    handleDelete(item) {
      this.$confirm(
        this.$t('tool.prompt.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          dangerouslyUseHTMLString: true,
          type: 'warning',
          center: true,
        },
      ).then(async () => {
        deleteCustomPrompt({
          customPromptId: item.customPromptId,
        }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            this.fetchList();
          }
        });
      });
    },
  },
};
</script>
<style lang="scss">
.card-logo {
  width: 50px;
  height: 50px;
  object-fit: cover;
}
.mcp-content-box {
  padding-top: 10px;
}
.mcp-content-box .noData {
  width: 100%;
  text-align: center;
  margin-top: -60px;
  ::v-deep .el-empty__description p {
    color: #b3b1bc;
  }
}
</style>
