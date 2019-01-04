package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/overmike/webterminal/terminal"
	"github.com/spf13/cobra"
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

type mss struct{}

// func (m *mss) isSessionRequest_Command() {
// 	return "ls -l"
// }
func runClient() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到 Server %v", err)
	}
	defer conn.Close()
	// r, err := c.Ping(ctx, &pb.PingRequest{Value: "Hello, World!"})
	c := terminal.NewTerminalClient(conn)
	stream, err := c.Session(context.Background())
	// bb := &terminal.SessionRequest{
	// 	Command: &terminal.SessionRequest_Message{"ls /"},
	// }
	// err = stream.Send(bb)
	// if err != nil {
	// 	s := gstatus.Convert(err)
	// 	// log.Printf("pingCheck Error: %s", s.Message())
	// 	fmt.Sprintf("send Error: %s", s.Message())
	// }
	// for {

	// 	resp, err := stream.Recv()
	// 	log.Println("111")

	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		s := gstatus.Convert(err)
	// 		// log.Printf("pingCheck Error: %s", s.Message())
	// 		fmt.Sprintf("Recv Error: %s", s.Message())
	// 		break
	// 	}

	// 	log.Println("in", string(resp.Message))
	// 	log.Println("222")

	// }

	// 启动一个 goroutine 接收命令行输入的指令
	go func() {
		log.Println("请输入消息...")
		input := bufio.NewReader(os.Stdin)
		for {
			// 获取 命令行输入的字符串， 以回车 \n 作为结束标志
			inmsg, _ := input.ReadString('\n')
			// 向服务端发送 指令
			if err := stream.Send(&terminal.SessionRequest{
				Command: &terminal.SessionRequest_Message{inmsg},
			}); err != nil {
				return
			}
		}
	}()
	for {
		// 接收从 服务端返回的数据流
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
		fmt.Printf(resp.Message)
	}

}
