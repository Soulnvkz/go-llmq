clean:
	rm -rf deps/llama.cpp/build

prepare-rocm:
	HIPCXX="$(hipconfig -l)/clang" HIP_PATH="$(hipconfig -R)" \                 
		cmake -S ./deps/llama.cpp -B ./deps/llama.cpp/build -DGGML_HIP=ON -DAMDGPU_TARGETS=gfx1030 -DCMAKE_BUILD_TYPE=Release -DLLAMA_CURL=ON 

build-rocm:
	HIPCXX="$(hipconfig -l)/clang" HIP_PATH="$(hipconfig -R)" \
		cmake --build ./deps/llama.cpp/build --config Release -j$(nproc)

prepare-cpu:
	cmake -S ./deps/llama.cpp -B ./deps/llama.cpp/build -DCMAKE_BUILD_TYPE=Release;

build-cpu:
	cmake --build ./deps/llama.cpp/build --config Release -j $(nproc)

run:
	LD_LIBRARY_PATH=deps/llama.cpp/build/bin \
	MODEL_PATH=/home/sol/programming/ai/models/SAINEMO-reMIX.i1-Q6_K.gguf \
	MQ_USER=admin \
	MQ_PASSWORD=admin \
	MQ_HOST=localhost \
	MQ_PORT=5672 \
	MQ_LLM_Q=llm_q \
	MQ_CANCEL_EX=llm_cancel_ex \
	./cmd/cmd