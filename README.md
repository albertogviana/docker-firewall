# docker-firewall [![Build Status](https://travis-ci.org/albertogviana/docker-firewall.svg?branch=master)](https://travis-ci.org/albertogviana/docker-firewall) [![Go Report Card](https://goreportcard.com/badge/github.com/albertogviana/docker-firewall)](https://goreportcard.com/report/github.com/albertogviana/docker-firewall)

A wrapper on top of Iptables to manage rules to block docker.

# Configuration

To use `docker-firewall` you need to create the folder  `/etc/docker-firewall`, and create the file `config.yml`. There is a sample confguration file on [example-config.yml](./example-config.yml).

It is possible to allow access from:

- interface such as `docker0` and `docker_gwbridge`

```yaml
- interface:
    - docker0
    - docker_gwbridge
```

- based on port

```yaml
- port: 5601
```

- based on IP and port

```yaml
- allow:
    - 192.168.1.15
    - 192.168.2.15
    - 192.168.3.15
    - 192.168.4.15
    port: 3000
```

- based on IP, protocol and port

```yaml
- allow:
    - 10.0.1.15
    - 10.1.0.25
  protocol: tcp
  port: 9100
```

# TODO
- Automate release process
- Validate config file and output if there is errors.
- Allow IP range
