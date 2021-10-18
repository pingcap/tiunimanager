package warehouse

type ZoneHostStock struct {
	ZoneBaseInfo
	SpecBaseInfo
	Count int
}

type ZoneBaseInfo struct {
	ZoneCode string
	ZoneName string
}

type SpecBaseInfo struct {
	SpecCode string
	SpecName string
}
