package addrdec

import (
	"fmt"
	"github.com/blocktree/openwallet/openwallet"
	"strings"

	"github.com/blocktree/go-owcdrivers/addressEncoder"
)

var (
	EOSPublicKeyPrefix       = "PUB_"
	EOSPublicKeyK1Prefix     = "PUB_K1_"
	EOSPublicKeyR1Prefix     = "PUB_R1_"
	EOSPublicKeyPrefixCompat = "EOS"

	//EOS stuff
	EOS_mainnetPublic               = addressEncoder.AddressType{"eos", addressEncoder.BTCAlphabet, "ripemd160", "", 33, []byte(EOSPublicKeyPrefixCompat), nil}
	EOS_mainnetPrivateWIF           = addressEncoder.AddressType{"base58", addressEncoder.BTCAlphabet, "doubleSHA256", "", 32, []byte{0x80}, nil}
	EOS_mainnetPrivateWIFCompressed = addressEncoder.AddressType{"base58", addressEncoder.BTCAlphabet, "doubleSHA256", "", 32, []byte{0x80}, []byte{0x01}}

	Default = AddressDecoderV2{}
)

//AddressDecoderV2
type AddressDecoderV2 struct {
	openwallet.AddressDecoderV2Base
	IsTestNet bool
}

// AddressDecode decode address
func (dec *AddressDecoderV2) AddressDecode(pubKey string, opts ...interface{}) ([]byte, error) {

	var pubKeyMaterial string
	if strings.HasPrefix(pubKey, EOSPublicKeyR1Prefix) {
		pubKeyMaterial = pubKey[len(EOSPublicKeyR1Prefix):] // strip "PUB_R1_"
	} else if strings.HasPrefix(pubKey, EOSPublicKeyK1Prefix) {
		pubKeyMaterial = pubKey[len(EOSPublicKeyK1Prefix):] // strip "PUB_K1_"
	} else if strings.HasPrefix(pubKey, EOSPublicKeyPrefixCompat) { // "EOS"
		pubKeyMaterial = pubKey[len(EOSPublicKeyPrefixCompat):] // strip "EOS"
	} else {
		return nil, fmt.Errorf("public key should start with [%q | %q] (or the old %q)", EOSPublicKeyK1Prefix, EOSPublicKeyR1Prefix, EOSPublicKeyPrefixCompat)
	}

	ret, err := addressEncoder.Base58Decode(pubKeyMaterial, addressEncoder.NewBase58Alphabet(EOS_mainnetPublic.Alphabet))
	if err != nil {
		return nil, addressEncoder.ErrorInvalidAddress
	}
	if addressEncoder.VerifyChecksum(ret, EOS_mainnetPublic.ChecksumType) == false {
		return nil, addressEncoder.ErrorInvalidAddress
	}

	return ret[:len(ret)-4], nil
}

// AddressEncode encode address
func (dec *AddressDecoderV2) AddressEncode(hash []byte, opts ...interface{}) (string, error) {
	data := addressEncoder.CatData(hash, addressEncoder.CalcChecksum(hash, EOS_mainnetPublic.ChecksumType))
	return string(EOS_mainnetPublic.Prefix) + addressEncoder.EncodeData(data, "base58", EOS_mainnetPublic.Alphabet), nil
}
