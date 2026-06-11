import React, { useState } from 'react';
import {
Rocket,
Server,
Activity,
CheckCircle,
GitBranch,
Package,
Loader,
Cpu,
HardDrive,
} from 'lucide-react';
import toast from 'react-hot-toast';
import api from '../api';

const DeploymentPage = () => {
const [loading, setLoading] = useState(false);

const [formData, setFormData] = useState({
service_name: '',
image: '',
replicas: 1,
namespace: 'default',
});

const [deployment, setDeployment] = useState(null);

const stats = [
{
title: 'Deployments',
value: '24',
icon: <Rocket size={22} />,
},
{
title: 'Running Pods',
value: '18',
icon: <Package size={22} />,
},
{
title: 'Namespaces',
value: '4',
icon: <Server size={22} />,
},
{
title: 'Success Rate',
value: '99.5%',
icon: <CheckCircle size={22} />,
},
];

const handleChange = (e) => {
const { name, value } = e.target;


setFormData((prev) => ({
  ...prev,
  [name]:
    name === 'replicas'
      ? parseInt(value || 1)
      : value,
}));


};

const handleDeploy = async (e) => {
e.preventDefault();


setLoading(true);

try {
  const response = await api.post(
    '/deploy/create',
    formData
  );

  setDeployment(response.data);

  toast.success('Deployment started');
} catch (error) {
  toast.error(
    error.response?.data?.error ||
      'Deployment failed'
  );
} finally {
  setLoading(false);
}


};

return ( <div className="space-y-8">


  <div className="rounded-3xl bg-gradient-to-r from-blue-600 to-indigo-700 p-8 text-white">
    <div className="flex items-center gap-4">
      <Rocket size={40} />
      <div>
        <h1 className="text-3xl font-bold">
          Kubernetes Deployment Platform
        </h1>

        <p className="text-blue-100 mt-2">
          Build, deploy and manage applications
          on Kubernetes.
        </p>
      </div>
    </div>
  </div>

  <div className="grid md:grid-cols-4 gap-5">
    {stats.map((item, index) => (
      <div
        key={index}
        className="card hover:shadow-xl transition-all"
      >
        <div className="flex justify-between items-center">
          <div>
            <p className="text-gray-500 text-sm">
              {item.title}
            </p>

            <h2 className="text-3xl font-bold mt-2">
              {item.value}
            </h2>
          </div>

          <div className="p-3 rounded-xl bg-blue-100 text-blue-600">
            {item.icon}
          </div>
        </div>
      </div>
    ))}
  </div>

  <div className="grid lg:grid-cols-3 gap-6">

    <div className="lg:col-span-2 card">
      <h2 className="text-xl font-bold mb-6">
        Create Deployment
      </h2>

      <form
        onSubmit={handleDeploy}
        className="space-y-5"
      >
        <input
          type="text"
          name="service_name"
          placeholder="Service Name"
          value={formData.service_name}
          onChange={handleChange}
          className="form-input w-full"
          required
        />

        <input
          type="text"
          name="image"
          placeholder="Docker Image"
          value={formData.image}
          onChange={handleChange}
          className="form-input w-full"
          required
        />

        <div className="grid md:grid-cols-2 gap-4">
          <input
            type="number"
            min="1"
            max="20"
            name="replicas"
            value={formData.replicas}
            onChange={handleChange}
            className="form-input"
          />

          <input
            type="text"
            name="namespace"
            value={formData.namespace}
            onChange={handleChange}
            className="form-input"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="btn-primary w-full"
        >
          {loading ? (
            <>
              <Loader
                size={18}
                className="animate-spin mr-2"
              />
              Deploying...
            </>
          ) : (
            'Deploy Application'
          )}
        </button>
      </form>
    </div>

    <div className="card">
      <h2 className="text-xl font-bold mb-5">
        Cluster Health
      </h2>

      <div className="space-y-5">

        <div>
          <div className="flex justify-between mb-2">
            <span>CPU Usage</span>
            <span>45%</span>
          </div>

          <div className="h-2 bg-gray-200 rounded-full">
            <div
              className="h-2 rounded-full bg-blue-600"
              style={{ width: '45%' }}
            />
          </div>
        </div>

        <div>
          <div className="flex justify-between mb-2">
            <span>Memory</span>
            <span>62%</span>
          </div>

          <div className="h-2 bg-gray-200 rounded-full">
            <div
              className="h-2 rounded-full bg-green-600"
              style={{ width: '62%' }}
            />
          </div>
        </div>

      </div>
    </div>

  </div>
</div>


);
};

export default DeploymentPage;
