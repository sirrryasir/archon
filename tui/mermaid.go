package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MermaidNode represents a node in the diagram
type MermaidNode struct {
	ID    string
	Label string
}

// MermaidSubgraph represents a grouped subgraph box in the diagram
type MermaidSubgraph struct {
	Title string
	Nodes []string
}

// MermaidLink represents a directional connection
type MermaidLink struct {
	From string
	To   string
}

// contains checks if a slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RenderMermaidInTerminal parses a Mermaid block and renders it using Unicode box characters and Lipgloss.
func RenderMermaidInTerminal(mermaidCode string, termWidth int) string {
	lines := strings.Split(mermaidCode, "\n")

	nodes := make(map[string]MermaidNode)
	var subgraphs []MermaidSubgraph
	var links []MermaidLink

	currentSubgraph := -1

	// Regex to extract all node declarations: ID["Label"], ID("Label"), ID[Label], or ID(Label)
	nodeDefReg := regexp.MustCompile(`([a-zA-Z0-9_-]+)[\[\(]+"(.*?)"[\]\)]+|([a-zA-Z0-9_-]+)[\[\(]+(.*?)[\]\)]+`)
	// Regex for link matching (assumes clean node IDs after cleaning labels)
	linkReg := regexp.MustCompile(`([a-zA-Z0-9_-]+)\s*-->\s*(?:\|.*?\|)?\s*([a-zA-Z0-9_-]+)`)
	// Regex for subgraph matching
	subReg := regexp.MustCompile(`subgraph\s+(.*?)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "graph ") || strings.HasPrefix(line, "flowchart ") || strings.HasPrefix(line, "%%") {
			continue
		}

		if line == "end" {
			currentSubgraph = -1
			continue
		}

		// Handle subgraph title
		if subMatches := subReg.FindStringSubmatch(line); len(subMatches) > 1 {
			title := strings.Trim(strings.TrimSpace(subMatches[1]), `"'`)
			subgraphs = append(subgraphs, MermaidSubgraph{Title: title, Nodes: []string{}})
			currentSubgraph = len(subgraphs) - 1
			continue
		}

		// 1. Extract ALL node definitions on this line first
		nodeMatches := nodeDefReg.FindAllStringSubmatch(line, -1)
		hasNodeDef := false
		for _, match := range nodeMatches {
			var id, label string
			if match[1] != "" {
				id = match[1]
				label = match[2]
			} else {
				id = match[3]
				label = match[4]
			}
			id = strings.TrimSpace(id)
			label = strings.Trim(strings.TrimSpace(label), `"'`)

			if id != "" {
				hasNodeDef = true
				nodes[id] = MermaidNode{ID: id, Label: label}
				if currentSubgraph >= 0 {
					if !contains(subgraphs[currentSubgraph].Nodes, id) {
						subgraphs[currentSubgraph].Nodes = append(subgraphs[currentSubgraph].Nodes, id)
					}
				}
			}
		}

		// 2. Clean the line by replacing label blocks to match links cleanly
		// Example: "A[Label] --> B[Label]" becomes "A --> B"
		cleanLine := nodeDefReg.ReplaceAllString(line, "$1$3")

		// 3. Parse links on the cleaned line
		linkMatches := linkReg.FindAllStringSubmatch(cleanLine, -1)
		hasLink := false
		for _, match := range linkMatches {
			if len(match) > 2 {
				from := strings.TrimSpace(match[1])
				to := strings.TrimSpace(match[2])
				if from != "" && to != "" {
					hasLink = true
					links = append(links, MermaidLink{From: from, To: to})
					// Add implicit nodes if they weren't defined with labels
					if _, exists := nodes[from]; !exists {
						nodes[from] = MermaidNode{ID: from, Label: from}
					}
					if _, exists := nodes[to]; !exists {
						nodes[to] = MermaidNode{ID: to, Label: to}
					}
					if currentSubgraph >= 0 {
						if !contains(subgraphs[currentSubgraph].Nodes, from) {
							subgraphs[currentSubgraph].Nodes = append(subgraphs[currentSubgraph].Nodes, from)
						}
						if !contains(subgraphs[currentSubgraph].Nodes, to) {
							subgraphs[currentSubgraph].Nodes = append(subgraphs[currentSubgraph].Nodes, to)
						}
					}
				}
			}
		}

		// 4. If no node definition and no link, check if it's a plain node ID
		if !hasNodeDef && !hasLink && currentSubgraph >= 0 {
			plainIDReg := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
			if plainIDReg.MatchString(line) {
				id := line
				if _, exists := nodes[id]; !exists {
					nodes[id] = MermaidNode{ID: id, Label: id}
				}
				if !contains(subgraphs[currentSubgraph].Nodes, id) {
					subgraphs[currentSubgraph].Nodes = append(subgraphs[currentSubgraph].Nodes, id)
				}
			}
		}
	}

	// If no subgraphs parsed, check if it forms a branching tree hierarchy
	if len(subgraphs) == 0 {
		return renderBranchingTreeOrFlow(nodes, links, termWidth)
	}

	return renderSubgraphLayout(nodes, subgraphs, links, termWidth)
}

func renderBranchingTreeOrFlow(nodes map[string]MermaidNode, links []MermaidLink, termWidth int) string {
	if len(nodes) == 0 {
		return ""
	}

	// Build adjacency list and track incoming links count
	adj := make(map[string][]string)
	incomingCount := make(map[string]int)
	for _, link := range links {
		adj[link.From] = append(adj[link.From], link.To)
		incomingCount[link.To]++
	}

	// Find root nodes
	var roots []string
	for id := range nodes {
		if incomingCount[id] == 0 {
			roots = append(roots, id)
		}
	}

	// If there are no roots or links, fall back to flat ordered print
	if len(links) == 0 || len(roots) == 0 {
		return renderFlatFlow(nodes)
	}

	// Render roots recursively as a hierarchy
	var blocks []string
	visited := make(map[string]bool)
	for _, root := range roots {
		blocks = append(blocks, renderTree(root, nodes, adj, "", true, visited))
	}

	// Print any orphaned nodes that were skipped (e.g. cycles or disjoint subgraphs)
	var orphans []string
	for id := range nodes {
		if !visited[id] {
			orphans = append(orphans, id)
		}
	}
	if len(orphans) > 0 {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(archonGray).
			Padding(0, 1)
		var orphanBlocks []string
		for _, id := range orphans {
			orphanBlocks = append(orphanBlocks, boxStyle.Render(nodes[id].Label))
		}
		blocks = append(blocks, "\n📦 Other Components:\n"+lipgloss.JoinHorizontal(lipgloss.Center, orphanBlocks...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func renderTree(rootID string, nodes map[string]MermaidNode, adj map[string][]string, prefix string, isLast bool, visited map[string]bool) string {
	visited[rootID] = true
	node := nodes[rootID]
	label := node.Label
	if label == "" {
		label = rootID
	}

	// Stylize label
	var styledLabel string
	if prefix == "" {
		// Root node style: Bold Cyan
		styledLabel = lipgloss.NewStyle().Foreground(archonCyan).Bold(true).Render("● " + label)
	} else {
		// Leaf/Branch node style: Standard White
		styledLabel = lipgloss.NewStyle().Foreground(archonWhite).Render(label)
	}

	var result []string
	if prefix == "" {
		result = append(result, styledLabel)
	} else {
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		result = append(result, prefix+marker+styledLabel)
	}

	children := adj[rootID]
	nextPrefix := prefix
	if prefix != "" {
		if isLast {
			nextPrefix += "    "
		} else {
			nextPrefix += "│   "
		}
	} else {
		nextPrefix += "    "
	}

	// Process children recursively
	for i, childID := range children {
		if !visited[childID] {
			childIsLast := i == len(children)-1
			result = append(result, renderTree(childID, nodes, adj, nextPrefix, childIsLast, visited))
		}
	}

	return strings.Join(result, "\n")
}

func renderFlatFlow(nodes map[string]MermaidNode) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(archonCyan).
		Padding(0, 2).
		MarginBottom(1)

	var blocks []string
	for _, node := range nodes {
		blocks = append(blocks, boxStyle.Render(node.Label))
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func renderSubgraphLayout(nodes map[string]MermaidNode, subgraphs []MermaidSubgraph, links []MermaidLink, termWidth int) string {
	// Style definitions
	subgraphStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(archonGray).
		Padding(1, 2).
		MarginBottom(1)

	nodeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(archonWhite).
		Padding(0, 1).
		MarginRight(1)

	arrowStyle := lipgloss.NewStyle().
		Foreground(archonBlue).
		Render("\n   │\n   ▼\n")

	var subBlocks []string

	// Calculate a safe width to wrap nodes inside the subgraph box
	safeWidth := termWidth - 16
	if safeWidth < 40 {
		safeWidth = 40
	}

	for _, sub := range subgraphs {
		var rows []string
		var currentRow []string
		currentRowWidth := 0

		for _, nodeID := range sub.Nodes {
			if node, exists := nodes[nodeID]; exists {
				renderedNode := nodeStyle.Render(node.Label)
				nodeW := lipgloss.Width(renderedNode)

				// If adding this node to the row exceeds safeWidth, wrap to a new row
				if currentRowWidth > 0 && currentRowWidth+nodeW > safeWidth {
					rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center, currentRow...))
					currentRow = []string{renderedNode}
					currentRowWidth = nodeW
				} else {
					currentRow = append(currentRow, renderedNode)
					currentRowWidth += nodeW
				}
			}
		}

		if len(currentRow) > 0 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center, currentRow...))
		}

		// Arrange rows inside subgraph vertically
		var nodesContent string
		if len(rows) > 0 {
			nodesContent = lipgloss.JoinVertical(lipgloss.Left, rows...)
		}

		// Title styled above the box
		titleStyle := lipgloss.NewStyle().Foreground(archonCyan).Bold(true).Render("📁 " + sub.Title)
		boxContent := subgraphStyle.Render(nodesContent)
		combined := lipgloss.JoinVertical(lipgloss.Left, titleStyle, boxContent)

		subBlocks = append(subBlocks, combined)
	}

	// Join all subgraph blocks vertically with arrows
	var layout []string
	for i, block := range subBlocks {
		layout = append(layout, block)
		if i < len(subBlocks)-1 {
			layout = append(layout, arrowStyle)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, layout...)
}
