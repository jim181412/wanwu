'use strict';

const path = require('path');
const { name } = require('./package.json');
const webpack = require('webpack');

function resolve(dir) {
  return path.join(__dirname, dir);
}

const CompressionWebpackPlugin = require('compression-webpack-plugin');
const isProdOrTest = process.env.NODE_ENV !== 'development';

const bffProxyUrl = process.env.VUE_APP_BFF_PROXY_URL || 'http://localhost:6668';
const workflowProxyUrl =
  process.env.VUE_APP_WORKFLOW_PROXY_URL || 'http://localhost:8999';
const workflowFileProxyUrl =
  process.env.VUE_APP_WORKFLOW_FILE_PROXY_URL || 'http://localhost:8998';

function createProxy(target, extra = {}) {
  return {
    target,
    changeOrigin: true,
    secure: false,
    ...extra,
  };
}

function createRewriteProxy(target, prefix) {
  return createProxy(target, {
    pathRewrite: {
      [`^${prefix}`]: '',
    },
  });
}

module.exports = {
  // 基础配置 详情看文档
  publicPath: process.env.VUE_APP_BASE_PATH + '/aibase',
  outputDir: 'dist',
  assetsDir: 'static',
  lintOnSave: process.env.NODE_ENV === 'development',
  productionSourceMap: false, //源码映射
  transpileDependencies: [
    'ml-matrix',
    '@antv/layout',
    '@antv/g6',
    '@antv/graphlib',
  ],
  chainWebpack(config) {
    config.module
      .rule('md')
      .test(/\.md$/)
      .use('html-loader')
      .loader('html-loader')
      .end()
      .use('markdown-loader')
      .loader('markdown-loader')
      .end();

    config.plugins.delete('prefetch');
    if (isProdOrTest) {
      // 对超过10kb的文件gzip压缩
      config.plugin('compressionPlugin').use(
        new CompressionWebpackPlugin({
          test: /\.(css|html)$/,
          threshold: 10240,
        }),
      );
    }

    config.module
      .rule('svg')
      .exclude.add(resolve('src/assets/icons')) //svg文件位置
      .end();
    config.module
      .rule('icons')
      .test(/\.svg$/)
      .include.add(resolve('src/assets/icons')) //svg文件位置
      .end()
      .use('svg-sprite-loader')
      .loader('svg-sprite-loader')
      .options({
        symbolId: 'icon-[name]',
      })
      .end();

    // 生产环境去掉 console 打印
    config.when(process.env.NODE_ENV === 'production', config => {
      config.optimization.minimize(true);
      config.optimization.minimizer('terser').tap(args => {
        args[0].terserOptions.compress.drop_console = true;
        return args;
      });
    });
  },
  devServer: {
    port: 8082,
    open: false,
    hot: true,
    compress: false,
    client: {
      // 禁用错误覆盖层
      overlay: {
        warnings: false,
        errors: false,
        runtimeErrors: false,
      },
    },
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
    proxy: {
      '/openAi': createProxy(bffProxyUrl),
      '/user/api': createRewriteProxy(bffProxyUrl, '/user/api'),
      '/service/url/openurl/v1': createRewriteProxy(
        bffProxyUrl,
        '/service/url/openurl/v1',
      ),
      '/service/api': createRewriteProxy(bffProxyUrl, '/service/api'),
      '/training/api': createRewriteProxy(bffProxyUrl, '/training/api'),
      '/resource/api': createRewriteProxy(bffProxyUrl, '/resource/api'),
      '/datacenter/api': createRewriteProxy(bffProxyUrl, '/datacenter/api'),
      '/modelprocess/api': createRewriteProxy(bffProxyUrl, '/modelprocess/api'),
      '/expand/api': createRewriteProxy(bffProxyUrl, '/expand/api'),
      '/record/api': createRewriteProxy(bffProxyUrl, '/record/api'),
      '/img': createProxy(bffProxyUrl),
      '/konwledgeServe': createProxy(bffProxyUrl),
      '/proxyupload': createProxy(bffProxyUrl),
      '/use/model/api': createRewriteProxy(bffProxyUrl, '/use/model/api'),
      '/prompt/api': createRewriteProxy(bffProxyUrl, '/prompt/api'),
      '/v1/static': createProxy(bffProxyUrl),
      '/workflow/api': createProxy(workflowProxyUrl, {
        pathRewrite: {
          '^/workflow/api': '',
        },
      }),
      '/api': createProxy(workflowProxyUrl),
      '/workflow/minio/presign': createProxy(workflowFileProxyUrl),
    },
  },
  css: {
    sourceMap: false,
    loaderOptions: {
      scss: {
        implementation: require('sass'),
        sassOptions: {
          outputStyle: 'compressed',
          sourceMap: false,
          quietDeps: true,
        },
        additionalData: `
            @use "@/style/theme/vars_blue.scss" as *;
            @use "@/style/theme/common.scss" as *;
        `, // 假设variables.scss位于src/styles目录下
      },
    },
  },
  parallel: false,
  configureWebpack: {
    cache: {
      type: 'filesystem',
      buildDependencies: {
        config: [__filename],
      },
      cacheDirectory: path.resolve(__dirname, 'node_modules/.cache/webpack'),
    },
    // @路径走src文件夹
    module: {
      rules: [
        {
          test: /\.swf$/,
          loader: 'url-loader',
          options: {
            limit: 10000,
            name: 'static/media/[name].[hash:7].[ext]',
          },
        },
      ],
    },
    resolve: {
      alias: {
        vue$: 'vue/dist/vue.esm.js',
        '@': resolve('src'),
        '@common': resolve('common'),
        '@antv/g6': path.resolve(__dirname, 'node_modules/@antv/g6'),
      },
    },
    output: {
      // 把子应用打包成 umd 库格式(必须)
      library: `${name}-[name]`,
      libraryTarget: 'umd',
      chunkLoadingGlobal: `webpackJsonp_${name}`,
    },
    plugins: [
      new webpack.optimize.LimitChunkCountPlugin({
        maxChunks: 10, // 来限制 chunk 的最大数量
      }),
      new webpack.optimize.MinChunkSizePlugin({
        minChunkSize: 50000, // Minimum number of characters
      }),
    ],
  },
};
