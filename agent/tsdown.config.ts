import { defineConfig } from "tsdown";

export default defineConfig({
  entry: ["./src/index.ts"],
  format: ["esm", "cjs"],
  target: "node20",
  dts: true,
  clean: true,
  bundle: true,
  skipNodeModulesBundle: true,
  sourcemap: true,
});
