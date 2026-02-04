import type { Preview } from '@storybook/react';
import '../src/styles/index.css'; // Import TailwindCSS and global styles

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/,
      },
    },
    backgrounds: {
      default: 'cream',
      values: [
        { name: 'cream', value: '#faf9f7' },
        { name: 'white', value: '#ffffff' },
        { name: 'navy', value: '#0d1829' },
        { name: 'sand', value: '#f5f1ed' },
      ],
    },
    viewport: {
      viewports: {
        mobile: { name: 'Mobile', styles: { width: '375px', height: '667px' } },
        tablet: { name: 'Tablet', styles: { width: '768px', height: '1024px' } },
        desktop: { name: 'Desktop', styles: { width: '1280px', height: '800px' } },
      },
    },
    a11y: {
      config: {
        rules: [
          { id: 'color-contrast', enabled: true },
          { id: 'label', enabled: true },
          { id: 'button-name', enabled: true },
        ],
      },
    },
  },
  tags: ['autodocs'],
};

export default preview;
