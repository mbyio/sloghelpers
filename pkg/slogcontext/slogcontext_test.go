package slogcontexthandler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func attrsLess(a, b slog.Attr) bool {
	return a.Key < b.Key
}

func TestEmptyGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrs := GetAttrs(ctx)
	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(attrs))
	}
}

func TestAddEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = AddAttrs(ctx)
	attrs := GetAttrs(ctx)
	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(attrs))
	}
}

func TestAdd(t *testing.T) {
	subTests := []struct {
		name          string
		attrs         []slog.Attr
		expectedAttrs []slog.Attr
	}{
		{
			name:          "empty",
			attrs:         []slog.Attr{},
			expectedAttrs: nil,
		},
		{
			name:          "one",
			attrs:         []slog.Attr{slog.String("foo", "bar")},
			expectedAttrs: []slog.Attr{slog.String("foo", "bar")},
		},
		{
			name:          "two",
			attrs:         []slog.Attr{slog.String("foo", "bar"), slog.String("baz", "qux")},
			expectedAttrs: []slog.Attr{slog.String("foo", "bar"), slog.String("baz", "qux")},
		},
		{
			name:          "duplicate",
			attrs:         []slog.Attr{slog.String("foo", "bar"), slog.String("foo", "baz")},
			expectedAttrs: []slog.Attr{slog.String("foo", "baz")},
		},
		{
			name:          "duplicate2",
			attrs:         []slog.Attr{slog.String("foo", "bar"), slog.String("abc", "123"), slog.String("foo", "qux")},
			expectedAttrs: []slog.Attr{slog.String("foo", "qux"), slog.String("abc", "123")},
		},
	}
	for _, subTest := range subTests {
		subTest := subTest
		t.Run(subTest.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx = AddAttrs(ctx, subTest.attrs...)

			attrs := GetAttrs(ctx)
			t.Log("attrs:", attrs)
			t.Log("expectedAttrs:", subTest.expectedAttrs)
			if diff := cmp.Diff(subTest.expectedAttrs, attrs, cmpopts.SortSlices(attrsLess), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unexpected attributes:\n%s", diff)
			}
		})
	}
}

func TestAddRepeated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = AddAttrs(ctx, slog.String("foo", "bar"))
	ctx = AddAttrs(ctx, slog.String("baz", "qux"))
	ctx = AddAttrs(ctx, slog.String("hello", "world"))
	ctx = AddAttrs(ctx, slog.String("foo", "bar"))
	ctx = AddAttrs(ctx, slog.String("baz", "qux"))
	ctx = AddAttrs(ctx, slog.String("abc", "123"))
	attrs := GetAttrs(ctx)
	expected := []slog.Attr{
		slog.String("foo", "bar"),
		slog.String("baz", "qux"),
		slog.String("hello", "world"),
		slog.String("abc", "123"),
	}
	if diff := cmp.Diff(expected, attrs, cmpopts.SortSlices(attrsLess), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("unexpected attributes:\n%s", diff)
	}
}

func TestSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = SetAttrs(ctx, slog.Int64("foo", 123))
	ctx = SetAttrs(ctx, slog.String("bar", "baz"))
	ctx = SetAttrs(ctx, slog.String("hello", "world"), slog.Bool("doTheThing", true))
	attrs := GetAttrs(ctx)
	expected := []slog.Attr{
		slog.String("hello", "world"),
		slog.Bool("doTheThing", true),
	}
	if diff := cmp.Diff(expected, attrs, cmpopts.SortSlices(attrsLess), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("unexpected attributes:\n%s", diff)
	}
}

type recordingHandler struct {
	records []slog.Record
}

func TestHandler(t *testing.T) {
	t.Parallel()

	// Log to this buffer and then check the contents.
	// Customize the inner handler to remove fields we don't care about for the test.
	var buf bytes.Buffer
	logger := slog.New(
		New(slog.NewJSONHandler(
			&buf,
			&slog.HandlerOptions{
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "level" || a.Key == "time" {
						return slog.Attr{}
					}
					return a
				},
			},
		)))

	logInputs := []struct {
		attrs          []slog.Attr
		msg            string
		expectedOutput map[string]any
	}{
		{
			attrs: nil,
			msg:   "no attributes",
			expectedOutput: map[string]any{
				"msg": "no attributes",
			},
		},
		{
			attrs: []slog.Attr{slog.String("foo", "bar")},
			msg:   "one attribute",
			expectedOutput: map[string]any{
				"msg": "one attribute",
				"foo": "bar",
			},
		},
		{
			attrs: []slog.Attr{slog.String("foo", "bar"), slog.Int64("baz", 123)},
			msg:   "two attributes",
			expectedOutput: map[string]any{
				"msg": "two attributes",
				"foo": "bar",
				"baz": 123.0,
			},
		},
		{
			attrs: []slog.Attr{slog.String("foo", "bar"), slog.Int64("baz", 123), slog.Bool("qux", true)},
			msg:   "three attributes",
			expectedOutput: map[string]any{
				"msg": "three attributes",
				"foo": "bar",
				"baz": 123.0,
				"qux": true,
			},
		},
		{
			attrs: nil,
			msg:   "no attributes again",
			expectedOutput: map[string]any{
				"msg": "no attributes again",
			},
		},
	}

	for _, li := range logInputs {
		logger.InfoContext(AddAttrs(context.Background(), li.attrs...), li.msg)
	}

	dec := json.NewDecoder(&buf)
	for i, li := range logInputs {
		t.Logf("reading log message %v (%v)", i, logInputs[i].msg)
		logMsg := map[string]any{}
		if err := dec.Decode(&logMsg); err != nil {
			t.Fatal("failed to decode log message:", err)
		}
		if diff := cmp.Diff(li.expectedOutput, logMsg); diff != "" {
			t.Errorf("Reading log message %v: %v", i, diff)
		}
	}

	if dec.More() {
		t.Error("expected no more log messages")
	}
}
