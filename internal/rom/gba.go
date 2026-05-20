package rom

func newGBAROM(fileName string, header []byte, sha1Hex string) *ROM {
	titleId := readString(header, 0xAC, 4)
	return newROM(fileName, GBA,
		WithHeader(header),
		WithSha1(sha1Hex),
		withTitleId(titleId),
	)
}
