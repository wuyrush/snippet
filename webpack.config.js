const path = require("path");
const HtmlWebPackPlugin = require("html-webpack-plugin"),
  CopyWebpackPlugin = require("copy-webpack-plugin"),
  HtmlWebpackIncludeAssetsPlugin = require('html-webpack-include-assets-plugin');

// Try the environment variable, otherwise use root
const ASSET_PATH = process.env.ASSET_PATH || '/';

module.exports = {
  mode: "production",
  entry: {
    'index': './frontend/src/js/index.js',
    'view': './frontend/src/js/view.js',
  },
  output: {
    path: path.resolve(__dirname, "frontend", "assets"),
    filename: "[name].js",
    // set how to refer to the assets. Try tweak it and inspect the resulting html file being
    // generated. For details see https://webpack.js.org/guides/public-path/
    publicPath: ASSET_PATH,
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      },
      {
        test: /\.html$/,
        use: [
          {
            loader: "html-loader"
          }
        ]
      },
      {
        test: /\.css$/,
        use: [
          {
            loader: "css-loader"
          }
        ]
      },
    ]
  },
  plugins: [
    // copy over assets from other places
    // https://github.com/webpack-contrib/copy-webpack-plugin#to
    new CopyWebpackPlugin([
      {
        from: "./node_modules/bulma/css/bulma.min.css"
        // the root dir of dest is set to that of output already
      },
      {
        from: "./node_modules/prismjs/themes/prism-tomorrow.css",
        to: "./prism.css"
      }
    ]),
    new HtmlWebPackPlugin({
      template: "./frontend/src/index.html",
      filename: "index.html",
      chunks: ['index'],
    }),
    new HtmlWebPackPlugin({
      template: "./frontend/src/view.html",
      filename: "view.html",
      chunks: ['view'],
    }),
    // so that we don't need to manually reference assets in our html markup
    new HtmlWebpackIncludeAssetsPlugin({
      assets: 'bulma.min.css',
      append: true
    }),
    new HtmlWebpackIncludeAssetsPlugin({
      assets: 'prism.css',
      append: true,
      files: 'view.html'
    })
  ]
}
