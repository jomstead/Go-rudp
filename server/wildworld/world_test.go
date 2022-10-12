package wildworld

import (
	"crypto/sha1"
	"log"
	"sync"
	"time"

	"github.com/jomstead/wildspace/server/wildnet"
	"github.com/xtaci/kcp-go/v5"

	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestPackets(t *testing.T) {
	// WaitGroup is used to wait for the program to finish goroutines.
	var wg sync.WaitGroup

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	// wait for server to become ready
	time.Sleep(time.Second)

	// dial to the echo server
	if sess, err := kcp.DialWithOptions("127.0.0.1:12345", block, 10, 3); err == nil {
		wg.Add(10000)
		for i := 0; i < 10000; i++ {
			go func() {
				defer wg.Done()

				data := []byte{byte(wildnet.SCAN)}
				n, e := sess.Write([]byte(data))
				if e != nil || n == 0 {
					println("not sent")
				}
				data = []byte{byte(wildnet.SCAN_TARGET)}
				n, e = sess.Write([]byte(data))
				if e != nil || n == 0 {
					println("not sent")
				}
				data = []byte{byte(wildnet.MOVE)}
				n, e = sess.Write([]byte(data))
				if e != nil || n == 0 {
					println("not sent")
				}
				data = []byte{byte(254)}
				n, e = sess.Write([]byte(data))
				if e != nil || n == 0 {
					println("not sent")
				}
				time.Sleep(time.Second)
			}()
		}
		wg.Wait()
	} else {
		log.Fatal(err)
	}
}
