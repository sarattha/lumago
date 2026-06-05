VULKAN_ENV := GODEBUG=cgocheck=0 DYLD_LIBRARY_PATH=/opt/homebrew/lib DYLD_FALLBACK_LIBRARY_PATH=/opt/homebrew/lib VK_ICD_FILENAMES=/opt/homebrew/Cellar/molten-vk/1.4.1/etc/vulkan/icd.d/MoltenVK_icd.json

.PHONY: fmt test run run-nop vet tidy shaders

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

shaders:
	mkdir -p shaders/bin
	glslc shaders/quad.vert -o shaders/bin/quad.vert.spv
	glslc shaders/quad.frag -o shaders/bin/quad.frag.spv

run: shaders
	$(VULKAN_ENV) go run ./cmd/sandbox

run-nop:
	LUMAGO_RENDERER=nop go run ./cmd/sandbox
