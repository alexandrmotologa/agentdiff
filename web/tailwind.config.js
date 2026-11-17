/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        background: "#090d16",
        surface: "#101726",
        surfaceBorder: "#1e293b",
        accent: "#38bdf8",
        success: "#22c55e",
        warning: "#eab308",
        danger: "#ef4444",
      },
    },
  },
  plugins: [],
};
