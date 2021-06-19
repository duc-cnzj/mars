const CracoAntDesignPlugin = require("craco-antd");

const plugins = process.env.NODE_ENV === "production" ? ["transform-remove-console"] : []

module.exports = {
  webpack: {
    configure: (webpackConfig, {
      env, paths
    }) => {
      webpackConfig.output = {
        ...webpackConfig.output,
        publicPath: process.env.NODE_ENV === 'production' ? '/resources' : '/',
      }
      return webpackConfig
    }
  },
  plugins: [
    {
      plugin: CracoAntDesignPlugin,
      options: {
        customizeTheme: {
          //   '@primary-color': '#1DA57A',
        },
      },
    },
  ],
  babel: {
    plugins: plugins,
  },
};
