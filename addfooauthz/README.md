# addfooauthz
> gokit microservice demo

## Content
- [addfooauthz](#addfooauthz)
  - [Content](#content)
  - [Setup](#setup)
    - [skaffold run addfooauthz](#skaffold-run-addfooauthz)
  - [services](#services)
    - [add service](#add-service)
    - [foo service](#foo-service)
    - [authz service](#authz-service)
  - [Development](#development)
  - [Test](#test)

## Setup

- istio 1.5.6 and psql database install guide please vist [gokit-istio-security README](../README.md)
- [Skaffold](https://skaffold.dev/) 13.1

### skaffold run addfooauthz

```bash
# apply addfoodauthz directly
$ k apply -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/addfooauthz/deployments/k8s/addfooauthz-all.yaml

# or
# run addfoodauthz by skaffold and build docker image locally
$ cd addfooauthz
$ skaffold run
```

If everything setup successfully. Thare are 4 pods (add/foo/authz/postgres)

```bash
$ k get po
NAME                      READY   STATUS    RESTARTS   AGE
add-7d89cc7868-xwr4d      2/2     Running   0          45m
authz-57b6f4f7ff-22zhr    2/2     Running   1          45m
foo-569494fbd-72j7g       2/2     Running   0          45m
postgresql-postgresql-0   2/2     Running   0          51m
```

## services
- add
- foo
- authz
  - postgres database

### add service

- POST: `/api/v1/add/sum`
- POST: `/api/v1/add/concat`

### foo service

- POST: `/api/v1/foo/foo`

### authz service

- GET: `/api/v1/authz/roles`


## Development

```bash
$ make
Usage: 

  test          run unit test
  integration   run integration test
  services      build all services
  dev_dockers   quick build all
  help          Prints this help message

  add                build add service
  authz              build authz service
  foo                build foo service
  dev_docker_add     build add docker image
  dev_docker_authz   build authz docker image
  dev_docker_foo     build foo docker image
```

## Test

```bash
# sum
$ curl -X "POST" "http://localhost:80/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
{"apiVersion":"1.0.0","data":{"res":37}}

# concat
$ curl -X "POST" "http://localhost:80/api/v1/add/concat" -H 'Content-Type: application/json' -d '{ "a": "3", "b": "34"}'
{"apiVersion":"1.0.0","data":{"res":"334"}}

# foo
$ curl -X "POST" "http://localhost:80/api/v1/foo/foo" -H 'Content-Type: application/json' -d '{ "s": "foo"}'
{"apiVersion":"1.0.0","data":{"res":"foo bar"}}

# authz
$ curl -X "GET" "http://localhost:80/api/v1/authz/roles" -H 'Content-Type: application/json'
{"apiVersion":"1.0.0","data":{"items":[{"id":"G0znZWT5ajITIT97v6WXi","name":"owner","role_permissions":[{"id":"2ZaY1E3vLYs09yHgUgmeH","resource":
...
```