// A prefix tree whose keys are path segments.
package walk

import (
	"io/fs"
)

type pathTrie struct {
	children map[string]*pathTrie
	entry    *fs.DirEntry
}

func (t *pathTrie) AddChild(entry fs.DirEntry) *pathTrie {
	if t.children == nil {
		t.children = map[string]*pathTrie{}
	}

	child := &pathTrie{entry: &entry}
	t.children[entry.Name()] = child
	return child
}
