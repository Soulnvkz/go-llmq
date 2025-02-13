package llama

/*
#cgo CFLAGS: -I${SRCDIR}/../../deps/llama.cpp/ggml/include -I${SRCDIR}/../../deps/llama.cpp/include
#cgo LDFLAGS: -L${SRCDIR}/../../deps/llama.cpp/build/bin -lllama -lggml -lstdc++ -ldl -lm
#include "llama.h"
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"unsafe"
)

type LLM struct {
	model     *C.struct_llama_model
	vocab     *C.struct_llama_vocab
	ctx       *C.struct_llama_context
	smpl      *C.struct_llama_sampler
	n_predict int
	n_ctx     int
	n_batch   int

	mu         sync.Mutex
	cancelList map[string]struct {
		cancel *context.CancelFunc
		ctx    *context.Context
	}
}

func NewLLM() *LLM {
	return &LLM{
		n_predict: 1024,
		n_ctx:     4096,
		n_batch:   512,

		mu: sync.Mutex{},
		cancelList: map[string]struct {
			cancel *context.CancelFunc
			ctx    *context.Context
		}{},
	}
}

func (llm *LLM) load_model(model_path string) error {
	if llm.model != nil {
		return fmt.Errorf("model has initilized already")
	}
	if llm.vocab != nil {
		return fmt.Errorf("model has initilized already")
	}

	model_params := C.llama_model_default_params()
	model_params.n_gpu_layers = 0

	model := C.llama_model_load_from_file(C.CString(model_path), model_params)
	if model == nil {
		return fmt.Errorf("can't initilize the model")
	}

	vocab := C.llama_model_get_vocab(model)
	if vocab == nil {
		return fmt.Errorf("can't initiliize vocab")
	}
	llm.model = model
	llm.vocab = vocab

	return nil
}

func (llm *LLM) initilize_context() error {
	// initialize the context

	ctx_params := C.llama_context_default_params()
	// n_ctx is the context size
	ctx_params.n_ctx = C.uint32_t(llm.n_ctx)
	// n_batch is the maximum number of tokens that can be processed in a single call to llama_decode
	ctx_params.n_batch = C.uint32_t(llm.n_batch)
	// enable performance counters
	ctx_params.no_perf = false

	ctx := C.llama_init_from_model(llm.model, ctx_params)
	if ctx == nil {
		return fmt.Errorf("can't initiliize context")
	}

	llm.ctx = ctx
	return nil
}

func (llm *LLM) initilize_sampler() error {
	sparams := C.llama_sampler_chain_default_params()
	sparams.no_perf = false
	smpl := C.llama_sampler_chain_init(sparams)
	if smpl == nil {
		return fmt.Errorf("can't initiliize sampler")
	}
	C.llama_sampler_chain_add(smpl, C.llama_sampler_init_greedy())

	llm.smpl = smpl
	return nil
}

func (llm *LLM) tokenize_promtp(prompt string) (int, []C.llama_token, error) {
	// find the number of tokens in the prompt
	n_prompt := int(-C.llama_tokenize(llm.vocab, C.CString(prompt), C.int(len(prompt)), nil, 0, true, true))
	prompt_tokens := make([]C.llama_token, n_prompt)
	if C.llama_tokenize(llm.vocab, C.CString(prompt), C.int(len(prompt)), &prompt_tokens[0], C.int(len(prompt)), true, true) < 0 {
		return 0, nil, fmt.Errorf("prompt tokenize failed")
	}

	return n_prompt, prompt_tokens, nil
}

func (llm *LLM) Initilize(model string) error {
	// init backend
	C.ggml_backend_load_all()
	err := llm.load_model(model)
	if err != nil {
		return err
	}
	err = llm.initilize_context()
	if err != nil {
		return err
	}
	err = llm.initilize_sampler()
	if err != nil {
		return err
	}

	return nil
}

func (llm *LLM) Clean() error {

	C.llama_sampler_free(llm.smpl)
	C.llama_free(llm.ctx)
	C.llama_model_free(llm.model)

	llm.smpl = nil
	llm.ctx = nil
	llm.model = nil

	return nil
}

func (llm *LLM) Cancel(req_id string) {
	c, ok := llm.cancelList[req_id]
	if !ok {
		log.Printf("Cancel: adding ctx for %s", req_id)
		ctx, cancel := context.WithCancel(context.Background())
		llm.cancelList[req_id] = struct {
			cancel *context.CancelFunc
			ctx    *context.Context
		}{
			ctx:    &ctx,
			cancel: &cancel,
		}
		return
	}

	log.Printf("Cancel: ctx %s already exists, call cancel", req_id)
	(*c.cancel)()

	// TODO: clean map
}

func (llm *LLM) ProccessNext(prompt string, req_id string, on_next func([]byte, error) error) (chan bool, error) {
	n_prompt, prompt_tokens, err := llm.tokenize_promtp(prompt)
	if err != nil {
		return nil, err
	}

	_, ok := llm.cancelList[req_id]
	var ctx context.Context
	var cancel context.CancelFunc
	if !ok {
		log.Printf("ProccessNext: adding ctx for %s", req_id)
		ctx, cancel = context.WithCancel(context.Background())
		llm.cancelList[req_id] = struct {
			cancel *context.CancelFunc
			ctx    *context.Context
		}{
			ctx:    &ctx,
			cancel: &cancel,
		}
	} else {
		log.Printf("ProccessNext: ctx %s already exists", req_id)
		return nil, errors.New("request has canceled already")
	}

	batch := C.llama_batch_get_one(&prompt_tokens[0], C.int(len(prompt_tokens)))
	n_decode := 0
	new_token_id := C.llama_token(0)

	n_pos := 0
	stop := make(chan bool)
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				stop <- true
				break loop
			default:
				if n_pos+int(batch.n_tokens) >= int(n_prompt)+llm.n_predict {
					stop <- true
					break loop
				}
				// evaluate the current batch with the transformer model
				if C.llama_decode(llm.ctx, batch) > 0 {
					on_next(nil, fmt.Errorf("failed to eval current batch"))
					stop <- true
					break loop
				}

				n_pos += int(batch.n_tokens)

				// sample the next token
				new_token_id = C.llama_sampler_sample(llm.smpl, llm.ctx, -1)

				// is it an end of generation?
				if C.llama_vocab_is_eog(llm.vocab, new_token_id) {
					stop <- true
					break loop
				}

				buf := make([]C.char, 128)

				n := C.llama_token_to_piece(llm.vocab, new_token_id, &buf[0], C.int(len(buf)), 0, true)
				if n < 0 {
					on_next(nil, fmt.Errorf("failed to convert token to piece"))
					stop <- true
					break loop
				}
				cstr := (*C.char)(unsafe.Pointer(&buf[0])) // Get pointer to the first element
				err = on_next([]byte(C.GoString(cstr)), nil)
				if err != nil {
					stop <- true
					break loop
				}
				// prepare the next batch with the sampled token
				batch = C.llama_batch_get_one(&new_token_id, 1)

				n_decode += 1
			}
		}
	}()

	return stop, nil
}
