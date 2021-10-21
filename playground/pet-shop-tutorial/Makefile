
.PHONY: deploy-contracts
deploy-contracts:
	truffle migrate --reset
	cp ./build/contracts/*.json ./js

.PHONY: serve-pet-shop
serve-pet-shop:
	npm run dev
