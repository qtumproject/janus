.PHONY: build build-openzeppelin openzeppelin
build: build-openzeppelin

# Build openzeppelin docker files with docker-compose
openzeppelin-docker-compose:
	cd openzeppelin && ./run-openzeppelin-tests.sh

openzeppelin-ganache-docker-compose:
	cd openzeppelin && DOCKER_COMPOSE=docker-compose-openzeppelin-ganache.yml ./run-openzeppelin-tests.sh

build-openzeppelin: build-truffle-parser-docker build-openzeppelin-docker

build-truffle-parser-docker:
	docker build -t qtum/janus-truffle-parser -f ./truffle-parser.Dockerfile .

build-openzeppelin-docker:
	cd openzeppelin && docker build -t qtum/janus-openzeppelin -f ./openzeppelin.Dockerfile .

init:
	git submodule init
	git submodule update
	cd openzeppelin/openzeppelin-contracts && yarn install
	cd openzeppelin/openzeppelin-contracts && yarn add mocha-spec-json-output-reporter

openzeppelin:
	cd openzeppelin/openzeppelin-contracts && truffle test --network testing

openzeppelin-development:
	cd openzeppelin/openzeppelin-contracts && truffle test --network development

openzeppelin-docker: build
	docker run --rm qtum/janus-openzeppelin

github-action-openzeppelin:
	docker logs ci_openzeppelin_1 | sed -n 's/.* \([0-9]\{1,\}\) passing.*/::set-output name=PASSING=::\1/p'
	docker logs ci_openzeppelin_1 | sed -n 's/.* \([0-9]\{1,\}\) pending.*/::set-output name=PENDING=::\1/p'
	docker logs ci_openzeppelin_1 | sed -n 's/.* \([0-9]\{1,\}\) failing.*/::set-output name=FAILING=::\1/p'

openzeppelin-truffle-parser-docker:
	cd openzeppelin && make -f ../Makefile truffle-parser-docker

truffle-parser-extract-result-docker:
	docker cp ${CONTAINER}:${INPUT} ./truffle-result-output.json

truffle-parser-docker:
	docker run --rm -v `pwd`:/output qtum/janus-truffle-parser \
        --expected /output/truffle-expected-output.json \
        --input /output/truffle-result-output.json \
        --output /output/truffle-pruned-output.json