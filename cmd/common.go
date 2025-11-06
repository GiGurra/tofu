package cmd

import "github.com/GiGurra/boa/pkg/boa"

func defaultParamEnricher() boa.ParamEnricher {
	return boa.ParamEnricherCombine(
		boa.ParamEnricherBool,
		boa.ParamEnricherName,
		boa.ParamEnricherShort,
	)
}
