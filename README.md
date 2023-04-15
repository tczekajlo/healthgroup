# healthgroup

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)

`healthgroup` is a tool to check the health of services with an auxiliary group of health checks.

If one of the checks in a group fails, healthgroup returns the `503` status code.

Use cases where `healthgroup` can be helpful are up to your imagination :) The typical use case would be a situation where you want to make an HTTP(S) Load Balancer aware of a service that is not placed directly in the backend, and between the LB and the service is placed another hop, e.g. proxy.

## Usage

```text
Usage of ./out/bin/healthgroup:
      --config string       config file (default is $HOME/.healthgroup.yaml)
      --in-cluster          use in-cluster config. Use always in a case when the app is running on a Kubernetes cluster
      --kubeconfig string   absolute path to the kubeconfig file (default "/Users/tczekajlo/.kube/config")
```

## Example of usage

```text
healthgroup --config ./test/testdata/config.yaml
[...]
{"level":"info","ts":1681571453.370432,"caller":"healthcheck/healthcheck.go:133","msg":"external health check","request_id":"131d2a2d-8e75-4926-8776-fe2e0470ce31","url":"https://google.com","status":200}
{"level":"info","ts":1681571453.3706791,"caller":"logger/logger.go:70","msg":"request","request_id":"131d2a2d-8e75-4926-8776-fe2e0470ce31","duration_ms":319,"status":200,"url":"/health/kubernetes/default/kubernetes","agent":"curl/7.86.0","ip":"127.0.0.1","method":"GET"}
```

## Configuration

In this section, you can learn how to configure `healthgroup`. The configuration is read in the following order:

- environment variables
- configuration file

The environment variables always take precedence over the configuration file.

### Environment variables

| Variable                         | Description                                                                                                                                            |
|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `HG_CONSUL_TIMEOUT`              | Timeout specifies a time limit for requests made to the Consul server. A Timeout of zero means no timeout.                                             |
| `HG_CONSUL_ADDRESS`              | This is the address of the Consul server.                                                                                                              |
| `HG_CONSUL_SCHEME`               | This is the URI scheme for the Consul server.                                                                                                          |
| `HG_CONSUL_TOKEN`                | This is the API access token required when access control lists (ACLs) are enabled.                                                                    |
| `HG_CONSUL_INSECURE_SKIP_VERIFY` | This is a boolean value (default `true`) to specify SSL certificate verification; setting this value to `false` is not recommended for production use. |
| `HG_CONSUL_KEY_FILE`             | Path to a client key file to use for TLS.                                                                                                              |
| `HG_CONSUL_CERT_FILE`            | Path to a client cert file to use for TLS.                                                                                                             |
| `HG_CONSUL_CA_FILE`              | Path to a CA file to use for TLS when communicating with Consul.                                                                                       |
| `HG_LOG_LEVEL`                   | Defines log level.                                                                                                                                     |
| `HG_CONSUL_ENABLED`              | Defines if requests to Consul should be enabled.                                                                                                       |
| `HG_KUBERNETES_ENABLED`          | Defines if request to Kubernetes should be enabled.                                                                                                    |
| `HG_CONCURRENCY`                 | Defines how many health checks can be executed in parallel.                                                                                            |
| `HG_SERVER_PORT`                 | Defines a port on which to listen to.                                                                                                                  |
| `HG_SERVER_ADDRESS`              | Defines a bind address of the server.                                                                                                                  |

### Configuration file

In this section you can find a configuration file with default values. The configuration file is read from the `$HOME/.healthgroup.yaml` location by default. You can use the `--config` flag to define a path to the configuration file.

Under the [`configs/config.yaml`](./configs/config.yaml) path, you can find an example of the configuration file with all parameters.

```yaml
# example
server:
  address: 127.0.0.1
  port: 8080
concurrency: 5
kubernetes:
  enabled: true
httpHealthCheck:
  - timeout: 3s
    type: https
    host: google.com
```

| Parameter                   | Description                                                                                                                       | Type                | Default          |
|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------|---------------------|------------------|
| `server.port`               | Defines a port on which to listen to                                                                                              | `int`               | `8080`           |
| `server.idleTimeout`        | The maximum amount of time to wait for the next request (when keep-alive is enabled)                                              | `string`            | `5s`             |
| `server.address`            | Settings bind address                                                                                                             | `string`            | `0.0.0.0`        |
| `kubernetes.enabled`        | Defines if Kubernetes discovery service should be enabled                                                                         | `bool`              | `true`           |
| `httpHealthCheck`           | Defines auxiliary HTTP(S) health checks                                                                                           | `httpHealthCheck[]` | `[]`             |
| `consul.token`              | The API access token                                                                                                              | `string`            | `""`             |
| `consul.timeout`            | Timeout specifies a time limit for requests made to the Consul server. A Timeout of zero means no timeout, e.g. `2s`, `30s`, `1h` | `string`            | `2s`             |
| `consul.scheme`             | The URI scheme for the Consul server (available: `http` \| `https`)                                                               | `string`            | `http`           |
| `consul.keyFile`            | Path to a client key file to use for TLS                                                                                          | `string`            | `""`             |
| `consul.insecureSkipVerify` | Specify SSL certificate verification                                                                                              | `bool`              | `false`          |
| `consul.enabled`            | Defines if Consul discovery service should be enabled                                                                             | `bool`              | `false`          |
| `consul.certFile`           | Path to a client cert file to use for TLS                                                                                         | `string`            | `""`             |
| `consul.caFile`             | Path to a CA file to use for TLS when communicating with Consul                                                                   | `string`            | `""`             |
| `consul.address`            | The address of the Consul server                                                                                                  | `string`            | `127.0.0.1:8500` |
| `concurrency`               | Defines how many health checks can be executed in parallel per request                                                            | `int`               | `5`              |

#### HTTP(s) health check specification

| Health check parameter | Description                                                                                               | Type     | Default |
|------------------------|-----------------------------------------------------------------------------------------------------------|----------|---------|
| `host`                 | Address of the target to which the probe should connect                                                   | `string` | `""`    |
| `insecureSkipVerify`   | Whether to verify SSL certificate                                                                         | `bool`   | `false` |
| `namespace`            | The namespace name that the check should be group with                                                    | `string` | `""`    |
| `port`                 | Port of the target host                                                                                   | `int`    | `null`  |
| `requestPath`          | The request path to which the probe should connect                                                        | `string` | `/`     |
| `service`              | The service name that the check should be group with                                                      | `string` | `""`    |
| `timeout`              | Timeout specifies a time limit for requests made to the Consul server. A Timeout of zero means no timeout | `string` | `0s`    |
| `type`                 | Type of the check, available: `http`, `https`, `http2`                                                    | `string` | `http`  |
| `discovery`            | Specifies a discovery service for which the check should be executed, available: `consul`, `kubernetes`   | `string` | `""`    |

## Health check grouping

By default, all auxiliary checks are executed along with the main one (every time when the `/health/*` endpoint is called). If you want to assign a particular check to a given service, namespace, or discovery, you can do it by defining the `service` or/and `namespace`, or/and `discovery` parameters in the health check specification.

For instance, if you'd like to check health of the `https://google.com` every time when you check the health of the `mytest` Kubernetes service that is located in the `staging` namespace. It's what the configuration would look like:

```yaml
httpHealthCheck:
  - type: https
    host: google.com
    namespace: staging
    service: mytest
    discovery: kubernetes
```

If one of the grouping parameters is omitted, it works like a wildcard.

```yaml
httpHealthCheck:
  # Execute the health check if the service is a Kubernetes service,
  # located in the staging namespace, and the name of the service is `mytest`.
  - type: https
    host: google.com
    namespace: staging
    service: mytest
    discovery: kubernetes
  # Execute the health check if the service name is `mytest`.
  - type: https
    host: example.com
    service: mytest
  # Execute the health check every time.
  - type: https
    host: github.com
```

## Endpoints

Below you can find a list of endpoints supported by `healthgroup`.

### Kubernetes

The Kubernetes endpoint checks the status of a Kubernetes service. If the Kubernetes services don't have any healthy endpoints then the endpoint returns the `503` status code.

If any of the auxiliary health checks failed, the endpoint returns the `503` status code.

| Method | Path                                     | Produces           |
|--------|------------------------------------------|--------------------|
| `GET`  | `/health/kubernetes/:namespace/:service` | `application/json` |

#### Path Parameters

- `namespace` `(string: <required>)` - Specifies the name of the namespace where the Kubernetes service is located.
- `service` `(string: <required>` - Specifies the name of the Kubernetes service.

#### Sample Request

```bash
curl -i http://localhost:8080/health/kubernetes/default/kubernetes
```

#### Sample Response

```text
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 53
X-Request-Id: 88ab63e5-5cbb-49eb-9840-e855e3fab1d6
```

```json
{
  "success": true,
  "message": "all health checks passed"
}
```

### Consul

The Consul endpoint checks the status of a Consul service. If the Consul services return zero healthy instances then the endpoint returns the `503` status code.

If any of the auxiliary health checks failed, the endpoint returns the `503` status code.

| Method | Path | Produces |
| -- | --| -- |
| `GET` | `/health/consul/:service` | `application/json` |
| `GET` | `/health/consul/:namespace/:service` | `application/json` |

#### Path Parameters

- `namespace` `(string: <optional>)` - Specifies the name of the namespace where the Consul service is located. The parameter works only with Consul Enterprise.
- `service` `(string: <required>)` - Specifies the name of the Consul service.

#### Query Parameters

- `tag` `(string: "")` - Specifies the tag to filter the list of instances for a given service.

#### Sample Request

```bash
curl -i http://127.0.0.1:8080/health/consul/redis
```

#### Sample Response

```text
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
Content-Length: 52
X-Request-Id: 84ab63e5-5cbb-49eb-9840-e855e3fab1d6
```

```json
{
  "success": false,
  "message": "Service is not healthy"
}
```

## Test & lint
Run linting

```bash
make lint
```

Run tests

```bash
make test
```

Whenever you need help regarding the available actions, use the `make help` command.