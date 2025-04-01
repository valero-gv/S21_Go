package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/api"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/api/warehousepb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/storage"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/warehouse"
)

type appConf struct {
	knownNode     string
	pulseInterval time.Duration
}

func main() {
	address := getListenAddress()

	conf, err := readFlagConf()
	if err != nil {
		log.Fatal("Err Read conf: ", err.Error())
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	myNode := warehouse.NewNode(
		context.Background(),
		address,
		conf.pulseInterval,
		storage.NewMemoStorage(),
	)
	log.Println("listening -", address, myNode.GetNodeUUID())

	tryAddKnown(myNode, conf.knownNode)

	grpcServer := grpc.NewServer()
	warehousepb.RegisterWarehouseServer(grpcServer, api.NewServer(myNode))

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("serve: %s", err)
	}
}

func getListenAddress() string {
	const (
		portStart = 20000
		portEnd   = 65000
	)

	port := portStart + rand.Intn(portEnd-portStart)
	address := "localhost:" + strconv.Itoa(port)

	return address
}

func readFlagConf() (appConf, error) {
	flagPulseInterval := flag.String("p", "", "pulse period")
	flagKnownInfo := flag.String("kn", "", "known host info")

	flag.Parse()

	pulseInterval, err := time.ParseDuration(*flagPulseInterval)
	if err != nil {
		return appConf{}, fmt.Errorf("parse pulse interval: %w", err)
	}

	return appConf{
		pulseInterval: pulseInterval,
		knownNode:     *flagKnownInfo,
	}, nil
}

func tryAddKnown(myNode *warehouse.Node, known string) {
	known = strings.TrimSpace(known)
	knownInfo := strings.Split(known, " ")

	if known != "" {
		if len(knownInfo) != 2 {
			log.Fatal("known wrong format")
		}

		id, err := uuid.Parse(knownInfo[1])
		if err != nil {
			log.Fatal("known uuid can't parse")
		}

		err = myNode.AddKnown(context.Background(), api.NewClient(knownInfo[0], id))
		if err != nil {
			log.Fatal("AddKnown: ", err)
		}
	}
}
