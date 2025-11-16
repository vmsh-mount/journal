/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/**/*.html",
    "./internal/handlers/**/*.go",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      maxWidth: {
        'content': '680px',
      },
      colors: {
        'text': '#111',
        'text-muted': '#555',
      },
    },
  },
  plugins: [],
}

