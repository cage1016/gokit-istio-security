<h1 align="center">Welcome to gokit-istio-security 👋</h1>
<p>
  <a href="d" target="_blank">
    <img alt="License: LICENSE" src="https://img.shields.io/badge/License-LICENSE-yellow.svg" />
  </a>
  <a href="https://twitter.com/CageChung" target="_blank">
    <img alt="Twitter: CageChung" src="https://img.shields.io/twitter/follow/CageChung.svg?style=social" />
  </a>
</p>

> demo how to implement Authentication and custom Authorization mixer adapter by istio 1.5.6 and Open Policy Agent


## Contents
- [Contents](#contents)
- [Infrastructure](#infrastructure)
- [Setup](#setup)
  - [Kubernetes](#kubernetes)
  - [Getting Started](#getting-started)
  - [Helm install istio](#helm-install-istio)
  - [Helm install authz psql database](#helm-install-authz-psql-database)
  - [Skaffold install addfooauthz](#skaffold-install-addfooauthz)
- [Scenarios](#scenarios)
  - [Helper commands](#helper-commands)
  - [Scenario1 addfooauthz](#scenario1-addfooauthz)
  - [Scenario2 addfooauthz with istio RequestAuthentication/AuthorizationPolicy](#scenario2-addfooauthz-with-istio-requestauthenticationauthorizationpolicy)
  - [Scenario3 add custom mixer adapter](#scenario3-add-custom-mixer-adapter)
- [Clean up](#clean-up)
- [Author](#author)
- [Contributing](#contributing)
- [Show your support](#show-your-support)
- [License](#license)

## Infrastructure

![infrastructure](./infrastructure.png)

## Setup

### Kubernetes

This demo all run on Mac local machine

```bash
$ docker -v
Docker version 19.03.5, build 633a0ea

$ k cluster-info 
Kubernetes master is running at https://kubernetes.docker.internal:6443
KubeDNS is running at https://kubernetes.docker.internal:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

### Getting Started 

```bash
# clone the repo 
$ git clone git@github.com:cage1016/gokit-istio-security.git
```

### Helm install istio
```bash
# git clone istio tag 1.5.6
$ git clone -b 1.5.6 --depth 1 git@github.com:istio/istio.git

$ helm install ./istio/install/kubernetes/helm/istio-init --name istio-init --namespace istio-system

# Wait for all Istio CRDs to be created (until 26 crd created):
$ kubectl get crds | grep 'istio.io' | wc -l

# install istio with custom values
$ helm install --name istio --namespace istio-system ./istio/install/kubernetes/helm/istio -f istio-values.yaml

# enable istio sidecar inject for default namespace
$ kubectl label namespace default istio-injection=enabled
```

### Helm install authz psql database

```bash
# helm install postgres
$ helm install --name postgresql \
    --set global.postgresql.postgresqlDatabase=authz \
    --set global.postgresql.postgresqlPassword=password \
    stable/postgresql
```

### Skaffold install addfooauthz

Please visit [addfooauthz README](./addfooauthz/README.md) for more detail

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

## Scenarios

### Helper commands

```bash
$ make 
Usage: 

  owner_token                                      owner token
  viewer_token                                     viewer token
  scenario1_add_sum                                scenario 1 add    sum API
  scenario1_add_concat                             scenario 1 add    concat API
  scenario1_foo_foo                                scenario 1 foo    foo API
  scenario1_authz_roles                            scenario 1 authz  roles API
  scenario2_add_sum_owner_token                    scenario 2 add    sum API with owner token
  scenario2_add_sum_viewer_token                   scenario 2 add    sum API with viewer token
  scenario2_add_concat_owner_token                 scenario 2 add    concat API with owner token
  scenario2_add_concat_viewer_token                scenario 2 add    concat API with viewer token
  scenario2_foo_foo_owner_token                    scenario 2 foo    foo API with owner token
  scenario2_foo_foo_viewer_token                   scenario 2 foo    foo API with viewer token
  scenario2_authz_roles_owner_token                scenario 2 authz  roles API with owner token
  scenario2_authz_roles_viewer_token               scenario 2 authz  roles API with viewer token
  scenario3_add_sum_owner_token                    scenario 3 add    sum API with owner token
  scenario3_add_sum_viewer_token_permission_deny   scenario 3 add    sum API with viewer token permission deny
  scenario3_authz_roles_viewer_token               scenario 3 authz  roles API with viewer token
```

### Scenario1 addfooauthz

> run add foo authz without check authn/authz

```bash
# sum
$ make scenario1_add_sum
curl -X "POST" "http://localhost:80/api/v1/add/sum" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
{"apiVersion":"1.0.0","data":{"res":37}}

# concat
$ make scenario1_add_concat
curl -X "POST" "http://localhost:80/api/v1/add/concat" \
                -H 'Content-Type: application/json' \
                -d '{ "a": "3", "b": "34"}'
{"apiVersion":"1.0.0","data":{"res":"334"}}

# foo
$ make scenario1_foo_foo
curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
                -H 'Content-Type: application/json' \
                -d '{ "s": "foo"}'
{"apiVersion":"1.0.0","data":{"res":"foo bar"}}

# authz
$ make scenario1_authz_roles
curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
                -H 'Content-Type: application/json'
{"apiVersion":"1.0.0","data":{"items":[{"id":"G0znZWT5ajITIT97v6WXi","name":"owner","role_permissions":[{"id":"2ZaY1E3vLYs09yHgUgmeH","resource":
...
```

### Scenario2 addfooauthz with istio RequestAuthentication/AuthorizationPolicy

Install istio RequestAuthentication and AuthorizationPolicy

`authz-jwt` will check JWT token

```bash
$ kubectl apply -f authz-jwt.yaml
requestauthentication.security.istio.io/jwt-example created
authorizationpolicy.security.istio.io/require-jwt created
```

```bash
# request will get 403 Forbidden without providing JWT token
$ make scenario1_add_sum
curl -X "POST" "http://localhost:80/api/v1/add/sum" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
RBAC: access denied
```

We will generate JWT token from `RSA` private key and setup pair public JWK at istio RequestAuthentication `jwtRules`

> you could generate new RSA public/private by `sh openssl_2way_auth.sh`

authz microservice has two roles. `owner` and `viewer` 

-  "userId": "eBenfKuCzAiAC_bfqETwY" is owner
-  "userId": "0tywi6h69mdPKNN_NkI7B" is viewer

you could get token ONLY by simple command

```bash
$ make owner_token 
docker run --rm -it -v /Users/cagechung/Documents/codes/gokit-istio-security/keys:/keys -e PRIVATE_KEY_FILE=/keys/server/private/server_key.pem -e PUBLIC_KEY_FILE=/keys/server/public/server_pubkey.pem -e PAYLOAD=ewogICAgInN1YiI6ICJvd25lckBleGFtcGxlLmNvbSIsCiAgICAiaXNzIjogInRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwKICAgICJyb2xlcyI6IFsKICAgICAgICAib3duZXIiCiAgICBdLAogICAgInVzZXJJZCI6ICJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kiCn0= cage1016/gokit-istio-security-generate-token:latest
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjc4MjgyNzMsImlhdCI6MTU5NzgyODI3MywiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJBMzFzNWJLNWM1UzZ4dFVWZXI3QzB3IiwibmJmIjoxNTk3ODI4MjczLCJyb2xlcyI6WyJvd25lciJdLCJzdWIiOiJvd25lckBleGFtcGxlLmNvbSIsInVzZXJJZCI6ImVCZW5mS3VDekFpQUNfYmZxRVR3WSJ9.ys1ZY85nSGkbA6DD-uWV-YLgZoMY0tTrMkP3vEAoVTJs1Kuz1zhltw6mbZxuk87dQXfDJWUtgVjNuL9N0JrlBhbZYWLLnIWyCwnySacL2Pb6V4d6juQ8Z7L3U_YZDaRtXXRu3EPfGnSse8YDT8johWL5pSSQmmBPxbKavamfWtWuwweiN8nX-B-vz0kVoF0PYhJ99EGZDyKs47V3Q54eUCJ9TFGY0Mp_P81Uw_N_zs8sFmWPkCyqUl14Dw-2BlwZ29sTPkBn7peHjastDTNvNpb6GviBzdHeci0d7h6NSbefRckVb327Aw_uNeoCU6KzP5_lb_Ee8XfmvttLxOeBGA

$ make viewer_token
docker run --rm -it -v /Users/cagechung/Documents/codes/gokit-istio-security/keys:/keys -e PRIVATE_KEY_FILE=/keys/server/private/server_key.pem -e PUBLIC_KEY_FILE=/keys/server/public/server_pubkey.pem -e PAYLOAD=ewogICAgInN1YiI6ICJ2aWV3ZXJAZXhhbXBsZS5jb20iLAogICAgImlzcyI6ICJ0ZXN0aW5nQHNlY3VyZS5pc3Rpby5pbyIsCiAgICAicm9sZXMiOiBbCiAgICAgICAgInZpZXdlciIKICAgIF0sCiAgICAidXNlcklkIjogIjB0eXdpNmg2OW1kUEtOTl9Oa0k3QiIKfQ== cage1016/gokit-istio-security-generate-token:latest
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjc4MjgzMDIsImlhdCI6MTU5NzgyODMwMiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJXNzV4MV9xTTN2MFhyLWg4MEh2dXNnIiwibmJmIjoxNTk3ODI4MzAyLCJyb2xlcyI6WyJ2aWV3ZXIiXSwic3ViIjoidmlld2VyQGV4YW1wbGUuY29tIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.PxemJIBxTbHThI3pidNsHdNBUuWkhoe2zYMBYUSOm2ZnCknhiJIu0Vy7YRa7xMn7TSoI_ZjpH7eDh3iOBym-f4QUR_tL0onuxtzsT3jd7dm8u0ABbwAJ3jkPpneMP2S8YSVT7ZidqS3fDYVQ_vslj-YMIpigGKi2oXTuNvbDuQLUYR593hNz0Xb_oKlNbELGYC5Hy_e6atoPOsnx51K7lptmEoddy8K95KU04XF3gsPRiBuxygG6g8LrkohT_HzhrL6KywpAOJtRnVhMxB6KHYYdOixek-IPg3Lop_0zyTw1AGci3ezvvSd6uY5z-ASpZuFajs7p-3vfvDAxZL0sJg
```

istio RequestAuthentication/AuthorizationPolicy will check JWT token only. So, what ever `owner` or `viewer` will pass JWK check by providing valid token

_owner_

```bash
# request 200 ok with valid JWT token
$ make scenario2_add_sum_owner_token
curl -X "POST" "http://localhost:80/api/v1/add/sum" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDMzNDAsImlhdCI6MTU5NjAwMzM0MCwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiItTlJqVlh4aXRnZEUxM2dLeWI0Q2F3IiwibmJmIjoxNTk2MDAzMzQwLCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsib3duZXIiXSwic3ViIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJ1c2VySWQiOiJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kifQ.Tm7Ig-2lzPUy-86Rn_JdG-HLRQOM8JfMwDO-HJNc2sJMWlEtGT2-3mkUHuRzt-JnwehQv08K7R2ujNOXTatHl_R-i5lAe63degbWnV4eUCIUrHX_93Ch_KkXIAZcUIyAbF0EmTG46gReM6IUyKusY3r_Db5zvmL9_u1Kk1pGevsDQcOugomYe4aRRowIciRPjXu9_UgmY5MkjFXQ-9iJ7xjBvQ44QkLglBGceTtd9JGopRT_A9P4lSYXj7qppoDq7c2fM6goVy4HJaqjUblMaQGZTTsht-uMgZqiAyV2woOGQp8tn3kHhbfzFa74SWLEU_fmI_UdBuL5yDnvlTB8ww" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
{"apiVersion":"1.0.0","data":{"res":37}}

$ make scenario2_add_concat_owner_token
curl -X "POST" "http://localhost:80/api/v1/add/concat" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDMzNzYsImlhdCI6MTU5NjAwMzM3NiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiIwaVJuRFdxQS1aRE1rUkFLZWkycDVBIiwibmJmIjoxNTk2MDAzMzc2LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsib3duZXIiXSwic3ViIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJ1c2VySWQiOiJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kifQ.a-ug3-E20hoHMTf9lSDAQKkfXNElaqFx9goWutnLLx_HGtg5ayj409aC5iFYJOruO3SEsWPhEXG183A6kaCi8taGD-4PssWajFYerqq0mvUHtOSxlQuJRBnOlCtxzOH93SA33NYg8BUnCEg4DAdajlxENl7scW14eSyHcLIjSZFJp1StSqSMNSzccYKRhDyehwa-rzBJ1mcgNpDTYInoYZRsH4Tvt0wBKZ_60tde6O6PE7MSAohhItde-mgHtSc2IW8SyXpxr2TkTwVNBiDaVxvEG1Po8MDTW-beismZO2q3-ed-fK5qZquWn_SgAHIne3FtFbR80CEmRWOaoKmo3Q" \
                -H 'Content-Type: application/json' \
                -d  '{ "a": "3", "b": "34"}'
{"apiVersion":"1.0.0","data":{"res":"334"}}

$ make scenario2_foo_foo_owner_token
curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDMzOTksImlhdCI6MTU5NjAwMzM5OSwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJvc3lPQ0xCMm95eFBWYWVlQ2pHbVNRIiwibmJmIjoxNTk2MDAzMzk5LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsib3duZXIiXSwic3ViIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJ1c2VySWQiOiJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kifQ.cH58oJbLTMGM1qQtdJ9H_frGrbtsT3fgGviwMoI9uMosDV9BqQQ0_UufnrdG618oxaNqNpN4YbyhF4YpFxuD_bLMhjaD-MAsItGP0YghLMCdbUO8YX6ry-fAzz6sPRyBHr8v1alpJ7wMYmOLNAMz1FKzpVF_bOyW2uckrpdpLDntT3hr0zgyXqzI8LQGtu_WZwz8CNN7nqJSJCyv8VqnixjP4RTKf6H-t1UPi8C8KzU3SwXN9vkoO083iAFFWr4L5suBZCqrziA1rkM5Gl15Mn4KVhWm4gz5ASfrGJzuAm0mgQ-roqc0KuvgmrPlYofOWiJI0OnRLLSxtLig_6cJgA" \
                -H 'Content-Type: application/json' \
                -d  '{ "s": "foo"}'
{"apiVersion":"1.0.0","data":{"res":"foo bar"}}

$ make scenario2_authz_roles_owner_token
curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDM2OTksImlhdCI6MTU5NjAwMzY5OSwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJiS3VJR2N3c2JGdmhDTUtycGNRbmZRIiwibmJmIjoxNTk2MDAzNjk5LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsib3duZXIiXSwic3ViIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJ1c2VySWQiOiJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kifQ.CxvSkcPRlwSAzPcEETIci9d5j_EexVyJstgJ3C2B4fErCy9uvfn67q0jcEkFqUN_m0wWTRG_rg_xMArUbSaO3UavKXC6JcqFeeShxf7w_53mzXZ2c-jLhCrN2bS5XEYwae1sCchQoE0Adb9jhzgped85sz0VS-U3t1HnQUM7y5cx4rO7s8xlwnKugS-ll00PvIvHsadtsMLx0Uu-cvBOgAmwk2Xq-A-FHZT1bxMA2uXJqa2PZizzGJ6i54GNYl6n11YP1gBZIUicetDYcaw7mclOheYjPOy6XTIh7IHICPGR3Z53wt6SbVSp267UrIXlKIrAAy78CtQopqIkOblbuA" \
                -H 'Content-Type: application/json'
{"apiVersion":"1.0.0","data":{"items":[{"id":"G0znZWT5ajITIT97v6WXi","name":"owner","role_permissions":[{"id":"2ZaY1E3vLYs09yHgUgmeH","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"h1CZJOQVP9KoVtABc2U47","name":"DELETE","description":"DELETE","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}},{"id":"HT8npR6ikexUaHUe5jpsW","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"c4Jn_pJxYpBO~RAcLK6vh","name":"POST","description":"POST","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}...                
```

_viewer_

```bash
$ make scenario2_add_sum_viewer_token
curl -X "POST" "http://localhost:80/api/v1/add/sum" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDMzNjMsImlhdCI6MTU5NjAwMzM2MywiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiIzcWVzRjdkVlJvQXJQVHV6ZEVvUURnIiwibmJmIjoxNTk2MDAzMzYzLCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.stL5f5hOofxhb0BE1GGTP6O4Mim5i_W514_JBfzuV7PUTX4fvwg4mYQJfPuTk2Uy5lAseft0ZkDWUylTjAlTRvxwDRStsalIcLcCb5w9NzlM1bHSS1yOBS8PIBxx3uD79oxRSm8nHhyEIvMqoM8TkRPTdknzaWix6kBmlObus3ZTKJzMTWXEBNy3tm2tuMj_fJsLzEi9wFpu2GmADDOg5UP11BV8Ulcuna2z4mQmZth_o-k6VxOC2hpI1DhTAU6MLXb3nu5_hlKoAsrhNN9jPm6vSznEzJALhHtvbhwtIuI-4z01WSb7bW7x4X-GmGVwrdrWX2B-aXFbpiISMmTqMA" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
{"apiVersion":"1.0.0","data":{"res":37}}

$ make scenario2_add_concat_viewer_token
curl -X "POST" "http://localhost:80/api/v1/add/concat" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDMzODgsImlhdCI6MTU5NjAwMzM4OCwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJBOFlFOHJ6Z2NjSkU2WGlrWVE5ZTdRIiwibmJmIjoxNTk2MDAzMzg4LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.YooAXQslb9TIVuXsYAKZhCgHf_nxKCKUl07jKVrR0o4z3-U3wyoxAk9d1sX8OznIgUklWs5ZGd0hAINkf7YnwmaU8jkivQeAStyx6EUGaHfvq-p5tSZxLu2U5WDS4Teu-yMKUv6jMDiQh0QYHfZwneicr4Zlw5j7mrELabIFZeDnXRosWK4r0E0bwbB8mDaQBP-q_T3aw85n9UVl2N41QfjDxiEFAb3kkLzpfsf6KifzszmwWky10YPPK5aAwJ2jZFQsTH2n1AVltbQajQ6T6tHV13UfNV-S-XBLkcT4f-M7yJ5rUhFkdSipwRvqJzdfeNb34M30HIxgbEN7XWC8XA" \
                -H 'Content-Type: application/json' \
                -d  '{ "a": "3", "b": "34"}'
{"apiVersion":"1.0.0","data":{"res":"334"}}

$ make scenario2_foo_foo_viewer_token
curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDM0MTMsImlhdCI6MTU5NjAwMzQxMywiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiItcTJaVmprOEdSR09CSDE3ZHgtbnBBIiwibmJmIjoxNTk2MDAzNDEzLCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.LQd8s7wscjSr1iwXfOo9mNWJNxBJ9WdFNNJ9h4j7LcisSab_IX2z-AeJJsyXT7B8-jT3MB3hoRAicd53Mzvg6hDr14Z9oPFilELO7M9xWvMPHO_zZF_hRHNL9w0X0wLd0CxVWFjecy7ksnkEUhX89Fn1O_YYHmwR6q14_ir4jHed5XIGTpe6PAuSoSzJ4-RnUvaMwL4gd3iHjJqvUCia1XTI3XLE6bsqLLkVEGl1zZmQmM2AH18-EZ3x0aPv8CkrQo5lfpp9AyreoVosIPFfJM9UgPVptr58nZOHTbpJy3WECAqiUUSm1r_PJlaBRf-9IWZPm-D87kClcSkVIApUMw" \
                -H 'Content-Type: application/json' \
                -d  '{ "s": "foo"}'
{"apiVersion":"1.0.0","data":{"res":"foo bar"}}

$ make scenario2_authz_roles_viewer_token
curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDM3NDIsImlhdCI6MTU5NjAwMzc0MiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiI2Zk4wcE9vbE5iOUxXdHZ0a196WFV3IiwibmJmIjoxNTk2MDAzNzQyLCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.caOeRjV04mMrlBcVGjjRT8dbBIZnHWnrmaeQp21WymHNwyyquJja2E_HrTYjki4v6-pzVTspo8B3IOf7H9NKkRf-pluiQ7jTCEuz2VFIZIVpCmoisquLHVYfim014YO_vDkYVhzYGlGO0h2lHpBrLnE5WgTa1lFR2eQI11-uRK8fXidF-Nxz7ly7lzu9pwMoRsKgyRdMzFht9fTY6xq2OHhNynV50EvlVJOwWz8tvy1HO_32Wa_fNmAIN1FeL1MAVOWnFYaQWMSNahoflWaZeHQbgIZh09UFBPD6yXEGccKbyAAO7L2M89UR9F6eANcIv-C7C8WWeB42MCTVI0uTOQ" \
                -H 'Content-Type: application/json'
{"apiVersion":"1.0.0","data":{"items":[{"id":"G0znZWT5ajITIT97v6WXi","name":"owner","role_permissions":[{"id":"2ZaY1E3vLYs09yHgUgmeH","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"h1CZJOQVP9KoVtABc2U47","name":"DELETE","description":"DELETE","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}},{"id":"HT8npR6ikexUaHUe5jpsW","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"c4Jn_pJxYpBO~RAcLK6vh","name":"POST","description":"POST","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}
```

### Scenario3 add custom mixer adapter

Please visit [authzopa (istio custom authorization mixer adapter) README](./authzopa/README.md) for more detail how to create custom istio mixier adpater

install custom mixer adapter `authzopa`

```bash
$ k apply -f authzopa/operatorconfig
attributemanifest.config.istio.io/istio-proxy created
attributemanifest.config.istio.io/kubernetes created
adapter.config.istio.io/authzopa-adapter created
service/authzopa-adapter-service created
deployment.extensions/authzopa-adapter created
handler.config.istio.io/h1 created
instance.config.istio.io/i1authorization created
rule.config.istio.io/r1 created
template.config.istio.io/authorization created
```

_OPA data_

```json
{
    "rolePermissions": {
        "editor": ...
        "owner": [
            {
                "method": "POST",
                "path": "/api/([^\\\\/]+)/foo/foo"
            },
            {
                "method": "GET",
                "path": "/api/([^\\\\/]+)/authz/roles"
            },
            {
                "method": "PUT",
                "path": "/api/([^\\\\/]+)/authz/roles"
            },
            {
                "method": "POST",
                "path": "/api/([^\\\\/]+)/authz/roles"
            },
            {
                "method": "DELETE",
                "path": "/api/([^\\\\/]+)/authz/roles"
            },
            {
                "method": "POST",
                "path": "/api/([^\\\\/]+)/add/sum"
            },
            {
                "method": "POST",
                "path": "/api/([^\\\\/]+)/add/concat"
            }
        ],
        "viewer": [
            {
                "method": "GET",
                "path": "/api/([^\\\\/]+)/authz/roles"
            }
        ]
    },
    "userRoles": {
        "0tywi6h69mdPKNN_NkI7B": [
            "viewer"
        ],
        "eBenfKuCzAiAC_bfqETwY": [
            "owner"
        ]
    }
}
```

As above OPA data describe. `owner` allow to access ALL APIs and `viewer` ONLY allow to access `/api/v1/authz/roles` API. Our custom istio authorization mixer adapter `authzopa` will verify REQUEST by calling `authz` `/pb.Authz/IsAuthorizedReq` GRPC

> You could try this [The Rego Playground: viewer ONLY allow access /api/v1/authz/roles](https://play.openpolicyagent.org/p/RktBtYp5Tx) for above data

```bash
# owner could access ALL APIs
$ make scenario3_add_sum_owner_token
curl -X "POST" "http://localhost:80/api/v1/add/sum" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDM4MjYsImlhdCI6MTU5NjAwMzgyNiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJsSHU5UU4xaE94dGJNU1pDWDEwS3JRIiwibmJmIjoxNTk2MDAzODI2LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsib3duZXIiXSwic3ViIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJ1c2VySWQiOiJlQmVuZkt1Q3pBaUFDX2JmcUVUd1kifQ.iA8c-K6b2vTABsjWohQ6z1sKlvzCFlnKd4-Rt_e830PxljLalMhw5wSqMsSgKoUxsdqMwbA4-2Tlds-3Xa00VJdhp7B8ZvtcxBEkRu0X48aah--70p8uCKeAqAxwgiHAj2mQChCbhXtWQwz94kZpS8t5QRbD5czWeJkn0kMgG0nv60dwsaDlPGE_DDwUrMK7oJ4cxU2oz3DeCzR1klX3mhq-ABUiKCZ9e_B9Os0JJhcTWeJGqThDMO0bQpm0jlwQPnvnaZfLopavjLrtixAs-DWGLxi21BB2E5ItV-ubY8P02dbbxSLahELHloMnBXXDoI_QZtew1SJZj495xDY0Nw" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
{"apiVersion":"1.0.0","data":{"res":37}}

# viewer could NOT access ANY APIs except /api/v1/authz/roles
$ make scenario3_add_sum_viewer_token_permission_deny
curl -v "POST" "http://localhost:80/api/v1/add/sum" \
                -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDc4NTUsImlhdCI6MTU5NjAwNzg1NSwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiItMWZ2TTdNUm1TSnZ0NFAzd1dhaEFnIiwibmJmIjoxNTk2MDA3ODU1LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.UJzDTlGBAchlOiX2CKqctWy2tKnWQI0W4GOOZSsoPWbAD7looEYa_4O1xFFmF-QhRIViWh4Wf9RLSU8iuj1UWr9zJsh948f-5hRjazQ_dwhe6rQLbxvskA7yf8ocSkAPJPmyR4FjivdlBQTopVlWlswcMcsrFcspKTzudmYviIdKnRhTJo6HFuDwcqmXMvlD73o1fzI7KXkH7k6xgoWiT2nGwg1Vw7CyQhihfuKlYn8ClOOAhp38o6vU0kRCpMfXdb3TIBulwQ8Xv3Q0aai00uya23qTTnyGna7A4yKtt469Dbe2IBzqcn3O-PzaSaMnnJ2qHRBcBRbm1Nddc8xkvw" \
                -H 'Content-Type: application/json' \
                -d '{ "a": 3, "b": 34}'
* Could not resolve host: POST
* Closing connection 0
curl: (6) Could not resolve host: POST
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 80 (#1)
> POST /api/v1/add/sum HTTP/1.1
> Host: localhost
> User-Agent: curl/7.64.1
> Accept: */*
> Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDc4NTUsImlhdCI6MTU5NjAwNzg1NSwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiItMWZ2TTdNUm1TSnZ0NFAzd1dhaEFnIiwibmJmIjoxNTk2MDA3ODU1LCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.UJzDTlGBAchlOiX2CKqctWy2tKnWQI0W4GOOZSsoPWbAD7looEYa_4O1xFFmF-QhRIViWh4Wf9RLSU8iuj1UWr9zJsh948f-5hRjazQ_dwhe6rQLbxvskA7yf8ocSkAPJPmyR4FjivdlBQTopVlWlswcMcsrFcspKTzudmYviIdKnRhTJo6HFuDwcqmXMvlD73o1fzI7KXkH7k6xgoWiT2nGwg1Vw7CyQhihfuKlYn8ClOOAhp38o6vU0kRCpMfXdb3TIBulwQ8Xv3Q0aai00uya23qTTnyGna7A4yKtt469Dbe2IBzqcn3O-PzaSaMnnJ2qHRBcBRbm1Nddc8xkvw
> Content-Type: application/json
> Content-Length: 18
> 
* upload completely sent off: 18 out of 18 bytes
< HTTP/1.1 403 Forbidden
< content-length: 71
< content-type: text/plain
< date: Wed, 29 Jul 2020 07:30:55 GMT
< server: istio-envoy
< 
* Connection #1 to host localhost left intact
PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny* Closing connection 1

# viewer ONLY allow to access /api/v1/authz/roles
$ make scenario3_authz_roles_viewer_token
curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
                        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjYwMDY4NDIsImlhdCI6MTU5NjAwNjg0MiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJISEpmaFdzUnpGeDZud0lxZGNvek93IiwibmJmIjoxNTk2MDA2ODQyLCJwcm92aWRlciI6Im15cW5hcGNsb3VkIiwicm9sZXMiOlsidmlld2VyIl0sInN1YiI6InRlc3RpbmdAc2VjdXJlLmlzdGlvLmlvIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.aVw_-yr4mUdoz9LYf19tCGSum7AexJFlv7hkcM1oOSdqQw-zwbHvkFaYdwO8zT1Ky1AfSUFEZ6DUqx8m35K6cfM9o6TaaeL1eRxc-eyKV6xYk3AeKk23KcraFRgN21RHACY3Rqgi7Yl_cOjkZZjHqAMIgxUlw9xCWriPxSatYKjJC5o8tEhm2-Qr-nN8imABVtJq53pxy4SzydvNIJc_sWyzRpBNeR_2Fd-vapy_tyfhOtatKffXpY5ZpAsIlfehjaYMxFVIUO-sq9rE4hh-cWReF7I2bOgdG7p89YRZKcKZiANJM1hCu1_AzWz36A61ZO0yti6XuE1_6A9DuvAEbQ" \
                        -H 'Content-Type: application/json'
{"apiVersion":"1.0.0","data":{"items":[{"id":"G0znZWT5ajITIT97v6WXi","name":"owner","role_permissions":[{"id":"2ZaY1E3vLYs09yHgUgmeH","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"h1CZJOQVP9KoVtABc2U47","name":"DELETE","description":"DELETE","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}},{"id":"HT8npR6ikexUaHUe5jpsW","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36:33Z"},"action":{"id":"c4Jn_pJxYpBO~RAcLK6vh","name":"POST","description":"POST","createdAt":"2019-07-10T16:06:58Z","updatedAt":"2019-07-10T16:06:58Z"}},{"id":"B9ZMONWtZgrPagOT1WEDw","resource":{"id":"gsK3pwCVlddkXUE8B_1mr","name":"/api/([^\\\\/]+)/authz/roles","createdAt":"2019-07-10T15:55:43Z","updatedAt":"2019-12-26T03:36.....
```

## Clean up

```bash
# delete custom mixer adapter
$ k delete -f authzopa/operatorconfig

# delete istion RequestAuthentication/AuthorizationPolicy
$ k delete -f authz-jwt.yaml

# delete addfooauthz
$ cd addfooauthz && skaffold delete

# delete postgres
$ helm del --purge postgresql

# delete istio
$ helm del --purge istio
$ helm del --purge istio-init
$ k delete ns istio-system
```

## Author

👤 **KAI CHU CHUNG**

* Website: https://kaichu.io/
* Twitter: [@CageChung](https://twitter.com/CageChung)
* Github: [@cage1016](https://github.com/cage1016)
* LinkedIn: [@kaichuchung](https://linkedin.com/in/kaichuchung)

## Contributing

Contributions, issues and feature requests are welcome!<br />Feel free to check [issues page](https://github.com/cage1016/gokit-istio-security/issues). 

## Show your support

Give a ⭐️ if this project helped you!

## License

Copyright © 2020 [KAI CHU CHUNG](https://github.com/cage1016).<br />
This project is [LICENSE](./LICENSE) licensed.

***
_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_