import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

if (process.env.QOTD_ENV !== "prod") {
  process.env.VITE_HEADER_TAGS = "";
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
});
