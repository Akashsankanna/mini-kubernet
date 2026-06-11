import React, { useEffect, useState } from 'react';
import { Plus, Github, FolderGit2, Loader } from 'lucide-react';
import toast from 'react-hot-toast';
import api from '../api';

const ProjectsPage = () => {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    github_url: '',
  });

  const fetchProjects = async () => {
    try {
      const response = await api.get('/projects');
      setProjects(response.data);
    } catch (error) {
      toast.error('Failed to load projects');
      console.error(error);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const handleChange = (e) => {
    const { name, value } = e.target;

    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleCreateProject = async (e) => {
    e.preventDefault();

    if (!formData.name || !formData.github_url) {
      toast.error('All fields are required');
      return;
    }

    try {
      setLoading(true);

      const response = await api.post('/projects', formData);

      setProjects((prev) => [...prev, response.data]);

      setFormData({
        name: '',
        github_url: '',
      });

      toast.success('Project created successfully');
    } catch (error) {
      toast.error(
        error.response?.data?.error ||
          'Failed to create project'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8">

      {/* Header */}
      <div className="rounded-3xl bg-gradient-to-r from-blue-600 to-indigo-700 p-8 text-white">
        <div className="flex items-center gap-4">
          <FolderGit2 size={40} />

          <div>
            <h1 className="text-3xl font-bold">
              Projects
            </h1>

            <p className="text-blue-100 mt-2">
              Create and manage GitHub projects
            </p>
          </div>
        </div>
      </div>

      {/* Create Project Form */}
      <div className="card">

        <h2 className="text-xl font-bold mb-5">
          Create New Project
        </h2>

        <form
          onSubmit={handleCreateProject}
          className="space-y-4"
        >

          <input
            type="text"
            name="name"
            placeholder="Project Name"
            value={formData.name}
            onChange={handleChange}
            className="form-input w-full"
            required
          />

          <input
            type="text"
            name="github_url"
            placeholder="https://github.com/user/repo.git"
            value={formData.github_url}
            onChange={handleChange}
            className="form-input w-full"
            required
          />

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
                Creating...
              </>
            ) : (
              <>
                <Plus size={18} />
                Create Project
              </>
            )}
          </button>

        </form>
      </div>

      {/* Projects List */}

      <div className="card">

        <h2 className="text-xl font-bold mb-5">
          My Projects
        </h2>

        {projects.length === 0 ? (
          <div className="text-center py-10 text-gray-500">
            No projects found
          </div>
        ) : (
          <div className="grid md:grid-cols-2 gap-4">

            {projects.map((project) => (

              <div
                key={project.id}
                className="border rounded-xl p-5 hover:shadow-lg transition-all"
              >

                <div className="flex justify-between items-start">

                  <div>
                    <h3 className="font-bold text-lg">
                      {project.name}
                    </h3>

                    <p className="text-sm text-gray-500 mt-2 break-all">
                      {project.github_url}
                    </p>
                  </div>

                  <Github size={20} />
                </div>

                <div className="mt-4 flex justify-between items-center">

                  <span className="px-3 py-1 rounded-full bg-green-100 text-green-700 text-sm">
                    {project.status}
                  </span>

                  <span className="text-xs text-gray-500">
                    #{project.id}
                  </span>

                </div>

              </div>

            ))}

          </div>
        )}

      </div>

    </div>
  );
};

export default ProjectsPage;