package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/wI2L/jsondiff"
)

const (
	CREATE string = "create"
	UPDATE string = "update"
	DELETE string = "delete"
	TAG    string = "tag"
	NOOP   string = "no-op"
)

var (
	resourcesList = make(map[string][]string)
)

func Parser(output []byte, showTags, showUnchanged, compact, useMarkdown bool) {
	var data tfjson.Plan
	if err := json.Unmarshal(output, &data); err != nil {
		fmt.Printf("Error unmarshalling plan: %v\n", err)
		return
	}

	for _, resourceChange := range data.ResourceChanges {
		processResourceChange(resourceChange, showTags)
	}

	PrintPlanSummary(showTags, showUnchanged, compact, useMarkdown)
}

func processResourceChange(resourceChange *tfjson.ResourceChange, showTags bool) {
	isUpdate := contains(resourceChange.Change.Actions, tfjson.ActionUpdate)

	if isUpdate {
		// Check if it's only a tag change
		isOnlyTagChange, err := checkOnlyTagChanges(resourceChange)
		if err != nil {
			fmt.Printf("Error checking for only tag changes: %v\n", err)
			return
		}
		if isOnlyTagChange && showTags {
			// Only add to TAG if it's only a tag change
			resourcesList[TAG] = append(resourcesList[TAG], resourceChange.Address)
			return
		}

		isTagChange := hasTagChanges(resourceChange)
		if err != nil {
			fmt.Printf("Error checking for tag changes: %v\n", err)
			return
		}

		if isTagChange && showTags {
			resourcesList[TAG] = append(resourcesList[TAG], resourceChange.Address)
		}

		detailedChanges := processDetailedChanges(resourceChange)
		if detailedChanges != "" {
			addActionToResourceListWithDetails(resourceChange.Change.Actions, resourceChange.Address, detailedChanges)
		} else {
			addActionToResourceList(resourceChange.Change.Actions, resourceChange.Address)
		}
		return
	} else {
		// Not an update, add to other categories as necessary
		addActionToResourceList(resourceChange.Change.Actions, resourceChange.Address)
	}
}

func hasTagChanges(resourceChange *tfjson.ResourceChange) bool {
	beforeRaw, err := json.Marshal(resourceChange.Change.Before)
	if err != nil {
		fmt.Printf("Error marshaling before state: %v\n", err)
		return false
	}
	afterRaw, err := json.Marshal(resourceChange.Change.After)
	if err != nil {
		fmt.Printf("Error marshaling after state: %v\n", err)
		return false
	}

	var beforeMap, afterMap map[string]interface{}
	if err := json.Unmarshal(beforeRaw, &beforeMap); err != nil {
		fmt.Printf("Error unmarshaling before state: %v\n", err)
		return false
	}
	if err := json.Unmarshal(afterRaw, &afterMap); err != nil {
		fmt.Printf("Error unmarshaling after state: %v\n", err)
		return false
	}

	if equal(beforeMap["tags"], afterMap["tags"]) && equal(beforeMap["tags_all"], afterMap["tags_all"]) {
		return false
	}

	return true
}

func addActionToResourceList(actions []tfjson.Action, address string) {
	for _, action := range actions {
		resourcesList[string(action)] = append(resourcesList[string(action)], address)
	}
}

func addActionToResourceListWithDetails(actions []tfjson.Action, address string, details string) {
	for _, action := range actions {
		resourceDetail := fmt.Sprintf("%s [%s]", address, details)
		resourcesList[string(action)] = append(resourcesList[string(action)], resourceDetail)
	}
}

func PrintResources(message string, resources []string, bulletSymbol string, color *color.Color, compact bool, useMarkdown bool) {
	if len(resources) != 0 {
		if useMarkdown {
			fmt.Printf("## %s\n\n", message) // Markdown header for the message
			for _, resource := range resources {
				var emoji string
				switch bulletSymbol {
				case "+":
					emoji = "✅" // Green check mark for create
				case "~":
					emoji = "⚠️" // Yellow warning sign for update
				case "-":
					emoji = "🧨" // Red circle for destroy
				case "#":
					emoji = "#️⃣" // Blue diamond for tag/untag
				case "•":
					emoji = "🔷" // Blue circle for unchanged
				default:
					emoji = "➡️" // Default arrow
				}
				fmt.Printf("* %s %s\n", emoji, resource)
			}
		} else {
			fmt.Println(message)
			for _, resource := range resources {
				color.Printf("  %s ", bulletSymbol)
				fmt.Println(resource)
			}
		}
		if !compact {
			fmt.Println()
		}
	}
}

func PrintPlanSummary(showTags, showUnchanged, compact, useMarkdown bool) {
	if showUnchanged {
		PrintResources("🔵 Unchanged:", resourcesList[NOOP], "•", color.New(color.FgBlue), compact, useMarkdown)
	}
	if showTags {
		PrintResources("🟣 Tag/Untag:", resourcesList[TAG], "#", color.New(color.FgMagenta), compact, useMarkdown)
	}
	PrintResources("🟢 Create:", resourcesList[CREATE], "+", color.New(color.FgGreen), compact, useMarkdown)
	PrintResources("🟡 Update:", resourcesList[UPDATE], "~", color.New(color.FgYellow), compact, useMarkdown)
	PrintResources("🔴 Destroy:", resourcesList[DELETE], "-", color.New(color.FgRed), compact, useMarkdown)
}

func processDetailedChanges(resourceChange *tfjson.ResourceChange) string {
	beforeRaw, _ := json.Marshal(resourceChange.Change.Before)
	afterRaw, _ := json.Marshal(resourceChange.Change.After)
	beforeStr, afterStr := string(beforeRaw), string(afterRaw)

	patch, err := generateJSONDiff(beforeStr, afterStr)
	if err != nil {
		// handle error
		return ""
	}

	// Use the custom formatter
	return formatPatch(patch)
}

func marshalIndentJSON(obj interface{}) (string, error) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func diffLines(a, b string) string {
	var result strings.Builder

	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	maxLen := len(aLines)
	if len(bLines) > maxLen {
		maxLen = len(bLines)
	}

	for i := 0; i < maxLen; i++ {
		var aLine, bLine string
		if i < len(aLines) {
			aLine = aLines[i]
		}
		if i < len(bLines) {
			bLine = bLines[i]
		}

		if aLine != bLine {
			result.WriteString(fmt.Sprintf("- %s\n+ %s\n", aLine, bLine))
		}
	}

	return result.String()
}

// func processDetailedChanges(resourceChange *tfjson.ResourceChange) string {
// 	var details []string

// 	// Type assert to map[string]interface{}
// 	before, okBefore := resourceChange.Change.Before.(map[string]interface{})
// 	after, okAfter := resourceChange.Change.After.(map[string]interface{})
// 	if !okBefore || !okAfter {
// 		return ""
// 	}

// 	for key := range after {
// 		beforeRaw, errBefore := json.Marshal(before[key])
// 		afterRaw, errAfter := json.Marshal(after[key])
// 		if errBefore != nil || errAfter != nil {
// 			continue // Skip this key if there is an error in marshaling
// 		}

// 		beforeStr, afterStr := string(beforeRaw), string(afterRaw)

// 		if beforeStr != afterStr {
// 			diff := generateJSONDiff(beforeStr, afterStr)
// 			if diff != "" {
// 				details = append(details, fmt.Sprintf("~%s:\n%s", key, diff))
// 			}
// 		}
// 	}

// 	return strings.Join(details, "\n")
// }

func formatPatch(patch []jsondiff.Operation) string {
	var details strings.Builder

	for _, op := range patch {
		switch op.Type {
		case jsondiff.OperationAdd:
			details.WriteString(fmt.Sprintf("\n    + %s: %v\n", op.Path, op.Value))
		case jsondiff.OperationRemove:
			details.WriteString(fmt.Sprintf("\n    - %s\n", op.Path))
		case jsondiff.OperationReplace:
			details.WriteString(fmt.Sprintf("\n    ~ %s: %v -> %v\n", op.Path, op.OldValue, op.Value))
			// Handle other cases like 'move' or 'copy' if necessary
		}
	}

	return details.String()
}

func generateJSONDiff(before, after string) ([]jsondiff.Operation, error) {
	var beforeObj, afterObj interface{}
	json.Unmarshal([]byte(before), &beforeObj)
	json.Unmarshal([]byte(after), &afterObj)

	patch, err := jsondiff.Compare(beforeObj, afterObj)
	if err != nil {
		return nil, err
	}
	return patch, nil
}

func checkOnlyTagChanges(resourceChange *tfjson.ResourceChange) (bool, error) {
	beforeRaw, err := json.Marshal(resourceChange.Change.Before)
	if err != nil {
		return false, fmt.Errorf("failed to marshal before state: %v", err)
	}
	afterRaw, err := json.Marshal(resourceChange.Change.After)
	if err != nil {
		return false, fmt.Errorf("failed to marshal after state: %v", err)
	}

	var beforeMap, afterMap map[string]interface{}
	if err := json.Unmarshal(beforeRaw, &beforeMap); err != nil {
		return false, fmt.Errorf("failed to unmarshal before state: %v", err)
	}
	if err := json.Unmarshal(afterRaw, &afterMap); err != nil {
		return false, fmt.Errorf("failed to unmarshal after state: %v", err)
	}

	if equal(beforeMap, afterMap) {
		return false, nil
	}

	for key := range beforeMap {
		if key != "tags" && key != "tags_all" {
			if vAfter, exists := afterMap[key]; exists {
				if !equal(beforeMap[key], vAfter) {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
	}

	for key := range afterMap {
		if key != "tags" && key != "tags_all" {
			if vBefore, exists := beforeMap[key]; exists {
				if !equal(vBefore, afterMap[key]) {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
	}

	return true, nil
}

func equal(a, b interface{}) bool {
	aJson, _ := json.Marshal(a)
	bJson, _ := json.Marshal(b)
	return string(aJson) == string(bJson)
}

func contains(slice []tfjson.Action, val tfjson.Action) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
