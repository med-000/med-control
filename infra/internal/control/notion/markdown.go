package notion

import "strings"

func blocksToMarkdown(blocks []map[string]any, depth int) string {
	var lines []string
	for index, block := range blocks {
		lines = append(lines, blockToMarkdown(block, depth, index+1)...)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func blockToMarkdown(block map[string]any, depth int, number int) []string {
	blockType := stringValue(block["type"])
	content := mapValue(block[blockType])
	indent := strings.Repeat("  ", depth)

	var lines []string
	switch blockType {
	case "paragraph":
		lines = appendIfNotEmpty(lines, indent+richTextPlainText(content["rich_text"]))
	case "heading_1":
		lines = appendIfNotEmpty(lines, "# "+richTextPlainText(content["rich_text"]))
	case "heading_2":
		lines = appendIfNotEmpty(lines, "## "+richTextPlainText(content["rich_text"]))
	case "heading_3":
		lines = appendIfNotEmpty(lines, "### "+richTextPlainText(content["rich_text"]))
	case "heading_4":
		lines = appendIfNotEmpty(lines, "#### "+richTextPlainText(content["rich_text"]))
	case "bulleted_list_item":
		lines = appendIfNotEmpty(lines, indent+"- "+richTextPlainText(content["rich_text"]))
	case "numbered_list_item":
		lines = appendIfNotEmpty(lines, indent+stringFromInt(number)+". "+richTextPlainText(content["rich_text"]))
	case "to_do":
		checked, _ := content["checked"].(bool)
		mark := " "
		if checked {
			mark = "x"
		}
		lines = appendIfNotEmpty(lines, indent+"- ["+mark+"] "+richTextPlainText(content["rich_text"]))
	case "quote":
		lines = appendIfNotEmpty(lines, indent+"> "+richTextPlainText(content["rich_text"]))
	case "code":
		language := stringValue(content["language"])
		lines = append(lines, "```"+language, richTextPlainText(content["rich_text"]), "```")
	case "divider":
		lines = append(lines, "---")
	case "callout":
		lines = appendIfNotEmpty(lines, indent+"> "+richTextPlainText(content["rich_text"]))
	case "toggle":
		lines = appendIfNotEmpty(lines, indent+"- "+richTextPlainText(content["rich_text"]))
	case "child_page":
		lines = appendIfNotEmpty(lines, indent+"## "+stringValue(content["title"]))
	case "bookmark", "embed", "link_preview":
		lines = appendIfNotEmpty(lines, indent+stringValue(content["url"]))
	case "image", "video", "file", "pdf", "audio":
		lines = appendIfNotEmpty(lines, indent+fileURL(content))
	default:
		lines = appendIfNotEmpty(lines, indent+richTextPlainText(content["rich_text"]))
	}

	children, _ := block["children"].([]map[string]any)
	if len(children) > 0 {
		childMarkdown := blocksToMarkdown(children, depth+1)
		if childMarkdown != "" {
			lines = append(lines, childMarkdown)
		}
	}

	return lines
}

func fileURL(content map[string]any) string {
	for _, key := range []string{"external", "file"} {
		value := mapValue(content[key])
		if value == nil {
			continue
		}
		if url := stringValue(value["url"]); url != "" {
			return url
		}
	}
	return ""
}

func appendIfNotEmpty(lines []string, line string) []string {
	if strings.TrimSpace(line) == "" {
		return lines
	}
	return append(lines, line)
}

func stringFromInt(value int) string {
	digits := "0123456789"
	if value == 0 {
		return "0"
	}

	var result string
	for value > 0 {
		result = string(digits[value%10]) + result
		value /= 10
	}
	return result
}
