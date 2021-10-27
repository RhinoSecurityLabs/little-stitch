package lib

import (
	"testing"
)

func TestPing(t *testing.T) {
	type args struct {
		ip []byte
		port int
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"returns true and no error when ran against a open port",
			args{
				ip: []byte{1,1,1,1},
				port: 80,
			},
			true,
			false,
		},
		{
			"returns false and an error when ran against a filtered port",
			args{
				ip: []byte{1,1,1,1},
				port: 1234,
			},
			false,
			true,
		},
		{
			"returns false and an error when ran against a closed port",
			args{
				ip: []byte{127, 0, 0, 1},
				port: 1234,
			},
			false,
			true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ping(tt.args.ip, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ping() got = %v, want %v", got, tt.want)
			}
		})
	}
}
