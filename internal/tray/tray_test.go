package tray

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		input    string
		max      int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"", 5, ""},
		{"test", 4, "test"},
		{"this is a very long string", 10, "this is a…"},
	}

	for _, c := range cases {
		got := truncate(c.input, c.max)
		if got != c.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.input, c.max, got, c.expected)
		}
	}
}

func TestReasonEmoji(t *testing.T) {
	cases := []struct {
		reason   string
		expected string
	}{
		{"comment", "💬"},
		{"mention", "📢"},
		{"team_mention", "📢"},
		{"review_requested", "👀"},
		{"author", "✏️"},
		{"subscribed", "🔔"},
		{"manual", "📌"},
		{"review_comment", "📝"},
		{"issue_comment", "💭"},
		{"unknown", "•"},
		{"", "•"},
	}

	for _, c := range cases {
		got := reasonEmoji(c.reason)
		if got != c.expected {
			t.Errorf("reasonEmoji(%q) = %q, want %q", c.reason, got, c.expected)
		}
	}
}

func TestHumanReason(t *testing.T) {
	cases := []struct {
		reason   string
		expected string
	}{
		{"comment", "Nuevo comentario en tu PR"},
		{"mention", "Te mencionaron en un comentario"},
		{"team_mention", "Mencionaron a tu equipo"},
		{"review_requested", "Te pidieron review"},
		{"author", "Actividad en un PR tuyo"},
		{"subscribed", "Actividad en PR suscripto"},
		{"manual", "Notificación manual"},
		{"review_comment", "Comentario en código"},
		{"issue_comment", "Comentario en PR"},
		{"unknown_reason", "unknown_reason"},
	}

	for _, c := range cases {
		got := humanReason(c.reason)
		if got != c.expected {
			t.Errorf("humanReason(%q) = %q, want %q", c.reason, got, c.expected)
		}
	}
}

func TestMenuItem_Struct(t *testing.T) {
	// Test que el struct menuItem existe y tiene los campos correctos
	// No podemos crear systray.MenuItem en tests, pero podemos verificar la estructura
	mi := menuItem{}
	if mi.parent != nil {
		t.Error("menuItem.parent should be nil by default")
	}
	if mi.open != nil {
		t.Error("menuItem.open should be nil by default")
	}
	if mi.resolve != nil {
		t.Error("menuItem.resolve should be nil by default")
	}
}

func TestMaxItems_Constant(t *testing.T) {
	if maxItems != 15 {
		t.Errorf("maxItems = %d, want 15", maxItems)
	}
}
