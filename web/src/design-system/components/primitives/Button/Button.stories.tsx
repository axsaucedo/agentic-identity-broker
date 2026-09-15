/**
 * Button Component Stories
 *
 * Demonstrates all Button variants, sizes, states, and usage patterns.
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta = {
  title: 'Design System/Primitives/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline', 'ghost', 'danger'],
      description: 'Visual style variant',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'primary' },
      },
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg', 'xl'],
      description: 'Button size',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'md' },
      },
    },
    isLoading: {
      control: 'boolean',
      description: 'Show loading spinner',
    },
    disabled: {
      control: 'boolean',
      description: 'Disable button interactions',
    },
    fullWidth: {
      control: 'boolean',
      description: 'Make button full width',
    },
    children: {
      control: 'text',
      description: 'Button content',
    },
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Default button with primary variant and medium size.
 * This is the most common button style used for primary actions.
 */
export const Default: Story = {
  args: {
    children: 'Primary Button',
    variant: 'primary',
    size: 'md',
  },
};

/**
 * All button variants side-by-side for visual comparison.
 * Choose variants based on action importance:
 * - Primary: Most important actions (submit, confirm)
 * - Secondary: Success/positive actions (approve, grant)
 * - Outline: Tertiary actions (view details, learn more)
 * - Ghost: Subtle actions (cancel, dismiss)
 * - Danger: Destructive actions (delete, revoke)
 */
export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="danger">Danger</Button>
    </div>
  ),
};

/**
 * All button sizes from sm to xl.
 * - sm: Compact spaces, secondary actions
 * - md: Default size for most actions
 * - lg: Prominent CTAs, landing pages
 * - xl: Hero sections, high-impact actions
 */
export const Sizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-4">
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
      <Button size="xl">Extra Large</Button>
    </div>
  ),
};

/**
 * Button states: normal, disabled, and loading.
 * Loading state shows spinner and disables interactions.
 * Disabled state reduces opacity and prevents clicks.
 */
export const States: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Normal</p>
        <div className="flex gap-4">
          <Button variant="primary">Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Disabled</p>
        <div className="flex gap-4">
          <Button variant="primary" disabled>
            Primary
          </Button>
          <Button variant="secondary" disabled>
            Secondary
          </Button>
          <Button variant="outline" disabled>
            Outline
          </Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Loading</p>
        <div className="flex gap-4">
          <Button variant="primary" isLoading>
            Primary
          </Button>
          <Button variant="secondary" isLoading>
            Secondary
          </Button>
          <Button variant="outline" isLoading>
            Outline
          </Button>
        </div>
      </div>
    </div>
  ),
};

/**
 * Buttons with icons before or after text.
 * Icons add visual clarity and help users quickly identify actions.
 */
export const WithIcons: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button
        variant="primary"
        iconBefore={
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
        }
      >
        Add New
      </Button>

      <Button
        variant="secondary"
        iconBefore={
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
        }
      >
        Approve
      </Button>

      <Button
        variant="outline"
        iconAfter={
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        }
      >
        Next
      </Button>

      <Button
        variant="danger"
        iconBefore={
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
            />
          </svg>
        }
      >
        Delete
      </Button>
    </div>
  ),
};

/**
 * Full width buttons that span the container width.
 * Useful for mobile layouts and form submissions.
 */
export const FullWidth: Story = {
  render: () => (
    <div className="w-96 space-y-4">
      <Button variant="primary" fullWidth>
        Full Width Primary
      </Button>
      <Button variant="outline" fullWidth>
        Full Width Outline
      </Button>
      <Button variant="secondary" fullWidth isLoading>
        Full Width Loading
      </Button>
    </div>
  ),
};

/**
 * Loading states showing spinner placement.
 * The button remains in its original size to prevent layout shift.
 */
export const LoadingStates: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Loading with text
        </p>
        <div className="flex gap-4">
          <Button variant="primary" isLoading>
            Saving...
          </Button>
          <Button variant="secondary" isLoading>
            Processing...
          </Button>
          <Button variant="outline" isLoading>
            Loading...
          </Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Different sizes
        </p>
        <div className="flex items-center gap-4">
          <Button size="sm" isLoading>
            Small
          </Button>
          <Button size="md" isLoading>
            Medium
          </Button>
          <Button size="lg" isLoading>
            Large
          </Button>
          <Button size="xl" isLoading>
            Extra Large
          </Button>
        </div>
      </div>
    </div>
  ),
};

/**
 * Danger variant for destructive actions.
 * Use sparingly and always with confirmation dialogs.
 */
export const DangerActions: Story = {
  render: () => (
    <div className="flex gap-4">
      <Button variant="danger">Delete Account</Button>
      <Button variant="danger" size="sm">
        Remove
      </Button>
      <Button variant="danger" isLoading>
        Revoking...
      </Button>
    </div>
  ),
};

/**
 * Interactive playground to experiment with all props.
 * Try different combinations of variant, size, loading, and disabled states.
 */
export const Playground: Story = {
  args: {
    children: 'Click me',
    variant: 'primary',
    size: 'md',
    isLoading: false,
    disabled: false,
    fullWidth: false,
  },
  parameters: {
    docs: {
      description: {
        story:
          'Experiment with all button props using the controls below. Try different variants, sizes, and states.',
      },
    },
  },
};

/**
 * Accessibility features demonstration.
 * All buttons support:
 * - Keyboard navigation (Tab to focus, Enter/Space to activate)
 * - Screen reader labels
 * - Visible focus indicators (ring on focus)
 * - Disabled state prevents interaction
 */
export const Accessibility: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Keyboard Navigation
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Press Tab to focus buttons, Enter or Space to activate
        </p>
        <div className="flex gap-4">
          <Button variant="primary">First Button</Button>
          <Button variant="secondary">Second Button</Button>
          <Button variant="outline">Third Button</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Focus Indicators
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Visible focus rings meet WCAG 2.1 AA contrast requirements
        </p>
        <div className="flex gap-4">
          <Button variant="primary">Focus Me</Button>
          <Button variant="danger">Focus Me Too</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Disabled State
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Disabled buttons cannot receive focus or be activated
        </p>
        <div className="flex gap-4">
          <Button variant="primary" disabled>
            Disabled Button
          </Button>
          <Button variant="secondary" disabled>
            Also Disabled
          </Button>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Buttons are fully accessible with keyboard support, visible focus indicators, and proper ARIA attributes. All interactive states are keyboard accessible.',
      },
    },
  },
};
