const path = require("node:path");
const {
  ModuleFederationPlugin,
} = require("@module-federation/enhanced/rspack");
const rspack = require("@rspack/core");

const frontendSrc = path.resolve(__dirname, "../../frontend/src");

/** @type {import('@rspack/cli').Configuration} */
module.exports = {
  entry: "./src/index.ts",
  mode: process.env.NODE_ENV === "production" ? "production" : "development",
  devtool: "source-map",
  output: {
    uniqueName: "temflowralCanvas",
    publicPath: "auto",
    path: path.resolve(__dirname, "dist"),
    clean: true,
  },
  resolve: {
    extensions: [".tsx", ".ts", ".jsx", ".js"],
    alias: {
      "@": frontendSrc,
    },
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: {
          loader: "builtin:swc-loader",
          options: {
            jsc: {
              parser: { syntax: "typescript", tsx: true },
              transform: { react: { runtime: "automatic" } },
            },
          },
        },
        type: "javascript/auto",
      },
      {
        test: /\.css$/,
        use: [
          "style-loader",
          "css-loader",
          {
            loader: "postcss-loader",
            options: {
              postcssOptions: {
                plugins: ["@tailwindcss/postcss"],
              },
            },
          },
        ],
        type: "javascript/auto",
      },
    ],
  },
  plugins: [
    new rspack.CopyRspackPlugin({
      patterns: [{ from: "public", to: "." }],
    }),
    new ModuleFederationPlugin({
      name: "temflowralCanvas",
      filename: "remoteEntry.js",
      exposes: {
        "./WorkflowBuilder": "./src/expose.tsx",
      },
      shared: {
        react: {
          singleton: true,
          requiredVersion: "19.1.0",
          eager: false,
        },
        "react-dom": {
          singleton: true,
          requiredVersion: "19.1.0",
          eager: false,
        },
        "@xyflow/react": {
          singleton: true,
          requiredVersion: "12.8.6",
        },
      },
      dts: false,
    }),
  ],
  devServer: {
    port: 3002,
    historyApiFallback: true,
    headers: {
      "Access-Control-Allow-Origin": "*",
    },
  },
};
