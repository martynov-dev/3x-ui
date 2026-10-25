import { useState, type ChangeEvent } from 'react';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Button,
  Card,
  Collapse,
  ConfigProvider,
  Form,
  Input,
  InputNumber,
  Layout,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tabs,
  Typography,
} from 'antd';
import { z } from 'zod';

import AppSidebar from '@/layouts/AppSidebar';
import { useTheme } from '@/hooks/useTheme';
import { HttpUtil } from '@/utils';
import { withBasePath } from '@/api/http-init';
import {
  FullBackupFileSchema,
  FullBackupSettingsSchema,
  FullBackupStatusSchema,
  NginxSiteSchema,
} from '@/generated/zod';
import type { FullBackupSettings, NginxSite } from '@/generated/types';

const endpoint = '/panel/api/maintenance';
const backupQueryKey = ['maintenance', 'backups'];
const backupResponse = z.object({
  settings: FullBackupSettingsSchema,
  status: FullBackupStatusSchema,
  files: z.array(FullBackupFileSchema),
});

async function getBackups() {
  const msg = await HttpUtil.get(`${endpoint}/backups`, undefined, { silent: true });
  if (!msg.success) throw new Error(msg.msg);
  return backupResponse.parse(msg.obj);
}

async function save(url: string, value?: unknown) {
  const msg = await HttpUtil.post(url, value, {
    silent: true,
    timeout: 30 * 60 * 1000,
    headers: { 'Content-Type': 'application/json' },
  });
  if (!msg.success) throw new Error(msg.msg);
}

function BackupForm({
  settings,
  mutation,
}: {
  settings: FullBackupSettings;
  mutation: UseMutationResult<void, Error, FullBackupSettings>;
}) {
  const { t } = useTranslation();
  const [form] = Form.useForm<FullBackupSettings>();
  return (
    <Form<FullBackupSettings>
      form={form}
      layout="vertical"
      validateMessages={{ required: t('maintenance.required') }}
      initialValues={settings}
      onFinish={(value) =>
        mutation.mutate({
          ...value,
          paths: value.paths.map((path) => path.trim()).filter(Boolean),
        })
      }
    >
      <Form.Item
        name="enabled"
        label={t('maintenance.enabled')}
        extra={t('maintenance.scheduleHelp')}
        valuePropName="checked"
      >
        <Switch />
      </Form.Item>
      <Space wrap align="start">
        <Form.Item
          name="intervalHours"
          label={t('maintenance.interval')}
          rules={[{ required: true }]}
        >
          <InputNumber min={1} max={8760} />
        </Form.Item>
        <Form.Item
          name="keep"
          label={t('maintenance.keep')}
          extra={t('maintenance.keepHelp')}
          rules={[{ required: true }]}
        >
          <InputNumber min={1} max={365} />
        </Form.Item>
      </Space>
      <Alert type="info" showIcon title={t('maintenance.archiveInfo')} />
      <Collapse
        defaultActiveKey={settings.remoteHost ? ['ssh'] : []}
        items={[
          {
            key: 'paths',
            label: t('maintenance.paths'),
            forceRender: true,
            children: (
              <Form.Item
                name="paths"
                label={t('maintenance.paths')}
                extra={t('maintenance.pathsHelp')}
                getValueProps={(value: string[]) => ({ value: value.join('\n') })}
                getValueFromEvent={(event: ChangeEvent<HTMLTextAreaElement>) =>
                  event.target.value.split('\n')
                }
              >
                <Input.TextArea rows={7} />
              </Form.Item>
            ),
          },
          {
            key: 'ssh',
            label: t('maintenance.ssh'),
            forceRender: true,
            children: (
              <>
                <Typography.Paragraph>{t('maintenance.sshInfo')}</Typography.Paragraph>
                <Alert type="info" showIcon title={t('maintenance.sshSetup')} />
                <Form.Item
                  name="remoteHost"
                  label={t('maintenance.host')}
                  extra={t('maintenance.hostHelp')}
                >
                  <Input placeholder="backup.example.com" />
                </Form.Item>
                <Space wrap align="start">
                  <Form.Item
                    name="remoteUser"
                    label={t('maintenance.user')}
                    dependencies={['remoteHost']}
                    rules={[({ getFieldValue }) => ({ required: !!getFieldValue('remoteHost') })]}
                  >
                    <Input />
                  </Form.Item>
                  <Form.Item
                    name="remotePort"
                    label={t('maintenance.port')}
                    dependencies={['remoteHost']}
                    rules={[({ getFieldValue }) => ({ required: !!getFieldValue('remoteHost') })]}
                  >
                    <InputNumber min={1} max={65535} />
                  </Form.Item>
                </Space>
                <Form.Item
                  name="remotePath"
                  label={t('maintenance.remotePath')}
                  extra={t('maintenance.remotePathHelp')}
                  dependencies={['remoteHost']}
                  rules={[({ getFieldValue }) => ({ required: !!getFieldValue('remoteHost') })]}
                >
                  <Input placeholder="/backups/3x-ui" />
                </Form.Item>
                <Form.Item
                  name="identityFile"
                  label={t('maintenance.keyPath')}
                  extra={t('maintenance.keyHelp')}
                  dependencies={['remoteHost']}
                  rules={[({ getFieldValue }) => ({ required: !!getFieldValue('remoteHost') })]}
                >
                  <Input placeholder="/root/.ssh/backup_ed25519" />
                </Form.Item>
              </>
            ),
          },
        ]}
      />
      {mutation.error && <Alert type="error" title={mutation.error.message} />}
      {mutation.isSuccess && <Alert type="success" title={t('maintenance.saved')} />}
      <Button type="primary" htmlType="submit" loading={mutation.isPending}>
        {t('save')}
      </Button>
    </Form>
  );
}

function BackupsTab() {
  const { t } = useTranslation();
  const client = useQueryClient();
  const query = useQuery({
    queryKey: backupQueryKey,
    queryFn: getBackups,
    refetchInterval: 30_000,
  });
  const run = useMutation({
    mutationFn: () => save(`${endpoint}/backups/run`),
    onSettled: () => client.invalidateQueries({ queryKey: backupQueryKey }),
  });
  const settingsMutation = useMutation({
    mutationFn: (value: FullBackupSettings) => save(`${endpoint}/backups/settings`, value),
    onSuccess: () => client.invalidateQueries({ queryKey: backupQueryKey }),
  });
  if (query.error) return <Alert type="error" title={query.error.message} />;
  if (!query.data) return <Spin />;
  const { settings, status, files } = query.data;
  return (
    <Space orientation="vertical" size="large" style={{ width: '100%' }}>
      <Card title={t('maintenance.archives')}>
        <Typography.Paragraph>{t('maintenance.quickStart')}</Typography.Paragraph>
        <Typography.Paragraph>{t('maintenance.downloadHelp')}</Typography.Paragraph>
        <Space wrap>
          <Button onClick={() => run.mutate()} loading={run.isPending}>
            {t('maintenance.run')}
          </Button>
          <Button onClick={() => query.refetch()}>{t('refresh')}</Button>
        </Space>
        <Typography.Paragraph>
          {t('maintenance.lastSuccess')}:{' '}
          {status.lastSuccess && !status.lastSuccess.startsWith('0001-')
            ? new Date(status.lastSuccess).toLocaleString()
            : '—'}
        </Typography.Paragraph>
        {run.isSuccess && <Alert type="success" showIcon title={t('maintenance.completed')} />}
        {(run.error || status.error) && (
          <Alert
            type="error"
            showIcon
            title={
              (run.error?.message || status.error).includes('local archive retained')
                ? t('maintenance.deliveryFailed')
                : run.error?.message || status.error
            }
            description={
              <details>
                <summary>{t('maintenance.errorDetails')}</summary>
                {run.error?.message || status.error}
              </details>
            }
          />
        )}
        <Table
          rowKey="name"
          dataSource={files}
          pagination={{ pageSize: 7 }}
          columns={[
            {
              title: t('maintenance.file'),
              dataIndex: 'name',
              render: (name: string) => (
                <a href={withBasePath(`${endpoint}/backups/download/${encodeURIComponent(name)}`)}>
                  {name}
                </a>
              ),
            },
            {
              title: t('maintenance.size'),
              dataIndex: 'size',
              render: (size: number) => `${(size / 1024 / 1024).toFixed(1)} MB`,
            },
          ]}
        />
      </Card>
      <Card title={t('maintenance.backupSettings')}>
        <BackupForm
          key={JSON.stringify(settings)}
          settings={settings}
          mutation={settingsMutation}
        />
      </Card>
    </Space>
  );
}

function SiteEditor({ site }: { site: NginxSite }) {
  const { t } = useTranslation();
  const [form] = Form.useForm<NginxSite>();
  const [domain, setDomain] = useState('');
  const [error, setError] = useState('');
  const sites = useQuery({
    queryKey: ['maintenance', 'nginx', 'sites'],
    queryFn: async () => {
      const msg = await HttpUtil.get(`${endpoint}/nginx/sites`, undefined, { silent: true });
      if (!msg.success) throw new Error(msg.msg);
      return z.array(NginxSiteSchema).parse(msg.obj);
    },
  });
  const apply = useMutation({ mutationFn: (value: NginxSite) => save(`${endpoint}/nginx`, value) });
  const load = useMutation({
    mutationFn: async () => {
      const values = await form.validateFields(['configPath', 'root']);
      const msg = await HttpUtil.get(
        `${endpoint}/nginx`,
        { configPath: values.configPath, root: values.root },
        { silent: true },
      );
      if (!msg.success) throw new Error(msg.msg);
      return NginxSiteSchema.parse(msg.obj);
    },
    onSuccess: (value) => form.setFieldsValue(value),
  });
  function template() {
    const root: unknown = form.getFieldValue('root');
    if (
      !/^[a-zA-Z0-9][a-zA-Z0-9.-]*$/.test(domain) ||
      typeof root !== 'string' ||
      !/^\/[a-zA-Z0-9_/.-]+$/.test(root)
    ) {
      setError(t('maintenance.invalidDomain'));
      return;
    }
    setError('');
    form.setFieldValue(
      'config',
      `server {\n    listen 80;\n    server_name ${domain};\n    root ${root};\n    index index.html;\n\n    location / {\n        try_files $uri $uri/ =404;\n    }\n}\n`,
    );
    form.setFieldValue(
      'html',
      '<!doctype html>\n<html lang="en"><meta charset="utf-8"><title>Welcome</title>\n<body><h1>Welcome</h1></body></html>\n',
    );
  }
  return (
    <Form<NginxSite>
      form={form}
      layout="vertical"
      initialValues={site}
      onFinish={(value) => apply.mutate(value)}
    >
      <Alert type="info" showIcon title={t('maintenance.nginxInfo')} />
      {sites.error && <Alert type="warning" title={sites.error.message} />}
      <Form.Item label={t('maintenance.detectedSites')} htmlFor="maintenance-detected-sites">
        <Select
          id="maintenance-detected-sites"
          allowClear
          options={(sites.data ?? []).map((value) => ({
            value: value.configPath,
            label: value.configPath,
          }))}
          onChange={(path: string | undefined) => {
            const selected = sites.data?.find((value) => value.configPath === path);
            if (selected)
              form.setFieldsValue({ configPath: selected.configPath, root: selected.root });
          }}
        />
      </Form.Item>
      <Form.Item name="configPath" label={t('maintenance.configPath')} rules={[{ required: true }]}>
        <Input />
      </Form.Item>
      <Form.Item name="root" label={t('maintenance.root')} rules={[{ required: true }]}>
        <Input />
      </Form.Item>
      <Space wrap>
        <Button loading={load.isPending} onClick={() => load.mutate()}>
          {t('maintenance.readFiles')}
        </Button>
        <Input
          aria-label={t('maintenance.domain')}
          placeholder="example.com"
          value={domain}
          onChange={(event) => setDomain(event.target.value)}
        />
        <Button onClick={template}>{t('maintenance.template')}</Button>
      </Space>
      <Typography.Paragraph>{t('maintenance.httpInfo')}</Typography.Paragraph>
      <Form.Item name="config" label={t('maintenance.nginxConfig')} rules={[{ required: true }]}>
        <Input.TextArea rows={14} spellCheck={false} />
      </Form.Item>
      <Form.Item name="html" label="index.html">
        <Input.TextArea rows={12} spellCheck={false} />
      </Form.Item>
      {(error || load.error || apply.error) && (
        <Alert type="error" title={error || load.error?.message || apply.error?.message} />
      )}
      {apply.isSuccess && <Alert type="success" title={t('maintenance.applied')} />}
      <Button type="primary" htmlType="submit" loading={apply.isPending}>
        {t('maintenance.apply')}
      </Button>
    </Form>
  );
}

function NginxTab() {
  const query = useQuery({
    queryKey: ['maintenance', 'nginx'],
    queryFn: async () => {
      const msg = await HttpUtil.get(`${endpoint}/nginx`, undefined, { silent: true });
      if (!msg.success) throw new Error(msg.msg);
      return NginxSiteSchema.parse(msg.obj);
    },
  });
  if (query.error) return <Alert type="error" title={query.error.message} />;
  return query.data ? <SiteEditor site={query.data} /> : <Spin />;
}

export default function MaintenancePage() {
  const { t } = useTranslation();
  const { isDark, isUltra, antdThemeConfig } = useTheme();
  return (
    <ConfigProvider theme={antdThemeConfig}>
      <Layout className={`settings-page${isDark ? ' is-dark' : ''}${isUltra ? ' is-ultra' : ''}`}>
        <AppSidebar />
        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area">
            <Card title={t('maintenance.title')}>
              <Tabs
                items={[
                  { key: 'backups', label: t('maintenance.backups'), children: <BackupsTab /> },
                  { key: 'nginx', label: t('maintenance.site'), children: <NginxTab /> },
                ]}
              />
            </Card>
          </Layout.Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
}
