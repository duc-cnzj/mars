const plugins = process.env.NODE_ENV === "production" ? ["transform-remove-console"] : []

module.exports = {
  webpack: {
    resolve: {
      fallback: {
        "util": require.resolve("util/"),
        "assert": require.resolve("assert/")
       }
    },
    configure: (webpackConfig, {
      env, paths
    }) => {
      webpackConfig.output = {
        ...webpackConfig.output,
        publicPath: process.env.NODE_ENV === 'production' ? '/resources/' : '/',
      }
      return webpackConfig
    }
  },
  babel: {
    plugins: [...plugins, "@emotion"],
  },
};
