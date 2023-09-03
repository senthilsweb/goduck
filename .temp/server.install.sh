#! /bin/sh

if [ -f /etc/systemd/system/goduck.service ]; then
    sudo systemctl stop goduck.service
    sudo systemctl disable goduck.service
else
    sudo \cp -f ./goduck.service /etc/systemd/system/goduck.service
fi

sudo systemctl enable goduck.service
sudo systemctl start goduck.service