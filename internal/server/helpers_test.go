package server

import "testing"

func TestHelpers_ShiftPath(t *testing.T) {
	tests := []struct {
		name,
		path,
		wantHead,
		wantTail string
	}{
		{
			name:     "1-simple",
			path:     "/foo/bar/baz",
			wantHead: "foo",
			wantTail: "/bar/baz",
		},
		{
			name:     "2-no-first-slash",
			path:     "foo/bar/baz",
			wantHead: "foo",
			wantTail: "/bar/baz",
		},
		{
			name:     "3-no-tail",
			path:     "/foo",
			wantHead: "foo",
			wantTail: "/",
		},
		{
			name:     "4-root",
			path:     "/",
			wantHead: "",
			wantTail: "/",
		},
		{
			name:     "5-empty-path",
			path:     "",
			wantHead: "",
			wantTail: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHead, gotTail := ShiftPath(tt.path)
			if gotHead != tt.wantHead {
				t.Errorf("ShiftPath(\"%s\") gotHead = %v, want %v", tt.path, gotHead, tt.wantHead)
			}
			if gotTail != tt.wantTail {
				t.Errorf("ShiftPath(\"%s\") gotTail = %v, want %v", tt.path, gotTail, tt.wantTail)
			}
		})
	}
}
