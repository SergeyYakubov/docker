#!/bin/sh


mkdir $HOME


mkdir -p $HOME/etc $HOME/var/run
echo "UsePrivilegeSeparation no" > $HOME/etc/sshd_config
echo "PidFile $HOME/var/run/sshd.pid" >> $HOME/etc/sshd_config
echo "HostKey $HOME/etc/ssh_host_rsa_key" >> $HOME/etc/sshd_config
echo "Port $DOCKER_SSHPORT" >> $HOME/etc/sshd_config
ssh-keygen -t rsa -f  $HOME/etc/ssh_host_rsa_key -N '' > /dev/null


mkdir $HOME/.ssh
cp /ssh_keys/id_rsa* $HOME/.ssh
mv $HOME/.ssh/id_rsa.pub $HOME/.ssh/authorized_keys
chown -R $UID:`id -g` $HOME/.ssh
chmod 600 $HOME/.ssh/id_rsa
chmod 700  $HOME/.ssh
chmod 600 $HOME/.ssh/authorized_keys
echo "Port $DOCKER_SSHPORT" > $HOME/.ssh/config
echo "LogLevel=ERROR" >> $HOME/.ssh/config
echo "StrictHostKeyChecking no" >> $HOME/.ssh/config
echo " . /etc/bashrc">$HOME/.bashrc
/usr/sbin/sshd -f $HOME/etc/sshd_config -D > /dev/null

