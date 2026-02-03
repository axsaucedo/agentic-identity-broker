/**
 * Table Component Stories
 *
 * Demonstrates all variants and use cases of the Table component.
 */

import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { Table, type TableColumn } from './Table';
import { Badge } from '@design-system/components/primitives/Badge';
import { Avatar } from '@design-system/components/primitives/Avatar';
import { Button } from '@design-system/components/primitives/Button';

const meta: Meta<typeof Table> = {
  title: 'Data Display/Table',
  component: Table,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof meta>;

// Sample data types
interface User extends Record<string, unknown> {
  id: number;
  name: string;
  email: string;
  role: string;
  status: 'active' | 'inactive' | 'pending';
  lastActive: string;
}

interface Permission extends Record<string, unknown> {
  id: number;
  scope: string;
  description: string;
  type: 'read' | 'write' | 'admin';
  granted: boolean;
}

interface Transaction extends Record<string, unknown> {
  id: string;
  date: string;
  description: string;
  amount: number;
  status: 'completed' | 'pending' | 'failed';
  category: string;
}

// Sample data
const sampleUsers: User[] = [
  {
    id: 1,
    name: 'Alice Johnson',
    email: 'alice@example.com',
    role: 'Admin',
    status: 'active',
    lastActive: '2 hours ago',
  },
  {
    id: 2,
    name: 'Bob Smith',
    email: 'bob@example.com',
    role: 'User',
    status: 'active',
    lastActive: '1 day ago',
  },
  {
    id: 3,
    name: 'Charlie Brown',
    email: 'charlie@example.com',
    role: 'User',
    status: 'inactive',
    lastActive: '1 week ago',
  },
];

const samplePermissions: Permission[] = [
  {
    id: 1,
    scope: 'user:read',
    description: 'Read user information',
    type: 'read',
    granted: true,
  },
  {
    id: 2,
    scope: 'user:write',
    description: 'Modify user information',
    type: 'write',
    granted: true,
  },
  {
    id: 3,
    scope: 'admin:access',
    description: 'Access admin dashboard',
    type: 'admin',
    granted: false,
  },
  {
    id: 4,
    scope: 'billing:read',
    description: 'View billing information',
    type: 'read',
    granted: true,
  },
  {
    id: 5,
    scope: 'billing:write',
    description: 'Modify billing settings',
    type: 'write',
    granted: false,
  },
];

const largeDataset: User[] = Array.from({ length: 25 }, (_, i) => ({
  id: i + 1,
  name: `User ${i + 1}`,
  email: `user${i + 1}@example.com`,
  role: i % 3 === 0 ? 'Admin' : 'User',
  status: (i % 3 === 0
    ? 'active'
    : i % 3 === 1
      ? 'inactive'
      : 'pending') as User['status'],
  lastActive: `${i + 1} days ago`,
}));

const sampleTransactions: Transaction[] = [
  {
    id: 'TXN-001',
    date: '2024-01-15',
    description: 'Monthly subscription',
    amount: 29.99,
    status: 'completed',
    category: 'Subscription',
  },
  {
    id: 'TXN-002',
    date: '2024-01-14',
    description: 'API usage overage',
    amount: 15.5,
    status: 'pending',
    category: 'Usage',
  },
  {
    id: 'TXN-003',
    date: '2024-01-10',
    description: 'Support ticket',
    amount: 50.0,
    status: 'failed',
    category: 'Support',
  },
];

// Basic columns
const basicColumns: TableColumn<User>[] = [
  {
    key: 'name',
    header: 'Name',
    accessor: (row) => row.name,
  },
  {
    key: 'email',
    header: 'Email',
    accessor: (row) => row.email,
  },
  {
    key: 'role',
    header: 'Role',
    accessor: (row) => row.role,
  },
];

/**
 * Basic Table
 * Simple table with 3 columns and sample data.
 */
export const BasicTable: Story = {
  render: () => (
    <Table<User>
      columns={basicColumns}
      data={sampleUsers}
      caption="User Directory"
    />
  ),
};

/**
 * Striped Table
 * Table with alternating row backgrounds for easier scanning.
 */
export const StripedTable: Story = {
  render: () => (
    <Table<User> columns={basicColumns} data={sampleUsers} striped />
  ),
};

/**
 * Hover Table
 * Rows highlight on mouse hover for better interaction feedback.
 */
export const HoverTable: Story = {
  render: () => <Table<User> columns={basicColumns} data={sampleUsers} hover />,
};

/**
 * Sortable Table
 * Table with sortable columns and sort indicators.
 * Click column headers to sort.
 */
export const SortableTable: Story = {
  render: () => {
    const [sortKey, setSortKey] = useState<string>('name');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
    const [sortedData, setSortedData] = useState(sampleUsers);

    const handleSort = (key: string) => {
      const newDirection =
        sortKey === key && sortDirection === 'asc' ? 'desc' : 'asc';
      setSortKey(key);
      setSortDirection(newDirection);

      const sorted = [...sampleUsers].sort((a, b) => {
        const aVal = a[key as keyof User] as string | number;
        const bVal = b[key as keyof User] as string | number;

        if (aVal < bVal) return newDirection === 'asc' ? -1 : 1;
        if (aVal > bVal) return newDirection === 'asc' ? 1 : -1;
        return 0;
      });

      setSortedData(sorted);
    };

    const sortableColumns: TableColumn<User>[] = [
      {
        key: 'name',
        header: 'Name',
        accessor: (row) => row.name,
        sortable: true,
      },
      {
        key: 'email',
        header: 'Email',
        accessor: (row) => row.email,
        sortable: true,
      },
      {
        key: 'role',
        header: 'Role',
        accessor: (row) => row.role,
        sortable: true,
      },
    ];

    return (
      <Table<User>
        columns={sortableColumns}
        data={sortedData}
        onSort={handleSort}
        sortKey={sortKey}
        sortDirection={sortDirection}
        striped
        hover
      />
    );
  },
};

/**
 * Compact Density
 * Table with smaller padding for denser layouts.
 */
export const CompactDensity: Story = {
  render: () => (
    <Table<User>
      columns={basicColumns}
      data={sampleUsers}
      density="compact"
      striped
    />
  ),
};

/**
 * Default Density
 * Table with standard padding (default variant).
 */
export const DefaultDensity: Story = {
  render: () => (
    <Table<User>
      columns={basicColumns}
      data={sampleUsers}
      density="default"
      striped
    />
  ),
};

/**
 * Sticky Header
 * Header stays at the top when scrolling through long tables.
 */
export const StickyHeader: Story = {
  render: () => (
    <div style={{ maxHeight: '400px', overflow: 'auto' }}>
      <Table<User>
        columns={basicColumns}
        data={largeDataset}
        stickyHeader
        striped
        hover
      />
    </div>
  ),
};

/**
 * With Custom Content
 * Table cells with icons, badges, buttons, and avatars.
 */
export const WithCustomContent: Story = {
  render: () => {
    const customColumns: TableColumn<User>[] = [
      {
        key: 'avatar',
        header: '',
        accessor: (row) => (
          <Avatar
            initials={row.name
              .split(' ')
              .map((n) => n[0])
              .join('')}
            size="sm"
          />
        ),
        width: '48px',
      },
      {
        key: 'name',
        header: 'Name',
        accessor: (row) => (
          <div>
            <div className="font-medium text-neutral-900">{row.name}</div>
            <div className="text-sm text-neutral-500">{row.email}</div>
          </div>
        ),
      },
      {
        key: 'role',
        header: 'Role',
        accessor: (row) => (
          <Badge variant={row.role === 'Admin' ? 'primary' : 'neutral'}>
            {row.role}
          </Badge>
        ),
      },
      {
        key: 'status',
        header: 'Status',
        accessor: (row) => (
          <Badge
            variant={
              row.status === 'active'
                ? 'success'
                : row.status === 'inactive'
                  ? 'neutral'
                  : 'warning'
            }
            showDot
          >
            {row.status}
          </Badge>
        ),
      },
      {
        key: 'lastActive',
        header: 'Last Active',
        accessor: (row) => (
          <span className="text-sm text-neutral-500">{row.lastActive}</span>
        ),
      },
      {
        key: 'actions',
        header: 'Actions',
        accessor: () => (
          <Button variant="ghost" size="sm">
            Edit
          </Button>
        ),
        align: 'right',
      },
    ];

    return <Table<User> columns={customColumns} data={sampleUsers} hover />;
  },
};

/**
 * Empty State
 * Table with no data, displaying empty state message.
 */
export const EmptyState: Story = {
  render: () => (
    <Table<User>
      columns={basicColumns}
      data={[]}
      empty={
        <div className="text-center py-8">
          <svg
            className="mx-auto h-12 w-12 text-neutral-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-neutral-900">
            No users found
          </h3>
          <p className="mt-1 text-sm text-neutral-500">
            Get started by adding a new user.
          </p>
          <div className="mt-6">
            <Button variant="primary" size="sm">
              Add User
            </Button>
          </div>
        </div>
      }
    />
  ),
};

/**
 * Loading State
 * Table displaying skeleton loading placeholders.
 */
export const LoadingState: Story = {
  render: () => <Table<User> columns={basicColumns} data={[]} loading />,
};

/**
 * Column Alignment
 * Table with left, center, and right-aligned columns.
 */
export const ColumnAlignment: Story = {
  render: () => {
    const alignmentColumns: TableColumn<Transaction>[] = [
      {
        key: 'id',
        header: 'ID',
        accessor: (row) => row.id,
        align: 'left',
      },
      {
        key: 'description',
        header: 'Description',
        accessor: (row) => row.description,
        align: 'left',
      },
      {
        key: 'category',
        header: 'Category',
        accessor: (row) => (
          <Badge variant="neutral" size="sm">
            {row.category}
          </Badge>
        ),
        align: 'center',
      },
      {
        key: 'amount',
        header: 'Amount',
        accessor: (row) => (
          <span className="font-medium">${row.amount.toFixed(2)}</span>
        ),
        align: 'right',
      },
    ];

    return (
      <Table<Transaction>
        columns={alignmentColumns}
        data={sampleTransactions}
        striped
      />
    );
  },
};

/**
 * Large Dataset
 * Table with 25+ rows demonstrating scroll behavior.
 */
export const LargeDataset: Story = {
  render: () => {
    const largeColumns: TableColumn<User>[] = [
      {
        key: 'id',
        header: 'ID',
        accessor: (row) => `#${row.id}`,
        width: '80px',
      },
      {
        key: 'name',
        header: 'Name',
        accessor: (row) => row.name,
        sortable: true,
      },
      {
        key: 'email',
        header: 'Email',
        accessor: (row) => row.email,
      },
      {
        key: 'role',
        header: 'Role',
        accessor: (row) => (
          <Badge
            variant={row.role === 'Admin' ? 'primary' : 'neutral'}
            size="sm"
          >
            {row.role}
          </Badge>
        ),
      },
      {
        key: 'status',
        header: 'Status',
        accessor: (row) => (
          <Badge
            variant={
              row.status === 'active'
                ? 'success'
                : row.status === 'inactive'
                  ? 'neutral'
                  : 'warning'
            }
            size="sm"
            showDot
          >
            {row.status}
          </Badge>
        ),
      },
    ];

    return (
      <div style={{ maxHeight: '500px', overflow: 'auto' }}>
        <Table<User>
          columns={largeColumns}
          data={largeDataset}
          striped
          hover
          density="compact"
        />
      </div>
    );
  },
};

/**
 * Permissions Table
 * Real-world example showing OAuth scopes and permissions.
 */
export const PermissionsTable: Story = {
  render: () => {
    const permissionsColumns: TableColumn<Permission>[] = [
      {
        key: 'scope',
        header: 'Scope',
        accessor: (row) => (
          <code className="text-sm font-mono bg-neutral-100 px-2 py-1 rounded">
            {row.scope}
          </code>
        ),
      },
      {
        key: 'description',
        header: 'Description',
        accessor: (row) => (
          <span className="text-sm text-neutral-600">{row.description}</span>
        ),
      },
      {
        key: 'type',
        header: 'Type',
        accessor: (row) => (
          <Badge
            variant={
              row.type === 'admin'
                ? 'error'
                : row.type === 'write'
                  ? 'warning'
                  : 'info'
            }
            size="sm"
          >
            {row.type}
          </Badge>
        ),
        align: 'center',
      },
      {
        key: 'granted',
        header: 'Access',
        accessor: (row) => (
          <Badge variant={row.granted ? 'success' : 'neutral'} size="sm">
            {row.granted ? 'Granted' : 'Denied'}
          </Badge>
        ),
        align: 'center',
      },
    ];

    return (
      <Table<Permission>
        columns={permissionsColumns}
        data={samplePermissions}
        caption="Application Permissions"
        striped
        hover
      />
    );
  },
};

/**
 * Responsive Table
 * Table with horizontal scroll on narrow viewports.
 */
export const ResponsiveTable: Story = {
  render: () => {
    const responsiveColumns: TableColumn<Transaction>[] = [
      {
        key: 'id',
        header: 'Transaction ID',
        accessor: (row) => row.id,
        width: '120px',
      },
      {
        key: 'date',
        header: 'Date',
        accessor: (row) => row.date,
        width: '120px',
      },
      {
        key: 'description',
        header: 'Description',
        accessor: (row) => row.description,
      },
      {
        key: 'category',
        header: 'Category',
        accessor: (row) => row.category,
        width: '120px',
      },
      {
        key: 'status',
        header: 'Status',
        accessor: (row) => (
          <Badge
            variant={
              row.status === 'completed'
                ? 'success'
                : row.status === 'pending'
                  ? 'warning'
                  : 'error'
            }
            size="sm"
            showDot
          >
            {row.status}
          </Badge>
        ),
        width: '120px',
      },
      {
        key: 'amount',
        header: 'Amount',
        accessor: (row) => (
          <span className="font-medium">${row.amount.toFixed(2)}</span>
        ),
        align: 'right',
        width: '100px',
      },
    ];

    return (
      <Table<Transaction>
        columns={responsiveColumns}
        data={sampleTransactions}
        striped
        hover
      />
    );
  },
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
};

/**
 * All Features Combined
 * Table showcasing all features: striped, hover, sortable, sticky header, custom content.
 */
export const AllFeaturesCombined: Story = {
  render: () => {
    const [sortKey, setSortKey] = useState<string>('name');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
    const [sortedData, setSortedData] = useState(largeDataset);

    const handleSort = (key: string) => {
      const newDirection =
        sortKey === key && sortDirection === 'asc' ? 'desc' : 'asc';
      setSortKey(key);
      setSortDirection(newDirection);

      const sorted = [...largeDataset].sort((a, b) => {
        const aVal = a[key as keyof User] as string | number;
        const bVal = b[key as keyof User] as string | number;

        if (aVal < bVal) return newDirection === 'asc' ? -1 : 1;
        if (aVal > bVal) return newDirection === 'asc' ? 1 : -1;
        return 0;
      });

      setSortedData(sorted);
    };

    const allFeaturesColumns: TableColumn<User>[] = [
      {
        key: 'id',
        header: 'ID',
        accessor: (row) => `#${row.id}`,
        width: '80px',
        sortable: true,
      },
      {
        key: 'avatar',
        header: '',
        accessor: (row) => (
          <Avatar
            initials={row.name
              .split(' ')
              .map((n) => n[0])
              .join('')}
            size="sm"
            status={row.status === 'active' ? 'online' : 'offline'}
          />
        ),
        width: '56px',
      },
      {
        key: 'name',
        header: 'Name',
        accessor: (row) => (
          <div>
            <div className="font-medium text-neutral-900">{row.name}</div>
            <div className="text-xs text-neutral-500">{row.email}</div>
          </div>
        ),
        sortable: true,
      },
      {
        key: 'role',
        header: 'Role',
        accessor: (row) => (
          <Badge
            variant={row.role === 'Admin' ? 'primary' : 'neutral'}
            size="sm"
          >
            {row.role}
          </Badge>
        ),
        align: 'center',
        sortable: true,
      },
      {
        key: 'status',
        header: 'Status',
        accessor: (row) => (
          <Badge
            variant={
              row.status === 'active'
                ? 'success'
                : row.status === 'inactive'
                  ? 'neutral'
                  : 'warning'
            }
            size="sm"
            showDot
          >
            {row.status}
          </Badge>
        ),
        align: 'center',
      },
      {
        key: 'lastActive',
        header: 'Last Active',
        accessor: (row) => (
          <span className="text-sm text-neutral-500">{row.lastActive}</span>
        ),
      },
      {
        key: 'actions',
        header: '',
        accessor: () => (
          <div className="flex gap-2 justify-end">
            <Button variant="ghost" size="sm">
              Edit
            </Button>
            <Button variant="ghost" size="sm">
              Delete
            </Button>
          </div>
        ),
        align: 'right',
      },
    ];

    return (
      <div style={{ maxHeight: '600px', overflow: 'auto' }}>
        <Table<User>
          columns={allFeaturesColumns}
          data={sortedData}
          caption="User Management Dashboard"
          striped
          hover
          stickyHeader
          density="compact"
          onSort={handleSort}
          sortKey={sortKey}
          sortDirection={sortDirection}
        />
      </div>
    );
  },
};
