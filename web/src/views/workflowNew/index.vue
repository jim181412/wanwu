<template>
  <div class="wp">
    <iframe id="inlineFrameExample" :src="workflowUrl"></iframe>
  </div>
</template>

<script>
export default {
  data() {
    const query = `?workflow_id=${this.$route.query.id}&space_id=${this.$store.state.user.userInfo.orgId}`;
    return {
      workflowUrl:
        process.env.NODE_ENV === 'development'
          ? `http://localhost:8282/workflow${query}`
          : window.location.origin + this.$basePath + `/workflow/${query}`,
    };
  },
};
</script>

<style lang="scss" scoped>
.wp {
  width: 100%;
  height: calc(100% - 3px);
  background: #f4f5f6;
  iframe {
    width: 100%;
    height: 100%;
    border: none;
  }
}
</style>
