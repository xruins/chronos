package chronos

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGenerateTemplateFuncMap(t *testing.T) {
	j := &Job{
		execution: []*Execution{
			{
				count:        1,
				executedTime: time.Now().Add(-3 * time.Hour),
				succeeded:    true,
			},
			{
				count:        2,
				executedTime: time.Now().Add(-2 * time.Hour),
				succeeded:    false,
			},
			{
				count:        3,
				executedTime: time.Now().Add(-1 * time.Hour),
				succeeded:    true,
			},
		},
	}

	tf := j.generateTemplateFuncMap(map[string]string{
		"foo":  "bar",
		"hoge": "fuga",
	})

	args1 := `
{{env "foo"}}
{{env "hoge"}}
{{count}}`

	got := applyTemplate(t, tf, args1)
	want := `
bar
fuga
3`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected applied value. diff: %s", diff)
	}

	args2 := `{{time "2006-01-02T15:04:05Z07:00"}}`
	applied := applyTemplate(t, tf, args2)
	gotTime, err := time.Parse("2006-01-02T15:04:05Z07:00", applied)
	if got != want {
		t.Errorf("failed to returned value as time. got: %s, err: %s", got, err)
	}
	wantTime := time.Now()

	if !cmp.Equal(gotTime, wantTime, cmpopts.EquateApproxTime(time.Minute)) {
		t.Errorf("got time differs from wanted one over 1 minute. got: %s, want: %s", gotTime, wantTime)
	}
}

func applyTemplate(t *testing.T, tf map[string]interface{}, example string) string {
	tmpl := template.Must(template.New("template").Funcs(tf).Parse(example))
	w := new(bytes.Buffer)
	err := tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("failed to apply template. example: %s, err: %s", example, err)
	}
	return w.String()
}
