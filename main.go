/*
Links:
	* https://tip.golang.org/blog/wasi
	* https://extism.org/docs/write-a-plugin/go-pdk
	* https://github.com/golang/go/issues/42372
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/extism/extism"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	// set some input data to provide to the plugin module
	var data []byte
	if len(os.Args) > 1 {
		data = []byte(strings.Join(os.Args[1:], " "))
	} else {
		data = []byte("testing from go -> wasm shared memory...")
	}

	runner1, err := NewExtismPlugin("./plugin/extism/plugin_extism.wasm")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer runner1.Close()

	fmt.Println("* Extism *")
	for i := 0; i < 3; i++ {
		doTest(runner1.Run, data)
	}

	runner2, err := NewWazeroPlugin("./plugin/plugin.wasm")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer runner2.Close()

	fmt.Println("\n* Wazero *")
	for i := 0; i < 3; i++ {
		doTest(runner2.Run, data)
	}
}

func doTest(fn func(data []byte) ([]byte, error), data []byte) {
	start := time.Now()
	out, err := fn(data)
	end := time.Since(start)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var dest map[string]int
	err = json.Unmarshal(out, &dest)
	if err != nil {
		fmt.Printf("error reading response: %v", err)
		os.Exit(1)
	}

	fmt.Printf("  Elapsed: %s, [Result] Count: %v\n", end, dest["count"])
}

type WazeroPlugin struct {
	rt     wazero.Runtime
	guest  wazero.CompiledModule
	cache  wazero.CompilationCache
	config wazero.ModuleConfig
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func NewWazeroPlugin(plugin string) (*WazeroPlugin, error) {
	p := &WazeroPlugin{}
	if err := p.Init(plugin); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *WazeroPlugin) Init(plugin string) error {
	p.Close()

	ctx := context.Background()

	p.cache = wazero.NewCompilationCache()
	rtc := wazero.NewRuntimeConfig().WithCompilationCache(p.cache)

	p.rt = wazero.NewRuntimeWithConfig(ctx, rtc)

	wasmData, err := os.ReadFile(plugin)
	if err != nil {
		return err
	}

	p.guest, err = p.rt.CompileModule(ctx, wasmData)
	if err != nil {
		return fmt.Errorf("error compiling wasm binary: %v", err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, p.rt)

	p.config = wazero.NewModuleConfig().
		WithStdout(&p.stdout).
		WithStderr(&p.stderr)

	return nil
}

func (p *WazeroPlugin) Run(data []byte) ([]byte, error) {
	if p.rt == nil {
		return nil, fmt.Errorf("Init() has not been called")
	}

	ctx := context.Background()
	conf := p.config.WithArgs("plugin.wasm", string(data))

	p.stdout.Reset()
	p.stderr.Reset()

	_, err := p.rt.InstantiateModule(ctx, p.guest, conf)
	if err != nil {
		return nil, err
	}

	if p.stdout.Len() == 0 && p.stderr.Len() > 0 {
		return nil, fmt.Errorf(string(p.stderr.Bytes()))
	}

	return p.stdout.Bytes(), nil
}

func (p *WazeroPlugin) Close() {
	ctx := context.Background()
	if p.rt != nil {
		p.rt.Close(ctx)
		p.rt = nil
	}
	if p.cache != nil {
		p.cache.Close(ctx)
		p.cache = nil
	}
}

type ExtismPlugin struct {
	manifest *extism.Manifest
	guest    *extism.Plugin
}

func NewExtismPlugin(plugin string) (*ExtismPlugin, error) {
	p := &ExtismPlugin{}
	if err := p.Init(plugin); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *ExtismPlugin) Init(plugin string) error {
	p.Close()

	manifest := extism.Manifest{Wasm: []extism.Wasm{extism.WasmFile{Path: plugin}}}
	p.manifest = &manifest

	// NOTE: if you encounter an error such as:
	// "Unable to load plugin: unknown import: wasi_snapshot_preview1::fd_write has not been defined"
	// change `false` to `true` in the following function to provide WASI imports to your plugin.
	guest, err := extism.NewPluginFromManifest(manifest, []extism.Function{}, true)
	if err != nil {
		return err
	}
	p.guest = &guest
	return nil
}

func (p *ExtismPlugin) Run(data []byte) ([]byte, error) {
	if p.guest == nil {
		return nil, fmt.Errorf("Init() has not been called")
	}
	// use the extism Go library to provide the input data to the plugin, execute it, and then
	// collect the plugin state and error if present
	out, err := p.guest.Call("count_vowels", data)
	if err != nil {
		return nil, err
	}
	// out is a zero-copy reference that will be invalid after we return
	return bytes.Clone(out), nil
}

func (p *ExtismPlugin) Close() {
	if p.guest != nil {
		p.guest.Free()
		p.guest = nil
	}
}
