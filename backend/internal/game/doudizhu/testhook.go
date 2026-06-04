//go:build cardtest

package doudizhu

// DebugDealOneCardEnabled 供 backend/test 集成测试读取 debug 开关（需 -tags cardtest）。
func DebugDealOneCardEnabled() bool { return debugDealOneCard }
