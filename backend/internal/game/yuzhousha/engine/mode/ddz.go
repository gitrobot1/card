package mode

const Solo3pDdz = "3p_ddz"

func Is3pDdz(ctx Context) bool {
	return ctx.ModeID() == Solo3pDdz
}

// LandlordContext exposes the landlord seat for 3p ddz team rules.
type LandlordContext interface {
	Context
	DdzLandlordSeat() int
}

func landlordSeat(ctx Context) (int, bool) {
	lc, ok := ctx.(LandlordContext)
	if !ok || !Is3pDdz(ctx) {
		return 0, false
	}
	return lc.DdzLandlordSeat(), true
}
