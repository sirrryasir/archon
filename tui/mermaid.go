package tui

import (
	"fmt"
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
	From  string
	To    string
	Label string
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
	nodeSubgraphs := make(map[string]string)
	var nodeIDsOrdered []string
	var subgraphs []MermaidSubgraph
	var links []MermaidLink

	currentSubgraph := -1

	// Regex to extract all node declarations: ID["Label"], ID("Label"), ID[Label], ID{Label}, or ID(Label)
	nodeDefReg := regexp.MustCompile(`([a-zA-Z0-9_-]+)[\[\({]+"(.*?)"[\]\)}]+|([a-zA-Z0-9_-]+)[\[\({]+(.*?)[\]\)}]+`)
	// Regex for link matching (assumes clean node IDs after cleaning labels)
	linkReg := regexp.MustCompile(`([a-zA-Z0-9_-]+)\s*(?:-->\|.*?\||--.*?-->|-->.*?|--.*?---|-.+?\.->|-\.-\>|==.*?==>|==>)\s*([a-zA-Z0-9_-]+)`)
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
			// Strip FontAwesome icon tags like fa:fa-folder, fab:fa-github, etc.
			faReg := regexp.MustCompile(`(?i)fa[a-z]?:fa-[a-zA-Z0-9_-]+\s*`)
			label = faReg.ReplaceAllString(label, "")

			label = strings.ReplaceAll(label, "<br/>", " / ")
			label = strings.ReplaceAll(label, "<br>", " / ")
			label = strings.ReplaceAll(label, "<br >", " / ")

			if id != "" {
				hasNodeDef = true
				if _, exists := nodes[id]; !exists {
					nodeIDsOrdered = append(nodeIDsOrdered, id)
				}
				nodes[id] = MermaidNode{ID: id, Label: label}
				if currentSubgraph >= 0 {
					nodeSubgraphs[id] = subgraphs[currentSubgraph].Title
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
				to := strings.TrimSpace(match[len(match)-1])
				if from != "" && to != "" {
					hasLink = true
					label := ""
					for _, m := range match[2 : len(match)-1] {
						if m != "" {
							label = strings.TrimSpace(m)
							break
						}
					}
					links = append(links, MermaidLink{From: from, To: to, Label: label})
					// Add implicit nodes if they weren't defined with labels
					if _, exists := nodes[from]; !exists {
						nodes[from] = MermaidNode{ID: from, Label: from}
						nodeIDsOrdered = append(nodeIDsOrdered, from)
					}
					if _, exists := nodes[to]; !exists {
						nodes[to] = MermaidNode{ID: to, Label: to}
						nodeIDsOrdered = append(nodeIDsOrdered, to)
					}
					if currentSubgraph >= 0 {
						nodeSubgraphs[from] = subgraphs[currentSubgraph].Title
						nodeSubgraphs[to] = subgraphs[currentSubgraph].Title
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
					nodeIDsOrdered = append(nodeIDsOrdered, id)
				}
				nodeSubgraphs[id] = subgraphs[currentSubgraph].Title
				if !contains(subgraphs[currentSubgraph].Nodes, id) {
					subgraphs[currentSubgraph].Nodes = append(subgraphs[currentSubgraph].Nodes, id)
				}
			}
		}
	}

	// Always render as branching tree or flow for maximum TUI clarity, readability, and DX
	return renderBranchingTreeOrFlow(nodes, nodeSubgraphs, links, nodeIDsOrdered, termWidth)
}

func renderBranchingTreeOrFlow(nodes map[string]MermaidNode, nodeSubgraphs map[string]string, links []MermaidLink, nodeIDsOrdered []string, termWidth int) string {
	if len(nodes) == 0 {
		return ""
	}

	// Build adjacency list, link labels and track incoming links count
	adj := make(map[string][]string)
	incomingCount := make(map[string]int)
	linkLabels := make(map[string]string)
	for _, link := range links {
		adj[link.From] = append(adj[link.From], link.To)
		incomingCount[link.To]++
		if link.Label != "" {
			linkLabels[link.From+"->"+link.To] = link.Label
		}
	}

	// Find root nodes in stable sequence
	var roots []string
	for _, id := range nodeIDsOrdered {
		if _, ok := nodes[id]; ok && incomingCount[id] == 0 {
			roots = append(roots, id)
		}
	}

	// If there are links but no roots (due to a cycle/loop), pick the first defined node as the root
	if len(links) > 0 && len(roots) == 0 {
		for _, id := range nodeIDsOrdered {
			if _, ok := nodes[id]; ok {
				roots = []string{id}
				break
			}
		}
	}

	// If there are still no roots or links, fall back to flat ordered print
	if len(links) == 0 || len(roots) == 0 {
		return renderFlatFlow(nodes, nodeIDsOrdered)
	}

	// Render roots recursively as a hierarchy
	var blocks []string
	visited := make(map[string]bool)
	for _, root := range roots {
		blocks = append(blocks, renderTree(root, "", "", nodes, nodeSubgraphs, adj, linkLabels, "", true, visited))
	}

	// Print any orphaned nodes that were skipped (e.g. cycles or disjoint subgraphs)
	var orphans []string
	for _, id := range nodeIDsOrdered {
		if _, ok := nodes[id]; ok && !visited[id] {
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
			label := nodes[id].Label
			if sub, hasSub := nodeSubgraphs[id]; hasSub && sub != "" {
				label = fmt.Sprintf("[%s] %s", sub, label)
			}
			orphanBlocks = append(orphanBlocks, boxStyle.Render(label))
		}
		blocks = append(blocks, "\nOther Components:\n"+lipgloss.JoinHorizontal(lipgloss.Center, orphanBlocks...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

// getPathAndBranches extracts the linear path starting at u and returns the path node IDs and any branching children at the end of the path.
func getPathAndBranches(u string, adj map[string][]string, visited map[string]bool) ([]string, []string) {
	path := []string{u}
	curr := u
	for {
		children := adj[curr]
		var unvisitedChildren []string
		for _, child := range children {
			if !visited[child] {
				unvisitedChildren = append(unvisitedChildren, child)
			}
		}

		if len(unvisitedChildren) == 1 {
			child := unvisitedChildren[0]
			path = append(path, child)
			visited[child] = true
			curr = child
		} else {
			return path, unvisitedChildren
		}
	}
}

func renderTree(rootID string, parentID string, parentSubgraph string, nodes map[string]MermaidNode, nodeSubgraphs map[string]string, adj map[string][]string, linkLabels map[string]string, prefix string, isLast bool, visited map[string]bool) string {
	visited[rootID] = true
	pathIDs, children := getPathAndBranches(rootID, adj, visited)

	var pathLabels []string
	for i, id := range pathIDs {
		node := nodes[id]
		label := node.Label
		if label == "" {
			label = id
		}
		if i > 0 {
			edgeLabel := linkLabels[pathIDs[i-1]+"->"+id]
			if edgeLabel != "" {
				label = fmt.Sprintf("[%s] %s", edgeLabel, label)
			}
		} else if parentID != "" {
			edgeLabel := linkLabels[parentID+"->"+id]
			if edgeLabel != "" {
				label = fmt.Sprintf("[%s] %s", edgeLabel, label)
			}
		}
		pathLabels = append(pathLabels, label)
	}

	subStyle := lipgloss.NewStyle().Foreground(archonCyan).Bold(true)
	nodeStyle := lipgloss.NewStyle().Foreground(archonWhite)
	rootStyle := lipgloss.NewStyle().Foreground(archonGreen).Bold(true)

	var pathStr string
	var subTitle string
	for _, id := range pathIDs {
		if sub, hasSub := nodeSubgraphs[id]; hasSub && sub != "" {
			subTitle = sub
			break
		}
	}

	joinedPath := strings.Join(pathLabels, " ➔ ")
	if subTitle != "" && subTitle != parentSubgraph {
		pathStr = subStyle.Render("["+subTitle+"] ") + nodeStyle.Render(joinedPath)
	} else {
		pathStr = nodeStyle.Render(joinedPath)
	}

	if prefix == "" {
		pathStr = rootStyle.Render("● ") + pathStr
	}

	var result []string
	if prefix == "" {
		result = append(result, pathStr)
	} else {
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		result = append(result, prefix+marker+pathStr)
	}

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

	// The parent for the children will be the last node in the prefix path
	lastPathID := pathIDs[len(pathIDs)-1]
	nextParentSubgraph := parentSubgraph
	if subTitle != "" {
		nextParentSubgraph = subTitle
	}
	for i, childID := range children {
		childIsLast := i == len(children)-1
		result = append(result, renderTree(childID, lastPathID, nextParentSubgraph, nodes, nodeSubgraphs, adj, linkLabels, nextPrefix, childIsLast, visited))
	}

	return strings.Join(result, "\n")
}

func renderFlatFlow(nodes map[string]MermaidNode, nodeIDsOrdered []string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(archonCyan).
		Padding(0, 2).
		MarginBottom(1)

	var blocks []string
	for _, id := range nodeIDsOrdered {
		if node, ok := nodes[id]; ok {
			blocks = append(blocks, boxStyle.Render(node.Label))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func renderSubgraphLayout(nodes map[string]MermaidNode, nodeSubgraphs map[string]string, subgraphs []MermaidSubgraph, links []MermaidLink, nodeIDsOrdered []string, termWidth int) string {
	subgraphStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(archonGray).
		Padding(1, 2).
		MarginBottom(1)

	nodeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(archonCyan).
		Padding(0, 1).
		Margin(0, 1)

	var blocks []string

	// 1. Identify global nodes (not in any subgraph) and render them at the top
	globalNodes := make(map[string]MermaidNode)
	var globalLinks []MermaidLink
	for _, id := range nodeIDsOrdered {
		if node, ok := nodes[id]; ok {
			if nodeSubgraphs[id] == "" {
				globalNodes[id] = node
			}
		}
	}
	for _, link := range links {
		if nodeSubgraphs[link.From] == "" && nodeSubgraphs[link.To] == "" {
			globalLinks = append(globalLinks, link)
		}
	}

	if len(globalNodes) > 0 {
		globalFlow := renderBranchingTreeOrFlow(globalNodes, nodeSubgraphs, globalLinks, nodeIDsOrdered, termWidth)
		flowTitle := lipgloss.NewStyle().Foreground(archonBlue).Bold(true).Render("Main Execution Flow:")
		blocks = append(blocks, flowTitle, globalFlow, "")
	}

	// 2. Render each subgraph side-by-side wrapped in rows
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

		nodesContent := lipgloss.JoinVertical(lipgloss.Left, rows...)
		titleStyle := lipgloss.NewStyle().Foreground(archonCyan).Bold(true).Render(sub.Title)
		boxContent := subgraphStyle.Render(nodesContent)
		combined := lipgloss.JoinVertical(lipgloss.Left, titleStyle, boxContent)

		blocks = append(blocks, combined)
	}

	// 3. Render cross-module connections
	var crossLinks []string
	for _, link := range links {
		subFrom := nodeSubgraphs[link.From]
		subTo := nodeSubgraphs[link.To]
		if subFrom != subTo {
			fromLabel := nodes[link.From].Label
			toLabel := nodes[link.To].Label
			if subFrom != "" {
				fromLabel = fmt.Sprintf("[%s] %s", subFrom, fromLabel)
			}
			if subTo != "" {
				toLabel = fmt.Sprintf("[%s] %s", subTo, toLabel)
			}
			crossLinks = append(crossLinks, fmt.Sprintf("  • %s  ➔  %s", fromLabel, toLabel))
		}
	}

	result := lipgloss.JoinVertical(lipgloss.Left, blocks...)

	if len(crossLinks) > 0 {
		crossTitle := lipgloss.NewStyle().Foreground(archonBlue).Bold(true).Render("\nCross-Module Connections:")
		crossContent := lipgloss.NewStyle().Foreground(archonWhite).Render(strings.Join(crossLinks, "\n"))
		result = lipgloss.JoinVertical(lipgloss.Left, result, crossTitle, crossContent)
	}

	return result
}

func renderSubgraphTree(rootID string, parentID string, nodes map[string]MermaidNode, adj map[string][]string, linkLabels map[string]string, prefix string, isLast bool, visited map[string]bool) string {
	visited[rootID] = true
	pathIDs, children := getPathAndBranches(rootID, adj, visited)

	var pathLabels []string
	for i, id := range pathIDs {
		node := nodes[id]
		label := node.Label
		if label == "" {
			label = id
		}
		if i > 0 {
			edgeLabel := linkLabels[pathIDs[i-1]+"->"+id]
			if edgeLabel != "" {
				label = fmt.Sprintf("[%s] %s", edgeLabel, label)
			}
		} else if parentID != "" {
			edgeLabel := linkLabels[parentID+"->"+id]
			if edgeLabel != "" {
				label = fmt.Sprintf("[%s] %s", edgeLabel, label)
			}
		}
		pathLabels = append(pathLabels, label)
	}

	nodeStyle := lipgloss.NewStyle().Foreground(archonWhite)
	rootStyle := lipgloss.NewStyle().Foreground(archonGreen).Bold(true)

	joinedPath := strings.Join(pathLabels, " ➔ ")
	pathStr := nodeStyle.Render(joinedPath)
	if prefix == "" {
		pathStr = rootStyle.Render("● ") + pathStr
	}

	var result []string
	if prefix == "" {
		result = append(result, pathStr)
	} else {
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		result = append(result, prefix+marker+pathStr)
	}

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

	lastPathID := pathIDs[len(pathIDs)-1]
	for i, childID := range children {
		childIsLast := i == len(children)-1
		result = append(result, renderSubgraphTree(childID, lastPathID, nodes, adj, linkLabels, nextPrefix, childIsLast, visited))
	}

	return strings.Join(result, "\n")
}

func childIsLast(i int, slice []string) bool {
	return i == len(slice)-1
}
