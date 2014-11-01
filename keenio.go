// Copyright (c) 2014 Bill Davis. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package beego_keenio implements a middleware to KeenIO from the beego framework.
package beego_keenio

import (
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/philpearl/keengo"
)

// KEENIO_QUEUE_KEY constant to identify the context key in the request
const (
	KEENIO_QUEUE_KEY = "keenio_queue"
)

var sender *keengo.Sender

// keenioQueue interface for queue
type keenioQueue interface {
	Len() int
	Push()
	Pop() (string, interface{})
}

type keenioEvent struct {
	collection string
	data       interface{}
	next       *keenioEvent
}

//	KeenioQueue is FIFO data stucture
type KeenioQueue struct {
	head  *keenioEvent
	tail  *keenioEvent
	count int
	lock  *sync.Mutex
}

//	Creates a new pointer to a new queue.
func newKeenioQueue() *KeenioQueue {
	q := &KeenioQueue{}
	q.lock = &sync.Mutex{}
	return q
}

// Len returns the number of events in the queue
func (q *KeenioQueue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

// Push adds event to the end of the queue
func (q *KeenioQueue) Push(collection string, item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &keenioEvent{collection: collection, data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
}

// Pop returns event from the top of the queue
func (q *KeenioQueue) Pop() (string, interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == nil {
		return "", nil
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.collection, n.data
}

// InitKeenioQueue initialize the queue structure for this request
func InitKeenioQueue(ctx *context.Context) {
	q := newKeenioQueue()
	ctx.Input.SetData(KEENIO_QUEUE_KEY, *q)
}

// ProcessKeenioQueue iterates the queue structure for this request
func ProcessKeenioQueue(ctx *context.Context) {

	if q, ok := ctx.Input.GetData(KEENIO_QUEUE_KEY).(KeenioQueue); ok {
		cnt := q.Len()
		for i := 0; i < cnt; i++ {
			coll, data := q.Pop()
			if coll != "" && data != nil {
				sender.Queue(coll, data)
			}
		}
	}
}

// InitKeenioFilter initializes the keengo sender in a go-routine
func InitKeenioFilter() {

	// validate the necessary configuration
	projectId := beego.AppConfig.String("KeenioProjectId")
	if projectId == "" {
		beego.Warn("Please specify Keenio Project ID in the application config: KeenioProjectId=53dfa0000000000000000002")
		return
	}

	writeKey := beego.AppConfig.String("KeenioWriteKey")
	// easy to get whitespace in the write key based on length
	writeKey = strings.Replace(writeKey, " ", "", -1)

	if writeKey == "" {
		beego.Warn("Please specify Keenio Write Key in the application config: KeenioWriteKey=d21785d8ade08c6f5116b39eed701ff4dbe978688333sefd1a550788e09486c1a40cf1d48f56f1feee730ea4710a081f6631bc1b649847e8937d8be2953e1df9dc8a89c5a69f6d6ad18c6381739f3ab21bd90c376e07f0bf0fdcb6e9cbb702db1ace3c9a 60d3530fffa18d84c65cb3ee")
		return
	}

	sender = keengo.NewSender(projectId, writeKey)

	beego.InsertFilter("*", beego.BeforeRouter, InitKeenioQueue)
	beego.InsertFilter("*", beego.FinishRouter, ProcessKeenioQueue)

	beego.Info("Keenio filter initialized")
}
