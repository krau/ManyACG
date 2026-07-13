package orb

import (
	"math"
	"sort"
)

type extractorNode struct {
	keys     []keyPoint
	ulX, ulY int
	urX      int
	blY      int
	noMore   bool
}

func (n *extractorNode) divide() (n1, n2, n3, n4 extractorNode) {
	halfX := int(math.Ceil(float64(n.urX-n.ulX) / 2))
	halfY := int(math.Ceil(float64(n.blY-n.ulY) / 2))

	n1.ulX, n1.ulY, n1.urX, n1.blY = n.ulX, n.ulY, n.ulX+halfX, n.ulY+halfY
	n2.ulX, n2.ulY, n2.urX, n2.blY = n.ulX+halfX, n.ulY, n.urX, n.ulY+halfY
	n3.ulX, n3.ulY, n3.urX, n3.blY = n.ulX, n.ulY+halfY, n.ulX+halfX, n.blY
	n4.ulX, n4.ulY, n4.urX, n4.blY = n.ulX+halfX, n.ulY+halfY, n.urX, n.blY

	midX := float32(n.ulX + halfX)
	midY := float32(n.ulY + halfY)
	for _, kp := range n.keys {
		if kp.x < midX {
			if kp.y < midY {
				n1.keys = append(n1.keys, kp)
			} else {
				n3.keys = append(n3.keys, kp)
			}
		} else if kp.y < midY {
			n2.keys = append(n2.keys, kp)
		} else {
			n4.keys = append(n4.keys, kp)
		}
	}
	return
}

func distributeOctTree(keys []keyPoint, minX, maxX, minY, maxY, n int) []keyPoint {
	if len(keys) == 0 {
		return nil
	}
	nIni := max(int(math.Round(float64(maxX-minX)/float64(maxY-minY))), 1)
	hX := float64(maxX-minX) / float64(nIni)

	var nodes []extractorNode
	nodes = make([]extractorNode, 0, nIni)
	for i := range nIni {
		var ni extractorNode
		ni.ulX = int(hX * float64(i))
		ni.urX = int(hX * float64(i+1))
		ni.ulY = 0
		ni.blY = maxY - minY
		nodes = append(nodes, ni)
	}
	for _, kp := range keys {
		idx := max(int(kp.x/float32(hX)), 0)
		if idx >= len(nodes) {
			idx = len(nodes) - 1
		}
		nodes[idx].keys = append(nodes[idx].keys, kp)
	}

	filtered := nodes[:0]
	for i := range nodes {
		switch len(nodes[i].keys) {
		case 0:
		case 1:
			nodes[i].noMore = true
			filtered = append(filtered, nodes[i])
		default:
			filtered = append(filtered, nodes[i])
		}
	}
	nodes = append([]extractorNode(nil), filtered...)

	for {
		prevSize := len(nodes)

		var expandableIdx []int
		for i := range nodes {
			if !nodes[i].noMore && len(nodes[i].keys) > 1 {
				expandableIdx = append(expandableIdx, i)
			}
		}
		if len(expandableIdx) == 0 {
			break
		}
		sort.SliceStable(expandableIdx, func(a, b int) bool {
			return len(nodes[expandableIdx[a]].keys) > len(nodes[expandableIdx[b]].keys)
		})

		expanded := make(map[int]bool)
		var children []extractorNode
		for _, ni := range expandableIdx {
			c1, c2, c3, c4 := nodes[ni].divide()
			for _, c := range []extractorNode{c1, c2, c3, c4} {
				if len(c.keys) == 0 {
					continue
				}
				if len(c.keys) == 1 {
					c.noMore = true
				}
				children = append(children, c)
			}
			expanded[ni] = true
			kept := len(nodes) - len(expanded)
			if kept+len(children) >= n {
				break
			}
		}

		var rebuilt []extractorNode
		for i := range nodes {
			if !expanded[i] {
				rebuilt = append(rebuilt, nodes[i])
			}
		}
		rebuilt = append(rebuilt, children...)
		nodes = rebuilt

		if len(nodes) >= n || len(nodes) == prevSize {
			break
		}
	}

	out := make([]keyPoint, 0, len(nodes))
	for i := range nodes {
		if len(nodes[i].keys) == 0 {
			continue
		}
		best := nodes[i].keys[0]
		for _, kp := range nodes[i].keys[1:] {
			if kp.response > best.response {
				best = kp
			}
		}
		out = append(out, best)
	}
	return out
}
