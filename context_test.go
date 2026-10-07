package azfunc

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/potatoattack/azfunc/data"
	"github.com/potatoattack/azfunc/output"
)

func TestNewContext_PreservesParent(t *testing.T) {
	type contextKey struct{}
	parent, cancel := context.WithTimeout(context.WithValue(context.Background(), contextKey{}, "fixture"), time.Hour)
	defer cancel()
	ctx := newContext(parent)
	wantDeadline, _ := parent.Deadline()
	if deadline, ok := ctx.Deadline(); !ok || !deadline.Equal(wantDeadline) {
		t.Fatalf("Deadline() = %v, %v; want %v, true", deadline, ok, wantDeadline)
	}
	if value := ctx.Value(contextKey{}); value != "fixture" {
		t.Fatalf("Value() = %v; want fixture", value)
	}
	cancel()
	select {
	case <-ctx.Done():
	default:
		t.Fatal("parent cancellation was not propagated")
	}
	if err := ctx.Err(); err != context.Canceled {
		t.Fatalf("Err() = %v; want %v", err, context.Canceled)
	}
}

func TestNewContext_ReusesExistingContext(t *testing.T) {
	parent := &Context{Context: context.Background()}
	if got := newContext(parent); got != parent {
		t.Fatal("an existing function context must be reused")
	}
}

type contextTestTransport func(*http.Request) (*http.Response, error)

func (f contextTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewContext_WithHTTPClient(t *testing.T) {
	ctx := newContext(context.Background())
	calls := 0
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: contextTestTransport(func(req *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("{}")),
				Request:    req,
			}, nil
		}),
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://fixture.invalid/discovery", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || calls != 1 {
		t.Fatalf("HTTP response = %d, transport calls = %d; want 200, 1", resp.StatusCode, calls)
	}
}

func TestContext_Output_HTTP_Write(t *testing.T) {
	ctx := Context{
		Outputs: &outputs{},
	}
	ctx.Outputs.Add(output.NewHTTP())
	ctx.Outputs.HTTP().Write([]byte(`{"message":"hello"}`))

	want := output.NewHTTP(func(o *output.HTTPOptions) {
		o.Body = data.Raw(`{"message":"hello"}`)
	})
	got := ctx.Outputs.HTTP()

	if diff := cmp.Diff(want.Data(), got.Data(), cmp.AllowUnexported(output.HTTP{})); diff != "" {
		t.Errorf("Write() = unexpected result (-want +got)\n%s\n", diff)
	}
}

func TestContext_Output_HTTP_WriteHeader(t *testing.T) {
	ctx := Context{
		Outputs: &outputs{},
	}
	ctx.Outputs.Add(output.NewHTTP())
	ctx.Outputs.HTTP().WriteHeader(http.StatusNotFound)

	want := output.NewHTTP()
	want.WriteHeader(http.StatusNotFound)
	got := ctx.Outputs.HTTP()

	if diff := cmp.Diff(want, got, cmp.AllowUnexported(output.HTTP{})); diff != "" {
		t.Errorf("Write() = unexpected result (-want +got)\n%s\n", diff)
	}
}

func TestContext_Output_HTTP_Header_Add(t *testing.T) {
	ctx := Context{
		Outputs: &outputs{},
	}
	ctx.Outputs.Add(output.NewHTTP())
	ctx.Outputs.HTTP().Header().Add("Content-Type", "application/json")

	want := output.NewHTTP()
	want.Header().Add("Content-Type", "application/json")
	got := ctx.Outputs.HTTP()

	if diff := cmp.Diff(want.Header(), got.Header(), cmp.AllowUnexported(output.HTTP{})); diff != "" {
		t.Errorf("Write() = unexpected result (-want +got)\n%s\n", diff)
	}
}

func TestContext_Output_Trigger_Write(t *testing.T) {
	ctx := Context{
		Outputs: &outputs{},
	}
	ctx.Outputs.Add(output.NewGeneric("queue"))
	ctx.Outputs.Binding("queue").Write([]byte(`{"message":"hello"}`))
	got := ctx.Outputs.Binding("queue")

	want := output.NewGeneric("queue")
	want.Write([]byte(`{"message":"hello"}`))

	if diff := cmp.Diff(want, got, cmp.AllowUnexported(output.Generic{})); diff != "" {
		t.Errorf("Write() = unexpected result (-want +got)\n%s\n", diff)
	}
}
