package proxy

import (
	"errors"
	"testing"
)

func TestErrors_Add(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		e    *Errors
		args args
		len  int
	}{
		{
			name: "Happy Path",
			e: &Errors{
				errs: []error{
					errors.New("1 error"),
				},
			},
			args: args{
				err: errors.New("2 error"),
			},
			len: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e.Add(tt.args.err)
			if len(tt.e.errs) != tt.len {
				t.Errorf("len(tt.e.errs) = %d != tt.len = %d", len(tt.e.errs), tt.len)
			}
			if tt.e.errs[tt.len-1].Error() != tt.args.err.Error() {
				t.Errorf("errors do not match")
			}
		})
	}
}
