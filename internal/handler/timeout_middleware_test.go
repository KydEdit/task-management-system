package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware_NextReceivesDeadlineContext(t *testing.T) {
	var hasDeadline bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasDeadline = r.Context().Deadline()
		w.WriteHeader(http.StatusOK)
	})

	h := TimeoutMiddleware(3*time.Second, next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !hasDeadline {
		t.Error("next handler should see a context with deadline")
	}
}

func TestTimeoutMiddleware_ContextIsChildOfRequestContext(t *testing.T) {
	parentCtx, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	var childCtx context.Context
	entered := make(chan struct{})
	childDone := make(chan struct{})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childCtx = r.Context()
		close(entered)
		<-r.Context().Done()
		close(childDone)
	})

	h := TimeoutMiddleware(time.Hour, next)

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(parentCtx)
	rec := httptest.NewRecorder()

	go h(rec, req)

	<-entered

	if childCtx == nil {
		t.Fatal("handler was not called or context was not captured")
	}

	select {
	case <-childCtx.Done():
		t.Fatalf("child context canceled before parent: %v", childCtx.Err())
	default:
	}

	cancelParent()

	select {
	case <-childDone:
		if !errors.Is(childCtx.Err(), context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", childCtx.Err())
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("child context should be canceled when parent is canceled")
	}
}

func TestTimeoutMiddleware_TimeoutFires(t *testing.T) {
	done := make(chan struct{})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(done)
	})

	h := TimeoutMiddleware(50*time.Millisecond, next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	select {
	case <-done:

	case <-time.After(1 * time.Second):
		t.Fatal("timeout did not fire within 1 second")
	}
}
