package gophermart

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	limitDefault = 1000
	limitDelta   = 1
)

type accrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type queueOrder struct {
	*Queue
	ctx   context.Context
	order *Order
}

// Do рутина должена запускаться через errgroup
func (qo *queueOrder) Do() error {
	ctx, cancel := context.WithTimeout(qo.ctx, 60*time.Second)
	defer cancel()
	order := qo.order
	url := fmt.Sprintf("%s%d", qo.url, order.ID)
	log.Println("[DEBUG] Making request:", url)

	ao := &accrualOrder{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Length", "0").
		SetContext(ctx).
		SetResult(&ao).
		Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusInternalServerError {
		return fmt.Errorf("internal server error, status code %d", resp.StatusCode())
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		n := uint32(rand.Intn(10)) + 2 // эмулятор переменного кол-во запросов
		// TODO: реализовать парсинг строки `No more than N requests per minute allowed`
		atomic.StoreUint32(&qo.limit, n)
		fmt.Println("[WARNING] Too many requests detected", string(resp.Body()))
		return ErrTooManyRequests
	}

	if resp.StatusCode() == http.StatusNoContent {
		// некритичная ошибка, отменять выполнение других воркеров не надо: выведем варнинг и выйдем из рутины
		log.Printf("[WARNING] No content for order %d\n", order.ID)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unknown status code %d", resp.StatusCode())
	}

	if fmt.Sprint(order.ID) != ao.Order {
		// некритичная ошибка
		log.Printf("[WARNING] Order ID not match, want %d, got %s\n", order.ID, ao.Order)
		return nil
	}

	if order.Status == ao.Status && order.Status == StatusProcessing {
		// заказ уже находится в обработке, выходим
		log.Printf("[DEBUG] Order %d already in processing\n", order.ID)
		return nil
	}

	if !isValidStatus(ao.Status) {
		//некритичная ошибка
		log.Printf("[WARNING] Unknown status detected: %s\n", ao.Status)
		return nil
	}

	order.Status = ao.Status
	order.Accrual = uint64(ao.Accrual * 100)

	// запрос успешно выполнен, обновим заказ
	if err = qo.storage.UpdateOrder(order); err != nil {
		return fmt.Errorf("failed to update order ID %d - %w", order.ID, err)
	}
	log.Printf("[DEBUG] Order successfully updated: order %v\n", order)

	return nil
}

type Queue struct {
	url       string
	storage   Storer
	limit     uint32
	needSleep int32
	pool      map[uint64]*Order
}

func NewQueue(st Storer, addr string) *Queue {

	return &Queue{
		limit:   limitDefault,
		url:     addr + "/api/orders/",
		storage: st,
	}
}

func (q *Queue) updatePool() {
	limit := atomic.LoadUint32(&q.limit)

	ors, err := q.storage.GetPullOrders(limit) // получаем заказы со статусом NEW и PROCESSING, отсортированные по дате поступления
	if err != nil {
		log.Println("[ERROR] Failed to get orders for pool -", err)
		return
	}

	count := uint32(0)
	pool := make(map[uint64]*Order, limit)

	for k, order := range ors {
		count++
		if count > limit {
			break
		}
		pool[k] = order
	}

	q.pool = pool
	log.Printf("[DEBUG] Orders pool updated, now in pool: %d", len(q.pool))
}

func (q *Queue) processor(ctx context.Context) {
	for {
		q.updatePool()

		g, _ := errgroup.WithContext(ctx) // используем errgroup
		for _, order := range q.pool {
			w := &queueOrder{Queue: q, ctx: ctx, order: order}
			g.Go(w.Do)
		}
		err := g.Wait()
		if err != nil {
			atomic.StoreInt32(&q.needSleep, 1) // случилась ошибка, выставим флаг сделать паузу
			if !errors.Is(err, ErrTooManyRequests) {
				// если это не ошибка с превышением кол-ва запросов, выставим лимит по умолчанию
				// иначе, лимит уже был выставлен после парсинга ответа
				atomic.StoreUint32(&q.limit, limitDefault)
			}
			log.Println("[ERROR] Accrual service request failed -", err)
		}

		sleep := 1 * time.Second // дадим секундную передышку сервису `accrual`
		if atomic.LoadInt32(&q.needSleep) == 0 {
			// так как делать большую паузу не нужно, увеличим лимит возможных запросов
			atomic.AddUint32(&q.limit, limitDelta)
		} else {
			// воркер столкнулся с ошибкой или был превышен лимит, сделаем паузу на минуту
			sleep = 60 * time.Second
			// новый лимит уже был выставлен воркером, первым столкнувшимся с ошибкой
			// поэтому просто обнулим флаг `needSleep`
			atomic.StoreInt32(&q.needSleep, 0)
		}
		log.Println("[DEBUG] Got new limit:", atomic.LoadUint32(&q.limit))
		log.Printf("[DEBUG] Sleeping for %s seconds\n", sleep)

		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}
	}
}

func (q *Queue) Start() {
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig
		cancel()
	}()

	q.processor(ctx)
}
