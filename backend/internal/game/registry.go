package game

type Type string

const (
	TypeDouDizhu  Type = "doudizhu"
	TypeZhajinhua Type = "zhajinhua"
	TypeDouNiu    Type = "douniu"
	TypeYuzhousha Type = "yuzhousha"
	TypeUNO       Type = "uno"
)

type Meta struct {
	Type        Type   `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func Catalog() []Meta {
	return []Meta{
		{Type: TypeDouDizhu, Name: "斗地主", Description: "三人扑克，抢地主对战", Enabled: true},
		{Type: TypeZhajinhua, Name: "扎金花", Description: "2-8人比牌，牌型倍率结算", Enabled: true},
		{Type: TypeDouNiu, Name: "斗牛", Description: "看牌抢庄，2-8人比牛结算", Enabled: true},
		{Type: TypeYuzhousha, Name: "宇宙杀", Description: "1v1 策略对战，基础杀闪桃", Enabled: true},
		{Type: TypeUNO, Name: "UNO", Description: "2-8人变色牌，先出完获胜", Enabled: true},
	}
}
