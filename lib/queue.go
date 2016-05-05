package lib

type PlayQueue struct {
	elements []string
}

func NewPlayQueue() *PlayQueue {
	return &PlayQueue{elements: make([]string, 0, 0)}
}

func (queue *PlayQueue) Append(track string) string {
	queue.elements = append(queue.elements, track)

	return track
}

func (queue *PlayQueue) Pop() string {
	if len(queue.elements) == 0 {
		return ""
	}

	track := queue.elements[0]
	queue.elements = queue.elements[1:len(queue.elements)]
	return track
}

func (queue *PlayQueue) IsEmpty() bool {
	return len(queue.elements) == 0
}
