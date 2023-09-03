# "goduck" installation procedure

### Untar the file

```bash
tar -xvf goduck.tar.gz
```

### Copy the unarchive folder "goduck" to /opt

```bash
cd to /opt/goduck
```

### Install

```bash
sh ./server.install.sh

### Launch application

```bash
http://<serveripaddress>:port
```

### Start / Restart / Stop and Status the services

```bash
systemctl daemon-reload
systemctl start goduck
systemctl status goduck
systemctl restart goduck
journalctl -u goduck
journalctl -f -u goduck
```