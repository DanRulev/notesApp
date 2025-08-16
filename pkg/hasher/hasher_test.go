package hasher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasher_GenerateHash(t *testing.T) {
	t.Parallel()

	type args struct {
		password string
	}
	tests := []struct {
		name    string
		h       *Hasher
		args    args
		wantErr bool
	}{
		{
			name: "success",
			h:    NewHasher(),
			args: args{
				password: "test_password",
			},
			wantErr: false,
		},
		{
			name: "different hashes",
			h:    NewHasher(),
			args: args{
				password: "test_password",
			},
			wantErr: false,
		},
		{
			name: "empty password",
			h:    NewHasher(),
			args: args{
				password: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.h.GenerateHash(tt.args.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, got)
			require.NotEqual(t, tt.args.password, got)

			if tt.name == "different hashes" {
				got1, err := tt.h.GenerateHash(tt.args.password)
				require.NoError(t, err)
				require.NotEqual(t, got, got1)
			}

			err = tt.h.ComparePassword(got, tt.args.password)
			require.NoError(t, err)
		})
	}
}

func TestHasher_ComparePassword(t *testing.T) {
	t.Parallel()

	type args struct {
		passForHash string
		password    string
	}
	tests := []struct {
		name    string
		h       *Hasher
		args    args
		wantErr bool
	}{
		{
			name: "success",
			h:    NewHasher(),
			args: args{
				passForHash: "test",
				password:    "test",
			},
			wantErr: false,
		},
		{
			name: "empty hash",
			h:    NewHasher(),
			args: args{
				password: "test",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			h:    NewHasher(),
			args: args{
				passForHash: "test",
			},
			wantErr: true,
		},
		{
			name: "wrong password",
			h:    NewHasher(),
			args: args{
				passForHash: "test",
				password:    "wrong",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, _ := tt.h.GenerateHash(tt.args.passForHash)

			err := tt.h.ComparePassword(hash, tt.args.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
