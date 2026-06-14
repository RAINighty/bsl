package collector

import "testing"

func TestDetectStreetlight(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantHit  bool
		wantNote string
	}{
		{
			name:     "exact match without note",
			content:  "#路灯",
			wantHit:  true,
			wantNote: "#路灯",
		},
		{
			name:     "with note content",
			content:  "#路灯 中山路和解放路交叉口路灯不亮",
			wantHit:  true,
			wantNote: "中山路和解放路交叉口路灯不亮",
		},
		{
			name:     "with leading spaces",
			content:  "   #路灯 测试内容",
			wantHit:  true,
			wantNote: "测试内容",
		},
		{
			name:     "with trailing spaces in note",
			content:  "#路灯   两边有空格   ",
			wantHit:  true,
			wantNote: "两边有空格",
		},
		{
			name:     "not a streetlight - regular danmaku",
			content:  "主播好厉害",
			wantHit:  false,
			wantNote: "",
		},
		{
			name:     "not a streetlight - partial match",
			content:  "路灯坏了",
			wantHit:  false,
			wantNote: "",
		},
		{
			name:     "empty string",
			content:  "",
			wantHit:  false,
			wantNote: "",
		},
		{
			name:     "only spaces",
			content:  "   ",
			wantHit:  false,
			wantNote: "",
		},
		{
			name:     "case sensitive - uppercase",
			content:  "#路燈",
			wantHit:  false,
			wantNote: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hit, note := DetectStreetlight(tt.content)
			if hit != tt.wantHit {
				t.Errorf("hit = %v, want %v", hit, tt.wantHit)
			}
			if note != tt.wantNote {
				t.Errorf("note = %q, want %q", note, tt.wantNote)
			}
		})
	}
}
