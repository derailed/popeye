package internal

type (
	GVR  int
	GVRs map[GVR]string
)

const (
	LrGVR GVR = iota
	SvcGVR
	EpGVR
	NoGVR
	NsGVR
	PoGVR
	CmGVR
	SecGVR
	SaGVR
	PvGVR
	PvcGVR
	DpGVR
	RsGVR
	DsGVR
	StsGVR
	NpGVR
	CrGVR
	CrbGVR
	RoGVR
	RobGVR
	IngGVR
	PdbGVR
	PspGVR
	HpaGVR
)
