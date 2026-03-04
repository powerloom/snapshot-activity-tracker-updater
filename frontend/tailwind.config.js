/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        orbitron: ['Orbitron', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      colors: {
        pl: {
          bg: '#161818',
          'bg-card': '#151818',
          'bg-nav': '#141717',
          'bg-elevated': '#1a1d1e',
          'bg-input': '#1b1d1e',
          accent: '#54e794',
          'text-muted': '#b4ccc5',
          border: '#384949',
          'border-subtle': '#303131',
        },
      },
      boxShadow: {
        'glow-green': '0 0 20px rgba(84, 231, 148, 0.2), 0 0 40px rgba(84, 231, 148, 0.08)',
      },
    },
  },
  plugins: [],
}
