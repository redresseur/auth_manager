package acl_condition

const (
	// and
	ANDPRICINPLE uint8 = 1 << iota

	// or
	ORPRICINPLE

	// m of n
	MOfNPRICINPLE

	BASEPRICINPLE
)

type ApiPolicy struct {
	Principle     string `json:"principle"`
	SubPolicies   []*ApiPolicy `json:"subPolicies"`
	PrincipleType uint8 `json:"principleType"`
	MOfN          int `json:"mOfN"`
}

func (p *ApiPolicy) check(authList map[string]bool) bool {
	res := true
	switch p.PrincipleType {
	case ANDPRICINPLE:
		for _, p := range p.SubPolicies {
			if !p.check(authList) {
				res = false
				break
			}
		}
	case ORPRICINPLE:
		for _, p := range p.SubPolicies {
			if p.check(authList) {
				res = true
				break
			} else {
				res = false
			}
		}
	case BASEPRICINPLE:
		if !authList[p.Principle] {
			res = false
		}
	case MOfNPRICINPLE:
		count := 0
		for _, p := range p.SubPolicies {
			if p.check(authList) {
				count++
			}
		}

		if count < p.MOfN {
			res = false
		}
	default:
		res = false
	}

	return res
}

func AND(policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.SubPolicies = append(res.SubPolicies, policies...)
	res.PrincipleType = ANDPRICINPLE
	return &res
}

func OR(policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.SubPolicies = append(res.SubPolicies, policies...)
	res.PrincipleType = ORPRICINPLE
	return &res
}

func MOfN(mOfN int, policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.SubPolicies = append(res.SubPolicies, policies...)
	res.PrincipleType = MOfNPRICINPLE
	res.MOfN = mOfN
	return &res
}

func BASE(principle string) *ApiPolicy {
	return &ApiPolicy{
		Principle:     principle,
		PrincipleType: BASEPRICINPLE,
	}
}

func Check(p *ApiPolicy, authList map[string]bool) bool {
	if p == nil {
		return false
	}

	return p.check(authList)
}
