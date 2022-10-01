package liteserver

import (
	"fmt"
	"net"
	"time"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"github.com/xxf098/lite-proxy/web"
	"google.golang.org/grpc"
)

type server struct {
	pb.TestProxyServer
}

// stream
func (s *server) StartTest(req *pb.TestRequest, stream pb.TestProxy_StartTestServer) error {
	// check data
	links, err := web.ParseLinks(req.Subscription)
	if err != nil {
		return err
	}
	// config
	p := web.ProfileTest{
		Writer:      nil,
		MessageType: web.ALLTEST,
		Links:       links,
		Options: &web.ProfileTestOptions{
			GroupName:     "Default",
			SpeedTestMode: "all",
			PingMethod:    "googleping",
			SortMethod:    "none",
			Concurrency:   2,
			TestMode:      2,
			Timeout:       15 * time.Second,
			Language:      "en",
			FontSize:      24,
		},
	}

	trafficChan := make(chan int64)
	nodeChan, err := p.TestAll(stream.Context(), trafficChan)
	count := 0
	linkCount := len(links)
	for count < linkCount {
		node := <-nodeChan
		reply := pb.TestReply{
			Id:        int32(node.Id),
			GroupName: node.Group,
			Remarks:   node.Remarks,
			Protocol:  node.Protocol,
			Ping:      node.Ping,
			AvgSpeed:  node.AvgSpeed,
			MaxSpeed:  node.MaxSpeed,
			IsOk:      node.IsOk,
			Traffic:   node.Traffic,
			Link:      node.Link,
		}
		if err := stream.Send(&reply); err != nil {
			return err
		}
		count += 1
	}
	return nil
}

func StartServer(port uint16) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterTestProxyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
