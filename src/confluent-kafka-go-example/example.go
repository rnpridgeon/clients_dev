package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	usage = `
	Usage: benchmark  -C|-P [-b <broker,broker..>] [-s <size>] <configuration file>
	
	Options:
  		-C | -P Consumer or Producer mode must be specified first 
		-b <brokers> Broker address list (host[:port],..)
  		-t <topic>   Topic to consume / produce
  		-s <size>    Message size (producer)
 
 	In Consumer mode:\n
  		consumes messages from topic cc-benchmarking
 	In Producer mode:\n
  		writes messages to cc-benchmarking of size -s <..>
`
)

var (
	timer     *time.Timer
	topicName = "cc-benchmarking"
)

type benchmarkRunner func(*kafka.ConfigMap, io.Writer)

type EventEmitter interface {
	Event() chan kafka.Event
}

type DataSource interface {
	Next()
}

type byteBuffer struct {
	buffer     []byte
	nextWrite  int
	checkpoint int
}

func NewByteBuffer(size int) *byteBuffer {
	return &byteBuffer{
		buffer:     make([]byte, size),
		nextWrite:  0,
		checkpoint: 0,
	}
}

// Returns writer position not bytes written
func (b *byteBuffer) Write(p []byte) (n int, err error) {
	b.nextWrite += copy(b.buffer[b.nextWrite:], p)
	return b.nextWrite, nil
}

func (b *byteBuffer) Cap() (n int) {
	return cap(b.buffer)
}

func (b *byteBuffer) Len() (n int) {
	return b.nextWrite - 1
}

func (b *byteBuffer) Prime() {
	payloadPadSize := cap(b.buffer) - (md5.Size * (cap(b.buffer) / 100))

	seed := make([]byte, 1)
	for pad := payloadPadSize; pad >= 0; pad-- {
		seed[0]++
		b.Write(seed)
	}
	b.Checkpoint()
}

func (b *byteBuffer) Checkpoint() {
	b.checkpoint = b.Len()
}

func (b *byteBuffer) Reset() {
	b.nextWrite = b.checkpoint
}


type dataSource struct {
	pool [16][]byte
}

func NewDataSource(hash_count int) *dataSource {
	d := &dataSource{}

	superDigest := make([]byte, 1)
	for i := 0; i < 16; i++ {
		for c := 0; c < hash_count; c++ {
			superDigest[0]++
			superDigest := md5.Sum(superDigest)
			d.pool[i] = append(d.pool[i][:], superDigest[:]...)
		}
	}

	return d
}

func (d *dataSource) Next() []byte {
	return d.pool[rand.Intn(16)][:]
}

func fromProperties(conf *kafka.ConfigMap, filename string) (err error) {
	var (
		buff []byte
		line string
		prop []string
	)

	if buff, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	reader := bytes.NewBuffer(buff)
	for reader.Len() > 0 {
		line, err = reader.ReadString('\n')

		line = strings.TrimSpace(line)
		if len(line) <= 0 || line[0] == '#' || line[0:1] == "//" {
			continue
		}

		if prop = strings.Split(line, "="); len(prop) == 2 {
			conf.SetKey(strings.TrimSpace(prop[0]), strings.TrimSpace(prop[1]))
			continue
		}

		log.Printf("WARN: Ignoring unrecognized property %s", string(line))
	}
	return err
}

func createTopic(conf *kafka.ConfigMap) {
	a, err := kafka.NewAdminClient(conf)
	if err != nil {
		log.Fatal("Failed to create Admin client: %s\n", err)
	}

	results, err := a.CreateTopics(
		context.Background(),
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1}},
		kafka.SetAdminOperationTimeout(10*time.Second),
	)

	if err != nil {
		log.Fatalf("Failed to create topic %v", err)
	}

	log.Printf("CreateTopics result: %v", results)

	a.Close()
}


func runProducer(msgSize *int) benchmarkRunner {
	return func(conf *kafka.ConfigMap, sink io.Writer) {

		var (
			err              error
			size             = *msgSize
			success, failure int64
			done             chan struct{}
		)

		createTopic(conf)

		var p *kafka.Producer
		if p, err = kafka.NewProducer(conf); err != nil {
			log.Fatal(err)
		}

		size = *msgSize
		dataGenerator := NewDataSource(size/100)

		valueBuffer := NewByteBuffer(size)
		valueBuffer.Prime()

		go func() {
			sink.Write([]byte("["))
			var nonce bool
			for e := range p.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						failure++
						log.Println(ev.TopicPartition.Error)
					} else {
						if success%1000000 == 0 {
							log.Printf("Sent %d messages", success)
						}
						success++
					}
				case *kafka.Stats:
					if nonce {
						sink.Write([]byte(","))
					}
					sink.Write(append([]byte{}, e.String()...))
					nonce = true
				}
			}
			sink.Write([]byte("]"))
			done <- struct{}{}
		}()

		start := time.Now()
		defer func() {
			log.Printf("Summary-> %v Successful sends per second with %v failures",
				float64(success)/time.Since(start).Seconds(), failure)
		}()

		for i := 0;; i++ {
			select {
			case <-timer.C:
				p.Flush(-1)
				p.Close()
				return

			default:
				valueBuffer.Write(dataGenerator.Next())
				msg := &kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
					Value:          append([]byte{}, valueBuffer.buffer...),
				}
				p.Produce(msg, nil)
			}
			// reset writer for new payload
			valueBuffer.Reset()
		}
		log.Println("Flushing event queue")
		<-done
	}
}

//func runConsumer(conf *kafka.ConfigMap, sink io.Writer) {
//	var (
//		err error
//		c   *kafka.Consumer
//		ev  kafka.Event
//	)
//
//	if err = fromProperties(conf, "consumer.properties"); err != nil {
//		return
//	}
//
//	c, _ = kafka.NewConsumer(conf)
//	c.SubscribeTopics([]string{"cc-benchmarking"}, nil)
//
//	for {
//		select {
//		case <-timer.C:
//			break
//
//		default:
//			ev = c.Poll(100)
//			if ev == nil {
//				continue
//			}
//			switch e := ev.(type) {
//			case *kafka.Message:
//				log.Printf("%% Message on %s:\n%s\n",
//					e.TopicPartition, string(e.Value))
//			case kafka.Error:
//				log.Printf("%% Error: %v\n", e)
//				break
//			default:
//				log.Printf("Ignored %v\n", e)
//			}
//		}
//	}
//}

func main() {

	if len(os.Args) == 1 {
		log.Printf(usage)
		os.Exit(1)
	}

	var (
		conf              = &kafka.ConfigMap{}
		benchmarkDuration time.Duration
		messageSize       int
		benchmarkMode     benchmarkRunner
		sink              *os.File
		err               error
	)

	// handle index out of bounds for bad options handling
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Invalid syntax %s\n%s", os.Args, usage)
		}
	}()

	arg := 1
	for ; arg < len(os.Args[1:]); arg++ {
		switch os.Args[arg] {
		case "--P":
			fallthrough
		case "-P":
			// taking the address allows us to ignore ordering
			benchmarkMode = runProducer(&messageSize)
		case "--b":
			fallthrough
		case "-b":
			arg++ // shift
			conf.SetKey("bootstrap.servers", os.Args[arg])
		case "--d":
			fallthrough
		case "-d":
			arg++ // shift
			if benchmarkDuration, err = time.ParseDuration(os.Args[arg]); err != nil {
				log.Fatalf("%s\n%s", err, usage)
			}
		case "--s":
			fallthrough
		case "-s":
			arg++ // shift
			if messageSize, err = strconv.Atoi(os.Args[arg]); err != nil {
				log.Fatalf("%s\n%s", err, usage)
			}
		default:
			log.Println(usage)
		}
	}

	fromProperties(conf, os.Args[arg])

	if sink, err = os.Create("results.json"); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting benchmark with configuration %s\n", conf)

	timer = time.NewTimer(benchmarkDuration)
	benchmarkMode(conf, sink)

	sink.Sync()
	sink.Close()
}
