<h1 align="center">Welcome to gokit-istio-security 👋</h1>
<p>
  <a href="d" target="_blank">
    <img alt="License: LICENSE" src="https://img.shields.io/badge/License-LICENSE-yellow.svg" />
  </a>
  <a href="https://twitter.com/CageChung" target="_blank">
    <img alt="Twitter: CageChung" src="https://img.shields.io/twitter/follow/CageChung.svg?style=social" />
  </a>
</p>

> 1. demo how to implement Authentication and custom Authorization mixer adapter by and Open Policy Agent
> 2. demo how to implement Authentication and custom Authorization with envoy external and Open Policy Agent


## Services

| Service | Description           | repo                                                                                               |
| ------- | --------------------- | -------------------------------------------------------------------------------------------------- |
| add     | Expose Sum method     | [cage1016/ms-demo: gokit microservice demo](https://github.com/cage1016/ms-demo)                   |
| tictac  | Expose Tic/Tac method | [cage1016/ms-demo: gokit microservice demo](https://github.com/cage1016/ms-demo)                   |
| authz   | authorization RBAC    | [cage1016/ms-demo-authz: gokit microservice demo authz](https://github.com/cage1016/ms-demo-authz) |

## Features

- **[Kubernetes](https://kubernetes.io)/[GKE](https://cloud.google.com/kubernetes-engine/):**
  The app is designed to run on Kubernetes (both locally on "Docker for
  Desktop", as well as on the cloud with GKE).
- **[gRPC](https://grpc.io):** Microservices use a high volume of gRPC calls to
  communicate to each other.
- **[Istio](https://istio.io):** Application works on Istio service mesh.
- **[Skaffold](https://skaffold.dev):** Application
  is deployed to Kubernetes with a single command using Skaffold.
- **[go-kit/kit](https://github.com/go-kit/kit):** Go kit is a programming toolkit for building microservices (or elegant monoliths) in Go. We solve common problems in distributed systems and application architecture so you can focus on delivering business value.
- **[open-policy-agent/opa](https://github.com/open-policy-agent/opa):** The Open Policy Agent (OPA) is an open source, general-purpose policy engine that enables unified, context-aware policy enforcement across the entire stack.

## Infrastructure
- Mixer
![infrastructure Mixer](./infrastructure-mixer.png)

- Envoy External
![infrastructure Envoy External](./infrastructure-ef.png)

## Prepare
1. istio require 1.5

## Scenario1: install ms-demo

> - install ms-demo 
> - install ms-demo-authz

1. Install [cage1016/ms-demo: gokit microservice demo](https://github.com/cage1016/ms-demo)
    ```sh
    # install ms-demo
    kubectl apply -f https://raw.githubusercontent.com/cage1016/ms-demo/master/deployments/all.yaml

    # install ms-demo istio
    kubectl apply -f https://raw.githubusercontent.com/cage1016/ms-demo/master/deployments/gateway.yaml
    ```
2. Install [cage1016/ms-demo-authz: gokit microservice demo authz](https://github.com/cage1016/ms-demo-authz)
    ```sh
    # install ms-demo-authz
    kubectl apply -f https://raw.githubusercontent.com/cage1016/ms-demo-authz/master/deployments/all.yaml

    # install ms-demo-authz istio
    kubectl apply -f https://raw.githubusercontent.com/cage1016/ms-demo-authz/master/deployments/gateway.yaml
    ```
3. Set the `GATEWAY_HTTP_URL/GATEWAY_GRPC_URL` environment variable in your shell to the public IP/port of the Istio Ingress gateway.
    ```sh
    export INGRESS_HTTP_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
    export INGRESS_GRPC_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
    export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    export GATEWAY_HTTP_URL=$INGRESS_HOST:$INGRESS_HTTP_PORT
    export GATEWAY_GRPC_URL=$INGRESS_HOST:$INGRESS_GRPC_PORT
    echo $GATEWAY_HTTP_URL
    echo $GATEWAY_GRPC_URL
    ```
4. Access by command
    - add service Sum method
    ```sh
    curl -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    ```
    - tictac service tic method
    ```sh
    curl -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"
    ```
    - tictac service tac method
    ```sh
    curl "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"
    ```
    - authz service roles method
    ```sh
    curl "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    ```

## Scenario2-1: istio RequestAuhentication/AutthorizationPolicy by isto mixer adapter

1. Install istio RequestAuthentication and AuthorizationPolicy `authz-jwt` will check JWT token
    ```sh
    kubectl apply -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/authz-jwt.yaml
    ```
2. Access by Command
    - any request without valid JWT token is forbidden
    ```sh
    curl -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    RBAC: access denied

    curl -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"
    RBAC: access denied

    curl "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"
    RBAC: access denied

    curl "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    RBAC: access denied
    ```
3. Access those API with any valid JWT token
    - export valid token for `owner`, `editor`, & `viwer`
    ```sh
    # temporary token DO NOT use in production
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat owner.json | base64)
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat editor.json | base64)
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat viewer.json | base64)

    export OWNER_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzMzQsImlhdCI6MTYwNDMwMDMzNCwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJLYVpSSk9jNjhoQ2FsaE1NanI1aWVBIiwibmJmIjoxNjA0MzAwMzM0LCJyb2xlcyI6WyJvd25lciJdLCJzdWIiOiJvd25lckBleGFtcGxlLmNvbSIsInVzZXJJZCI6ImVCZW5mS3VDekFpQUNfYmZxRVR3WSJ9.u2XYn2h7R32peWmPEfRNRVpdmSDVYuHDCn-eJpqTICyLzSH0VuRSD3AIQmTuJkDXtG9HNrqoGLkvA9Ai5DmiCICkT_fXmTJw2IZY8TvHxcK8czoLO1qgVYWkVe37Mz0ReWSDryjRIRQXdeebxi0-ooZaKLAgXZERpzaRe81ZEKuU834DOVayOoZgB6jbB0prOWQd6PzBgVOgNcNTWOwMtaCnURYSmfFV-oIHsetvt4-C98KPyrA5GqLyBWi7IzDLDvLrEtVyobL7Z4qYAZHflAdzJLKe62KuGGT_fXGFET8X8U9IZCi1ceHjNljVJyLlb_rrBZIKL3yziCuLr_n-iw
    export EDITOR_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzNjMsImlhdCI6MTYwNDMwMDM2MywiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJPYlhJYnlMQkNnUTlvV1piRUdmSUJRIiwibmJmIjoxNjA0MzAwMzYzLCJyb2xlcyI6WyJlZGl0b3IiXSwic3ViIjoiZWRpdG9yQGV4YW1wbGUuY29tIiwidXNlcklkIjoiQmExVmRob0IzN2haTE9pVkJiX2hLIn0.pvT6Ytg0bKCj01NiNaNp7KK4NGf3jXkjYLmXs0cbbTA515cN3JZLwcIG_cHOon48rG-WkovAkwxKgtxgz2ZyY-GFt3pURA2wBWIFVfMv5gcyu0LWJP7LF7g9NWpdf_iqZPVYuyGW2bBknbXisfWof99_Ngu_QkM11wlCe2zA8ts3y4nYUn7Bz8FjIRVaa0X65HVT6FqBYuL-Pz4AoXwVtrX-L_qKnKVXV4-IXt1SbGmgVViRDHTaggp1HOmu1bL5cBII1h75fiaWXNU4nvtRMoBifmy9rUB5D8sW81VkJSrQfHhWkCEgHK-UP1-ZNA0Wqe8wKcGUOtKDJSm2UWr2vg
    export VIEWER_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzODYsImlhdCI6MTYwNDMwMDM4NiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJNaXh1cko1WEM2N0lheWJNenZmbFNBIiwibmJmIjoxNjA0MzAwMzg2LCJyb2xlcyI6WyJ2aWV3ZXIiXSwic3ViIjoidmlld2VyQGV4YW1wbGUuY29tIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.SWE3SFaf1BBjKCxUiS4cxAt9Q-AP5hP4U_HZvIT853KxGP5ii-8s2EqwEBcumdfsUxkMIr8pfv37VXGYCxFE40YtTsSfHj5qSQWmnCkXUwg3OrY_eSIUK9_M4JpD_2jsnvABEf6n23z-FdCdnPy5m89CeRco3zPpA9P5K21abgp0QdQ9zIfhE7ChlP4pB9_g1OWPmX6uW2slCLLBV414GJj_tAL3Mp2GqwYwEfkGbYyeQF2N_IiW6HnTptjlCcsH1rnsGCy6ei1aMdDdsce5U2Fg6RtXUvBLaeSeq_hVTg9h93RlZZMjB64wiK8bkYBNKKDsh1BZ025hPKX5ITuebw
    ```
    - Access those API with valid token
    > you could use $EDITOR_TOKEN or $VIEWER_TOKEN as you want
    ```sh
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    ```
4. Install custom mixer adpater
    > Please visit [authzopa (istio custom authorization mixer adapter) README](./authzopa/README.md) for more detail how to create custom istio mixier adpater
    ```sh
    kubectl apply -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/authzopa/deployments/all.yaml
    ```

    - _OPA data_

    ```json
    {
        "rolePermissions": {
            "editor": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/tictac/tic"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/tictac/tac"
                }
            ],
            "owner": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/add/sum"
                },
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/tictac/tic"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/tictac/tac"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/authz/roles"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/authz/roles/[a-zA-Z0-9_-~]{21}"
                }
            ],
            "viewer": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/add/sum"
                }
            ]
        },
        "userRoles": {
            "0tywi6h69mdPKNN_NkI7B": [
                "viewer"
            ],
            "Ba1VdhoB37hZLOiVBb_hK": [
                "editor"
            ],
            "eBenfKuCzAiAC_bfqETwY": [
                "owner"
            ]
        }
    }
    ```
    - As above OPA data describe. `owner` allow to access ALL APIs, `editor` ONLY allow to access `/api/v1/tictac/*` APIs and `viewer` ONLY allow to access `/api/v1/add/sum` API. Our custom istio authorization mixer adapter `authzopa` will verify REQUEST by calling `authz` `/pb.Authz/IsAuthorizedReq` GRPC
    - owner could access any APis
    ```sh
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    ```
    - editor ONLY allow to access `tictac` APIs
    ```sh
    curl -H "Authorization: Bearer $EDITOR_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    
    curl -H "Authorization: Bearer $EDITOR_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"

    curl -H "Authorization: Bearer $EDITOR_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"

    curl -H "Authorization: Bearer $EDITOR_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    ```
    - viewer ONLY allow to access `add` APis
    ```sh
    curl -H "Authorization: Bearer $VIEWER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    
    curl -H "Authorization: Bearer $VIEWER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny

    curl -H "Authorization: Bearer $VIEWER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny

    curl -H "Authorization: Bearer $VIEWER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    ```
    -  Clean up
    ```bash
    # delete custom mixer adapter
    kubectl delete -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/authzopa/deployments/all.yaml

    # delete istion RequestAuthentication/AuthorizationPolicy
    kubectl delete -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/authz-jwt.yaml
    ```

## Scenario2-2: istio RequestAuhentication/AutthorizationPolicy by isto external authorization

1. Install istio external authorization implementation
   > Please visit [extauthz README](./extauthz/README.md) for more detail how to create istio external authorization 
    ```sh
    kubectl apply -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/extauthz/deployments/all.yaml
    ```
2.  Declare Authorization by OPA 
    - _OPA data_

    ```json
    {
        "rolePermissions": {
            "editor": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/tictac/tic"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/tictac/tac"
                }
            ],
            "owner": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/add/sum"
                },
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/tictac/tic"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/tictac/tac"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/authz/roles"
                },
                {
                    "method": "GET",
                    "path": "/api/([^\\/]+)/authz/roles/[a-zA-Z0-9_-~]{21}"
                }
            ],
            "viewer": [
                {
                    "method": "POST",
                    "path": "/api/([^\\/]+)/add/sum"
                }
            ]
        },
        "userRoles": {
            "0tywi6h69mdPKNN_NkI7B": [
                "viewer"
            ],
            "Ba1VdhoB37hZLOiVBb_hK": [
                "editor"
            ],
            "eBenfKuCzAiAC_bfqETwY": [
                "owner"
            ]
        }
    }
    ```
    - As above OPA data describe. `owner` allow to access ALL APIs, `editor` ONLY allow to access `/api/v1/tictac/*` APIs and `viewer` ONLY allow to access `/api/v1/add/sum` API. Our custom istio authorization mixer adapter `authzopa` will verify REQUEST by calling `authz` `/pb.Authz/IsAuthorizedReq` GRPC
    - export valid token for `owner`, `editor`, & `viwer`
    ```sh
    # temporary token DO NOT use in production
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat owner.json | base64)
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat editor.json | base64)
    # python scripts/app.py keys/server/public/server_pubkey.pem keys/server/private/server_key.pem $(cat viewer.json | base64)

    export OWNER_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzMzQsImlhdCI6MTYwNDMwMDMzNCwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJLYVpSSk9jNjhoQ2FsaE1NanI1aWVBIiwibmJmIjoxNjA0MzAwMzM0LCJyb2xlcyI6WyJvd25lciJdLCJzdWIiOiJvd25lckBleGFtcGxlLmNvbSIsInVzZXJJZCI6ImVCZW5mS3VDekFpQUNfYmZxRVR3WSJ9.u2XYn2h7R32peWmPEfRNRVpdmSDVYuHDCn-eJpqTICyLzSH0VuRSD3AIQmTuJkDXtG9HNrqoGLkvA9Ai5DmiCICkT_fXmTJw2IZY8TvHxcK8czoLO1qgVYWkVe37Mz0ReWSDryjRIRQXdeebxi0-ooZaKLAgXZERpzaRe81ZEKuU834DOVayOoZgB6jbB0prOWQd6PzBgVOgNcNTWOwMtaCnURYSmfFV-oIHsetvt4-C98KPyrA5GqLyBWi7IzDLDvLrEtVyobL7Z4qYAZHflAdzJLKe62KuGGT_fXGFET8X8U9IZCi1ceHjNljVJyLlb_rrBZIKL3yziCuLr_n-iw
    export EDITOR_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzNjMsImlhdCI6MTYwNDMwMDM2MywiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJPYlhJYnlMQkNnUTlvV1piRUdmSUJRIiwibmJmIjoxNjA0MzAwMzYzLCJyb2xlcyI6WyJlZGl0b3IiXSwic3ViIjoiZWRpdG9yQGV4YW1wbGUuY29tIiwidXNlcklkIjoiQmExVmRob0IzN2haTE9pVkJiX2hLIn0.pvT6Ytg0bKCj01NiNaNp7KK4NGf3jXkjYLmXs0cbbTA515cN3JZLwcIG_cHOon48rG-WkovAkwxKgtxgz2ZyY-GFt3pURA2wBWIFVfMv5gcyu0LWJP7LF7g9NWpdf_iqZPVYuyGW2bBknbXisfWof99_Ngu_QkM11wlCe2zA8ts3y4nYUn7Bz8FjIRVaa0X65HVT6FqBYuL-Pz4AoXwVtrX-L_qKnKVXV4-IXt1SbGmgVViRDHTaggp1HOmu1bL5cBII1h75fiaWXNU4nvtRMoBifmy9rUB5D8sW81VkJSrQfHhWkCEgHK-UP1-ZNA0Wqe8wKcGUOtKDJSm2UWr2vg
    export VIEWER_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5MDQzMDAzODYsImlhdCI6MTYwNDMwMDM4NiwiaXNzIjoidGVzdGluZ0BzZWN1cmUuaXN0aW8uaW8iLCJqdGkiOiJNaXh1cko1WEM2N0lheWJNenZmbFNBIiwibmJmIjoxNjA0MzAwMzg2LCJyb2xlcyI6WyJ2aWV3ZXIiXSwic3ViIjoidmlld2VyQGV4YW1wbGUuY29tIiwidXNlcklkIjoiMHR5d2k2aDY5bWRQS05OX05rSTdCIn0.SWE3SFaf1BBjKCxUiS4cxAt9Q-AP5hP4U_HZvIT853KxGP5ii-8s2EqwEBcumdfsUxkMIr8pfv37VXGYCxFE40YtTsSfHj5qSQWmnCkXUwg3OrY_eSIUK9_M4JpD_2jsnvABEf6n23z-FdCdnPy5m89CeRco3zPpA9P5K21abgp0QdQ9zIfhE7ChlP4pB9_g1OWPmX6uW2slCLLBV414GJj_tAL3Mp2GqwYwEfkGbYyeQF2N_IiW6HnTptjlCcsH1rnsGCy6ei1aMdDdsce5U2Fg6RtXUvBLaeSeq_hVTg9h93RlZZMjB64wiK8bkYBNKKDsh1BZ025hPKX5ITuebw
    ```
    - owner could access any APis
    ```sh
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    
    curl -H "Authorization: Bearer $OWNER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"

    curl -H "Authorization: Bearer $OWNER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    ```
    - editor ONLY allow to access `tictac` APIs
    ```sh
    curl -H "Authorization: Bearer $EDITOR_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    
    curl -H "Authorization: Bearer $EDITOR_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"

    curl -H "Authorization: Bearer $EDITOR_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"

    curl -H "Authorization: Bearer $EDITOR_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    ```
    - viewer ONLY allow to access `add` APis
    ```sh
    curl -H "Authorization: Bearer $VIEWER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/add/sum" -H 'Content-Type: application/json' -d '{ "a": 3, "b": 34}'
    
    curl -H "Authorization: Bearer $VIEWER_TOKEN" -X "POST" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tic"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny

    curl -H "Authorization: Bearer $VIEWER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/tictac/tac"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny

    curl -H "Authorization: Bearer $VIEWER_TOKEN" "http://$GATEWAY_HTTP_URL/api/v1/authz/roles"
    PERMISSION_DENIED:h1.handler.istio-system:Unauthorized: permission deny
    ```
    -  Clean up
    ```bash
    # delete custom mixer adapter
    kubectl delete -f https://raw.githubusercontent.com/cage1016/gokit-istio-security/master/extauthz/deployments/all.yaml
    ```

## CleanUP
- ms-demo-authz
```sh
# install ms-demo-authz
kubectl delete -f https://raw.githubusercontent.com/cage1016/ms-demo-authz/master/deployments/all.yaml

# install ms-demo-authz istio
kubectl delete -f https://raw.githubusercontent.com/cage1016/ms-demo-authz/master/deployments/gateway.yaml
```
- ms-demo-authz
```sh
# install ms-demo
kubectl delete -f https://raw.githubusercontent.com/cage1016/ms-demo/master/deployments/all.yaml

# install ms-demo istio
kubectl delete -f https://raw.githubusercontent.com/cage1016/ms-demo/master/deployments/gateway.yaml
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