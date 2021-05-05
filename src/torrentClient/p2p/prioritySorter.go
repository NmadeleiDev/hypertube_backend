package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type prioritySorter struct {
	mu sync.Mutex
	Pieces	[]*pieceWork
	PriorityUpdates	<- chan int64

	topPiece *pieceWork
}

func (s *prioritySorter) InitSorter(ctx context.Context) (topPriorityPieceChan, returnedPiecesChan chan *pieceWork) {
	topPriorityPieceChan = make(chan *pieceWork)
	returnedPiecesChan = make(chan *pieceWork, 50)

	latestTopIdx := int64(0)

	s.mu.Lock()
	s.topPiece, _ = s.findClosestPiece(s.Pieces, latestTopIdx)
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <- ctx.Done():
				close(topPriorityPieceChan)
				close(returnedPiecesChan)
				return
			case newTopIdx := <- s.PriorityUpdates:
				latestTopIdx = newTopIdx
				if len(s.Pieces) == 0 {
					s.topPiece = nil
					continue
				}
				s.mu.Lock()
				s.topPiece, _ = s.findClosestPiece(s.Pieces, latestTopIdx)
				fmt.Printf("Got priority update=%v; Found new top piece idx=%v\n", newTopIdx, s.topPiece.index)
				s.mu.Unlock()
			case returnedPiece := <- returnedPiecesChan: // нам вернули часть, которую не получилось скачать
				s.mu.Lock()
				if len(s.Pieces) == 0 {
					s.Pieces = append(s.Pieces, returnedPiece)
					s.topPiece = returnedPiece
					s.mu.Unlock()
					continue
				} else {
					for i, piece := range s.Pieces { // важно вставить часть на свое место
						if piece.index > returnedPiece.index {
							if i == 0 {
								s.Pieces = append([]*pieceWork{returnedPiece}, s.Pieces...)
								break
							}
							s.Pieces = append(append(s.Pieces[:i], returnedPiece), s.Pieces[i:]...)
						}
					}
				}
				s.topPiece, _ = s.findClosestPiece(s.Pieces, latestTopIdx)
				s.mu.Unlock()
			default:
				if s.topPiece == nil {
					continue
				}
				topPriorityPieceChan <- s.topPiece
				if len(s.Pieces) == 1 { // значит, мы эту единственную часть только что и отдали, больше не осталось
					//close(topPriorityPieceChan) все равно закроем после контекста
					//return
					logrus.Debugf("All pieces are sent to workers!")
					s.Pieces = s.Pieces[:0]
				} else {
					// удаляю отданную часть
					pLen := len(s.Pieces)
					for i, piece := range s.Pieces {
						if piece.index == s.topPiece.index {
							if i == pLen {
								s.Pieces = s.Pieces[:]
								break
							}
							copy(s.Pieces[i:], s.Pieces[i + 1:])
							s.Pieces = s.Pieces[:len(s.Pieces) - 1]
							break
						}
					}
					//fmt.Printf("Deleted piece in sorter, new len=%v\n", len(s.Pieces))
				}
				if len(s.Pieces) == 0 {
					s.topPiece = nil
					continue
				}
				s.mu.Lock()
				s.topPiece, _ = s.findClosestPiece(s.Pieces, latestTopIdx)
				s.mu.Unlock()
			}
		}
	}()

	return topPriorityPieceChan, returnedPiecesChan
}

func (s *prioritySorter) findClosestPiece(pieces []*pieceWork, ideal int64) (piece *pieceWork, distance int64) {
	numPieces := len(pieces)
	if numPieces == 1 {
		return pieces[0], int64(pieces[0].index) - ideal
	}

	center := numPieces / 2
	hereFoundPiece := pieces[center]
	hereFoundDistance := int64(hereFoundPiece.index) - ideal

	var thereFoundPiece *pieceWork
	var thereFoundDistance int64

	if int64(hereFoundPiece.index) == ideal {
		return hereFoundPiece, 0
	} else if int64(hereFoundPiece.index) > ideal || center + 1 >= numPieces {
		thereFoundPiece, thereFoundDistance = s.findClosestPiece(pieces[:center], ideal)
	} else {
		thereFoundPiece, thereFoundDistance = s.findClosestPiece(pieces[center + 1:], ideal)
	}

	if thereFoundDistance == 0 {
		return thereFoundPiece, thereFoundDistance
	} else if thereFoundDistance > 0 {
		if hereFoundDistance > 0 && hereFoundDistance < thereFoundDistance {
			return hereFoundPiece, hereFoundDistance
		} else {
			return thereFoundPiece, thereFoundDistance
		}
	} else {
		if hereFoundDistance < 0 && thereFoundDistance > hereFoundDistance {
			return thereFoundPiece, thereFoundDistance
		} else {
			return hereFoundPiece, hereFoundDistance
		}
	}
}
