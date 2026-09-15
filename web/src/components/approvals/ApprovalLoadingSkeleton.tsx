/**
 * ApprovalLoadingSkeleton - Skeleton loading state for approval review page.
 */

import { Skeleton } from '@components/ui/Skeleton';

export function ApprovalLoadingSkeleton() {
  return (
    <div className="max-w-2xl mx-auto p-6 space-y-6" aria-busy="true">
      {/* Header skeleton */}
      <div className="space-y-3">
        <Skeleton width="60%" height="2rem" />
        <Skeleton width="40%" height="1rem" />
      </div>

      {/* Tool call card skeleton */}
      <div className="card p-6 space-y-4">
        <div className="flex items-center gap-3">
          <Skeleton width="40px" height="40px" className="rounded-lg" />
          <div className="flex-1 space-y-2">
            <Skeleton width="50%" height="1.25rem" />
            <Skeleton width="30%" height="0.875rem" />
          </div>
        </div>
        <Skeleton width="100%" height="1rem" />
        <Skeleton width="80%" height="1rem" />
        <div className="space-y-2 pt-2">
          <Skeleton width="100%" height="3rem" />
          <Skeleton width="100%" height="3rem" />
        </div>
      </div>

      {/* Action buttons skeleton */}
      <div className="flex gap-3">
        <Skeleton width="120px" height="2.5rem" className="rounded-lg" />
        <Skeleton width="120px" height="2.5rem" className="rounded-lg" />
      </div>
    </div>
  );
}

export default ApprovalLoadingSkeleton;
