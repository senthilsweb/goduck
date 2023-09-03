# "{{app}}" installation procedure

### Untar the file

```bash
tar -xvf {{tar_file}}
```

### Copy the unarchive folder "{{app}}" to /opt

```bash
cd to /opt/{{app}}
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
systemctl start {{app}}
systemctl status {{app}}
systemctl restart {{app}}
journalctl -u {{app}}
journalctl -f -u {{app}}
```