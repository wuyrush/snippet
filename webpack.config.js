const path = require("path");
const HtmlWebPackPlugin = require("html-webpack-plugin"),
  CopyWebpackPlugin = require("copy-webpack-plugin"),
  HtmlWebpackIncludeAssetsPlugin = require('html-webpack-include-assets-plugin');

module.exports = {
  mode: "production",
  entry: "./frontend/src/js/index.js",
  output: {
    path: path.resolve(__dirname, "frontend", "assets"),
    filename: "index.js"
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
      }
    ]),
    new HtmlWebPackPlugin({
      template: "./frontend/src/index.html",
      filename: "index.html"
    }),
    // so that we don't need to manually reference assets in our html markup
    new HtmlWebpackIncludeAssetsPlugin({
      assets: 'bulma.min.css',
      append: true
    })
  ]
}
