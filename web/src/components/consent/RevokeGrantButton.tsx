/**
 * RevokeGrantButton - Triggers the revoke-all-access confirmation flow.
 *
 * Renders a danger-variant button that opens RevokeGrantDialog on click.
 * Handles the API call and delegates the success callback to the parent.
 */

import { useState } from 'react';
import { Button } from '@design-system/components/primitives/Button';
import { useToast } from '@components/ui/Toast';
import { consentApi } from '@services/api';
import { RevokeGrantDialog } from './RevokeGrantDialog';

export interface RevokeGrantButtonProps {
  /** Unique agent identifier */
  agentId: string;
  /** Display name of the agent (used in aria-label and dialog text) */
  agentName: string;
  /** Called after a successful revocation */
  onRevoked: () => void;
}

/**
 * Button that opens a confirmation dialog and calls deleteGrant on confirm.
 * Calls onRevoked after a successful API response.
 */
export function RevokeGrantButton({
  agentId,
  agentName,
  onRevoked,
}: RevokeGrantButtonProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const { showToast } = useToast();

  const handleConfirm = async () => {
    setIsLoading(true);
    try {
      await consentApi.deleteGrant(agentId);
      setIsDialogOpen(false);
      onRevoked();
    } catch (err: unknown) {
      const message =
        err && typeof err === 'object' && 'message' in err
          ? String((err as { message: string }).message)
          : 'Failed to revoke access. Please try again.';
      showToast(message, 'error');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <Button
        variant="danger"
        aria-label={`Revoke all access for ${agentName}`}
        onClick={() => setIsDialogOpen(true)}
      >
        Revoke All Access
      </Button>

      <RevokeGrantDialog
        agentName={agentName}
        isOpen={isDialogOpen}
        onConfirm={handleConfirm}
        onClose={() => {
          if (!isLoading) setIsDialogOpen(false);
        }}
        isLoading={isLoading}
      />
    </>
  );
}

export default RevokeGrantButton;
