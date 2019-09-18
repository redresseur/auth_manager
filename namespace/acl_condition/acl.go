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
	pricinple     string
	subPolicies   []*ApiPolicy
	pricinpleType uint8
	mOfN          int
}

func (p *ApiPolicy) check(authList map[string]bool) bool {
	res := true
	switch p.pricinpleType {
	case ANDPRICINPLE:
		for _, p := range p.subPolicies {
			if !p.check(authList) {
				res = false
				break
			}
		}
	case ORPRICINPLE:
		for _, p := range p.subPolicies {
			if p.check(authList) {
				res = true
				break
			} else {
				res = false
			}
		}
	case BASEPRICINPLE:
		if !authList[p.pricinple] {
			res = false
		}
	case MOfNPRICINPLE:
		count := 0
		for _, p := range p.subPolicies {
			if p.check(authList) {
				count++
			}
		}

		if count < p.mOfN {
			res = false
		}
	default:
		res = false
	}

	return res
}

func AND(policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.subPolicies = append(res.subPolicies, policies...)
	res.pricinpleType = ANDPRICINPLE
	return &res
}

func OR(policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.subPolicies = append(res.subPolicies, policies...)
	res.pricinpleType = ORPRICINPLE
	return &res
}

func MOfN(mOfN int, policies ...*ApiPolicy) *ApiPolicy {
	res := ApiPolicy{}

	res.subPolicies = append(res.subPolicies, policies...)
	res.pricinpleType = MOfNPRICINPLE
	res.mOfN = mOfN
	return &res
}

func BASE(pricinple string) *ApiPolicy {
	return &ApiPolicy{
		pricinple:     pricinple,
		pricinpleType: BASEPRICINPLE,
	}
}

func Check(p *ApiPolicy, authList map[string]bool) bool {
	if p == nil {
		return false
	}

	return p.check(authList)
}
