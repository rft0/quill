package quill

import "math"

// Compute runs the layout algorithm on the tree rooted at node,
// given the available container size. After this call, every node
// in the tree has its Layout field populated with absolute coordinates.
func Compute(node *Node, availWidth, availHeight float64) {
	// Resolve the root node's outer size.
	node.Layout.X = node.Style.Margin.Left
	node.Layout.Y = node.Style.Margin.Top
	node.Layout.Width = resolveSize(node.Style.Width, availWidth, availWidth) - node.Style.Margin.Horizontal()
	node.Layout.Height = resolveSize(node.Style.Height, availHeight, availHeight) - node.Style.Margin.Vertical()

	// If root width/height are auto, use all available space.
	if node.Style.Width.Unit == Auto {
		node.Layout.Width = availWidth - node.Style.Margin.Horizontal()
	}
	if node.Style.Height.Unit == Auto {
		node.Layout.Height = availHeight - node.Style.Margin.Vertical()
	}

	layoutNode(node)
}

// childInfo holds per-child layout information during flex computation.
type childInfo struct {
	node       *Node
	baseMain   float64
	baseCross  float64
	flexGrow   float64
	flexShrink float64
	frozen     bool
	mainSize   float64 // final main axis size
	crossSize  float64 // final cross axis size
	marginMain float64
}

// layoutNode recursively lays out a node's children using flexbox rules.
func layoutNode(node *Node) {
	if node.IsLeaf() {
		return
	}

	style := &node.Style
	bw := style.BorderWidth()
	innerW := node.Layout.Width - style.Padding.Horizontal() - bw*2
	innerH := node.Layout.Height - style.Padding.Vertical() - bw*2

	// Separate flow and absolute children.
	var flowChildren []*Node
	var absChildren []*Node
	for _, child := range node.Children {
		if child.Style.Position == PositionAbsolute {
			absChildren = append(absChildren, child)
		} else {
			flowChildren = append(flowChildren, child)
		}
	}

	// Position absolute children relative to parent's content box.
	for _, child := range absChildren {
		cs := &child.Style
		contentX := node.Layout.X + style.Padding.Left + bw
		contentY := node.Layout.Y + style.Padding.Top + bw

		// Resolve size first (needed for right/bottom positioning).
		if cs.Width.Unit != Auto {
			child.Layout.Width = resolveSize(cs.Width, innerW, innerW)
		} else {
			child.Layout.Width = innerW
		}
		if cs.Height.Unit != Auto {
			child.Layout.Height = resolveSize(cs.Height, innerH, innerH)
		} else {
			child.Layout.Height = innerH
		}

		// Horizontal: Left takes priority over Right.
		if cs.Left.Unit != Auto {
			child.Layout.X = contentX + resolveSize(cs.Left, innerW, 0)
		} else if cs.Right.Unit != Auto {
			child.Layout.X = contentX + innerW - child.Layout.Width - resolveSize(cs.Right, innerW, 0)
		} else {
			child.Layout.X = contentX
		}

		// Vertical: Top takes priority over Bottom.
		if cs.Top.Unit != Auto {
			child.Layout.Y = contentY + resolveSize(cs.Top, innerH, 0)
		} else if cs.Bottom.Unit != Auto {
			child.Layout.Y = contentY + innerH - child.Layout.Height - resolveSize(cs.Bottom, innerH, 0)
		} else {
			child.Layout.Y = contentY
		}
		layoutNode(child)
	}

	if len(flowChildren) == 0 {
		return
	}

	isRow := style.Direction == FlexRow

	var mainSize, crossSize float64
	if isRow {
		mainSize, crossSize = innerW, innerH
	} else {
		mainSize, crossSize = innerH, innerW
	}

	// --- Phase 1: Determine each child's base (hypothetical) main size ---
	infos := measureBaseChildren(flowChildren, isRow, mainSize, crossSize, style)

	// Delegate to wrap or no-wrap layout.
	if style.Wrap == FlexWrapWrap {
		layoutWrap(node, infos, isRow, mainSize, crossSize, style, bw)
	} else {
		layoutNoWrap(node, infos, isRow, mainSize, crossSize, style, bw)
	}
}

// measureBaseChildren computes the base (hypothetical) sizes for all flow children.
func measureBaseChildren(flowChildren []*Node, isRow bool, mainSize, crossSize float64, style *Style) []childInfo {
	infos := make([]childInfo, len(flowChildren))

	for i, child := range flowChildren {
		cs := &child.Style
		grow := cs.FlexGrow
		shrink := cs.FlexShrink

		var marginMain float64
		if isRow {
			marginMain = cs.Margin.Horizontal()
		} else {
			marginMain = cs.Margin.Vertical()
		}

		// Determine base main size from flex-basis, then width/height, then content.
		base := 0.0
		basis := cs.FlexBasis

		if basis.Unit != Auto {
			base = resolveSize(basis, mainSize, mainSize)
		} else {
			// Use explicit width/height if set.
			var explicit Dimension
			if isRow {
				explicit = cs.Width
			} else {
				explicit = cs.Height
			}
			if explicit.Unit != Auto {
				base = resolveSize(explicit, mainSize, mainSize)
			} else {
				// Measure content for leaf nodes.
				base = measureChildMain(child, isRow, mainSize, crossSize)
			}
		}

		// Determine cross size.
		var crossDim Dimension
		if isRow {
			crossDim = cs.Height
		} else {
			crossDim = cs.Width
		}
		cross := 0.0
		if crossDim.Unit != Auto {
			cross = resolveSize(crossDim, crossSize, crossSize)
		} else {
			cross = measureChildCross(child, isRow, base, crossSize)
		}

		infos[i] = childInfo{
			node:       child,
			baseMain:   base,
			baseCross:  cross,
			flexGrow:   grow,
			flexShrink: shrink,
			mainSize:   base,
			crossSize:  cross,
			marginMain: marginMain,
		}
	}

	return infos
}

// layoutNoWrap lays out all children on a single flex line (the default).
func layoutNoWrap(node *Node, infos []childInfo, isRow bool, mainSize, crossSize float64, style *Style, bw float64) {
	// --- Phase 2: Flex grow/shrink ---
	flexGrowShrink(infos, isRow, mainSize, style.Gap)

	// --- Phase 3: Cross axis sizing (align-items / align-self) ---
	for i := range infos {
		resolveCrossSize(&infos[i], isRow, crossSize, style.Align)
	}

	// --- Phase 4: Main axis positioning (justify-content) ---
	totalGaps := 0.0
	if len(infos) > 1 {
		totalGaps = style.Gap * float64(len(infos)-1)
	}
	mainOffset, gap := justifyMain(infos, mainSize, totalGaps, style.Justify)

	// --- Phase 5: Assign absolute positions ---
	positionChildren(node, infos, isRow, mainOffset, gap, crossSize, style, bw)
}

// flexGrowShrink distributes free space among children via flex-grow/shrink.
func flexGrowShrink(infos []childInfo, isRow bool, mainSize, styleGap float64) {
	totalBaseMain := 0.0
	for i := range infos {
		totalBaseMain += infos[i].baseMain + infos[i].marginMain
	}
	totalGaps := 0.0
	if len(infos) > 1 {
		totalGaps = styleGap * float64(len(infos)-1)
	}
	freeSpace := mainSize - totalBaseMain - totalGaps

	if freeSpace > 0 {
		totalGrow := 0.0
		for i := range infos {
			totalGrow += infos[i].flexGrow
		}
		if totalGrow > 0 {
			for i := range infos {
				if infos[i].flexGrow > 0 {
					infos[i].mainSize = infos[i].baseMain + freeSpace*(infos[i].flexGrow/totalGrow)
				}
			}
		}
	} else if freeSpace < 0 {
		totalShrinkScaled := 0.0
		for i := range infos {
			totalShrinkScaled += infos[i].flexShrink * infos[i].baseMain
		}
		if totalShrinkScaled > 0 {
			overflowAbs := -freeSpace
			for i := range infos {
				ratio := (infos[i].flexShrink * infos[i].baseMain) / totalShrinkScaled
				shrinkAmount := overflowAbs * ratio
				infos[i].mainSize = math.Max(0, infos[i].baseMain-shrinkAmount)
			}
		}
	}

	// Apply min/max constraints on main axis.
	for i := range infos {
		cs := &infos[i].node.Style
		var minDim, maxDim Dimension
		if isRow {
			minDim, maxDim = cs.MinWidth, cs.MaxWidth
		} else {
			minDim, maxDim = cs.MinHeight, cs.MaxHeight
		}
		if minDim.Unit == Fixed {
			infos[i].mainSize = math.Max(infos[i].mainSize, minDim.Value)
		}
		if maxDim.Unit == Fixed {
			infos[i].mainSize = math.Min(infos[i].mainSize, maxDim.Value)
		}
	}
}

// resolveCrossSize resolves cross-axis sizing for a single child, including stretch.
func resolveCrossSize(info *childInfo, isRow bool, lineCross float64, parentAlign AlignItems) {
	cs := &info.node.Style
	effAlign := parentAlign
	switch cs.AlignSelf {
	case AlignSelfFlexStart:
		effAlign = AlignFlexStart
	case AlignSelfFlexEnd:
		effAlign = AlignFlexEnd
	case AlignSelfCenter:
		effAlign = AlignCenter
	case AlignSelfStretch:
		effAlign = AlignStretch
	}

	if effAlign == AlignStretch {
		var crossDim Dimension
		if isRow {
			crossDim = cs.Height
		} else {
			crossDim = cs.Width
		}
		if crossDim.Unit == Auto {
			var crossMargin float64
			if isRow {
				crossMargin = cs.Margin.Vertical()
			} else {
				crossMargin = cs.Margin.Horizontal()
			}
			info.crossSize = lineCross - crossMargin
		}
	}

	// Apply min/max on cross axis.
	var minDim, maxDim Dimension
	if isRow {
		minDim, maxDim = cs.MinHeight, cs.MaxHeight
	} else {
		minDim, maxDim = cs.MinWidth, cs.MaxWidth
	}
	if minDim.Unit == Fixed {
		info.crossSize = math.Max(info.crossSize, minDim.Value)
	}
	if maxDim.Unit == Fixed {
		info.crossSize = math.Min(info.crossSize, maxDim.Value)
	}
}

// justifyMain computes the starting offset and inter-item gap for justify-content.
func justifyMain(infos []childInfo, mainSize, totalGaps float64, justify JustifyContent) (mainOffset, gap float64) {
	usedMain := 0.0
	for i := range infos {
		usedMain += infos[i].mainSize + infos[i].marginMain
	}
	remaining := mainSize - usedMain - totalGaps
	if remaining < 0 {
		remaining = 0
	}

	n := len(infos)
	switch justify {
	case JustifyFlexStart:
		mainOffset = 0
	case JustifyFlexEnd:
		mainOffset = remaining
	case JustifyCenter:
		mainOffset = remaining / 2
	case JustifySpaceBetween:
		if n > 1 {
			gap = remaining / float64(n-1)
		}
	case JustifySpaceAround:
		if n > 0 {
			gap = remaining / float64(n)
			mainOffset = gap / 2
		}
	}
	return
}

// positionChildren assigns absolute x/y positions to children along a single line.
func positionChildren(node *Node, infos []childInfo, isRow bool, mainOffset, gap, crossSize float64, style *Style, bw float64) {
	cursor := mainOffset
	for i := range infos {
		child := infos[i].node
		cs := &child.Style

		var marginMainStart, marginCrossStart float64
		if isRow {
			marginMainStart = cs.Margin.Left
			marginCrossStart = cs.Margin.Top
		} else {
			marginMainStart = cs.Margin.Top
			marginCrossStart = cs.Margin.Left
		}

		cursor += marginMainStart
		mainPos := cursor

		// Resolve effective cross alignment (AlignSelf overrides parent).
		effAlign := style.Align
		switch cs.AlignSelf {
		case AlignSelfFlexStart:
			effAlign = AlignFlexStart
		case AlignSelfFlexEnd:
			effAlign = AlignFlexEnd
		case AlignSelfCenter:
			effAlign = AlignCenter
		case AlignSelfStretch:
			effAlign = AlignStretch
		}

		crossPos := 0.0
		childCross := infos[i].crossSize
		switch effAlign {
		case AlignFlexStart:
			crossPos = marginCrossStart
		case AlignFlexEnd:
			crossPos = crossSize - childCross - marginCrossStart
		case AlignCenter:
			crossPos = (crossSize - childCross) / 2
		case AlignStretch:
			crossPos = marginCrossStart
		}

		// Convert main/cross to x/y, offset by padding + border.
		if isRow {
			child.Layout.X = node.Layout.X + style.Padding.Left + bw + mainPos
			child.Layout.Y = node.Layout.Y + style.Padding.Top + bw + crossPos
			child.Layout.Width = infos[i].mainSize
			child.Layout.Height = childCross
		} else {
			child.Layout.X = node.Layout.X + style.Padding.Left + bw + crossPos
			child.Layout.Y = node.Layout.Y + style.Padding.Top + bw + mainPos
			child.Layout.Width = childCross
			child.Layout.Height = infos[i].mainSize
		}

		// Advance cursor: remaining margin + justify gap + style gap.
		childGap := gap + style.Gap
		if i == len(infos)-1 {
			childGap = 0 // no gap after last child
		}
		cursor += infos[i].mainSize + (infos[i].marginMain - marginMainStart) + childGap

		// Recurse into children.
		layoutNode(child)
	}
}

// layoutWrap lays out children across multiple flex lines, wrapping when
// children exceed the available main-axis space.
func layoutWrap(node *Node, infos []childInfo, isRow bool, mainSize, crossSize float64, style *Style, bw float64) {
	// Break children into lines.
	type flexLine struct {
		items []childInfo
	}
	var lines []flexLine
	var currentLine []childInfo
	lineMainUsed := 0.0

	for i := range infos {
		childMain := infos[i].baseMain + infos[i].marginMain
		gapSize := 0.0
		if len(currentLine) > 0 {
			gapSize = style.Gap
		}
		if len(currentLine) > 0 && lineMainUsed+gapSize+childMain > mainSize {
			lines = append(lines, flexLine{items: currentLine})
			currentLine = []childInfo{infos[i]}
			lineMainUsed = childMain
		} else {
			lineMainUsed += gapSize + childMain
			currentLine = append(currentLine, infos[i])
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, flexLine{items: currentLine})
	}

	// Process each line: grow/shrink, cross sizing, and positioning.
	crossOffset := 0.0
	for _, line := range lines {
		// Phase 2: Flex grow/shrink within line.
		flexGrowShrink(line.items, isRow, mainSize, style.Gap)

		// Determine line cross size = max child cross in this line.
		lineCross := 0.0
		for j := range line.items {
			c := line.items[j].baseCross
			if isRow {
				c += line.items[j].node.Style.Margin.Vertical()
			} else {
				c += line.items[j].node.Style.Margin.Horizontal()
			}
			if c > lineCross {
				lineCross = c
			}
		}

		// Phase 3: Cross axis sizing within line.
		for j := range line.items {
			resolveCrossSize(&line.items[j], isRow, lineCross, style.Align)
		}

		// Phase 4: Justify within line.
		totalGaps := 0.0
		if len(line.items) > 1 {
			totalGaps = style.Gap * float64(len(line.items)-1)
		}
		mainOffset, gap := justifyMain(line.items, mainSize, totalGaps, style.Justify)

		// Phase 5: Position children, offset by crossOffset for this line.
		cursor := mainOffset
		for j := range line.items {
			child := line.items[j].node
			cs := &child.Style

			var marginMainStart, marginCrossStart float64
			if isRow {
				marginMainStart = cs.Margin.Left
				marginCrossStart = cs.Margin.Top
			} else {
				marginMainStart = cs.Margin.Top
				marginCrossStart = cs.Margin.Left
			}

			cursor += marginMainStart
			mainPos := cursor

			effAlign := style.Align
			switch cs.AlignSelf {
			case AlignSelfFlexStart:
				effAlign = AlignFlexStart
			case AlignSelfFlexEnd:
				effAlign = AlignFlexEnd
			case AlignSelfCenter:
				effAlign = AlignCenter
			case AlignSelfStretch:
				effAlign = AlignStretch
			}

			crossPos := 0.0
			childCross := line.items[j].crossSize
			switch effAlign {
			case AlignFlexStart:
				crossPos = marginCrossStart
			case AlignFlexEnd:
				crossPos = lineCross - childCross - marginCrossStart
			case AlignCenter:
				crossPos = (lineCross - childCross) / 2
			case AlignStretch:
				crossPos = marginCrossStart
			}

			if isRow {
				child.Layout.X = node.Layout.X + style.Padding.Left + bw + mainPos
				child.Layout.Y = node.Layout.Y + style.Padding.Top + bw + crossOffset + crossPos
				child.Layout.Width = line.items[j].mainSize
				child.Layout.Height = childCross
			} else {
				child.Layout.X = node.Layout.X + style.Padding.Left + bw + crossOffset + crossPos
				child.Layout.Y = node.Layout.Y + style.Padding.Top + bw + mainPos
				child.Layout.Width = childCross
				child.Layout.Height = line.items[j].mainSize
			}

			childGap := gap + style.Gap
			if j == len(line.items)-1 {
				childGap = 0
			}
			cursor += line.items[j].mainSize + (line.items[j].marginMain - marginMainStart) + childGap

			layoutNode(child)
		}

		crossOffset += lineCross
	}
}

// ComputeInline runs the layout algorithm with fixed width and content-based
// height, matching the inline rendering model of Ink/iocraft. Returns the
// computed content height.
func ComputeInline(node *Node, availWidth float64) float64 {
	node.Layout.X = node.Style.Margin.Left
	node.Layout.Y = node.Style.Margin.Top
	if node.Style.Width.Unit != Auto {
		node.Layout.Width = resolveSize(node.Style.Width, availWidth, availWidth) - node.Style.Margin.Horizontal()
	} else if node.Style.FlexGrow > 0 {
		node.Layout.Width = availWidth - node.Style.Margin.Horizontal()
	} else {
		// Use natural content width so inline elements don't stretch
		// to the full terminal width.
		nw := naturalWidth(node)
		maxW := availWidth - node.Style.Margin.Horizontal()
		if nw > maxW {
			nw = maxW
		}
		node.Layout.Width = nw
	}

	// Measure natural content height.
	h := measureContentHeight(node)
	node.Layout.Height = h

	layoutNode(node)
	return h
}

// measureContentHeight computes the natural height of a node given its
// Layout.Width is already set. This recurses into children.
func measureContentHeight(node *Node) float64 {
	style := &node.Style
	bw := style.BorderWidth()

	if node.IsLeaf() {
		if node.MeasureFunc != nil {
			innerW := node.Layout.Width - style.Padding.Horizontal() - bw*2
			_, h := node.MeasureFunc(innerW, 0)
			return h + style.Padding.Vertical() + bw*2
		}
		return style.Padding.Vertical() + bw*2
	}

	innerW := node.Layout.Width - style.Padding.Horizontal() - bw*2
	isRow := style.Direction == FlexRow

	if isRow {
		maxH := 0.0
		for _, child := range node.Children {
			h := childNaturalHeight(child, innerW) + child.Style.Margin.Vertical()
			if h > maxH {
				maxH = h
			}
		}
		return maxH + style.Padding.Vertical() + bw*2
	}

	// Column: height = sum of children heights + gaps.
	totalH := 0.0
	for i, child := range node.Children {
		totalH += childNaturalHeight(child, innerW) + child.Style.Margin.Vertical()
		if i < len(node.Children)-1 {
			totalH += style.Gap
		}
	}
	return totalH + style.Padding.Vertical() + bw*2
}

// childNaturalHeight returns one child's natural height given the parent's
// inner width. It handles leaf nodes, explicit dimensions, and nested containers.
func childNaturalHeight(child *Node, parentInnerW float64) float64 {
	cs := &child.Style

	// Explicit height always wins.
	if cs.Height.Unit != Auto {
		return resolveSize(cs.Height, 0, 0)
	}

	// Flex-basis on a Column parent acts as height.
	if cs.FlexBasis.Unit != Auto {
		return resolveSize(cs.FlexBasis, 0, 0)
	}

	// Determine child width for measurement.
	childW := parentInnerW - cs.Margin.Horizontal()
	if cs.Width.Unit != Auto {
		childW = resolveSize(cs.Width, parentInnerW, parentInnerW)
	}

	// Leaf with measure func.
	if child.IsLeaf() && child.MeasureFunc != nil {
		_, h := child.MeasureFunc(childW-cs.Padding.Horizontal(), 0)
		return h + cs.Padding.Vertical()
	}

	// Nested container: recurse.
	if !child.IsLeaf() {
		child.Layout.Width = childW
		return measureContentHeight(child)
	}

	return 0
}

// resolveSize resolves a Dimension to a concrete value.
func resolveSize(dim Dimension, containerSize, fallback float64) float64 {
	switch dim.Unit {
	case Fixed:
		return dim.Value
	case Percent:
		return containerSize * dim.Value / 100
	default:
		return fallback
	}
}

// naturalWidth computes the intrinsic content width of a node by recursing
// into its children. This is used for content-sized boxes.
func naturalWidth(node *Node) float64 {
	cs := &node.Style

	if cs.Width.Unit != Auto {
		return resolveSize(cs.Width, 0, 0)
	}

	bw := cs.BorderWidth()

	if node.IsLeaf() {
		if node.MeasureFunc != nil {
			w, _ := node.MeasureFunc(math.MaxFloat64, 0)
			return w + cs.Padding.Horizontal() + bw*2
		}
		return cs.Padding.Horizontal() + bw*2
	}

	if cs.Direction == FlexRow {
		totalW := 0.0
		for i, child := range node.Children {
			totalW += naturalWidth(child) + child.Style.Margin.Horizontal()
			if i < len(node.Children)-1 {
				totalW += cs.Gap
			}
		}
		return totalW + cs.Padding.Horizontal() + bw*2
	}

	// Column: width = max child width.
	maxW := 0.0
	for _, child := range node.Children {
		w := naturalWidth(child) + child.Style.Margin.Horizontal()
		if w > maxW {
			maxW = w
		}
	}
	return maxW + cs.Padding.Horizontal() + bw*2
}

// measureChildMain returns the intrinsic main-axis size of a child.
func measureChildMain(child *Node, isRow bool, availMain, availCross float64) float64 {
	if child.MeasureFunc != nil {
		// MeasureFunc expects (availWidth, availHeight) regardless of flex direction.
		var w, h float64
		if isRow {
			w, h = child.MeasureFunc(availMain, availCross)
		} else {
			w, h = child.MeasureFunc(availCross, availMain)
		}
		if isRow {
			return w + child.Style.Padding.Horizontal()
		}
		return h + child.Style.Padding.Vertical()
	}
	if !child.IsLeaf() {
		if isRow {
			return naturalWidth(child)
		}
		// Column main = height: compute natural height.
		child.Layout.Width = availCross
		return measureContentHeight(child)
	}
	return 0
}

// measureChildCross returns the intrinsic cross-axis size of a child.
func measureChildCross(child *Node, isRow bool, mainSize, availCross float64) float64 {
	if child.MeasureFunc != nil {
		// MeasureFunc expects (availWidth, availHeight).
		var w, h float64
		if isRow {
			w, h = child.MeasureFunc(mainSize, availCross)
		} else {
			w, h = child.MeasureFunc(availCross, mainSize)
		}
		if isRow {
			return h + child.Style.Padding.Vertical()
		}
		return w + child.Style.Padding.Horizontal()
	}
	if !child.IsLeaf() {
		if isRow {
			// Cross is height: compute natural height given main (width).
			child.Layout.Width = mainSize
			return measureContentHeight(child)
		}
		// Cross is width: compute natural width.
		return naturalWidth(child)
	}
	return 0
}
