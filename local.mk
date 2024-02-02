.PHONY: deploy
deploy: create_services
	@if test -n "$(shell docker compose ls -q)"; then      \
		docker compose -f ${RESOURCE_DIR}/compose.yaml stop; \
	fi
	@docker compose -f ${RESOURCE_DIR}/compose.yaml up -d

export DCKR_ENV_FLAGS :=--net=host --memory=8g
