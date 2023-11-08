package server

//go run main.go server wfq
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/src/model"
	"os"
	"strconv"
	"time"

	"github.com/lucas-clemente/quic-go"
)

var tileGlobal int = 1

type Server struct {
	serverURL   string
	serverPort  int
	queuePolicy QueuePolicy
}

func NewServer(serverURL string, serverPort int, queuePolicy string) *Server {
	return &Server{
		serverURL:   serverURL,
		serverPort:  serverPort,
		queuePolicy: QueuePolicy(queuePolicy),
	}
}

func (s *Server) Start() {

	url := fmt.Sprintf("%s:%d", s.serverURL, s.serverPort)
	config := &quic.Config{
		MaxIdleTimeout:       5 * time.Minute,  // Set a longer maximum idle timeout
		HandshakeIdleTimeout: 10 * time.Second, // Set the receive connection flow control window size to 20 MB
	}
	listener, err := quic.ListenAddr(url, generateTLSConfig(), config)
	if err != nil {
		log.Println(err)
	}

	log.Println("Server listening on", url)

	for {

		connection, err := listener.Accept(context.Background())
		if err != nil {
			log.Println(err)
		}
		if err == nil {
			go s.handleStream(connection)
		}

	}
}

func (s *Server) handleStream(connection quic.Connection) {
	// open stream

	for {
		stream, err := connection.AcceptStream(context.Background())
		if err != nil {
			log.Println(err)
		}
		if streamFinished := connection.Context().Err(); streamFinished != nil {
			return
		}
		if err == nil {
			go func() {
				defer stream.Close()
				// receive file request
				req := s.receiveData(stream)
				zaroreq := model.VideoPacketRequest{}
				if req == zaroreq {
					return
				}
				log.Println("i:", req.Segment)
				// read file
				data := s.readFile(req.Bitrate, req.Segment, req.Tile)

				// send file response
				s.sendData(stream, req.Priority, req.Bitrate, req.Segment, req.Tile, data)
				log.Println("i:", req.Segment)

			}()
		}
	}

}

// Handle request
func (s *Server) handleRequest(req model.VideoPacketRequest, responses chan model.VideoPacketResponse) {
	// read file
	data := s.readFile(req.Bitrate, req.Segment, req.Tile)

	// send response on channel
	res := model.VideoPacketResponse{
		Priority: req.Priority,
		Bitrate:  req.Bitrate,
		Segment:  req.Segment,
		Tile:     req.Tile,
		Data:     data,
	}
	responses <- res
}

// Read file
func (s *Server) readFile(bitrate model.Bitrate, segment int, tile int) []byte {
	basePath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	// TODO check the file name logic
	//data, err := os.ReadFile(basePath + fmt.Sprintf("/data/segments/video_tiled_%d_dash_track%d_%d.m4s", bitrate, segment, tile))
	data, err := os.ReadFile(basePath + fmt.Sprintf("/data/segments/video_tiled_10_dash_track10_%d.m4s", segment))

	str := strconv.Itoa(tileGlobal)

	fmt.Printf("tile")
	fmt.Printf(str)
	tileGlobal += 1
	if err != nil {
		log.Println(err)
	}
	return data
}

// Receive file request
func (s *Server) receiveData(stream quic.Stream) (req model.VideoPacketRequest) {
	if err := json.NewDecoder(stream).Decode(&req); err != nil {
		//log.Fatal(err)

	}
	// streamId := stream.StreamID()
	// fmt.Printf("Server stream %d: Got '%+v'\n", streamId, req)
	return
}

// Send file response
func (s *Server) sendData2(stream quic.Stream, priority model.Priority, bitrate model.Bitrate, segment int, tile int, data []byte) {
	res := model.VideoPacketResponse{
		Priority: priority,
		Bitrate:  bitrate,
		Segment:  segment,
		Tile:     tile,
		Data:     data,
	}

	// encode the response as JSON
	resBytes, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
	}

	// send the response data to the client
	_, err = stream.Write(resBytes)
	if err != nil {
		log.Println(err)
	}
}

// Send file response
func (s *Server) sendData(stream quic.Stream, priority model.Priority, bitrate model.Bitrate, segment int, tile int, data []byte) {
	// streamId := stream.StreamID()
	// fmt.Printf("Server stream %d: Sending '%+v'\n", streamId, res)
	res := model.VideoPacketResponse{
		Priority: priority,
		Bitrate:  bitrate,
		Segment:  segment,
		Tile:     tile,
		Data:     data,
	}
	if err := json.NewEncoder(stream).Encode(&res); err != nil {
		log.Println(err)
	}
}
