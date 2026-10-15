import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import MaintenancePage from '@/pages/maintenance/MaintenancePage';
import { HttpUtil, Msg } from '@/utils';

vi.mock('@/layouts/AppSidebar', () => ({ default: () => null }));
vi.mock('@/hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false, isUltra: false, antdThemeConfig: {} }),
}));

const settings = {
  enabled: false,
  intervalHours: 24,
  keep: 7,
  paths: ['/etc/nginx', '/var/www'],
  remoteHost: '',
  remotePort: 22,
  remoteUser: '',
  remotePath: '',
  identityFile: '',
};
const backupData = {
  settings,
  status: {
    lastAttempt: '0001-01-01T00:00:00Z',
    lastSuccess: '0001-01-01T00:00:00Z',
    error: '',
    file: '',
  },
  files: [],
};

function showPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <MaintenancePage />
    </QueryClientProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

describe('maintenance', () => {
  it('saves the chosen backup schedule as JSON and reports failed backups', async () => {
    vi.spyOn(HttpUtil, 'get').mockResolvedValue(new Msg(true, '', backupData));
    const post = vi.spyOn(HttpUtil, 'post').mockResolvedValue(new Msg(true));
    showPage();
    const interval = await screen.findByLabelText('Interval (hours)');
    fireEvent.change(interval, { target: { value: '12' } });
    const paths = screen.getByLabelText('Additional paths (one per line)');
    fireEvent.change(paths, { target: { value: '/etc/nginx\n/var/www\n' } });
    expect(paths).toHaveProperty('value', '/etc/nginx\n/var/www\n');
    fireEvent.change(paths, { target: { value: '/etc/nginx\n/var/www\n /srv/site \n' } });
    fireEvent.click(screen.getByRole('switch', { name: 'Automatic backups' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith(
        '/panel/api/maintenance/backups/settings',
        expect.objectContaining({
          enabled: true,
          intervalHours: 12,
          paths: ['/etc/nginx', '/var/www', '/srv/site'],
        }),
        expect.objectContaining({ headers: { 'Content-Type': 'application/json' } }),
      ),
    );
    post.mockResolvedValue(new Msg(false, 'SSH upload failed: destination unavailable'));
    fireEvent.click(screen.getByRole('button', { name: 'Back up now' }));
    expect(await screen.findByText('SSH upload failed: destination unavailable')).toBeTruthy();
  });

  it('reads existing site files and submits the edited config without replacing it with a template', async () => {
    const site = {
      configPath: '/etc/nginx/conf.d/site.conf',
      root: '/var/www/site',
      config: 'server { listen 8080; }',
      html: '<h1>Existing site</h1>',
    };
    vi.spyOn(HttpUtil, 'get').mockImplementation(async (url) => {
      if (url.endsWith('/backups')) return new Msg(true, '', backupData);
      if (url.endsWith('/sites')) return new Msg(true, '', [site]);
      return new Msg(true, '', site);
    });
    const post = vi.spyOn(HttpUtil, 'post').mockResolvedValue(new Msg(true));
    showPage();
    fireEvent.click(screen.getByRole('tab', { name: 'nginx website' }));
    const config = await screen.findByLabelText('nginx site configuration');
    expect(config).toHaveProperty('value', site.config);
    fireEvent.change(config, { target: { value: 'server { listen 8081; }' } });
    fireEvent.click(screen.getByRole('button', { name: 'Test and apply' }));
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith(
        '/panel/api/maintenance/nginx',
        { ...site, config: 'server { listen 8081; }' },
        expect.objectContaining({ headers: { 'Content-Type': 'application/json' } }),
      ),
    );
    expect(await screen.findByText('Configuration tested and nginx reloaded')).toBeTruthy();
  });
});
