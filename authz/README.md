# gokit microservice demo - authz
> gokit authz extend for [cage1016/ms-demo: gokit microservice demo](https://github.com/cage1016/ms-demo)

| Service | Description        |
| ------- | ------------------ |
| authz   | authorization RBAC |

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


## Install

1. Run `skaffold run` (first time will be slow)
2. Set the `AUTHZ_HTTP_LB_URL/AUTHZ_GRPC_LB_URL` environment variable in your shell to the public IP/port of the Kubernetes loadBalancer
    ```sh
    export AUTHZ_HTTP_LB_PORT=$(kubectl get service authz-external -o jsonpath='{.spec.ports[?(@.name=="http")].port}')
    export AUTHZ_GRPC_LB_PORT=$(kubectl get service authz-external -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
    export AUTHZ_LB_HOST=$(kubectl get service authz-external -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    export AUTHZ_HTTP_LB_URL=$AUTHZ_LB_HOST:$AUTHZ_HTTP_LB_PORT
    export AUTHZ_GRPC_LB_URL=$AUTHZ_LB_HOST:$AUTHZ_GRPC_LB_PORT
    echo $AUTHZ_HTTP_LB_URL
    echo $AUTHZ_GRPC_LB_URL
    ```
3. Access by command
    - authz roles method
    ```sh
    curl $AUTHZ_HTTP_LB_URL/roles
    
    or
    
    grpcurl -plaintext -proto ./pb/authz/authz.proto $AUTHZ_GRPC_LB_URL pb.Authz.ListRoles
    ```
    - get authz role
    ```sh
    curl $AUTHZ_HTTP_LB_URL/roles/G0znZWT5ajITIT97v6WXi
    
    or
    
    grpcurl -d '{"role_id": "G0znZWT5ajITIT97v6WXi"}' -plaintext -proto ./pb/authz/authz.proto $AUTHZ_GRPC_LB_URL pb.Authz.GetRole
    ```
4. Apply istio manifests `kubectl apply -f deployments/istio-manifests`
5. Set the `GATEWAY_HTTP_URL/GATEWAY_GRPC_URL` environment variable in your shell to the public IP/port of the Istio Ingress gateway.
    ```sh
    export INGRESS_HTTP_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
    export INGRESS_GRPC_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
    export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    export GATEWAY_HTTP_URL=$INGRESS_HOST:$INGRESS_HTTP_PORT
    export GATEWAY_GRPC_URL=$INGRESS_HOST:$INGRESS_GRPC_PORT
    echo $GATEWAY_HTTP_URL
    echo $GATEWAY_GRPC_URL
    ```
6. Access by command
    - authz roles method
    ```sh
    curl $GATEWAY_HTTP_URL/api/v1/authz/roles
    
    or
    
    grpcurl -plaintext -proto ./pb/authz/authz.proto $GATEWAY_GRPC_URL pb.Authz.ListRoles
    ```
    - get authz role
    ```sh
    curl $GATEWAY_HTTP_URL/api/v1/authz/roles/G0znZWT5ajITIT97v6WXi
    
    or
    
    grpcurl -d '{"role_id": "G0znZWT5ajITIT97v6WXi"}' -plaintext -proto ./pb/authz/authz.proto $GATEWAY_GRPC_URL pb.Authz.GetRole
    ```

## CleanUP

`skaffold delete`

or 

`kubectl delete -f deployments/istio-manifests`