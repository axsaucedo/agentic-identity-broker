import React, {useEffect} from 'react';

export default function LegacyDocRedirect({to}) {
  useEffect(() => {
    window.location.replace(to);
  }, [to]);

  return <a href={to}>Continue to the linked resource</a>;
}
