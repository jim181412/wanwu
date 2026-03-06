<template>
  <div class="container-wrapper page-wrapper">
    <div class="header">
      <div class="header-left">
        <span
          class="el-icon-arrow-left go-back"
          @click="goBack('/prompt')"
        ></span>
        <h3>{{ $t('promptEvaluate.title') }}</h3>
      </div>
    </div>
    <div class="conditions">
      <div>
        <label for="">{{ $t('promptEvaluate.modelSelectLabel') }}：</label>
        <modelSelect
          v-model="modelId"
          :options="options"
          :placeholder="$t('promptEvaluate.modelSelectPlaceholder')"
          @visible-change="visibleChange"
          style="width: 400px; margin-right: 10px"
        />
      </div>
      <div>
        <el-button
          class="btn_add"
          size="mini"
          icon="el-icon-plus"
          @click="addDomain"
          :disabled="dynamicValidateForm.domains.length >= MAX_NUM || loading"
        >
          {{ $t('promptEvaluate.addPrompt') }}
        </el-button>
        <el-button
          :loading="loading"
          type="primary"
          size="mini"
          @click="submitForm('promptEvaluateForm')"
        >
          {{ $t('promptEvaluate.startEvaluate') }}
        </el-button>
      </div>
    </div>

    <div class="content">
      <el-form
        :model="dynamicValidateForm"
        ref="promptEvaluateForm"
        label-width="100%"
        label-position="left"
        class="promptEvaluateForm"
      >
        <div
          class="item"
          :style="{
            width: `calc((100% - ${(dynamicValidateForm.domains.length - 1) * 10}px) / ${dynamicValidateForm.domains.length})`,
          }"
          v-for="(domain, index) in dynamicValidateForm.domains"
          :key="domain.key"
          :element-loading-text="$t('promptEvaluate.evaluating')"
          element-loading-spinner="el-icon-loading"
          element-loading-background="rgba(255, 255, 255, 0.7)"
        >
          <span>{{ $t('promptEvaluate.prompt') }}{{ index + 1 }}</span>
          <i
            v-if="dynamicValidateForm.domains.length > 1"
            class="el-icon-close"
            @click="removeDomain(domain)"
          ></i>
          <el-form-item
            :label="$t('promptEvaluate.inputPrompt')"
            :prop="'domains.' + index + '.prompt'"
            :rules="{
              required: true,
              message: $t('promptEvaluate.ruleMsg'),
              trigger: 'blur',
            }"
          >
            <el-input
              type="textarea"
              :autosize="{ minRows: 5, maxRows: 5 }"
              :placeholder="$t('common.input.placeholder')"
              v-model="domain.prompt"
              resize="none"
            ></el-input>
          </el-form-item>
          <el-form-item
            :label="$t('promptEvaluate.resultLabel')"
            :prop="'domains.' + index + '.expectedOutput'"
            :rules="{
              required: true,
              message: $t('promptEvaluate.ruleResult'),
              trigger: 'blur',
            }"
          >
            <el-input
              type="textarea"
              :autosize="{ minRows: 10, maxRows: 10 }"
              :placeholder="$t('common.input.placeholder')"
              v-model="domain.expectedOutput"
              resize="none"
            ></el-input>
          </el-form-item>
          <el-form-item
            :label="$t('promptEvaluate.exportResult')"
            :prop="'domains.' + index + '.modelId'"
          >
            <el-input
              :disabled="true"
              type="textarea"
              :autosize="{ minRows: 5, maxRows: 15 }"
              v-model="resultList[index].outPutResult"
              resize="none"
            ></el-input>
          </el-form-item>
          <el-form-item
            :label="$t('promptEvaluate.evaluateResult')"
            :prop="'domains.' + index + '.modelId'"
          >
            <el-input
              :disabled="true"
              type="textarea"
              :autosize="{ minRows: 5, maxRows: 15 }"
              v-model="resultList[index].evaluateResult"
              resize="none"
            ></el-input>
          </el-form-item>
        </div>
      </el-form>
    </div>
  </div>
</template>
<script>
import { selectModelList } from '@/api/modelAccess';
import ModelSelect from '@/components/modelSelect.vue';
import sseMethod from '@/mixins/sseMethod';
import { goBack } from '@/utils/util';

export default {
  components: { ModelSelect },
  mixins: [sseMethod],
  data() {
    return {
      MAX_NUM: 3, // 提示词评估最大添加数量
      loading: false,
      resultList: [
        {
          outPutResult: '',
          evaluateResult: '',
        },
      ], // 评估结果数组
      dynamicValidateForm: {
        domains: [
          {
            modelId: '',
            prompt: '',
            expectedOutput: '',
          },
        ],
      },
      options: [],
      modelId: '',
    };
  },
  created() {
    this.getModellist();
  },
  methods: {
    goBack,
    getModellist(type = '') {
      selectModelList()
        .then(res => {
          if (res.code === 0) {
            this.options = res.data.list || [];
            if (res.data.list.length > 0 && type === '') {
              this.modelId = res.data.list[0].modelId;
            }
          }
        })
        .catch(err => {
          this.$message.error(err);
        });
    },
    visibleChange(val) {
      if (val) {
        this.getModellist('refresh');
      }
    },
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (valid) {
          this.loading = true;
          this.concurRequest(this.dynamicValidateForm.domains);
        } else {
          return false;
        }
      });
    },
    removeDomain(item) {
      var index = this.dynamicValidateForm.domains.indexOf(item);
      if (index !== -1) {
        this.dynamicValidateForm.domains.splice(index, 1);
        this.resultList.splice(index, 1);
      }
    },
    addDomain() {
      this.dynamicValidateForm.domains.push({
        modelId: '',
        key: Date.now(),
        prompt: '',
        expectedOutput: '',
      });
      this.resultList.push({
        key: Date.now(),
        outPutResult: '',
        evaluateResult: '',
      });
    },
    concurRequest(urls) {
      if (urls.length === 0) return;

      this.loading = true;
      const total = urls.length;
      const completedSet = new Set();

      // 检查是否所有任务都已完成
      const checkAllComplete = () => {
        if (completedSet.size === total) {
          this.loading = false;
        }
      };

      // 标记单个任务完成
      const markTaskComplete = index => {
        if (!completedSet.has(index)) {
          completedSet.add(index);
          checkAllComplete();
        }
      };

      urls.forEach((url, index) => {
        // 获取 prompt 推理结果
        this.sendEventStreamIsolation(
          '/prompt/reason',
          {
            prompt: url.prompt,
            modelId: this.modelId,
          },
          {
            onProgress: data => {
              this.resultList[index].outPutResult = data;
            },
            onComplete: reasonData => {
              // 获取评估结果
              this.sendEventStreamIsolation(
                '/prompt/evaluate',
                {
                  answer: reasonData,
                  modelId: this.modelId,
                  expectedOutput: url.expectedOutput,
                },
                {
                  onProgress: data => {
                    this.resultList[index].evaluateResult = data;
                  },
                  onComplete: data => {
                    this.resultList[index].evaluateResult = data;
                    markTaskComplete(index);
                  },
                  onError: error => {
                    this.resultList[index].evaluateResult = this.$t(
                      'promptEvaluate.errorTips',
                    );
                    markTaskComplete(index);
                  },
                },
              );
            },
            onError: error => {
              this.resultList[index].outPutResult = this.$t(
                'promptEvaluate.errorTips',
              );
              this.resultList[index].evaluateResult = this.$t(
                'promptEvaluate.errorTips',
              );
              markTaskComplete(index);
            },
          },
        );
      });
    },
  },
};
</script>
<style lang="scss">
.container-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  padding: 0 10px;
  box-sizing: border-box;

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 24px 10px 14px 10px;
    border-bottom: 1px solid #eaeaea;
    .header-left {
      display: flex;
      align-items: center;
      .go-back {
        font-size: 18px;
        cursor: pointer;
      }
      h3 {
        font-size: 18px;
        font-weight: 800;
        color: #434c6c;
        margin-left: 10px;
      }
    }
  }

  .conditions {
    display: flex;
    justify-content: space-between;
    padding: 10px;

    .btn_add {
      border-color: $color;
      .el-icon-plus {
        color: $color;
      }
    }

    .is-disabled {
      border-color: #c0c4cc;
      .el-icon-plus {
        color: #c0c4cc;
      }
    }
  }
  .content {
    margin-bottom: 10px;
    .promptEvaluateForm {
      width: 100%;
      display: flex;
      .item {
        margin-right: 10px;
        position: relative;
        padding: 15px 10px;
        background: #fff;
        border-radius: 10px;
        background: rgba(242, 247, 255, 0.5607843137);

        &:last-child {
          margin-right: 0;
        }

        .el-icon-close {
          position: absolute;
          right: 10px;
          top: 15px;
          font-size: 16px;

          &:hover {
            cursor: pointer;
          }
        }
        .el-form-item {
          &:last-child {
            margin-bottom: 0;
          }
        }
        .el-form-item__label {
          display: block;
          float: none;
        }
        .el-form-item__content {
          margin-left: 0 !important;

          .el-textarea {
            font-size: 12px;
          }
          .el-textarea.is-disabled .el-textarea__inner {
            color: #606266;

            &:hover {
              cursor: pointer;
            }
          }
        }
      }
      .el-loading-text,
      .el-icon-loading {
        color: $color;
      }
      .el-loading-mask {
        border-radius: 10px;
      }
    }
  }
}
</style>
