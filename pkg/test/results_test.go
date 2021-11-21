// Copyright 2021 Comcast. All rights reserved.

package test

import (
	"fmt"
	"testing"
)

func TestCheckTestResults(t *testing.T) {
	type args struct {
		got     interface{}
		want    interface{}
		err     error
		wantErr interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test WantRegexpError No Error",
			args: args{
				err:     fmt.Errorf("General part: system dependent"),
				wantErr: NewWantRegexpError("(General part:).*"),
			},
			wantErr: false,
		},
		{
			name: "Test WantRegexpError With Error",
			args: args{
				err:     fmt.Errorf("General part does not match: system dependent"),
				wantErr: NewWantRegexpError("(General part:).*"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckTestResults(tt.args.got, tt.args.want, tt.args.err, tt.args.wantErr); (err != nil) != tt.wantErr {
				t.Errorf("CheckTestResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
