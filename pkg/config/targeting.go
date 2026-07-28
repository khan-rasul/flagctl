package config

import (
	"fmt"
)

type RuleType string

const (
	RuleTypeDenylist  RuleType = "DENYLIST"
	RuleTypeAllowlist RuleType = "ALLOWLIST"
	RuleTypeSegment   RuleType = "SEGMENT"
	RuleTypeLaunch    RuleType = "LAUNCH"
)

type RuleItem struct {
	Index     int         `json:"index"`
	Type      RuleType    `json:"type"`
	Attribute string      `json:"attribute,omitempty"`
	Operator  string      `json:"operator,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Variant   string      `json:"variant"`
	IsCohort  bool        `json:"isCohort"`
}

func (c *FlagConfig) AddTargetRule(flag *Flag, attr string, op string, val interface{}, variant string, isTop bool) error {
	cond := map[string]interface{}{
		op: []interface{}{
			map[string]interface{}{"var": attr},
			val,
		},
	}

	if flag.Targeting == nil {
		flag.Targeting = map[string]interface{}{
			"if": []interface{}{cond, variant, flag.DefaultVariant},
		}
		return nil
	}

	tMap, ok := flag.Targeting.(map[string]interface{})
	if !ok {
		flag.Targeting = map[string]interface{}{
			"if": []interface{}{cond, variant, flag.DefaultVariant},
		}
		return nil
	}

	ifList, ok := tMap["if"].([]interface{})
	if !ok {
		flag.Targeting = map[string]interface{}{
			"if": []interface{}{cond, variant, flag.DefaultVariant},
		}
		return nil
	}

	if isTop || variant == "off" || variant == "false" {
		// Insert at very beginning
		newIf := append([]interface{}{cond, variant}, ifList...)
		tMap["if"] = newIf
	} else {
		// Insert before final fallback element
		lastIdx := len(ifList) - 1
		newIf := append([]interface{}{}, ifList[:lastIdx]...)
		newIf = append(newIf, cond, variant, ifList[lastIdx])
		tMap["if"] = newIf
	}

	flag.Targeting = tMap
	return nil
}

func (c *FlagConfig) AddLaunchRamp(flag *Flag, attr string, val interface{}, splits map[string]int, bucketBy string) error {
	if bucketBy == "" {
		bucketBy = "userId"
	}

	var weights []interface{}
	for k, w := range splits {
		weights = append(weights, []interface{}{k, w})
	}

	fracRule := map[string]interface{}{
		"fractional": append([]interface{}{map[string]interface{}{"var": bucketBy}}, weights...),
	}

	if attr != "" {
		// Cohort launch rule
		return c.AddTargetRule(flag, attr, "==", val, fmt.Sprintf("%v", fracRule), false)
	}

	// Global Launch (replaces or sets fallback at the end of the if chain)
	if flag.Targeting == nil {
		flag.Targeting = map[string]interface{}{
			"if": []interface{}{fracRule, flag.DefaultVariant},
		}
		return nil
	}

	tMap, ok := flag.Targeting.(map[string]interface{})
	if !ok {
		flag.Targeting = map[string]interface{}{
			"if": []interface{}{fracRule, flag.DefaultVariant},
		}
		return nil
	}

	ifList, ok := tMap["if"].([]interface{})
	if !ok || len(ifList) == 0 {
		tMap["if"] = []interface{}{fracRule, flag.DefaultVariant}
	} else {
		// Set or replace fallback element
		ifList[len(ifList)-1] = fracRule
		tMap["if"] = ifList
	}

	flag.Targeting = tMap
	return nil
}

func (c *FlagConfig) GetTargetRules(flag *Flag) []RuleItem {
	if flag.Targeting == nil {
		return nil
	}

	tMap, ok := flag.Targeting.(map[string]interface{})
	if !ok {
		return nil
	}

	ifList, ok := tMap["if"].([]interface{})
	if !ok {
		return nil
	}

	var rules []RuleItem
	idx := 1

	for i := 0; i < len(ifList); i++ {
		elem := ifList[i]

		// Check if elem is a map condition or fractional
		if cMap, ok := elem.(map[string]interface{}); ok {
			if _, hasFrac := cMap["fractional"]; hasFrac {
				rules = append(rules, RuleItem{
					Index:   idx,
					Type:    RuleTypeLaunch,
					Variant: "fractional",
				})
				idx++
				continue
			}

			// Boolean condition -> next element in array is outcome
			if i+1 < len(ifList) {
				outcome := fmt.Sprintf("%v", ifList[i+1])
				item := RuleItem{
					Index:   idx,
					Variant: outcome,
				}

				for op, val := range cMap {
					item.Operator = op
					if args, ok := val.([]interface{}); ok && len(args) == 2 {
						if vMap, ok := args[0].(map[string]interface{}); ok {
							if attr, ok := vMap["var"].(string); ok {
								item.Attribute = attr
							}
						}
						item.Value = args[1]
					}
				}

				if item.Variant == "off" || item.Variant == "false" {
					item.Type = RuleTypeDenylist
				} else {
					item.Type = RuleTypeAllowlist
				}

				rules = append(rules, item)
				idx++
				i++ // Skip outcome item
			}
		}
	}

	return rules
}

func (c *FlagConfig) RemoveTargetRule(flag *Flag, index int) error {
	if flag.Targeting == nil {
		return fmt.Errorf("flag has no targeting rules")
	}

	tMap, ok := flag.Targeting.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid targeting structure")
	}

	ifList, ok := tMap["if"].([]interface{})
	if !ok || len(ifList) <= 1 {
		return fmt.Errorf("no rules to remove")
	}

	rules := c.GetTargetRules(flag)
	if index < 1 || index > len(rules) {
		return fmt.Errorf("rule index %d out of bounds (1 to %d)", index, len(rules))
	}

	// Single rule or fallback
	if len(rules) == 1 {
		flag.Targeting = nil
		return nil
	}

	// Remove corresponding elements from ifList
	ruleIndex := (index - 1) * 2
	if ruleIndex < len(ifList)-1 {
		newIf := append([]interface{}{}, ifList[:ruleIndex]...)
		newIf = append(newIf, ifList[ruleIndex+2:]...)
		if len(newIf) <= 1 {
			flag.Targeting = nil
		} else {
			tMap["if"] = newIf
			flag.Targeting = tMap
		}
	}

	return nil
}
