package ws

import (
	"encoding/hex"
	"github.com/golang/snappy"
	"github.com/patrickmn/go-cache"
	"gofly/pkg/cipher"
	"gofly/pkg/utils"
	"testing"
)

func BenchmarkEncryptNoXOR(b *testing.B) {
	cipher.SetKey("asdjakflrdeghyirtoy54ytiohjgfkbfjghklfjhfkitht")
	c := make([]byte, 1500)
	for i := 0; i < 1500; i++ {
		c[i] = byte(i + 2%255)
	}
	for i := 0; i < b.N; i++ {
		c = snappy.Encode(nil, c)
	}
}

func BenchmarkEncryptNoCompress(b *testing.B) {
	cipher.SetKey("asdjakflrdeghyirtoy54ytiohjgfkbfjghklfjhfkitht")
	c := make([]byte, 1500)
	for i := 0; i < 1500; i++ {
		c[i] = byte(i + 2%255)
	}
	for i := 0; i < b.N; i++ {
		c = cipher.XOR(c)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	cipher.SetKey("asdjakflrdeghyirtoy54ytiohjgfkbfjghklfjhfkitht")
	c := make([]byte, 1500)
	for i := 0; i < 1500; i++ {
		c[i] = byte(i + 2%255)
	}
	for i := 0; i < b.N; i++ {
		c = cipher.XOR(c)
		c = snappy.Encode(nil, c)
	}
}

func BenchmarkGetSrcKey(b *testing.B) {
	ipv6Packet, _ := hex.DecodeString("1ed52ffd72ac007087e004f486dd6001e9ed00200640240e037926cb4a000000000000000635200148380000001b000000000000020199aa0050bceb72fdd9aaa1568010008a090300000101080af3184d09ecce598b")
	for i := 0; i < b.N; i++ {
		utils.GetSrcKey(ipv6Packet)
	}
}

func BenchmarkGetDstKey(b *testing.B) {
	ipv6Packet, _ := hex.DecodeString("1ed52ffd72ac007087e004f486dd6001e9ed00200640240e037926cb4a000000000000000635200148380000001b000000000000020199aa0050bceb72fdd9aaa1568010008a090300000101080af3184d09ecce598b")
	for i := 0; i < b.N; i++ {
		utils.GetDstKey(ipv6Packet)
	}
}

func BenchmarkCache(b *testing.B) {
	ca := cache.New(cache.DefaultExpiration, cache.NoExpiration)
	for i := 0; i < 10000; i++ {
		ca.Set("aaa", i, cache.DefaultExpiration)
	}
	for i := 0; i < b.N; i++ {
		ca.Set("aaa", i, cache.DefaultExpiration)
		ca.Get("aaa")
	}
}
