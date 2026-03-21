/**
 * RevokeGrantDialog - Confirmation dialog for revoking all agent grants.
 *
 * Accessible confirmation dialog that:
 * - Traps focus (via Modal/Headless UI)
 * - Closes on Escape key (via Modal/Headless UI)
 * - Confirms on Enter key (via form submit)
 * - Shows loading state during async operation
 * - Uses error-primary semantic token for destructive CTA
 * - WCAG 2.1 AA compliant
 */

import { Modal } from '@design-system/components/overlays/Modal';
import { Button } from '@design-system/components/primitives/Button';

export interface RevokeGrantDialogProps {
  /** Display name of the agent whose access will be revoked */
  agentName: string;
  /** Whether the dialog is visible */
  isOpen: boolean;
  /** Called when the user confirms revocation */
  onConfirm: () => Promise<void>;
  /** Called when the user dismisses the dialog */
  onClose: () => void;
  /** Whether a revocation request is in flight */
  isLoading?: boolean;
}

/**
 * Confirmation dialog for revoking all agent grants.
 *
 * The Escape key dismisses (via Headless UI Dialog).
 * The Enter key confirms (via HTML form submit).
 */
export function RevokeGrantDialog({
  agentName,
  isOpen,
  onConfirm,
  onClose,
  isLoading = false,
}: RevokeGrantDialogProps) {
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onConfirm();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="sm"
      title="Revoke All Access"
      closeOnBackdropClick={!isLoading}
    >
      <form onSubmit={handleSubmit}>
        {/* Warning icon + body */}
        <div className="flex items-start gap-4">
          <div className="flex-shrink-0 w-12 h-12 bg-error-light rounded-full flex items-center justify-center">
            <svg
              className="w-6 h-6 text-error-primary"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
              />
            </svg>
          </div>
          <p className="text-neutral-700 mt-1">
            This will remove all permissions for{' '}
            <strong className="font-semibold text-trust-deep">{agentName}</strong>.{' '}
            Any connected services (e.g., GitHub, Google) remain active — they
            are not affected by this action.
          </p>
        </div>

        {/* Actions */}
        <div className="flex gap-3 mt-6">
          <Button
            variant="ghost"
            type="button"
            onClick={onClose}
            disabled={isLoading}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            variant="danger"
            type="submit"
            isLoading={isLoading}
            className="flex-1"
          >
            Revoke All Access
          </Button>
        </div>
      </form>
    </Modal>
  );
}

export default RevokeGrantDialog;
