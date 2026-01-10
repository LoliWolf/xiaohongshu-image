import { useState, useEffect } from 'react';
import { apiClient, Task } from '../../lib/api';
import Link from 'next/link';

export default function TaskDetailPage({ params }: { params: { id: string } }) {
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTask();
    const interval = setInterval(loadTask, 5000);
    return () => clearInterval(interval);
  }, [params.id]);

  const loadTask = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getTask(parseInt(params.id));
      setTask(data);
    } catch (error) {
      console.error('Failed to load task:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'PENDING':
        return 'bg-gray-100 text-gray-800';
      case 'EXTRACTED':
        return 'bg-blue-100 text-blue-800';
      case 'SUBMITTED':
        return 'bg-yellow-100 text-yellow-800';
      case 'RUNNING':
        return 'bg-indigo-100 text-indigo-800';
      case 'SUCCEEDED':
        return 'bg-green-100 text-green-800';
      case 'EMAILED':
        return 'bg-emerald-100 text-emerald-800';
      case 'FAILED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading && !task) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (!task) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-red-600">Task not found</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <Link href="/tasks" className="text-blue-600 hover:text-blue-800 mb-4 inline-block">
            ← Back to Tasks
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">Task #{task.id}</h1>
        </div>

        <div className="space-y-6">
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Status</h2>
            <div className="flex items-center gap-4">
              <span className={`px-3 py-1 inline-flex text-sm leading-5 font-semibold rounded-full ${getStatusColor(task.status)}`}>
                {task.status}
              </span>
              <span className="text-sm text-gray-500">
                Created: {new Date(task.created_at).toLocaleString()}
              </span>
              <span className="text-sm text-gray-500">
                Updated: {new Date(task.updated_at).toLocaleString()}
              </span>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Request Details</h2>
            <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
              <div>
                <dt className="text-sm font-medium text-gray-500">Type</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {task.request_type === 'image' ? '🖼️ Image' : '🎬 Video'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Email</dt>
                <dd className="mt-1 text-sm text-gray-900">{task.email || '-'}</dd>
              </div>
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-gray-500">Prompt</dt>
                <dd className="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded">
                  {task.prompt || '-'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Confidence</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {task.confidence !== undefined ? `${(task.confidence * 100).toFixed(1)}%` : '-'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Retry Count</dt>
                <dd className="mt-1 text-sm text-gray-900">{task.retry_count}</dd>
              </div>
            </dl>
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Provider</h2>
            <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
              <div>
                <dt className="text-sm font-medium text-gray-500">Provider Name</dt>
                <dd className="mt-1 text-sm text-gray-900">{task.provider_name || '-'}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Provider Job ID</dt>
                <dd className="mt-1 text-sm text-gray-900 font-mono">{task.provider_job_id || '-'}</dd>
              </div>
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-gray-500">Result URL</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {task.result_url ? (
                    <a
                      href={task.result_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:text-blue-800 break-all"
                    >
                      {task.result_url}
                    </a>
                  ) : '-'}
                </dd>
              </div>
              {task.error && (
                <div className="sm:col-span-2">
                  <dt className="text-sm font-medium text-gray-500">Error</dt>
                  <dd className="mt-1 text-sm text-red-600 bg-red-50 p-3 rounded">
                    {task.error}
                  </dd>
                </div>
              )}
            </dl>
          </div>

          {task.comment && (
            <div className="bg-white shadow rounded-lg p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Original Comment</h2>
              <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <div>
                  <dt className="text-sm font-medium text-gray-500">User</dt>
                  <dd className="mt-1 text-sm text-gray-900">{task.comment.user_name || '-'}</dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-gray-500">Comment UID</dt>
                  <dd className="mt-1 text-sm text-gray-900 font-mono">{task.comment.comment_uid}</dd>
                </div>
                <div className="sm:col-span-2">
                  <dt className="text-sm font-medium text-gray-500">Content</dt>
                  <dd className="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded">
                    {task.comment.content}
                  </dd>
                </div>
              </dl>
            </div>
          )}

          {task.deliveries && task.deliveries.length > 0 && (
            <div className="bg-white shadow rounded-lg p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Email Deliveries</h2>
              <div className="space-y-4">
                {task.deliveries.map((delivery) => (
                  <div key={delivery.id} className="border-l-4 pl-4">
                    <dl className="grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-3">
                      <div>
                        <dt className="text-xs font-medium text-gray-500">To</dt>
                        <dd className="mt-1 text-sm text-gray-900">{delivery.email_to}</dd>
                      </div>
                      <div>
                        <dt className="text-xs font-medium text-gray-500">Status</dt>
                        <dd className="mt-1">
                          <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                            delivery.status === 'SENT' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}>
                            {delivery.status}
                          </span>
                        </dd>
                      </div>
                      <div>
                        <dt className="text-xs font-medium text-gray-500">Sent At</dt>
                        <dd className="mt-1 text-sm text-gray-900">
                          {delivery.sent_at ? new Date(delivery.sent_at).toLocaleString() : '-'}
                        </dd>
                      </div>
                      {delivery.error && (
                        <div className="sm:col-span-3">
                          <dt className="text-xs font-medium text-gray-500">Error</dt>
                          <dd className="mt-1 text-sm text-red-600">{delivery.error}</dd>
                        </div>
                      )}
                    </dl>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
