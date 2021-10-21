img_tag = "hayeah/qtumportal:testjanus"
container_name = "qtumportal-testjanus"

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o ./docker/janus-linux-amd64 github.com/qtumproject/janus/cli/janus
	docker build ./docker -t $(img_tag)

.PHONY: run
run:
	cd .. && make run-janus-https

.PHONY: stop
stop:
	docker stop $(container_name)

.PHONY: exec
exec:
	docker exec -it $(container_name) sh

.PHONY: init-container
init-container:
	docker exec -it qtumportal-testjanus sh init-container.sh

.PHONY: logs
logs:
	docker logs --tail 1000  -f $(container_name)

.PHONY: clean-qtum-data
clean-qtum-data:
	rm -rf ./docker/.qtum

.PHONY: deploy-contracts
deploy-contracts:
	cd pet-shop-tutorial && truffle migrate --reset
	cp pet-shop-tutorial/build/contracts/*.json pet-shop-tutorial/js

.PHONY: serve-pet-shop
serve-pet-shop:
	cd pet-shop-tutorial; yarn
	echo "Two prefund accounts you can login with: \n 0x7926223070547d2d15b2ef5e7383e541c338ffe9 \n 0x2352be3db3177f0a07efbe6da5857615b8c9901d"
	cd pet-shop-tutorial; yarn dev &> /dev/null
