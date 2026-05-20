package rom

func newDSiROM(fileName string, header []byte, sha1Hex string) *ROM {
	title := readString(header, 0x00, 12)
	titleId := readString(header, 0x0C, 4)
	var regionId byte
	if len(header) > 0x0F {
		regionId = header[0x0F]
	}
	_ = title
	return newROM(fileName, DSi,
		WithHeader(header),
		WithSha1(sha1Hex),
		withTitleId(titleId),
		withRegionId(regionId),
	)
}
