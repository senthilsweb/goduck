#! /bin/sh

if [ -f /etc/systemd/system/{{app}}.service ]; then
    sudo systemctl stop {{app}}.service
    sudo systemctl disable {{app}}.service
else
    sudo \cp -f ./{{app}}.service /etc/systemd/system/{{app}}.service
fi

sudo systemctl enable {{app}}.service
sudo systemctl start {{app}}.service