package goudpkit

func (kit *GoUDPKit) EncryptData(data []byte, key []byte) []byte {
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted
}

func (kit *GoUDPKit) DecryptData(data []byte, key []byte) []byte {
	return kit.EncryptData(data, key)
}
