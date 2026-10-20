#!/bin/sh
# Keys and credentials here belong only to the isolated test environment.
set -eu
rm -f /test-ssh/ready
ssh-keygen -A
id backup >/dev/null 2>&1 || adduser -D backup
echo 'backup:Unused-test-password-2026' | chpasswd
mkdir -p /home/backup/.ssh /srv/backups/3x-ui
if [ ! -f /test-ssh/backup_ed25519 ]; then
    ssh-keygen -q -t ed25519 -N '' -f /test-ssh/backup_ed25519
fi
cp /test-ssh/backup_ed25519.pub /home/backup/.ssh/authorized_keys
chmod 700 /home/backup/.ssh
chmod 600 /home/backup/.ssh/authorized_keys
chown -R backup:backup /home/backup/.ssh /srv/backups
printf 'backup-target %s\n' "$(cat /etc/ssh/ssh_host_ed25519_key.pub)" > /test-ssh/known_hosts
cat > /etc/ssh/sshd_config <<'EOF'
Port 22
HostKey /etc/ssh/ssh_host_ed25519_key
PasswordAuthentication no
PermitRootLogin no
AllowUsers backup
Subsystem sftp internal-sftp
EOF
/usr/sbin/sshd
touch /test-ssh/ready
exec tail -f /dev/null
