all: help

owner_payload=$(shell cat owner.json | base64)
viewer_payload=$(shell cat viewer.json | base64)

## owner_token: owner token
owner_token:
	$(call fnToken,$(owner_payload))

## viewer_token: viewer token
viewer_token:
	$(call fnToken,$(viewer_payload))

## scenario1_sum: scenario 1 add:sum API
scenario1_add_sum:
	curl -X "POST" "http://localhost:80/api/v1/add/sum" \
		-H 'Content-Type: application/json' \
		-d '{ "a": 3, "b": 34}'

## scenario1_add_concat: scenario 1 add:concat API
scenario1_add_concat:
	curl -X "POST" "http://localhost:80/api/v1/add/concat" \
		-H 'Content-Type: application/json' \
		-d '{ "a": "3", "b": "34"}'

## scenario1_foo_foo: scenario 1 foo:foo API
scenario1_foo_foo:
	curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
		-H 'Content-Type: application/json' \
		-d '{ "s": "foo"}'

## scenario1_authz_roles: scenario 1 authz:roles API
scenario1_authz_roles:
	curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
		-H 'Content-Type: application/json'

## scenario2_add_sum_owner_token: scenario 2 add:sum API with owner token
scenario2_add_sum_owner_token:
	curl -X "POST" "http://localhost:80/api/v1/add/sum" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(owner_payload)))" \
		-H 'Content-Type: application/json' \
		-d '{ "a": 3, "b": 34}'

## scenario2_add_sum_viewer_token: scenario 2 add:sum API with viewer token
scenario2_add_sum_viewer_token:
	curl -X "POST" "http://localhost:80/api/v1/add/sum" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
		-H 'Content-Type: application/json' \
		-d '{ "a": 3, "b": 34}'		

## scenario2_add_concat_owner_token: scenario 2 add:concat API with owner token
scenario2_add_concat_owner_token:
	curl -X "POST" "http://localhost:80/api/v1/add/concat" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(owner_payload)))" \
		-H 'Content-Type: application/json' \
		-d  '{ "a": "3", "b": "34"}'

## scenario2_add_concat_viewer_token: scenario 2 add:concat API with viewer token
scenario2_add_concat_viewer_token:
	curl -X "POST" "http://localhost:80/api/v1/add/concat" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
		-H 'Content-Type: application/json' \
		-d  '{ "a": "3", "b": "34"}'

## scenario2_foo_foo_owner_token: scenario 2 foo:foo API with owner token
scenario2_foo_foo_owner_token:
	curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(owner_payload)))" \
		-H 'Content-Type: application/json' \
		-d  '{ "s": "foo"}'

## scenario2_foo_foo_viewer_token: scenario 2 foo:foo API with viewer token
scenario2_foo_foo_viewer_token:
	curl -X "POST" "http://localhost:80/api/v1/foo/foo" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
		-H 'Content-Type: application/json' \
		-d  '{ "s": "foo"}'	

## scenario2_authz_roles_owner_token: scenario 2 authz:roles API with owner token
scenario2_authz_roles_owner_token:
	curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(owner_payload)))" \
		-H 'Content-Type: application/json'

## scenario2_authz_roles_viewer_token: scenario 2 authz:roles API with viewer token
scenario2_authz_roles_viewer_token:
	curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
		-H 'Content-Type: application/json'

## scenario3_add_sum_owner_token: scenario 3 add:sum API with owner token
scenario3_add_sum_owner_token:
	curl -X "POST" "http://localhost:80/api/v1/add/sum" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(owner_payload)))" \
		-H 'Content-Type: application/json' \
		-d '{ "a": 3, "b": 34}'

## scenario3_add_sum_viewer_token_permission_deny: scenario 3 add:sum API with viewer token permission deny
scenario3_add_sum_viewer_token_permission_deny:
	curl -v "POST" "http://localhost:80/api/v1/add/sum" \
		-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
		-H 'Content-Type: application/json' \
		-d '{ "a": 3, "b": 34}'	

## scenario3_authz_roles_viewer_token: scenario 3 authz:roles API with viewer token
scenario3_authz_roles_viewer_token:
	curl -X "GET" "http://localhost:80/api/v1/authz/roles" \
			-H "Authorization: Bearer $(shell $(call fnToken,$(viewer_payload)))" \
			-H 'Content-Type: application/json'


define fnToken
	docker run --rm -it -v $(PWD)/keys:/keys \
			-e PRIVATE_KEY_FILE=/keys/server/private/server_key.pem \
			-e PUBLIC_KEY_FILE=/keys/server/public/server_pubkey.pem \
			-e PAYLOAD=$(1) \
			cage1016/gokit-istio-security-generate-token:latest	
endef

help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
	@echo ""
	@for var in $(helps); do \
		echo $$var; \
	done | column -t -s ':' |  sed -e 's/^/  /'		