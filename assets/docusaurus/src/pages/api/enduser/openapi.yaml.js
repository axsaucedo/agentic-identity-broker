import useBaseUrl from '@docusaurus/useBaseUrl';
import LegacyDocRedirect from '@site/src/components/LegacyDocRedirect';

export default function EndUserOpenAPIRedirect() {
  return <LegacyDocRedirect to={useBaseUrl('/api/enduser')} />;
}
