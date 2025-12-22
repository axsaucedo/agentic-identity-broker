/**
 * Accordion Component Stories
 *
 * Demonstrates all variants and use cases of the Accordion component.
 */

import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { Accordion, type AccordionItem } from './Accordion';

// Mock icons for stories
const DocumentIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
    />
  </svg>
);

const UserIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
    />
  </svg>
);

const SettingsIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
    />
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
    />
  </svg>
);

const SecurityIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
    />
  </svg>
);

const BellIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
    />
  </svg>
);

const Badge: React.FC<{ children: React.ReactNode; variant?: 'primary' | 'neutral' }> = ({
  children,
  variant = 'neutral',
}) => (
  <span
    className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
      variant === 'primary'
        ? 'bg-trust-deep text-white'
        : 'bg-gray-100 text-gray-700'
    }`}
  >
    {children}
  </span>
);

const meta = {
  title: 'Advanced/Accordion',
  component: Accordion,
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        component:
          'Accordion component for organizing collapsible content sections with smooth animations and full keyboard accessibility.',
      },
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Accordion>;

export default meta;
type Story = StoryObj<typeof meta>;

const basicItems: AccordionItem[] = [
  {
    id: '1',
    title: 'What is the Agentic Identity Broker?',
    content: (
      <p className="text-gray-700">
        The Agentic Identity Broker is a secure OAuth 2.0/OIDC proxy that enables
        AI agents to access protected resources on behalf of users while maintaining
        strict consent and scope management.
      </p>
    ),
  },
  {
    id: '2',
    title: 'How does consent management work?',
    content: (
      <div className="space-y-3">
        <p className="text-gray-700">
          Users grant explicit consent for agents to access specific scopes. All
          consents are tracked and can be revoked at any time.
        </p>
        <ul className="list-disc list-inside space-y-1 text-gray-700">
          <li>Granular scope-based permissions</li>
          <li>Expiration time controls</li>
          <li>Real-time revocation</li>
          <li>Audit trail of all access</li>
        </ul>
      </div>
    ),
  },
  {
    id: '3',
    title: 'What security measures are in place?',
    content: (
      <div className="space-y-3">
        <p className="text-gray-700">
          The broker implements multiple layers of security to protect user data:
        </p>
        <ul className="list-disc list-inside space-y-1 text-gray-700">
          <li>OAuth 2.0 and OpenID Connect standards</li>
          <li>Secure session management with short-lived tokens</li>
          <li>Rate limiting and anomaly detection</li>
          <li>Encrypted storage of sensitive data</li>
          <li>Comprehensive audit logging</li>
        </ul>
      </div>
    ),
  },
];

/**
 * Default accordion with multiple items and exclusive mode (only one open at a time).
 */
export const Default: Story = {
  args: {
    items: basicItems,
    exclusive: true,
    defaultOpen: '1',
  },
};

/**
 * Exclusive mode ensures only one panel can be open at any given time.
 * Opening a new panel automatically closes the previously open one.
 */
export const Exclusive: Story = {
  args: {
    items: basicItems,
    exclusive: true,
  },
};

/**
 * Non-exclusive mode allows multiple panels to be open simultaneously.
 * Users can expand as many sections as needed.
 */
export const NonExclusive: Story = {
  args: {
    items: basicItems,
    exclusive: false,
    defaultOpen: ['1', '2'],
  },
};

/**
 * Small size variant with compact spacing for dense layouts.
 */
export const SmallSize: Story = {
  args: {
    items: basicItems.slice(0, 3),
    size: 'sm',
    defaultOpen: '1',
  },
};

/**
 * Medium size variant (default) with comfortable spacing.
 */
export const MediumSize: Story = {
  args: {
    items: basicItems,
    size: 'md',
    defaultOpen: '1',
  },
};

/**
 * Accordion items with optional description text in the header.
 */
export const WithDescriptions: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Account Settings',
        description: 'Manage your profile, email preferences, and account security',
        content: (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Display Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                defaultValue="John Doe"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email Address
              </label>
              <input
                type="email"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                defaultValue="john@example.com"
              />
            </div>
          </div>
        ),
      },
      {
        id: '2',
        title: 'Privacy & Security',
        description: 'Control who can see your information and how it\'s used',
        content: (
          <div className="space-y-3">
            <label className="flex items-center gap-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-sm text-gray-700">
                Allow agents to access my profile information
              </span>
            </label>
            <label className="flex items-center gap-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-sm text-gray-700">
                Send security alerts via email
              </span>
            </label>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="rounded" />
              <span className="text-sm text-gray-700">
                Require 2FA for sensitive operations
              </span>
            </label>
          </div>
        ),
      },
      {
        id: '3',
        title: 'Notification Preferences',
        description: 'Choose what updates you want to receive',
        content: (
          <div className="space-y-3">
            <label className="flex items-center gap-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-sm text-gray-700">Email notifications</span>
            </label>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="rounded" />
              <span className="text-sm text-gray-700">Push notifications</span>
            </label>
            <label className="flex items-center gap-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-sm text-gray-700">Weekly digest</span>
            </label>
          </div>
        ),
      },
    ],
    defaultOpen: '1',
  },
};

/**
 * Accordion items with icons to enhance visual hierarchy.
 */
export const WithIcons: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Profile Information',
        icon: <UserIcon />,
        content: (
          <p className="text-gray-700">
            Update your name, photo, and other personal details visible to other users.
          </p>
        ),
      },
      {
        id: '2',
        title: 'Security Settings',
        icon: <SecurityIcon />,
        content: (
          <p className="text-gray-700">
            Manage passwords, two-factor authentication, and active sessions.
          </p>
        ),
      },
      {
        id: '3',
        title: 'Notification Settings',
        icon: <BellIcon />,
        content: (
          <p className="text-gray-700">
            Control how and when you receive notifications from the platform.
          </p>
        ),
      },
      {
        id: '4',
        title: 'Documentation',
        icon: <DocumentIcon />,
        content: (
          <p className="text-gray-700">
            Access guides, API documentation, and integration examples.
          </p>
        ),
      },
    ],
    defaultOpen: '1',
  },
};

/**
 * Accordion items with badges showing counts or status indicators.
 */
export const WithBadges: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Active Agents',
        badge: <Badge variant="primary">3</Badge>,
        content: (
          <div className="space-y-2">
            <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
              <span className="text-sm text-gray-700">Research Assistant</span>
              <span className="text-xs text-gray-500">Last active 5m ago</span>
            </div>
            <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
              <span className="text-sm text-gray-700">Data Analyzer</span>
              <span className="text-xs text-gray-500">Last active 1h ago</span>
            </div>
            <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
              <span className="text-sm text-gray-700">Content Generator</span>
              <span className="text-xs text-gray-500">Last active 2h ago</span>
            </div>
          </div>
        ),
      },
      {
        id: '2',
        title: 'Pending Requests',
        badge: <Badge>7</Badge>,
        content: (
          <p className="text-gray-700">
            You have 7 pending consent requests waiting for your approval.
          </p>
        ),
      },
      {
        id: '3',
        title: 'Revoked Permissions',
        badge: <Badge>12</Badge>,
        content: (
          <p className="text-gray-700">
            View history of revoked agent permissions and access logs.
          </p>
        ),
      },
    ],
    defaultOpen: '1',
  },
};

/**
 * Some accordion items can be disabled to prevent user interaction.
 */
export const WithDisabled: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Available Feature',
        content: <p className="text-gray-700">This feature is available for use.</p>,
      },
      {
        id: '2',
        title: 'Coming Soon',
        description: 'This feature is currently under development',
        disabled: true,
        content: <p className="text-gray-700">Content not accessible.</p>,
      },
      {
        id: '3',
        title: 'Premium Only',
        description: 'Upgrade to access this feature',
        disabled: true,
        content: <p className="text-gray-700">Content not accessible.</p>,
      },
      {
        id: '4',
        title: 'Another Available Feature',
        content: <p className="text-gray-700">This feature is also available.</p>,
      },
    ],
    defaultOpen: '1',
  },
};

/**
 * Accordions can be nested within other accordion panels.
 */
export const Nested: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Authentication & Authorization',
        icon: <SecurityIcon />,
        content: (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 mb-4">
              Configure authentication providers and authorization rules.
            </p>
            <Accordion
              size="sm"
              items={[
                {
                  id: '1-1',
                  title: 'OAuth Providers',
                  content: (
                    <p className="text-sm text-gray-700">
                      Configure Google, GitHub, and other OAuth providers.
                    </p>
                  ),
                },
                {
                  id: '1-2',
                  title: 'SAML Configuration',
                  content: (
                    <p className="text-sm text-gray-700">
                      Set up enterprise SAML SSO integration.
                    </p>
                  ),
                },
                {
                  id: '1-3',
                  title: 'Role-Based Access',
                  content: (
                    <p className="text-sm text-gray-700">
                      Define roles and permissions for your organization.
                    </p>
                  ),
                },
              ]}
            />
          </div>
        ),
      },
      {
        id: '2',
        title: 'Agent Management',
        icon: <SettingsIcon />,
        content: (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 mb-4">
              Manage AI agents and their access permissions.
            </p>
            <Accordion
              size="sm"
              items={[
                {
                  id: '2-1',
                  title: 'Agent Registry',
                  content: (
                    <p className="text-sm text-gray-700">
                      View and manage registered AI agents.
                    </p>
                  ),
                },
                {
                  id: '2-2',
                  title: 'Scope Configuration',
                  content: (
                    <p className="text-sm text-gray-700">
                      Define available scopes and permissions.
                    </p>
                  ),
                },
              ]}
            />
          </div>
        ),
      },
    ],
    defaultOpen: '1',
  },
};

/**
 * Rich content with forms, lists, and other components inside accordion panels.
 */
export const CustomContent: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'Agent Configuration',
        icon: <SettingsIcon />,
        content: (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Agent Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-trust-deep focus:border-transparent"
                placeholder="My Research Assistant"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Description
              </label>
              <textarea
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-trust-deep focus:border-transparent"
                rows={3}
                placeholder="Describe what this agent does..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Allowed Scopes
              </label>
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm">profile:read</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm">email:read</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm">documents:write</span>
                </label>
              </div>
            </div>
            <div className="flex gap-3 pt-2">
              <button className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-800 transition-colors">
                Save Changes
              </button>
              <button className="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 transition-colors">
                Cancel
              </button>
            </div>
          </div>
        ),
      },
      {
        id: '2',
        title: 'Access Logs',
        icon: <DocumentIcon />,
        badge: <Badge>24</Badge>,
        content: (
          <div className="space-y-2">
            {[
              { action: 'Profile Read', time: '2 minutes ago', status: 'success' },
              { action: 'Email Access', time: '15 minutes ago', status: 'success' },
              { action: 'Document Write', time: '1 hour ago', status: 'denied' },
              { action: 'Profile Read', time: '3 hours ago', status: 'success' },
            ].map((log, index) => (
              <div
                key={index}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-md"
              >
                <div>
                  <div className="text-sm font-medium text-gray-900">
                    {log.action}
                  </div>
                  <div className="text-xs text-gray-500">{log.time}</div>
                </div>
                <span
                  className={`px-2 py-1 text-xs font-medium rounded ${
                    log.status === 'success'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  }`}
                >
                  {log.status}
                </span>
              </div>
            ))}
          </div>
        ),
      },
    ],
  },
};

/**
 * Controlled mode allows parent components to programmatically control
 * which items are open via state management.
 */
export const ControlledMode: Story = {
  render: () => {
    const [openIds, setOpenIds] = useState<string[]>(['1']);

    return (
      <div className="space-y-4">
        <div className="flex gap-2">
          <button
            onClick={() => setOpenIds(['1'])}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Open First
          </button>
          <button
            onClick={() => setOpenIds(['2'])}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Open Second
          </button>
          <button
            onClick={() => setOpenIds(['1', '2', '3'])}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Open All
          </button>
          <button
            onClick={() => setOpenIds([])}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Close All
          </button>
        </div>
        <Accordion
          items={basicItems}
          exclusive={false}
          defaultOpen={openIds}
          onChange={setOpenIds}
        />
        <div className="text-sm text-gray-600">
          Open items: {openIds.length > 0 ? openIds.join(', ') : 'none'}
        </div>
      </div>
    );
  },
};

/**
 * Empty state when no items are provided.
 */
export const EmptyState: Story = {
  args: {
    items: [],
  },
};

/**
 * Real-world FAQ layout with questions and answers.
 */
export const FAQLayout: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'How do I register a new AI agent?',
        content: (
          <div className="space-y-3">
            <p className="text-gray-700">
              To register a new AI agent, follow these steps:
            </p>
            <ol className="list-decimal list-inside space-y-2 text-gray-700">
              <li>Navigate to the Agent Management section</li>
              <li>Click the "Register New Agent" button</li>
              <li>Provide agent name, description, and callback URL</li>
              <li>Select the required OAuth scopes</li>
              <li>Save and copy your client credentials</li>
            </ol>
            <p className="text-gray-700 pt-2">
              Store your client secret securely - it won't be shown again.
            </p>
          </div>
        ),
      },
      {
        id: '2',
        title: 'Can I revoke agent access at any time?',
        content: (
          <div className="space-y-3">
            <p className="text-gray-700">
              Yes, you have complete control over agent access. You can revoke
              permissions at any time through the Consent Management dashboard.
            </p>
            <p className="text-gray-700">
              When you revoke access, all active sessions for that agent are
              immediately terminated, and any access tokens are invalidated.
            </p>
          </div>
        ),
      },
      {
        id: '3',
        title: 'What happens when a consent expires?',
        content: (
          <div className="space-y-3">
            <p className="text-gray-700">
              When a consent reaches its expiration time:
            </p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>The agent can no longer access your protected resources</li>
              <li>Active access tokens are invalidated</li>
              <li>The agent must request consent again to restore access</li>
              <li>All activity is logged in the audit trail</li>
            </ul>
          </div>
        ),
      },
      {
        id: '4',
        title: 'How are my credentials stored?',
        content: (
          <p className="text-gray-700">
            All sensitive data is encrypted at rest using industry-standard AES-256
            encryption. Access tokens are short-lived and stored with additional
            encryption layers. We follow OWASP best practices for credential management.
          </p>
        ),
      },
      {
        id: '5',
        title: 'What is scope-based access control?',
        content: (
          <div className="space-y-3">
            <p className="text-gray-700">
              Scopes define granular permissions for what an agent can access. Instead
              of giving blanket access to your account, you grant specific capabilities:
            </p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li><strong>profile:read</strong> - View basic profile information</li>
              <li><strong>email:read</strong> - Access email address</li>
              <li><strong>documents:read</strong> - Read your documents</li>
              <li><strong>documents:write</strong> - Create or modify documents</li>
            </ul>
            <p className="text-gray-700 pt-2">
              You can grant or revoke individual scopes without affecting others.
            </p>
          </div>
        ),
      },
      {
        id: '6',
        title: 'Is there an audit log of agent activity?',
        content: (
          <div className="space-y-3">
            <p className="text-gray-700">
              Yes, we maintain comprehensive audit logs of all agent activity:
            </p>
            <ul className="list-disc list-inside space-y-1 text-gray-700">
              <li>Every API request made by agents on your behalf</li>
              <li>Consent grants and revocations</li>
              <li>Authentication events</li>
              <li>Failed access attempts</li>
            </ul>
            <p className="text-gray-700 pt-2">
              Access your complete audit trail in the Activity section of your dashboard.
            </p>
          </div>
        ),
      },
    ],
    exclusive: true,
  },
};

/**
 * Real-world settings panel with categorized configuration options.
 */
export const SettingsPanel: Story = {
  args: {
    items: [
      {
        id: '1',
        title: 'General Settings',
        description: 'Basic configuration and preferences',
        icon: <SettingsIcon />,
        content: (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Organization Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                defaultValue="Acme Corporation"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Default Token Lifetime
              </label>
              <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option>15 minutes</option>
                <option>30 minutes</option>
                <option selected>1 hour</option>
                <option>2 hours</option>
                <option>4 hours</option>
              </select>
            </div>
            <div className="pt-2">
              <label className="flex items-center gap-2">
                <input type="checkbox" defaultChecked className="rounded" />
                <span className="text-sm text-gray-700">
                  Enable automatic consent renewal
                </span>
              </label>
            </div>
          </div>
        ),
      },
      {
        id: '2',
        title: 'Security & Authentication',
        description: 'Manage authentication methods and security policies',
        icon: <SecurityIcon />,
        badge: <Badge variant="primary">Important</Badge>,
        content: (
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">
                Authentication Methods
              </h4>
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">OAuth 2.0</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">OpenID Connect</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">SAML 2.0</span>
                </label>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">
                Security Policies
              </h4>
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">
                    Require 2FA for admin actions
                  </span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">
                    Enforce strong passwords
                  </span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">
                    Enable IP whitelisting
                  </span>
                </label>
              </div>
            </div>
          </div>
        ),
      },
      {
        id: '3',
        title: 'Notifications & Alerts',
        description: 'Configure notification preferences',
        icon: <BellIcon />,
        content: (
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">
                Email Notifications
              </h4>
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">
                    New consent requests
                  </span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked className="rounded" />
                  <span className="text-sm text-gray-700">
                    Security alerts
                  </span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">
                    Weekly activity summary
                  </span>
                </label>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Alert Email Address
              </label>
              <input
                type="email"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                defaultValue="security@acme.com"
              />
            </div>
          </div>
        ),
      },
      {
        id: '4',
        title: 'API & Integrations',
        description: 'Manage API keys and third-party integrations',
        icon: <DocumentIcon />,
        content: (
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">
                API Configuration
              </h4>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    API Base URL
                  </label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                    value="https://api.example.com"
                    readOnly
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Rate Limit
                  </label>
                  <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                    <option>100 requests/minute</option>
                    <option selected>1000 requests/minute</option>
                    <option>10000 requests/minute</option>
                    <option>Unlimited</option>
                  </select>
                </div>
              </div>
            </div>
            <button className="w-full px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-800 transition-colors">
              Generate New API Key
            </button>
          </div>
        ),
      },
    ],
    defaultOpen: '1',
    size: 'md',
  },
};
