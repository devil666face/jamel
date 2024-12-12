package queue

import (
	"errors"
	"sync"
	"time"

	"jamel/gen/go/jamel"
)

const busyWait = 10 * time.Millisecond

var (
	ErrNotFound = errors.New("not found id")
	ErrTimeout  = errors.New("timeout")
)

type Queue struct {
	s     map[string]*jamel.TaskResponse
	mutex sync.Mutex
}

func New() *Queue {
	return &Queue{
		s: make(map[string]*jamel.TaskResponse),
	}
}

func (q *Queue) Get(id string) (*jamel.TaskResponse, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	resp, ok := q.s[id]
	if !ok {
		return nil, ErrNotFound
	}
	delete(q.s, id)
	return resp, nil
}

func (q *Queue) Set(resp *jamel.TaskResponse) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.s[resp.TaskId] = resp
}

func (q *Queue) WaitResp(id string, timeoutsec ...time.Duration) (*jamel.TaskResponse, error) {
	var timech <-chan time.Time

	if len(timeoutsec) > 0 && timeoutsec[0] > 0 {
		timech = time.After(timeoutsec[0] * time.Second)
	}

	for {
		select {
		case <-timech:
			if len(timeoutsec) > 0 && timeoutsec[0] > 0 {
				return nil, ErrTimeout
			}
		default:
			resp, err := q.Get(id)
			if errors.Is(err, ErrNotFound) {
				time.Sleep(busyWait)
				continue
			}
			return resp, err
		}
	}
}
