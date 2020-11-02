all: help

owner_payload=$(shell cat owner.json | base64)
viewer_payload=$(shell cat viewer.json | base64)
editor_payload=$(shell cat editor.json | base64)

## owner_token: create owner token
owner_token:
	$(call fnToken,$(owner_payload))

## viewer_token: create viewer token
viewer_token:
	$(call fnToken,$(viewer_payload))

## editor_token: create editor token
editor_token:
	$(call fnToken,$(editor_payload))

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