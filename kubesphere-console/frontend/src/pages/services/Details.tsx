// ...existing code...
import ServiceProxy from './ServiceProxy';
// ...existing code...

const ServiceDetails: React.FC = () => {
  const { namespace, service } = useParams();
  // ...existing code...
  return (
    <div>
      {/* ...existing UI... */}
      <ServiceProxy namespace={namespace} service={service} />
    </div>
  );
};

export default ServiceDetails;
// ...existing code...

