<template>
  <div class="card" @click.stop="handleClick">
    <div class="card-title">
      <img
        class="card-logo"
        v-if="info.avatar && info.avatar.path"
        :src="avatarSrc(info.avatar.path)"
      />
      <div class="mcp_detailBox">
        <span class="mcp_name">{{ info.name }}</span>
        <span class="mcp_from">
          <label>{{ $t('tempSquare.author') }}：{{ info.author }}</label>
        </span>
      </div>
    </div>
    <div class="card-des">{{ info.desc }}</div>
    <div class="card-bottom" style="justify-content: flex-end">
      <div class="card-bottom-right">
        <el-tooltip
          v-if="type === 3"
          :content="$t('tempSquare.skills.sendCustom')"
          placement="top"
        >
          <i class="el-icon-s-promotion" @click.stop="sendToResource"></i>
        </el-tooltip>

        <el-tooltip :content="$t('tempSquare.download')" placement="top">
          <i class="el-icon-download" @click.stop="downloadTemplate"></i>
        </el-tooltip>

        <!-- 自定义类型显示更多操作 -->
        <el-dropdown v-if="type == 2" placement="bottom">
          <span class="el-dropdown-link">
            <i class="el-icon-more" @click.stop />
          </span>
          <el-dropdown-menu slot="dropdown" style="margin-top: -10px">
            <el-dropdown-item @click.native="handleDelete">
              {{ $t('common.button.delete') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>

<script>
import { SKILL, SKILLCUSTOM } from '../constants';
import { avatarSrc } from '@/utils/util';

export default {
  name: 'SkillCard',
  props: {
    info: {
      type: Object,
      default: () => ({}),
    },
    // 1:内置，2:自定义，3:对话中的技能
    type: {
      type: Number,
      default: 1,
    },
  },
  methods: {
    avatarSrc,
    handleClick() {
      if (![1, 2].includes(this.type)) return;
      const path = '/skill/detail';
      const type = this.type === 2 ? SKILLCUSTOM : SKILL;
      this.$router.push({
        path,
        query: { templateSquareId: this.info.skillId, type },
      });
    },
    downloadTemplate() {
      this.$emit('download', this.info);
    },
    handleDelete() {
      this.$emit('delete', this.info);
    },
    // 发送到资源库
    sendToResource() {
      this.$emit('sendToResource', this.info);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tempSquare.scss';

.card {
  position: relative;
  padding: 20px 16px;
  border-radius: 12px;
  height: fit-content;
  background: #fff url('@/assets/imgs/card_bg.png');
  background-size: 100% 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: calc((100% / 4) - 20px);
  margin: 0 10px 20px;
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(0, 0, 0, 0);
  &:hover {
    cursor: pointer;
    box-shadow:
      0 2px 8px #171a220d,
      0 4px 16px #0000000f;
    border: 1px solid $border_color;

    .action-icon {
      display: block;
    }
  }
  .card-title {
    display: flex;
    width: 100%;
    padding-bottom: 7px;
    .svg-icon {
      width: 50px;
      height: 50px;
    }
    .mcp_detailBox {
      width: calc(100% - 70px);
      margin-left: 10px;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      padding: 3px 0;
      .mcp_name {
        display: block;
        font-size: 15px;
        font-weight: 700;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
        color: #5d5d5d;
      }
      .mcp_from {
        label {
          padding: 3px 7px;
          font-size: 12px;
          color: $tag_color;
          background: $tag_bg;
          border-radius: 3px;
          display: block;
          height: 22px;
          width: 100%;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }
    }

    margin-bottom: 13px;
  }
  .card-des {
    width: 100%;
    display: -webkit-box;
    text-overflow: ellipsis;
    color: #5d5d5d;
    font-weight: 400;
    overflow: hidden;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    font-size: 13px;
    height: 36px;
    word-wrap: break-word;
  }
  .card-bottom {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 14px;
    margin-bottom: -6px;
    .card-bottom-left {
      color: #888;
    }
    .card-bottom-right {
      i {
        margin-left: 5px;
        cursor: pointer;
      }
    }
  }
}
.card-logo {
  width: 50px;
  height: 50px;
  object-fit: cover;
}

.card-bottom-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
