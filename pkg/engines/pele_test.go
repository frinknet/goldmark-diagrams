package engines

import (
	"reflect"
	"testing"
)

func TestXYChartLabels(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "band",
			source: "xychart-beta\n  x-axis [Jan, Feb, Mar]\n  bar [1, 2, 3]\n",
			want:   []string{"Jan", "Feb", "Mar"},
		},
		{
			name:   "titled band",
			source: "xychart-beta\n  x-axis \"Month\" [Jan, Feb]\n  line [1, 2]\n",
			want:   []string{"Jan", "Feb"},
		},
		{
			name:   "quoted comma",
			source: "xychart-beta\n  x-axis [\"Jan, W1\", Feb]\n  bar [1, 2]\n",
			want:   []string{"Jan, W1", "Feb"},
		},
		{
			name:   "numeric range",
			source: "xychart-beta\n  x-axis 0 --> 10\n  bar [1, 2]\n",
			want:   nil,
		},
		{
			name:   "not an xy chart",
			source: "flowchart TD\n  A --> B\n",
			want:   nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := xychartLabels(tc.source); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("xychartLabels(%q) = %v, want %v", tc.source, got, tc.want)
			}
		})
	}
}
