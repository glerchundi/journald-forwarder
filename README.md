# journald-forwarder

## Using

* systemd (or [fleet](https://github.com/coreos/fleet))

```
[Unit]
Description=journald forwarder
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
ExecStartPre=-/usr/bin/docker kill journald-forwarder
ExecStartPre=-/usr/bin/docker rm journald-forwarder
ExecStartPre=/usr/bin/docker pull quay.io/glerchundi/journald-forwarder-loggly
ExecStart=/usr/bin/docker run \
--name journald-forwarder \
-v /lib64:/lib64:ro \
-v /var/log/journal:/var/log/journal:ro \
-v /usr/share/ca-certificates:/etc/ssl/certs:ro \
quay.io/glerchundi/journald-forwarder-loggly \
--loggly-token abcdefgh-ijkl-mnop-qrst-uvwxyzabcdef

[Install]
WantedBy=multi-user.target

[X-Fleet]
Global=true
```

* k8s specification

```
apiVersion: v1
kind: Pod
metadata:
  name: journald-forwarder
spec:
  hostNetwork: true
  containers:
  - name: journald-forwarder
    image: quay.io/glerchundi/journald-forwarder-loggly
    args: [ "--loggly-token", "abcdefgh-ijkl-mnop-qrst-uvwxyzabcdef" ]
    volumeMounts:
    - mountPath: /lib64
      name: lib64-host
      readOnly: true
    - mountPath: /var/log/journal
      name: journal-host
      readOnly: true
    - mountPath: /etc/ssl/certs
      name: ssl-certs-host
      readOnly: true
  volumes:
  - hostPath:
      path: /lib64
    name: lib64-host
  - hostPath:
      path: /var/log/journal
    name: journal-host
  - hostPath:
      path: /usr/share/ca-certificates
    name: ssl-certs-host
```