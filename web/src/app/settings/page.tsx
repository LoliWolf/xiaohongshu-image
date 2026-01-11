'use client';

import { useState, useEffect } from 'react';
import { apiClient, Setting } from '@/src/lib/api';

export default function SettingsPage() {
  const [settings, setSettings] = useState<Setting | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getSettings();
      setSettings(data);
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to load settings' });
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!settings) return;

    try {
      setSaving(true);
      await apiClient.updateSettings(settings);
      setMessage({ type: 'success', text: 'Settings saved successfully' });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to save settings' });
    } finally {
      setSaving(false);
    }
  };

  const handleRunPoll = async () => {
    try {
      await apiClient.runPoll();
      setMessage({ type: 'success', text: 'Poll job enqueued successfully' });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to run poll' });
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-red-600">Failed to load settings</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
          <p className="mt-2 text-gray-600">Configure your Xiaohongshu image generation system</p>
        </div>

        {message && (
          <div className={`mb-6 p-4 rounded-md ${
            message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
          }`}>
            {message.text}
          </div>
        )}

        <div className="bg-white shadow rounded-lg p-6 space-y-6">
          <div>
            <h2 className="text-lg font-medium text-gray-900 mb-4">Xiaohongshu Connector</h2>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Connector Mode</label>
                <select
                  value={settings.connector_mode}
                  onChange={(e) => setSettings({ ...settings, connector_mode: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="mock">Mock (Demo)</option>
                  <option value="mcp">MCP (Real)</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Note Target (URL or ID)</label>
                <input
                  type="text"
                  value={settings.note_target}
                  onChange={(e) => setSettings({ ...settings, note_target: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://www.xiaohongshu.com/explore/12345678"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Polling Interval (seconds)</label>
                <input
                  type="number"
                  value={settings.polling_interval_sec}
                  onChange={(e) => setSettings({ ...settings, polling_interval_sec: parseInt(e.target.value) || 120 })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="10"
                />
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-lg font-medium text-gray-900 mb-4">LLM Configuration</h2>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Base URL</label>
                <input
                  type="text"
                  value={settings.llm_base_url || ''}
                  onChange={(e) => setSettings({ ...settings, llm_base_url: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://api.openai.com/v1"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">API Key</label>
                <input
                  type="password"
                  value={settings.llm_api_key || ''}
                  onChange={(e) => setSettings({ ...settings, llm_api_key: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="sk-..."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
                <input
                  type="text"
                  value={settings.llm_model || ''}
                  onChange={(e) => setSettings({ ...settings, llm_model: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="gpt-4o-mini"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Intent Threshold</label>
                <input
                  type="number"
                  step="0.01"
                  value={settings.intent_threshold}
                  onChange={(e) => setSettings({ ...settings, intent_threshold: parseFloat(e.target.value) || 0.7 })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="0"
                  max="1"
                />
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-lg font-medium text-gray-900 mb-4">SMTP Configuration</h2>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Host</label>
                <input
                  type="text"
                  value={settings.smtp_host || ''}
                  onChange={(e) => setSettings({ ...settings, smtp_host: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="mailhog"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
                <input
                  type="number"
                  value={settings.smtp_port || ''}
                  onChange={(e) => setSettings({ ...settings, smtp_port: parseInt(e.target.value) || 1025 })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="1"
                  max="65535"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">From</label>
                <input
                  type="email"
                  value={settings.smtp_from || ''}
                  onChange={(e) => setSettings({ ...settings, smtp_from: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="noreply@example.com"
                />
              </div>
            </div>
          </div>

          <div className="flex gap-4">
            <button
              onClick={handleSave}
              disabled={saving}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
            <button
              onClick={handleRunPoll}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
            >
              Run Poll Now
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
