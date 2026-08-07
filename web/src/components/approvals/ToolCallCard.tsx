/**
 * ToolCallCard - Displays the user-facing approval summary for review.
 *
 * Mirrors the Tool Authorizations pending-card hierarchy so the detail page
 * presents the same approval information in the same order.
 */

import { Card } from '@design-system/components/data-display/Card';
import type { ToolApprovalDetail } from '../../types/approval';
import { ApprovalRequestSummary } from './ApprovalRequestSummary';

interface ToolCallCardProps {
  approval: ToolApprovalDetail;
  framed?: boolean;
  showExpiry?: boolean;
  showSessionContext?: boolean;
}

export function ToolCallCard({
  approval,
  framed = false,
  showExpiry = true,
  showSessionContext = true,
}: ToolCallCardProps) {
  const content = (
    <ApprovalRequestSummary
      approval={approval}
      showExpiry={showExpiry}
      showSessionContext={showSessionContext}
    />
  );

  if (!framed) {
    return content;
  }

  return (
    <Card padding="default" border="subtle">
      {content}
    </Card>
  );
}

export default ToolCallCard;
