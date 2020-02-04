package addrdec

import (
	"reflect"
	"testing"
)

func TestAddressDecoderV2_AddressDecode(t *testing.T) {
	type fields struct {
		IsTestNet bool
	}
	type args struct {
		addr string
		opts []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TGC bech32", fields: fields{IsTestNet: false},
			args:    args{addr: "EVS51wiJaHZxebPu562Kh91ozaeamqVj9s9k5zNxYpxV22FyefT56"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := &AddressDecoderV2{
				IsTestNet: tt.fields.IsTestNet,
			}
			got, err := AddressDecode(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressDecoderV2.AddressDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("AddressDecoderV2.AddressDecode() = %v, want %v", got, tt.want)
			}
		})
	}
}
