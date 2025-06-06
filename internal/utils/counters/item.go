// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package counters

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
)

const spanMaxValue = 10_000_000
const maxSpans = 10

type Item[T SupportedUIntType] struct {
	spans          [maxSpans + 1]T
	lastUpdateTime int64
	lifeSeconds    int64
	spanSeconds    int64
}

func NewItem[T SupportedUIntType](lifeSeconds int) Item[T] {
	if lifeSeconds <= 0 {
		lifeSeconds = 60
	}
	var spanSeconds = lifeSeconds / maxSpans
	if spanSeconds < 1 {
		spanSeconds = 1
	} else if lifeSeconds > maxSpans && lifeSeconds%maxSpans != 0 {
		spanSeconds++
	}

	return Item[T]{
		lifeSeconds:    int64(lifeSeconds),
		spanSeconds:    int64(spanSeconds),
		lastUpdateTime: fasttime.Now().Unix(),
	}
}

func (this *Item[T]) Increase() (result T) {
	var currentTime = fasttime.Now().Unix()
	var currentSpanIndex = this.calculateSpanIndex(currentTime)

	// return quickly
	if this.lastUpdateTime == currentTime {
		if this.spans[currentSpanIndex] < spanMaxValue {
			this.spans[currentSpanIndex]++
		}
		for _, count := range this.spans {
			result += count
		}
		return
	}

	if this.lastUpdateTime > 0 {
		if currentTime-this.lastUpdateTime > this.lifeSeconds {
			for index := range this.spans {
				this.spans[index] = 0
			}
		} else {
			var lastSpanIndex = this.calculateSpanIndex(this.lastUpdateTime)

			if lastSpanIndex != currentSpanIndex {
				var countSpans = len(this.spans)

				// reset values between LAST and CURRENT
				for index := lastSpanIndex + 1; ; index++ {
					var realIndex = index % countSpans
					this.spans[realIndex] = 0
					if realIndex == currentSpanIndex {
						break
					}
				}
			}
		}
	}

	if this.spans[currentSpanIndex] < spanMaxValue {
		this.spans[currentSpanIndex]++
	}
	this.lastUpdateTime = currentTime

	for _, count := range this.spans {
		result += count
	}

	return
}

func (this *Item[T]) Sum() (result T) {
	if this.lastUpdateTime == 0 {
		return 0
	}

	var currentTime = fasttime.Now().Unix()
	var currentSpanIndex = this.calculateSpanIndex(currentTime)

	if currentTime-this.lastUpdateTime > this.lifeSeconds {
		return 0
	} else {
		var lastSpanIndex = this.calculateSpanIndex(this.lastUpdateTime)
		var countSpans = len(this.spans)
		for index := currentSpanIndex + 1; ; index++ {
			var realIndex = index % countSpans
			result += this.spans[realIndex]
			if realIndex == lastSpanIndex {
				break
			}
		}
	}

	return result
}

func (this *Item[T]) Reset() {
	for index := range this.spans {
		this.spans[index] = 0
	}
}

func (this *Item[T]) IsExpired(currentTime int64) bool {
	return this.lastUpdateTime < currentTime-this.lifeSeconds-this.spanSeconds
}

func (this *Item[T]) calculateSpanIndex(timestamp int64) int {
	var index = int(timestamp % this.lifeSeconds / this.spanSeconds)
	if index > maxSpans-1 {
		return maxSpans - 1
	}
	return index
}

func (this *Item[T]) IsOk() bool {
	return this.lifeSeconds > 0
}
