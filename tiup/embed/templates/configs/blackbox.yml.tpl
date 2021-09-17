modules:
  http_2xx:
    prober: http
    http:
      method: GET
  http_post_2xx:
    prober: http
    http:
      method: POST
  tcp_connect:
    prober: tcp
  pop3s_banner:
    prober: tcp
    tcp:
      query_response:
        - expect: '^+OK'
      tls: true
      tls_config:
        insecure_skip_verify: false
  ssh_banner:
    prober: tcp
    tcp:
      query_response:
        - expect: '^SSH-2.0-'
  irc_banner:
    prober: tcp
    tcp:
      query_response:
        - send: 'NICK prober'
        - send: 'USER prober prober prober :prober'
        - expect: 'PING :([^ ]+)'
          send: 'PONG ${1}'
        - expect: '^:[^ ]+ 001'
  icmp:
    prober: icmp
    timeout: 5s
    icmp:
      preferred_ip_protocol: 'ip4'
