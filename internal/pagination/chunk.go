package pagination

// chunkCollector merges pages in fixed-size chunks to reduce peak memory during growth.
type chunkCollector struct {
	chunkSize int
	chunks    [][]map[string]any
	current   []map[string]any
}

func newChunkCollector(chunkSize int) *chunkCollector {
	if chunkSize <= 0 {
		chunkSize = ChunkSize
	}
	return &chunkCollector{
		chunkSize: chunkSize,
		current:   make([]map[string]any, 0, chunkSize),
	}
}

func (c *chunkCollector) add(items []map[string]any) {
	for _, item := range items {
		c.current = append(c.current, item)
		if len(c.current) >= c.chunkSize {
			c.flush()
		}
	}
}

func (c *chunkCollector) flush() {
	if len(c.current) == 0 {
		return
	}
	chunk := make([]map[string]any, len(c.current))
	copy(chunk, c.current)
	c.chunks = append(c.chunks, chunk)
	c.current = c.current[:0]
}

func (c *chunkCollector) result() []map[string]any {
	c.flush()
	total := 0
	for _, ch := range c.chunks {
		total += len(ch)
	}
	out := make([]map[string]any, 0, total)
	for _, ch := range c.chunks {
		out = append(out, ch...)
	}
	c.chunks = nil
	c.current = nil
	return out
}
