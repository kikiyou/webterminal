package cmd

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/overmike/webterminal/terminal"
	"github.com/spf13/cobra"
	sshterminal "golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

// serveCmd represents the serve command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run Web Terminal client",
	Long:  `Run Web Terminal client`,
	Run: func(cmd *cobra.Command, args []string) {
		runClient()
	},
}

type Session struct {
	session terminal.Terminal_SessionClient
	reader  io.Reader
}

func newSession(s terminal.Terminal_SessionClient) *Session {
	return &Session{session: s}
}

func (s *Session) Write(p []byte) (int, error) {
	err := s.session.Send(&terminal.SessionRequest{
		Command: &terminal.SessionRequest_Message{p},
	})
	if err != nil {
		return 0, err
	}
	return len(p), err
}

// func (s *Session) Reader(p []byte) (int, error) {
// 	n, err := s.reader.Read(p)
// 	if err != nil {
// 		return n, err
// 	}
// 	buf := make([]byte, n)
// }

func runClient() {
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到 Server %v", err)
	}
	defer conn.Close()
	// r, err := c.Ping(ctx, &pb.PingRequest{Value: "Hello, World!"})
	c := terminal.NewTerminalClient(conn)
	stream, err := c.Session(context.Background())
	session := newSession(stream)

	log.Println("请输入消息...")

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGWINCH, syscall.SIGCLD)
	go func() {
		for {
			select {
			case _ = <-sig:
				width, height, err := sshterminal.GetSize(0)
				tsize := terminal.TerminalResize{
					Columns: int32(width),
					Rows:    int32(height),
				}
				err = stream.Send(&terminal.SessionRequest{
					Command: &terminal.SessionRequest_Resize{Resize: &tsize},
				})
				if err != nil {
					log.Println("TerminalResize: ", err)
				}
			}

		}
	}()

	go func() { io.Copy(session, os.Stdin) }()
	// go func() { io.Copy(session, os.Stderr) }()

	// go func() { io.Copy(os.Stdout, session) }()
	// proverbs := new(bytes.Buffer)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("⚠️ 收到服务端的结束信号")
			break //如果收到结束信号，则退出“接收循环”，结束客户端程序
		}
		if err != nil {
			// TODO: 处理接收错误
			log.Println("接收数据出错:", err)
		}
		// 没有错误的情况下，打印来自服务端的消息
		// fmt.Printf(string(resp.Message))
		io.Copy(os.Stdout, bytes.NewBuffer(resp.Message))

		// }
		// if _, err := io.Copy(os.Stdout, resp.Message); err != nil {
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// }
	}

}
