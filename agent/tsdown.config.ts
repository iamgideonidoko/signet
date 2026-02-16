import { defineConfig } from "tsdown";

export default defineConfig({
  entry: ["./src/index.ts"],
  format: ["esm", "iife"],
  target: "es2020",
  dts: true,
  clean: true,
  bundle: true,
  skipNodeModulesBundle: false,
  minify: true,
  sourcemap: true,
  globalName: "Signet",
});
