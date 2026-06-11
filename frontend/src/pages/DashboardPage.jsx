import React from "react";
import { BarChart, Users, Activity, GitBranch } from "lucide-react";
import useAuthStore from "../store/authStore";
import { useNavigate } from "react-router-dom";

const DashboardPage = () => {
  const user = useAuthStore((state) => state.user);
  const navigate = useNavigate();
  const stats = [
    {
      label: "Total Deployments",
      value: "24",
      icon: <GitBranch className="text-primary-600" />,
      trend: "+12%",
    },
    {
      label: "Active Services",
      value: "18",
      icon: <Activity className="text-green-600" />,
      trend: "+5%",
    },
    {
      label: "Team Members",
      value: "12",
      icon: <Users className="text-blue-600" />,
      trend: "+2",
    },
    {
      label: "System Health",
      value: "99.9%",
      icon: <BarChart className="text-purple-600" />,
      trend: "Optimal",
    },
  ];

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="card bg-gradient-to-r from-primary-600 to-primary-700 text-white">
        <h2 className="text-2xl font-bold mb-2">
          Welcome back, {user?.first_name}! 👋
        </h2>
        <p className="text-primary-100">
          You have 3 active deployments and everything is running smoothly.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat, index) => (
          <div key={index} className="card">
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 bg-gray-100 rounded-lg">{stat.icon}</div>
              <span className="text-sm font-medium text-green-600">
                {stat.trend}
              </span>
            </div>
            <h3 className="text-gray-600 text-sm mb-1">{stat.label}</h3>
            <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Main Content Grid */}
     {/* Main Content Grid */}
<div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

  {/* Recent Projects */}
  <div className="lg:col-span-2 card">
    <h3 className="text-lg font-bold text-gray-900 mb-6">
      Recent Projects
    </h3>

    <div className="space-y-4">
      {[1, 2, 3].map((item) => (
        <div
          key={item}
          className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
        >
          <div className="flex-1">
            <h4 className="font-medium text-gray-900">
              Project #{item}
            </h4>

            <p className="text-sm text-gray-600">
              GitHub Repository Deployment
            </p>
          </div>

          <span className="badge badge-success">
            Active
          </span>
        </div>
      ))}
    </div>
  </div>

  {/* Quick Actions */}
  <div className="card">
    <h3 className="text-lg font-bold text-gray-900 mb-6">
      Quick Actions
    </h3>

    <div className="space-y-3">

      <button
        onClick={() => navigate("/projects")}
        className="btn-primary w-full justify-center"
      >
        New Project
      </button>

      <button
        onClick={() => navigate("/projects")}
        className="btn-secondary w-full justify-center"
      >
        View Projects
      </button>

      <button
        onClick={() => navigate("/deploy")}
        className="btn-secondary w-full justify-center"
      >
        Deployments
      </button>

      <button
        className="btn-secondary w-full justify-center"
      >
        View Logs
      </button>

      <button
        className="btn-secondary w-full justify-center"
      >
        Settings
      </button>

    </div>
  </div>

</div>
      {/* Services Section */}
      <div className="card">
        <h3 className="text-lg font-bold text-gray-900 mb-6">
          Services Overview
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {["API Gateway", "Auth Service", "Build Service"].map(
            (service, index) => (
              <div
                key={index}
                className="p-4 border border-gray-200 rounded-lg"
              >
                <h4 className="font-medium text-gray-900 mb-2">{service}</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">CPU Usage</span>
                    <span className="font-medium">45%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary-600 h-2 rounded-full"
                      style={{ width: "45%" }}
                    ></div>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Memory</span>
                    <span className="font-medium">62%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary-600 h-2 rounded-full"
                      style={{ width: "62%" }}
                    ></div>
                  </div>
                </div>
              </div>
            ),
          )}
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
