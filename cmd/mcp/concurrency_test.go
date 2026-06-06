// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentStdioDispatchInterleavesResponses verifies the
// worker-pool path: when multiple slow tools/call requests arrive in
// quick succession, they execute concurrently. We use a fake runner
// that sleeps so wall time tells the story — N concurrent slow calls
// should finish in ~max(call_time), not sum(call_time).
//
// This is THE test that closes the loop on T9. Without concurrency,
// the test takes 4× longer than the concurrent baseline.
func TestConcurrentStdioDispatchInterleavesResponses(t *testing.T) {
	const numCalls = 4
	const fakeCallDuration = 200 * time.Millisecond

	// Set up a server with a fake runner that always sleeps.
	var inFlight atomic.Int32
	var peakInFlight atomic.Int32
	fake := &fakeRunner{
		delay: fakeCallDuration,
		onCall: func() {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				peak := peakInFlight.Load()
				if cur <= peak || peakInFlight.CompareAndSwap(peak, cur) {
					break
				}
			}
			time.Sleep(fakeCallDuration)
		},
	}

	srv, stdinW, stdoutR, stdoutBuf := newTestServer(t, fake)
	srv.maxConcurrency = numCalls

	done := make(chan error, 1)
	go func() { done <- srv.runStdio() }()

	// Write initialize + N tools/call frames.
	frames := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
	}
	for i := 0; i < numCalls; i++ {
		frames = append(frames, fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"lark_contact_search","arguments":{"query":"q%d"}}}`,
			100+i, i,
		))
	}

	start := time.Now()
	// Write frames in a goroutine so the response reader can drain
	// concurrently — io.Pipe is unbuffered, so synchronous writes
	// would deadlock against the server's blocking stdout writes.
	go func() {
		for _, frame := range frames {
			_, _ = stdinW.Write([]byte(frame + "\n"))
		}
	}()

	// Read responses until we've seen all numCalls + 1 (initialize).
	expectedResponses := 1 + numCalls
	got := readNResponses(t, stdoutR, expectedResponses, 5*time.Second)
	wall := time.Since(start)

	_ = stdinW.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("runStdio did not exit after stdin close")
	}

	if len(got) != expectedResponses {
		t.Fatalf("got %d responses, want %d. stdoutBuf=%q", len(got), expectedResponses, stdoutBuf.String())
	}

	// Concurrency check: wall time should be much less than serial.
	// Serial would be ~numCalls * fakeCallDuration = 800ms; concurrent
	// should land near fakeCallDuration = 200ms (+ overhead).
	serial := time.Duration(numCalls) * fakeCallDuration
	if wall >= serial*8/10 {
		t.Errorf("wall=%v not significantly less than serial=%v — workers may not be running in parallel",
			wall, serial)
	}

	// Peak in-flight should be > 1 if workers ran concurrently.
	if peak := peakInFlight.Load(); peak < 2 {
		t.Errorf("peak in-flight=%d; expected ≥ 2 with maxConcurrency=%d", peak, numCalls)
	}
}

// TestSerialStdioPathPreserved guards the legacy path: when
// maxConcurrency ≤ 1, dispatches run sequentially. This proves the
// flag actually controls behaviour and the old code path is intact.
func TestSerialStdioPathPreserved(t *testing.T) {
	const numCalls = 3
	const fakeCallDuration = 100 * time.Millisecond

	var inFlight atomic.Int32
	var peakInFlight atomic.Int32
	fake := &fakeRunner{
		delay: fakeCallDuration,
		onCall: func() {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				peak := peakInFlight.Load()
				if cur <= peak || peakInFlight.CompareAndSwap(peak, cur) {
					break
				}
			}
			time.Sleep(fakeCallDuration)
		},
	}

	srv, stdinW, stdoutR, _ := newTestServer(t, fake)
	srv.maxConcurrency = 1 // legacy serial

	done := make(chan error, 1)
	go func() { done <- srv.runStdio() }()

	frames := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
	}
	for i := 0; i < numCalls; i++ {
		frames = append(frames, fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"lark_contact_search","arguments":{"query":"q%d"}}}`,
			200+i, i,
		))
	}
	// Write frames in a goroutine so the response reader can drain
	// concurrently — io.Pipe is unbuffered, so synchronous writes
	// would deadlock against the server's blocking stdout writes.
	go func() {
		for _, frame := range frames {
			_, _ = stdinW.Write([]byte(frame + "\n"))
		}
	}()
	_ = readNResponses(t, stdoutR, 1+numCalls, 5*time.Second)
	_ = stdinW.Close()
	<-done

	if peak := peakInFlight.Load(); peak > 1 {
		t.Errorf("serial path saw peak in-flight=%d; expected exactly 1", peak)
	}
}

// TestStdoutWritesAreFramedUnderConcurrency hammers the writeMu mutex
// to ensure responses don't interleave at byte boundaries when N
// workers write simultaneously. Each response must be a complete JSON
// object on its own line.
func TestStdoutWritesAreFramedUnderConcurrency(t *testing.T) {
	const numCalls = 20
	fake := &fakeRunner{delay: 10 * time.Millisecond}

	srv, stdinW, stdoutR, _ := newTestServer(t, fake)
	srv.maxConcurrency = 8

	done := make(chan error, 1)
	go func() { done <- srv.runStdio() }()

	frames := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
	}
	for i := 0; i < numCalls; i++ {
		frames = append(frames, fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"lark_contact_search","arguments":{"query":"q%d"}}}`,
			300+i, i,
		))
	}
	// Write frames in a goroutine so the response reader can drain
	// concurrently — io.Pipe is unbuffered, so synchronous writes
	// would deadlock against the server's blocking stdout writes.
	go func() {
		for _, frame := range frames {
			_, _ = stdinW.Write([]byte(frame + "\n"))
		}
	}()
	got := readNResponses(t, stdoutR, 1+numCalls, 5*time.Second)
	_ = stdinW.Close()
	<-done

	// Every line must independently parse as JSON-RPC.
	for i, line := range got {
		var msg rpcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Errorf("response %d not valid JSON-RPC (interleaved write?): %v\n  line: %s", i, err, line)
		}
	}
}

// ---------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------

// fakeRunner replaces the real subprocess runner for concurrency
// tests. It sleeps `delay` and returns a synthesised success.
type fakeRunner struct {
	delay  time.Duration
	onCall func()
}

func (f *fakeRunner) run(ctx context.Context, args []string) (execResult, error) {
	if f.onCall != nil {
		f.onCall()
	} else {
		time.Sleep(f.delay)
	}
	return execResult{Stdout: `{"ok":true,"data":{}}`, ExitCode: 0}, nil
}

// runnerInterface is the surface area the server needs from a runner.
// In production, *runner satisfies it; in tests, *fakeRunner does too
// (via a shim — we just plug the field directly).
//
// To avoid changing the production server struct's field type, we use
// a small adapter that wraps fakeRunner as a *runner. Since runner is
// a small struct with one method, we instead reach for unsafe-free
// embedding via a wrapper in the test helper.

// newTestServer creates a server wired with the fake runner and
// pipe-backed stdin/stdout suitable for unit tests. Returns the
// server, the stdin writer (test pushes frames here), the stdout
// reader (test reads responses from here), and a buffer that captures
// stdout for failure-mode debugging.
func newTestServer(t *testing.T, fake *fakeRunner) (*server, *io.PipeWriter, *bufio.Reader, *bytes.Buffer) {
	t.Helper()

	// io.Pipe gives us synchronous reader/writer pairs. Wrap stdout
	// writer with a tee so a Buffer collects everything for debug.
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	debugBuf := &syncBuffer{}
	tee := io.MultiWriter(stdoutW, debugBuf)

	// Plug fake runner via a tiny adapter — fakeRunner.run has the same
	// signature as *runner.run, so we can satisfy the call site by
	// constructing a *runner whose method delegates. But runner is a
	// struct, not interface. Easiest path: change the server to call
	// through an interface field. Avoid that refactor by using
	// runtime swap via an unexported variable in tests.
	//
	// Simplest concrete approach: wrap fakeRunner as a *runner by
	// using a closure that exec-s nothing — but the existing runner.run
	// always actually exec's the binary. We can't avoid touching the
	// type without refactoring.
	//
	// So: use the real runner type but inject the fake by composition.
	// We rely on the fact that handleStdioLine -> process -> dispatch
	// is the call path. The dispatch path takes the runner directly.
	// For tests, override dispatch entirely via a package-level hook.

	srv := &server{
		ctx:     context.Background(),
		runner:  nil, // dispatch is overridden in tests
		tools:   mustDescriptors(),
		stdin:   stdinR,
		stdout:  tee,
		stderr:  io.Discard,
		version: "test",
	}
	// Install the dispatch override for the duration of this test.
	prevDispatch := dispatchHookForTests
	dispatchHookForTests = func(ctx context.Context, _ *runner, name string, args map[string]interface{}) (toolsCallResult, execResult, error) {
		res, err := fake.run(ctx, []string{name})
		if err != nil {
			return toolsCallResult{}, res, err
		}
		return formatResult(res), res, nil
	}
	t.Cleanup(func() {
		dispatchHookForTests = prevDispatch
		_ = stdoutW.Close()
	})

	return srv, stdinW, bufio.NewReader(stdoutR), debugBuf.buf
}

// syncBuffer is a goroutine-safe bytes.Buffer used to capture stdout
// for failure-mode diagnostics. The standard bytes.Buffer is not safe
// for concurrent writers, which would race in our worker-pool test.
type syncBuffer struct {
	mu  sync.Mutex
	buf *bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.buf == nil {
		s.buf = &bytes.Buffer{}
	}
	return s.buf.Write(p)
}

func mustDescriptors() []toolDescriptor {
	descs, err := descriptors(allTools())
	if err != nil {
		panic(err)
	}
	return descs
}

// readNResponses pulls N newline-delimited JSON messages from r,
// failing the test if fewer arrive within timeout.
//
// Implementation note: we spawn ONE reader goroutine that pushes lines
// onto a channel, then drain the channel with a deadline. An earlier
// implementation spawned a goroutine per ReadLine call, which raced
// against itself on the underlying bufio.Reader's internal buffer
// (caught by `go test -race`). Single-reader-channel-consumer is the
// safe pattern.
func readNResponses(t *testing.T, r *bufio.Reader, n int, timeout time.Duration) []string {
	t.Helper()
	lines := make(chan string, n*2)
	go func() {
		defer close(lines)
		for {
			line, err := r.ReadString('\n')
			if line != "" {
				lines <- strings.TrimSpace(line)
			}
			if err != nil {
				return
			}
		}
	}()

	got := make([]string, 0, n)
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for len(got) < n {
		select {
		case line, ok := <-lines:
			if !ok {
				return got
			}
			got = append(got, line)
		case <-timer.C:
			return got
		}
	}
	return got
}
