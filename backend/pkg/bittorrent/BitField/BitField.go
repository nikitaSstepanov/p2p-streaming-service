package BitField;

type BT []byte;

func (b BT) Has(ind int) bool {
	ind /= 8;
	if ind < 0 || ind > len(b) {
		return false;
	}
	return 1 & (b[ind] >> uint(7 - (ind% 8))) == 1;
}

func (b BT) Set(ind int) {
	ind /= 8;
	if ind < 0 || ind > len(b) {
		return ;
	}
	b[ind] |= 1 << uint(7 - (ind % 8));
	return ;
}