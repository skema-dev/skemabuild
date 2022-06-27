api:
  proto: {{ .ServiceLowerCase }}.proto
  stubs:
    - grpc-go
    - openapi
  version: v0.0.1
  path: https://github.com/<your_github_name>/<your_repo>/<path_in_repo>

service:
  template:
    models:
      - user
      - message
    values:
      - mysql_server_name: mysql-svc
      - mysql_password: abcd1234
      - grpc_port: "9991"
      - http_port: "9992"



