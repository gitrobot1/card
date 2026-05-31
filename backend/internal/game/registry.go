package game

type Type string

const (
	TypeDouDizhu  Type = "doudizhu"
	TypeZhajinhua Type = "zhajinhua"
	TypeDouNiu    Type = "douniu"
	TypeSanguosha Type = "sanguosha"
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
		{Type: TypeDouNiu, Name: "斗牛", Description: "凑十比点数，敬请期待", Enabled: false},
		{Type: TypeSanguosha, Name: "三国杀", Description: "身份策略卡牌，敬请期待", Enabled: false},
		{Type: TypeUNO, Name: "UNO", Description: "经典变色牌，敬请期待", Enabled: false},
	}
}
