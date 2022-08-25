module.exports = {
    content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
    theme: {
      extend: {},
      fontFamily: {
        sans: [
          'Inter',
          'Avenir',
          'Roboto',
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Ubuntu',
          '"Helvetica Neue"',
          'Arial',
          '"Noto Sans"',
          'sans-serif',
          '"Apple Color Emoji"',
          '"Segoe UI Emoji"',
          '"Segoe UI Symbol"',
          '"Noto Color Emoji"',
        ],
      },
    },
    plugins: [require('@tailwindcss/forms')],
  };
  