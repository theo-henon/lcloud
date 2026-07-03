import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        canvas: "#0a0a0a",
        primary: {
          DEFAULT: "#faff69",
          active: "#e6eb52",
        },
        ink: "#ffffff",
        body: {
          DEFAULT: "#cccccc",
          strong: "#e6e6e6",
        },
        muted: {
          DEFAULT: "#888888",
          soft: "#5a5a5a",
        },
        hairline: {
          DEFAULT: "#2a2a2a",
          strong: "#3a3a3a",
        },
        surface: {
          soft: "#121212",
          card: "#1a1a1a",
          elevated: "#242424",
        },
        "on-primary": "#0a0a0a",
        accent: {
          emerald: "#22c55e",
          rose: "#ef4444",
          amber: "#f59e0b",
        },
        success: "#22c55e",
        error: "#ef4444",
      },
      fontFamily: {
        sans: ["Inter", "sans-serif"],
        mono: ["JetBrains Mono", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
