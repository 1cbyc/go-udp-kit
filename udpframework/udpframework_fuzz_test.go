//go:build go1.18

package goudpkit

import (
	"testing"
)

func FuzzCompressDecompress(f *testing.F) {
	f.Add([]byte("aaabbbccccccdddddddeee"))
	f.Fuzz(func(t *testing.T, data []byte) {
		kit, err := NewGoUDPKit(":0", RetryConfig{1, 1, 1.0}, QoSConfig{1, make([][]Packet, 1)}, BufferConfig{1, 1})
		if err != nil {
			t.Skip()
		}
		defer kit.Close()
		compressed := kit.Compress(data)
		decompressed := kit.Decompress(compressed)
		if string(decompressed) != string(data) {
			t.Fatalf("Compress/Decompress mismatch: got '%s', want '%s'", string(decompressed), string(data))
		}
	})
}

func FuzzEncryptDecrypt(f *testing.F) {
	f.Add([]byte("secret"), []byte("encrypt-this-data"))
	f.Fuzz(func(t *testing.T, key []byte, data []byte) {
		kit, err := NewGoUDPKit(":0", RetryConfig{1, 1, 1.0}, QoSConfig{1, make([][]Packet, 1)}, BufferConfig{1, 1})
		if err != nil {
			t.Skip()
		}
		defer kit.Close()
		cipher := kit.EncryptData(data, key)
		decrypted := kit.DecryptData(cipher, key)
		if string(decrypted) != string(data) {
			t.Fatalf("Encrypt/Decrypt mismatch: got '%s', want '%s'", string(decrypted), string(data))
		}
	})
}
