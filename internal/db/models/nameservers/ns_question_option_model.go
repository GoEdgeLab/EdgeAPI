package nameservers

// NSQuestionOption DNS请求选项
type NSQuestionOption struct {
	Id        uint64 `field:"id"`        // ID
	Name      string `field:"name"`      // 选项名
	Values    string `field:"values"`    // 选项值
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type NSQuestionOptionOperator struct {
	Id        interface{} // ID
	Name      interface{} // 选项名
	Values    interface{} // 选项值
	CreatedAt interface{} // 创建时间
}

func NewNSQuestionOptionOperator() *NSQuestionOptionOperator {
	return &NSQuestionOptionOperator{}
}
