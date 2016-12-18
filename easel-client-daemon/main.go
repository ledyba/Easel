package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
)

const (
	reqStatusEnqueued   = 0
	reqStatusInProgress = 1
	reqStatusDone       = 2
	reqStatusError      = 3
)

/* Server to work with */
var server *string = flag.String("server", "localhost:3000", "server to connect")
var dbAddr *string = flag.String("db", "user:password@tcp(host:port)/dbname", "db address")

var workers *int = flag.Int("workers", 10, "workers to run")

/* Filter Flags */
var filter *string = flag.String("filter", "lanczos", "applied filter name.")
var lobes *int = flag.Int("filter_lobes", 10, "lobes parameter")

/* General */
var help *bool = flag.Bool("help", false, "Print help and exit")

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage of %s:
	%s [OPTIONS] IN OUT
Options:
`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	printStartupBanner()
	if *help {
		usage()
		return
	}

	/* messaging queue */
	requestQueue := make(chan *ResampleRequest, 100)
	notifyQueue := make(chan *ResampleRequest, 100)
	/* chan to controll worker counts */
	workerRestartChan := make(chan bool, *workers)
	for i := 0; i < *workers; i++ {
		workerRestartChan <- true
	}
	fetcherRestartChan := make(chan bool, 1)
	fetcherRestartChan <- true
	notifierRestartChan := make(chan bool, 1)
	notifierRestartChan <- true
	for {
		select {
		case <-workerRestartChan:
			go (func() {
				var err error
				w := newWorker()
				defer w.destroy()
				defer (func() {
					workerRestartChan <- true
				})()
				err = w.connect()
				if err != nil {
					log.Errorf("[%d] Error on connect: %v", w.name, err)
					return
				}
				err = w.init()
				if err != nil {
					log.Errorf("[%d] Error on initialize: %v", w.name, err)
					return
				}
				var output []byte
				for {
					select {
					case r, ok := <-requestQueue:
						if !ok {
							log.Infof("[%d] buffer closed", w.name)
							return
						}
						output, err = w.render(r)
						if err != nil {
							r.err = err
							notifyQueue <- r
						} else {
							ioutil.WriteFile(r.dst, output, os.ModePerm)
							notifyQueue <- r
						}
					}
				}
			})()
		case <-fetcherRestartChan:
			go (func() {
				defer (func() {
					time.Sleep(3 * time.Second)
					fetcherRestartChan <- true
				})()
				var db *sql.DB
				var err error
				db, err = sql.Open("mysql", *dbAddr)
				if err != nil {
					log.Errorf("Error on connecting DB: %v", err)
					return
				}
				defer db.Close()
				timer := time.NewTicker(time.Second * 60)
				var rows *sql.Rows
				for {
					select {
					case <-timer.C:
						err = (func() error {
							rows, err = db.Query("select `id`,`src`,`dst`,`dst_width`,`dst_height`,`dst_quality` from ResampleRequest where status = %d", reqStatusEnqueued)
							if err != nil {
								log.Errorf("Error on selecting db: %v", err)
								return err
							}
							r := ResampleRequest{}
							defer rows.Close()
							for rows.Next() {
								err = rows.Scan(&r.id, &r.src, &r.dst, &r.dstWidth, &r.dstHeight, &r.dstQuality)
								if err != nil {
									log.Errorf("Error on selecting db: %v", err)
									return err
								}
								var q sql.Result
								q, err = db.Exec("update ResampleRequest SET `status`=? where `id`=? and `status`=?", reqStatusInProgress, r.id, reqStatusEnqueued)
								if err != nil {
									log.Errorf("Error on selecting db: %v", err)
									return err
								}
								c, _ := q.RowsAffected()
								ok := c == 1
								if ok {
									requestQueue <- &r
								}
							}
							return nil
						})()
						if err != nil {
							return
						}
					case r := <-requestQueue:
						if r.err == nil {
							var q sql.Result
							q, err = db.Exec("update EaselRequest SET `status`=2 where `id`=?", r.id)
							if err != nil {
								log.Errorf("Error on selecting db: %v", err)
								break
							}
							c, _ := q.RowsAffected()
							if c == 1 {
								log.Errorf("Error on writing db: %v", err)
							}
						}
					}
				}
			})()
		case <-notifierRestartChan:
			go (func() {
				defer (func() {
					time.Sleep(3 * time.Second)
					notifierRestartChan <- true
				})()
				var db *sql.DB
				var err error
				db, err = sql.Open("mysql", *dbAddr)
				if err != nil {
					log.Errorf("Error on connecting DB: %v", err)
					return
				}
				defer db.Close()
				for {
					select {
					case r := <-notifyQueue:
						var q sql.Result
						if r.err != nil {
							q, err = db.Exec("update ResampleRequest SET `status`=? where `id`=?", reqStatusDone, r.id)
						} else {
							q, err = db.Exec("update ResampleRequest SET `status`=?, `message`=? where `id`=?", reqStatusError, r.err.Error(), r.id)
						}
						if err != nil {
							log.Errorf("Error on selecting db: %v", err)
							return
						}
						q.RowsAffected()
					}
				}
			})()
		}
	}
}