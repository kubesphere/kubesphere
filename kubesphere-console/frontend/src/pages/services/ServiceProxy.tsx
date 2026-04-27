import React, { useState } from 'react';
import { Button } from '@kubesphere/console';
import { notify } from '@kubesphere/console';

interface ServiceProxyProps {
  namespace: string;
  service: string;
}

const ServiceProxy: React.FC<ServiceProxyProps> = ({ namespace, service }) => {
  const [loading, setLoading] = useState(false);

  const handleProxyClick = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/proxy/services/${namespace}/${service}/`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });
      if (response.ok) {
        window.open(response.url, '_blank');
      } else {
        notify.error('Failed to access service');
      }
    } catch (error) {
      notify.error('Error proxying service');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button loading={loading} onClick={handleProxyClick}>
      Access Service via Proxy
    </Button>
  );
};

export default ServiceProxy;

