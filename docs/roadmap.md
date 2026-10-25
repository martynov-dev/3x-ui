# Fork roadmap

These items are planned and are not available in the current release. No release date is set.

## Manage tgwebproxy through the panel

- Inspect the existing application's configuration and supported options before designing the form.
- Show service status and provide start, stop and restart actions and logs.
- Configure the domain, local port and supported application settings using a form.
- Manage the corresponding nginx reverse proxy and HTTPS configuration, validating changes and rolling back failed updates.
- Include application configuration, binaries, service definitions and required data in full backups.
- Restore the service account, permissions, dependencies and systemd service, then enable and start tgwebproxy on a clean Ubuntu/Debian server.
- Verify the complete setup, backup and restore flow on a clean VM.

Current limitation: additional backup paths can preserve tgwebproxy files, but the restore wizard does not automatically configure and start arbitrary additional services.
