import useBaseUrl from '@docusaurus/useBaseUrl';
import LegacyDocRedirect from '@site/src/components/LegacyDocRedirect';

export default function AdminOpenAPIRedirect() {
  return <LegacyDocRedirect to={useBaseUrl('/api/admin')} />;
}
