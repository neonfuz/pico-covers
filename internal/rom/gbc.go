package rom

func newGBCROM(fileName string, header []byte, sha1Hex string) *ROM {
	r := newROM(fileName, GBC,
		WithHeader(header),
		WithSha1(sha1Hex),
	)

	if len(header) > 0x143 && header[0x143] == 0xC0 {
		r.TitleId = readString(header, 0x13F, 4)
	}

	return r
}
