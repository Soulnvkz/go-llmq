package main

/*
#cgo CFLAGS: -I/home/sol/programming/ai/llama.cpp/ggml/include -I/home/sol/programming/ai/llama.cpp/src -I/home/sol/programming/ai/llama.cpp/include -I/home/sol/programming/ai/llama.cpp/src -I/home/sol/programming/ai/llama.cpp/common
#cgo LDFLAGS: -L/home/sol/programming/ai/llama.cpp/build/bin -Wl,-rpath=/home/sol/programming/ai/llama.cpp/build/bin -lllama -lggml -lm -lstdc++
#include "llama.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	C.ggml_backend_load_all()
	model_params := C.llama_model_default_params()
	model_params.n_gpu_layers = 99 // default max value
	fmt.Println(model_params)

	model_path := "/home/sol/programming/ai/models/SAINEMO-reMIX.i1-Q6_K.gguf"
	model := C.llama_model_load_from_file(C.CString(model_path), model_params)
	vocab := C.llama_model_get_vocab(model)

	if vocab == nil {
		fmt.Println("error: can't get vocab")
		panic(1)
	}

	if model == nil {
		fmt.Println("error: can't load the model")
		panic(1)
	}

	fmt.Println("model load sucessfully")

	prompt := "Привет. Как тебя зовут? Расскажи историю про грибы в лесу"

	// tokenize the prompt
	// find the number of tokens in the prompt
	n_prompt := -C.llama_tokenize(vocab, C.CString(prompt), C.int(len(prompt)), nil, 0, true, true)
	fmt.Println("n_prompt")

	prompt_tokens := make([]C.llama_token, n_prompt)
	if C.llama_tokenize(vocab, C.CString(prompt), C.int(len(prompt)), &prompt_tokens[0], C.int(len(prompt)), true, true) < 0 {
		fmt.Println("error: failed to llama_tokenize")
		panic(1)
	}

	fmt.Println(prompt_tokens)

	// initialize the context
	n_predict := 1024
	ctx_params := C.llama_context_default_params()
	// n_ctx is the context size
	ctx_params.n_ctx = C.uint32_t(n_prompt-1) + C.uint32_t(n_predict)
	// n_batch is the maximum number of tokens that can be processed in a single call to llama_decode
	ctx_params.n_batch = C.uint32_t(n_prompt)
	// enable performance counters
	ctx_params.no_perf = false

	ctx := C.llama_init_from_model(model, ctx_params)
	if ctx == nil {
		fmt.Println("%s: error: failed to create the llama_context\n")
	}

	// initialize the sampler

	sparams := C.llama_sampler_chain_default_params()
	sparams.no_perf = false
	smpl := C.llama_sampler_chain_init(sparams)
	C.llama_sampler_chain_add(smpl, C.llama_sampler_init_greedy())

	//    // print the prompt token-by-token

	//    for (auto id : prompt_tokens) {
	//     char buf[128];
	//     int n = llama_token_to_piece(vocab, id, buf, sizeof(buf), 0, true);
	//     if (n < 0) {
	//         fprintf(stderr, "%s: error: failed to convert token to piece\n", __func__);
	//         return 1;
	//     }
	//     std::string s(buf, n);
	//     printf("%s", s.c_str());
	// }

	batch := C.llama_batch_get_one(&prompt_tokens[0], C.int(len(prompt_tokens)))

	// main loop

	// const auto t_main_start = ggml_time_us();
	n_decode := 0
	new_token_id := C.llama_token(0)

	for n_pos := 0; n_pos+int(batch.n_tokens) < int(n_prompt)+n_predict; {
		// evaluate the current batch with the transformer model
		if C.llama_decode(ctx, batch) > 0 {
			fmt.Printf("failed to eval")
			break
		}

		n_pos += int(batch.n_tokens)

		// sample the next token
		{
			new_token_id = C.llama_sampler_sample(smpl, ctx, -1)

			// is it an end of generation?
			if C.llama_vocab_is_eog(vocab, new_token_id) {
				break
			}

			buf := make([]C.char, 128)

			n := C.llama_token_to_piece(vocab, new_token_id, &buf[0], C.int(len(buf)), 0, true)
			if n < 0 {
				fmt.Printf("error: failed to convert token to piece")
				break
			}
			cstr := (*C.char)(unsafe.Pointer(&buf[0])) // Get pointer to the first element
			goStr := C.GoString(cstr)
			fmt.Print(goStr)

			// prepare the next batch with the sampled token
			batch = C.llama_batch_get_one(&new_token_id, 1)

			n_decode += 1
		}
	}
}
