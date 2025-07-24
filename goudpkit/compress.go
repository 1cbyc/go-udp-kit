package goudpkit

func (kit *GoUDPKit) Compress(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	var compressed []byte
	count := 1
	current := data[0]

	for i := 1; i < len(data); i++ {
		if data[i] == current && count < 255 {
			count++
		} else {
			compressed = append(compressed, byte(count), current)
			count = 1
			current = data[i]
		}
	}
	compressed = append(compressed, byte(count), current)

	return compressed
}

func (kit *GoUDPKit) Decompress(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	var decompressed []byte

	for i := 0; i < len(data); i += 2 {
		count := int(data[i])
		value := data[i+1]
		for j := 0; j < count; j++ {
			decompressed = append(decompressed, value)
		}
	}

	return decompressed
}
