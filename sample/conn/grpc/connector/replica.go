// Copyright (c) 2019 NEC Laboratories Europe GmbH.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connector

import (
	"context"
	"io"
	"log"

	"google.golang.org/grpc"

	pb "github.com/hyperledger-labs/minbft/sample/conn/grpc/proto"
)

type replica struct {
	id        uint32
	rpcClient pb.ChannelClient
}

func (r *replica) HandleMessageStream(in <-chan []byte) <-chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)

		stream, err := r.rpcClient.Chat(context.Background(), grpc.WaitForReady(true))
		if err != nil {
			log.Printf("Error making RPC call to replica %d: %s\n", r.id, err)
			return
		}

		go r.handleIn(stream, in)

		r.handleOut(stream, out)
	}()

	return out
}

func (r *replica) handleIn(stream pb.Channel_ChatClient, in <-chan []byte) {
	for msg := range in {
		m := &pb.Message{Payload: msg}
		if err := stream.Send(m); err != nil {
			log.Printf("Error sending to replica %d: %s\n", r.id, err)
			return
		}
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("Error closing RPC stream of replica %d: %s\n", r.id, err)
	}
}

func (r *replica) handleOut(stream pb.Channel_ChatClient, out chan<- []byte) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Printf("Error receiving from replica %d: %s\n", r.id, err)
			return
		}
		out <- msg.Payload
	}
}
