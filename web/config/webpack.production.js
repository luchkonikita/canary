const webpack = require('webpack')
const merge = require('webpack-merge')
const UglifyJSPlugin = require('uglifyjs-webpack-plugin')
const CompressionPlugin = require("compression-webpack-plugin")
const base = require('./webpack.base.js')

module.exports = merge(base, {
  output: {
    publicPath: 'public'
  },
  devtool: 'source-map',
  plugins: [
    new UglifyJSPlugin(),
    new CompressionPlugin(),
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify('production')
    }),
    new webpack.optimize.CommonsChunkPlugin({
      name: 'manifest'
    })
  ]
})
