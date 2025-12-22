import type { Config } from 'tailwindcss'

export default {
  content: [
    './src/**/*.{js,jsx,ts,tsx}',
    './index.html',
  ],
  theme: {
    extend: {
      colors: {
        // Premium warm neutral base - invokes trust and sophistication
        cream: '#faf9f7',
        sand: '#f5f1ed',
        taupe: '#e8e3de',
        slate: '#d4cfc8',

        // Authoritative primary - confident but not aggressive
        navy: {
          50: '#f7f9fc',
          100: '#edf2f8',
          200: '#d5dff0',
          300: '#a8bee2',
          400: '#7a9dd0',
          500: '#4d7cbc',
          600: '#2d5a9f',
          700: '#1f3f6d',
          800: '#15294a',
          900: '#0d1829',
        },

        // Accent colors - for granted permissions and important actions
        emerald: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#134e4a',
        },

        amber: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
        },

        secondary: {
          50: '#f8f7f5',
          100: '#f1edea',
          200: '#ddd4d0',
          300: '#c9bbb5',
          400: '#8b7b73',
          500: '#6b5b53',
          600: '#5a4b43',
          700: '#4a3b33',
          800: '#3a2b23',
          900: '#2a1b13',
        },

      },
      fontFamily: {
        display: ['Crimson Pro', 'serif'],
        sans: ['Manrope', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      boxShadow: {
        // Premium subtle shadows for depth
        'sm-premium': '0 1px 2px 0 rgb(0 0 0 / 0.03)',
        'md-premium': '0 4px 12px 0 rgb(0 0 0 / 0.08)',
        'lg-premium': '0 10px 25px -5px rgb(0 0 0 / 0.1)',
        'xl-premium': '0 20px 40px -10px rgb(0 0 0 / 0.12)',

        // Elevated card styling
        'card': '0 2px 8px 0 rgb(0 0 0 / 0.04), inset 0 1px 0 0 rgb(255 255 255 / 1)',
        'card-hover': '0 8px 20px -4px rgb(0 0 0 / 0.1), inset 0 1px 0 0 rgb(255 255 255 / 1)',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-out',
        'slide-up': 'slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1)',
        'scale-in': 'scaleIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1)',
        'progress-stripes': 'progressStripes 1s linear infinite',
        'progress-indeterminate': 'progressIndeterminate 1.5s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(12px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        progressStripes: {
          '0%': { backgroundPosition: '1rem 0' },
          '100%': { backgroundPosition: '0 0' },
        },
        progressIndeterminate: {
          '0%': { left: '-30%' },
          '50%': { left: '100%' },
          '100%': { left: '100%' },
        },
      },
      typography: {
        DEFAULT: {
          css: {
            fontFamily: 'Manrope, sans-serif',
          },
        },
      },
    },
  },
  plugins: [],
} satisfies Config
