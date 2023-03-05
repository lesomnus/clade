package clade

import (
	"golang.org/x/exp/maps"
)

type BuildPlan struct {
	Iterations [][][]string `json:"iterations"`
}

func NewBuildPlan(bg *BuildGraph) BuildPlan {
	snapshot := bg.Snapshot()

	by_level := make(map[uint][]string)
	for ref, entry := range snapshot {
		iteration, ok := by_level[entry.Level]
		if !ok {
			iteration = make([]string, 0)
		}

		by_level[entry.Level] = append(iteration, ref)
	}

	// Remove root nodes.
	delete(by_level, 0)

	iterations := make([][][]string, len(by_level))
	for level, refs := range by_level {
		type Collection struct {
			Index  uint
			Groups map[uint]bool
		}

		groups := make(map[uint]*Collection)
		index := uint(0)
		for _, ref := range refs {
			entry, ok := snapshot[ref]
			if !ok {
				panic("entry must be exists")
			}

			var coll_curr *Collection

			// Find existing collection.
			for id := range entry.Group {
				coll, ok := groups[id]
				if !ok {
					continue
				}

				coll_curr = coll
				break
			}

			if coll_curr == nil {
				// Create new collection if there is no existing one.
				coll_curr = &Collection{
					Index:  index,
					Groups: maps.Clone(entry.Group),
				}
				index++
			} else {
				// Add new group IDs in the collection.
				for id := range entry.Group {
					coll_curr.Groups[id] = true
				}
			}

			for id := range entry.Group {
				coll_next, ok := groups[id]
				if !ok {
					// Assign current collection if the group is not collected.
					groups[id] = coll_curr
					continue
				}
				if coll_next == coll_curr {
					// Already collected.
					continue
				}

				// Merge collections.
				// In the case of:
				//  A(a) - C(a, b)
				//  B(b) /
				// A and B are merged by C.
				// But this case will not happen in the build graph
				// since C's group ID is propagated to A and B.
				for id := range coll_next.Groups {
					coll_curr.Groups[id] = true
					groups[id] = coll_curr
				}
			}
		}

		collections := make(map[uint]*Collection)
		for _, collection := range groups {
			coll, ok := collections[collection.Index]
			if ok && coll != collection {
				panic("collection duplicated")
			}

			collections[collection.Index] = collection
		}

		// Re-indexing.
		index = 0
		for _, collection := range collections {
			collection.Index = index
			index++
		}

		// Group references by collection.
		ref_groups := make([][]string, len(collections))
		for _, ref := range refs {
			entry, ok := snapshot[ref]
			if !ok {
				panic("entry must be exists")
			}

			index := uint(0)
			for id := range entry.Group {
				index = groups[id].Index
				break
			}

			ref_group := ref_groups[index]
			if ref_group == nil {
				ref_group = make([]string, 0, 1)
			}
			ref_groups[index] = append(ref_group, ref)
		}

		// Note that level starts from 1.
		iterations[level-1] = ref_groups
	}

	return BuildPlan{
		Iterations: iterations,
	}
}
