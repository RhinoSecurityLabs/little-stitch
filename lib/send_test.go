package lib

import (
	"syscall"
	"testing"
)

func TestPing(t *testing.T) {
	type args struct {
		addr syscall.Sockaddr
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
				&syscall.SockaddrInet4{
					Port: 80,
					Addr: [4]byte{1,1,1,1},
				},
			},
			true,
			false,
		},
		{
			"returns false and an error when ran against a filtered port",
			args{
				&syscall.SockaddrInet4{
					Port: 1234,
					Addr: [4]byte{1,1,1,1},
				},
			},
			true,
			false,
		},
		{
			"returns false and an error when ran against a closed port",
			args{
				&syscall.SockaddrInet4{
					Port: 1234,
					Addr: [4]byte{127, 0, 0, 1},
				},
			},
			true,
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ping(tt.args.addr)
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
