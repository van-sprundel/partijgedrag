import type { Config } from 'tailwindcss';

export default {
  content: ['./app/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        'search-panel-element': '#EDEDED',
      },
    },
  },
  darkMode: 'class',
  plugins: [],
} satisfies Config;
