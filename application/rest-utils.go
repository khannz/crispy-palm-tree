package application

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// func customPortServerApplicationValidation(sl validator.StructLevel) {
// 	sA := sl.Current().Interface().(ServerApplication)
// 	port, err := strconv.Atoi(sA.ServerPort)
// 	if err != nil {
// 		sl.ReportError(sA.ServerPort, "serverPort", "ServerPort", "port must be number", "")
// 	}
// 	if !(port > 0) || !(port < 20000) {
// 		sl.ReportError(sA.ServerPort, "serverPort", "ServerPort", "port must gt=0 and lt=20000", "")
// 	}
// }

// func customServiceHealthcheckValidation(sl validator.StructLevel) {
// 	sHc := sl.Current().Interface().(ServiceHealthcheck)
// 	// sA := sl.Current().Interface().([]ServerApplication) // FIXME: broken
// 	switch sHc.Type {
// 	case "tcp":
// 	case "http":
// 	case "icmp":
// 	case "http-advanced":
// 		// if sA != nil && len(sA) > 0 {
// 		// 	for i, applicationServer := range sA {
// 		// 		switch applicationServer.ServerHealthcheck.TypeOfCheck {
// 		// 		case "http-advanced-json":
// 		// 			if sA[i].ServerHealthcheck.AdvancedHealthcheckParameters == nil || len(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters) == 0 {
// 		// 				sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters,
// 		// 					"advancedHealthcheckParameters",
// 		// 					"AdvancedHealthcheckParameters",
// 		// 					"At healthcheck type 'http-advanced-json' http advanced parameters must be set",
// 		// 					"")
// 		// 			}
// 		// 			for j, advancedHealthcheckParameters := range sA[i].ServerHealthcheck.AdvancedHealthcheckParameters {
// 		// 				if len(advancedHealthcheckParameters.UserDefinedData) == 0 {
// 		// 					sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters[j].UserDefinedData,
// 		// 						"userDefinedData",
// 		// 						"UserDefinedData",
// 		// 						"At healthcheck type 'http-advanced-json' at http advanced parameters user defined data must be set",
// 		// 						"")
// 		// 				}
// 		// 			}
// 		// 		default:
// 		// 			sl.ReportError(sA[i].ServerHealthcheck.TypeOfCheck,
// 		// 				"typeOfCheck",
// 		// 				"TypeOfCheck",
// 		// 				"Unsupported type of http advanced healthcheck",
// 		// 				"")
// 		// 		}
// 		// 	}
// 		// }
// 	default:
// 		sl.ReportError(sHc.Type, "type", "Type", "unsupported healthcheck type", "")
// 	}
// 	if sHc.Timeout >= sHc.RepeatHealthcheck {
// 		sl.ReportError(sHc.Timeout, "timeout", "Timeout", "timeout can't be more than repeat healthcheck", "")
// 	}
// }

// Only for modify go-playground/validator!
func modifyValidateError(validateError error) error {
	var errorsString string
	for _, err := range validateError.(validator.ValidationErrors) {
		errorsString += fmt.Sprintf("at data %v got %v, can't validate in for rule %v %v;\n",
			err.Field(),
			err.Value(),
			err.ActualTag(),
			err.Param())
	}
	return fmt.Errorf("validate fail: %v", errorsString)
}

// func transformSliceToString(slice []string) string {
// 	var resultString string
// 	for _, el := range slice {
// 		resultString += "\n" + el
// 	}
// 	return resultString
// }

func validateServiceBalanceType(balanceType string) error {
	switch balanceType { // maybe range by array is better?
	case "rr":
	case "wrr":
	case "lc":
	case "wlc":
	case "lblc":
	case "sh":
	case "mhf":
	case "mhp":
	case "dh":
	case "fo":
	case "ovf":
	case "lblcr":
	case "sed":
	case "nq":
	default:
		return fmt.Errorf("unknown balance type for service: %v; supported types: rr|wrr|lc|wlc|lblc|sh|mh|dh|fo|ovf|lblcr|sed|nq0", balanceType)
	}
	return nil
}

func validateServiceRoutingType(routingType string) error {
	switch routingType { // maybe range by array is better?
	case "masquerading":
	case "tunneling":
	default:
		return fmt.Errorf("unknown routing type: %v; supported types: masquerading|tunneling", routingType)
	}
	return nil
}

func validateServiceProtocol(protocol string) error {
	switch protocol { // maybe range by array is better?
	case "tcp":
	case "udp":
	default:
		return fmt.Errorf("unknown protocol for service: %v; supported types: tcp|udp", protocol)
	}
	return nil
}
