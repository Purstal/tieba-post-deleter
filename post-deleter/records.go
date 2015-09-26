package post_deleter

type Records struct {
	WaterThread_Tids      map[uint64]struct{}
	ServerListThread_Tids map[uint64]struct{} //mc吧专用
	RulesThread_Tids      map[uint64]struct{}
}
