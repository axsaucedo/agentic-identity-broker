/**
 * ApprovalPage - Route-level page for /approvals/:id
 *
 * Fetches approval by ID from URL params, renders:
 * - Loading skeleton while fetching
 * - Error banner on failure
 * - ApprovalReviewPage for active approvals
 * - Confirmation for already-resolved approvals
 */

import { useParams } from 'react-router-dom';
import { AppLayout } from '@components/layout/AppLayout';
import { ApprovalLoadingSkeleton } from '@components/approvals/ApprovalLoadingSkeleton';
import { ApprovalErrorBanner } from '@components/approvals/ApprovalErrorBanner';
import { ApprovalReviewPage } from '@components/approvals/ApprovalReviewPage';
import { useApproval } from '../hooks/useApproval';

export function ApprovalPage() {
  const { id } = useParams<{ id: string }>();

  // Handle missing/invalid route param
  if (!id) {
    return (
      <AppLayout>
        <ApprovalErrorBanner errorCode="NOT_FOUND" message="No approval ID provided in the URL." />
      </AppLayout>
    );
  }

  return <ApprovalPageContent approvalId={id} />;
}

function ApprovalPageContent({ approvalId }: { approvalId: string }) {
  const {
    approval,
    loading,
    submitting,
    errorCode,
    errorMessage,
    approveResult,
    denyResult,
    approve,
    deny,
    refetch,
  } = useApproval(approvalId);

  if (loading) {
    return (
      <AppLayout>
        <ApprovalLoadingSkeleton />
      </AppLayout>
    );
  }

  if (errorCode && !approval) {
    return (
      <AppLayout>
        <ApprovalErrorBanner
          errorCode={errorCode}
          message={errorMessage}
          onRetry={refetch}
        />
      </AppLayout>
    );
  }

  if (!approval) {
    return (
      <AppLayout>
        <ApprovalErrorBanner errorCode="NOT_FOUND" />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <ApprovalReviewPage
        approval={approval}
        submitting={submitting}
        errorCode={errorCode}
        errorMessage={errorMessage}
        approveResult={approveResult}
        denyResult={denyResult}
        onApprove={approve}
        onDeny={deny}
        onRetry={refetch}
      />
    </AppLayout>
  );
}

export default ApprovalPage;
