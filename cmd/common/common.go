package common

import "github.com/GiGurra/boa/pkg/boa"

func DefaultParamEnricher() boa.ParamEnricher {
	return boa.ParamEnricherCombine(
		boa.ParamEnricherBool,
		boa.ParamEnricherName,
		boa.ParamEnricherShort,
	)
}
